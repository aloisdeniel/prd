package prd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ListFeatures returns all features found under basePath.
func ListFeatures(basePath string) ([]FeatureEntry, error) {
	matches, err := filepath.Glob(filepath.Join(basePath, "*", "prd.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)

	var entries []FeatureEntry
	for _, path := range matches {
		dir := filepath.Base(filepath.Dir(path))
		parts := strings.SplitN(dir, "-", 2)
		id := parts[0]
		name := ""
		if len(parts) > 1 {
			name = parts[1]
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		feature, err := Parse(string(content))
		if err != nil {
			continue
		}

		displayName := feature.Name
		if displayName == "" {
			displayName = name
		}

		completed := 0
		for _, us := range feature.UserStories {
			if IsStoryCompleted(us) {
				completed++
			}
		}

		entries = append(entries, FeatureEntry{
			ID:        id,
			Name:      displayName,
			Path:      path,
			Completed: completed,
			Total:     len(feature.UserStories),
		})
	}
	return entries, nil
}

// LoadFeature loads and parses a feature by ID from basePath.
func LoadFeature(basePath, id string) (path string, feature *Feature, err error) {
	matches, err := filepath.Glob(filepath.Join(basePath, id+"-*", "prd.md"))
	if err != nil {
		return "", nil, err
	}
	if len(matches) == 0 {
		return "", nil, fmt.Errorf("no PRD found for feature ID %q", id)
	}

	path = matches[0]
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	f, err := Parse(string(content))
	if err != nil {
		return "", nil, err
	}
	return path, &f, nil
}

// CreateFeature creates a new feature directory and template file.
func CreateFeature(basePath, name, desc string) (string, error) {
	matches, _ := filepath.Glob(filepath.Join(basePath, "*"))
	maxID := 0
	for _, m := range matches {
		base := filepath.Base(m)
		parts := strings.SplitN(base, "-", 2)
		if id, err := strconv.Atoi(parts[0]); err == nil && id > maxID {
			maxID = id
		}
	}
	newID := fmt.Sprintf("%03d", maxID+1)

	slug := Slugify(name)
	dirName := newID + "-" + slug
	dirPath := filepath.Join(basePath, dirName)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if desc == "" {
		desc = "TODO: Add description"
	}

	content := fmt.Sprintf(`# %s

%s

## Goals

TODO: Define goals

## User Stories

## Functional Requirements

TODO: Define functional requirements
`, name, desc)

	prdPath := filepath.Join(dirPath, "prd.md")
	if err := writePRD(prdPath, content); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return newID, nil
}

// CreateUserStory appends a new user story to the PRD file.
func CreateUserStory(prdPath, name, desc string, criteria []string, techConsider string) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}

	feature, err := Parse(string(content))
	if err != nil {
		return err
	}

	nextID := 1
	for _, us := range feature.UserStories {
		if us.ID >= nextID {
			nextID = us.ID + 1
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### US-%d | %s\n", nextID, name))
	if desc != "" {
		sb.WriteString("\n" + desc + "\n")
	}
	sb.WriteString("\n#### Acceptance Criteria\n\n")
	if len(criteria) > 0 {
		for _, ac := range criteria {
			sb.WriteString("- [ ] " + ac + "\n")
		}
	} else {
		sb.WriteString("- [ ] TODO\n")
	}
	if techConsider != "" {
		sb.WriteString("\n#### Technical Considerations\n\n" + techConsider + "\n")
	}

	newSection := sb.String()
	text := string(content)
	lines := SplitLines(text)

	insertIdx := -1
	inUserStories := false
	for i, line := range lines {
		if strings.HasPrefix(line, "## User Stories") {
			inUserStories = true
			continue
		}
		if inUserStories && strings.HasPrefix(line, "## ") {
			insertIdx = i
			break
		}
	}

	var result strings.Builder
	if insertIdx >= 0 {
		for i := 0; i < insertIdx; i++ {
			result.WriteString(lines[i] + "\n")
		}
		result.WriteString(newSection + "\n")
		for i := insertIdx; i < len(lines); i++ {
			result.WriteString(lines[i] + "\n")
		}
	} else {
		result.WriteString(text)
		if !strings.HasSuffix(text, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n" + newSection)
	}

	return writePRD(prdPath, result.String())
}

// CompleteUserStory marks all acceptance criteria as completed.
func CompleteUserStory(prdPath string, usID int) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	text := string(content)
	lines := SplitLines(text)
	section := ExtractUserStorySectionRange(text, usID)
	if section.Start < 0 {
		return fmt.Errorf("user story US-%d not found", usID)
	}

	for i := section.Start; i < section.End; i++ {
		lines[i] = strings.Replace(lines[i], "- [ ] ", "- [x] ", 1)
	}

	return writeLines(prdPath, lines, text)
}

// CompleteAcceptanceCriterion marks a single acceptance criterion as completed.
func CompleteAcceptanceCriterion(prdPath string, usID, acNum int) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	text := string(content)
	lines := SplitLines(text)
	section := ExtractUserStorySectionRange(text, usID)
	if section.Start < 0 {
		return fmt.Errorf("user story US-%d not found", usID)
	}

	count := 0
	found := false
	for i := section.Start; i < section.End; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ") {
			count++
			if count == acNum {
				lines[i] = strings.Replace(lines[i], "- [ ] ", "- [x] ", 1)
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("acceptance criterion %d not found in US-%d", acNum, usID)
	}

	return writeLines(prdPath, lines, text)
}

// ToggleAcceptanceCriterion toggles a single acceptance criterion.
func ToggleAcceptanceCriterion(prdPath string, usID, acNum int) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	text := string(content)
	lines := SplitLines(text)
	section := ExtractUserStorySectionRange(text, usID)
	if section.Start < 0 {
		return fmt.Errorf("user story US-%d not found", usID)
	}

	count := 0
	found := false
	for i := section.Start; i < section.End; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- [ ] ") {
			count++
			if count == acNum {
				lines[i] = strings.Replace(lines[i], "- [ ] ", "- [x] ", 1)
				found = true
				break
			}
		} else if strings.HasPrefix(trimmed, "- [x] ") {
			count++
			if count == acNum {
				lines[i] = strings.Replace(lines[i], "- [x] ", "- [ ] ", 1)
				found = true
				break
			}
		}
	}
	if !found {
		return fmt.Errorf("acceptance criterion %d not found in US-%d", acNum, usID)
	}

	return writeLines(prdPath, lines, text)
}

// DeleteUserStory removes a user story from the PRD file.
func DeleteUserStory(prdPath string, usID int) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	text := string(content)
	lines := SplitLines(text)
	section := ExtractUserStorySectionRange(text, usID)
	if section.Start < 0 {
		return fmt.Errorf("user story US-%d not found", usID)
	}

	end := section.End
	for end < len(lines) && lines[end] == "" {
		end++
	}

	var result strings.Builder
	for i := 0; i < section.Start; i++ {
		result.WriteString(lines[i] + "\n")
	}
	for i := end; i < len(lines); i++ {
		result.WriteString(lines[i] + "\n")
	}

	return writePRD(prdPath, result.String())
}

// AddNote adds a dated note to the PRD file.
func AddNote(prdPath, noteContent string) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	text := string(content)

	date := time.Now().Format("2006-01-02")
	noteBlock := fmt.Sprintf("### %s\n\n%s\n", date, noteContent)

	lines := SplitLines(text)

	insertIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## Notes") {
			insertIdx = i + 1
			for insertIdx < len(lines) && lines[insertIdx] == "" {
				insertIdx++
			}
			break
		}
	}

	var result strings.Builder
	if insertIdx >= 0 {
		for i := 0; i < insertIdx; i++ {
			result.WriteString(lines[i] + "\n")
		}
		result.WriteString("\n" + noteBlock)
		for i := insertIdx; i < len(lines); i++ {
			result.WriteString(lines[i] + "\n")
		}
	} else {
		footerIdx := -1
		for i, line := range lines {
			if strings.TrimSpace(line) == "---" {
				footerIdx = i
				break
			}
		}
		if footerIdx >= 0 {
			for i := 0; i < footerIdx; i++ {
				result.WriteString(lines[i] + "\n")
			}
			result.WriteString("## Notes\n\n" + noteBlock + "\n")
			for i := footerIdx; i < len(lines); i++ {
				result.WriteString(lines[i] + "\n")
			}
		} else {
			result.WriteString(text)
			if !strings.HasSuffix(text, "\n") {
				result.WriteString("\n")
			}
			result.WriteString("\n## Notes\n\n" + noteBlock)
		}
	}

	return writePRD(prdPath, result.String())
}

// Slugify converts a string to a URL-friendly slug.
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// IsStoryCompleted returns true if all acceptance criteria are completed.
func IsStoryCompleted(us UserStory) bool {
	if len(us.AcceptanceCriteria) == 0 {
		return false
	}
	for _, ac := range us.AcceptanceCriteria {
		if !ac.Completed {
			return false
		}
	}
	return true
}

// SplitLines splits a string into lines without trailing newlines.
func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// LineRange represents a range of line indices [Start, End).
type LineRange struct {
	Start, End int
}

// ExtractUserStorySectionRange returns the line range for a user story.
func ExtractUserStorySectionRange(content string, usID int) LineRange {
	lines := SplitLines(content)

	matchesUS := func(line string) bool {
		const h = "### US-"
		if len(line) < len(h) || line[:len(h)] != h {
			return false
		}
		rest := line[len(h):]
		numEnd := 0
		for numEnd < len(rest) && rest[numEnd] >= '0' && rest[numEnd] <= '9' {
			numEnd++
		}
		if numEnd == 0 {
			return false
		}
		n, err := strconv.Atoi(rest[:numEnd])
		if err != nil {
			return false
		}
		if n != usID {
			return false
		}
		return numEnd == len(rest) || rest[numEnd] == ' '
	}

	start := -1
	for i, line := range lines {
		if matchesUS(line) {
			start = i
			break
		}
	}
	if start < 0 {
		return LineRange{-1, -1}
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if IsHeadingLevel(lines[i], 3) {
			end = i
			break
		}
	}

	for end > start && lines[end-1] == "" {
		end--
	}

	return LineRange{start, end}
}

// ExtractUserStorySection returns the markdown text for a user story.
func ExtractUserStorySection(content string, usID int) string {
	r := ExtractUserStorySectionRange(content, usID)
	if r.Start < 0 {
		return ""
	}
	lines := SplitLines(content)
	var result strings.Builder
	for i := r.Start; i < r.End; i++ {
		result.WriteString(lines[i] + "\n")
	}
	return result.String()
}

// IsHeadingLevel returns true if the line is a markdown heading of the given level or lower.
func IsHeadingLevel(line string, level int) bool {
	hashes := 0
	for _, c := range line {
		if c == '#' {
			hashes++
		} else {
			break
		}
	}
	return hashes >= 1 && hashes <= level && len(line) > hashes && line[hashes] == ' '
}

func writeLines(path string, lines []string, originalContent string) error {
	var result strings.Builder
	for i, line := range lines {
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	if strings.HasSuffix(originalContent, "\n") && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString("\n")
	}
	return writePRD(path, result.String())
}

// writePRD writes content to the PRD file, updating the version footer.
func writePRD(path, content string) error {
	content = updateFooter(content)
	return os.WriteFile(path, []byte(content), 0o644)
}

// updateFooter updates or appends the document version footer.
// It increments the patch version and sets the date to today.
func updateFooter(content string) string {
	today := time.Now().Format("2006-01-02")
	lines := SplitLines(content)

	// Find existing footer: look for "---" line followed by version/date lines
	footerStart := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			// Check if next lines look like version/date footer
			if i+1 < len(lines) && strings.Contains(lines[i+1], "Document Version:") {
				footerStart = i
				break
			}
		}
	}

	newVersion := "1.0"
	if footerStart >= 0 {
		// Extract current version and increment
		for j := footerStart; j < len(lines); j++ {
			if strings.Contains(lines[j], "Document Version:") {
				// Parse version like *Document Version: 1.1*
				s := lines[j]
				s = strings.TrimSpace(s)
				s = strings.TrimPrefix(s, "*")
				s = strings.TrimSuffix(s, "*")
				s = strings.TrimPrefix(s, "Document Version:")
				s = strings.TrimSpace(s)
				newVersion = incrementVersion(s)
				break
			}
		}

		// Remove old footer lines (from --- to end of version/date block)
		footerEnd := footerStart + 1
		for footerEnd < len(lines) {
			line := strings.TrimSpace(lines[footerEnd])
			if line == "" || strings.HasPrefix(line, "*") {
				footerEnd++
			} else {
				break
			}
		}
		// Trim trailing empty lines before footer
		for footerStart > 0 && lines[footerStart-1] == "" {
			footerStart--
		}
		lines = append(lines[:footerStart], lines[footerEnd:]...)
	}

	// Rebuild content without trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var result strings.Builder
	for _, line := range lines {
		result.WriteString(line + "\n")
	}
	result.WriteString("\n---\n")
	result.WriteString(fmt.Sprintf("*Document Version: %s*\n", newVersion))
	result.WriteString(fmt.Sprintf("*Last Updated: %s*\n", today))

	return result.String()
}

func incrementVersion(v string) string {
	parts := strings.SplitN(v, ".", 2)
	if len(parts) != 2 {
		return "1.0"
	}
	major := parts[0]
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return major + ".1"
	}
	return fmt.Sprintf("%s.%d", major, minor+1)
}
