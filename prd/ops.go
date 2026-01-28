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

// ProgressPath returns the path to progress.md next to the given prd.md.
func ProgressPath(prdPath string) string {
	return filepath.Join(filepath.Dir(prdPath), "progress.md")
}

// LoadProgress parses progress.md and returns a map of user story ID to completion status.
func LoadProgress(prdPath string) (map[int]bool, error) {
	progPath := ProgressPath(prdPath)
	content, err := os.ReadFile(progPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[int]bool{}, nil
		}
		return nil, err
	}
	return parseProgressChecklist(string(content)), nil
}

func parseProgressChecklist(content string) map[int]bool {
	result := make(map[int]bool)
	for _, line := range SplitLines(content) {
		line = strings.TrimSpace(line)
		completed := false
		var rest string
		if strings.HasPrefix(line, "- [x] US-") {
			completed = true
			rest = strings.TrimPrefix(line, "- [x] US-")
		} else if strings.HasPrefix(line, "- [ ] US-") {
			rest = strings.TrimPrefix(line, "- [ ] US-")
		} else {
			continue
		}
		// Extract the number
		numStr := ""
		for _, c := range rest {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if id, err := strconv.Atoi(numStr); err == nil {
			result[id] = completed
		}
	}
	return result
}

// ensureProgress creates or syncs progress.md from prd.md user stories.
func ensureProgress(prdPath string, feature *Feature) error {
	progPath := ProgressPath(prdPath)
	existing, _ := LoadProgress(prdPath)

	var sb strings.Builder
	sb.WriteString("# Progress\n\n")
	for _, us := range feature.UserStories {
		check := " "
		if existing[us.ID] {
			check = "x"
		}
		sb.WriteString(fmt.Sprintf("- [%s] US-%d\n", check, us.ID))
	}

	// Preserve existing notes section
	content, err := os.ReadFile(progPath)
	if err == nil {
		text := string(content)
		if idx := strings.Index(text, "\n## Notes"); idx >= 0 {
			sb.WriteString(text[idx:])
		}
	}

	return os.WriteFile(progPath, []byte(sb.String()), 0o644)
}

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

		progress, _ := LoadProgress(path)
		completed := 0
		for _, us := range feature.UserStories {
			if progress[us.ID] {
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
// CreateFeature creates a new feature directory and template file.
// If allSections is true, all sections from the document format are included.
// If false, only the required sections (Goals, User Stories, Functional Requirements) are included.
func CreateFeature(basePath, name, desc string, allSections bool) (string, error) {
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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n%s\n\n", name, desc))
	sb.WriteString("## Goals\n\nTODO: Define goals\n\n")
	sb.WriteString("## User Stories\n\n")
	sb.WriteString("## Functional Requirements\n\nTODO: Define functional requirements\n")

	if allSections {
		sb.WriteString("\n## Non-Goals\n\nTODO: Define non-goals\n")
		sb.WriteString("\n## Technical Considerations\n\nTODO: Define technical considerations\n")
		sb.WriteString("\n## Analytics & Instrumentation\n\nTODO: Define analytics\n")
		sb.WriteString("\n## Risks & Mitigations\n\nTODO: Define risks\n")
		sb.WriteString("\n## Success Metrics\n\nTODO: Define success metrics\n")
		sb.WriteString("\n## Open Questions\n\nTODO: Define open questions\n")
	}

	prdPath := filepath.Join(dirPath, "prd.md")
	if err := writePRD(prdPath, sb.String()); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return newID, nil
}

// CreateUserStory appends a new user story to the PRD file.
// Priority should be 1-5 (use 0 to default to P3).
func CreateUserStory(prdPath, name, desc string, criteria []string, techConsider string, priority int) error {
	if priority < 1 || priority > 5 {
		priority = 3
	}
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
	sb.WriteString(fmt.Sprintf("### US-%d | P%d | %s\n", nextID, priority, name))
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

	if err := writePRD(prdPath, result.String()); err != nil {
		return err
	}

	// Re-parse to get updated feature and sync progress.md
	updated, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	feat, err := Parse(string(updated))
	if err != nil {
		return err
	}
	return ensureProgress(prdPath, &feat)
}

// CompleteUserStory toggles the completion of a user story in progress.md.
func CompleteUserStory(prdPath string, usID int) error {
	// Verify story exists in prd.md
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	feature, err := Parse(string(content))
	if err != nil {
		return err
	}
	found := false
	for _, us := range feature.UserStories {
		if us.ID == usID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("user story US-%d not found", usID)
	}

	// Ensure progress.md exists
	if err := ensureProgress(prdPath, &feature); err != nil {
		return err
	}

	// Toggle in progress.md
	progPath := ProgressPath(prdPath)
	progContent, err := os.ReadFile(progPath)
	if err != nil {
		return err
	}
	progText := string(progContent)
	lines := SplitLines(progText)

	target := fmt.Sprintf("US-%d", usID)
	toggled := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [ ] "+target) {
			lines[i] = strings.Replace(line, "- [ ] ", "- [x] ", 1)
			toggled = true
			break
		} else if strings.HasPrefix(trimmed, "- [x] "+target) {
			lines[i] = strings.Replace(line, "- [x] ", "- [ ] ", 1)
			toggled = true
			break
		}
	}
	if !toggled {
		return fmt.Errorf("user story US-%d not found in progress.md", usID)
	}

	var result strings.Builder
	for i, line := range lines {
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	if strings.HasSuffix(progText, "\n") && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString("\n")
	}
	return os.WriteFile(progPath, []byte(result.String()), 0o644)
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

	if err := writePRD(prdPath, result.String()); err != nil {
		return err
	}

	// Remove from progress.md
	progPath := ProgressPath(prdPath)
	progContent, err := os.ReadFile(progPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	target := fmt.Sprintf("US-%d", usID)
	progLines := SplitLines(string(progContent))
	var progResult strings.Builder
	for _, line := range progLines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, target) && (strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "- [x] ")) {
			continue
		}
		progResult.WriteString(line + "\n")
	}
	return os.WriteFile(progPath, []byte(progResult.String()), 0o644)
}

// AddNote adds a dated note to progress.md.
func AddNote(prdPath, noteContent string) error {
	progPath := ProgressPath(prdPath)

	// Ensure progress.md exists
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	feature, err := Parse(string(content))
	if err != nil {
		return err
	}
	if err := ensureProgress(prdPath, &feature); err != nil {
		return err
	}

	progContent, err := os.ReadFile(progPath)
	if err != nil {
		return err
	}
	text := string(progContent)

	date := time.Now().Format("2006-01-02")
	noteBlock := fmt.Sprintf("### %s\n\n%s\n", date, noteContent)

	lines := SplitLines(text)

	// Find the "## Notes" heading
	notesIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## Notes") {
			notesIdx = i
			break
		}
	}

	var result strings.Builder
	if notesIdx >= 0 {
		// Write everything up to and including "## Notes"
		for i := 0; i <= notesIdx; i++ {
			result.WriteString(lines[i] + "\n")
		}
		// Blank line after heading, then the new note
		result.WriteString("\n" + noteBlock)
		// Skip any blank lines after the heading in the original
		rest := notesIdx + 1
		for rest < len(lines) && lines[rest] == "" {
			rest++
		}
		// Add remaining existing notes with a blank line separator
		if rest < len(lines) {
			result.WriteString("\n")
			for i := rest; i < len(lines); i++ {
				result.WriteString(lines[i] + "\n")
			}
		}
	} else {
		result.WriteString(text)
		if !strings.HasSuffix(text, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n## Notes\n\n" + noteBlock)
	}

	return os.WriteFile(progPath, []byte(result.String()), 0o644)
}

// LoadProgressNotes parses notes from progress.md.
func LoadProgressNotes(prdPath string) ([]Note, error) {
	progPath := ProgressPath(prdPath)
	content, err := os.ReadFile(progPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	text := string(content)
	lines := SplitLines(text)

	var notes []Note
	inNotes := false
	var currentNote *Note
	for _, line := range lines {
		if strings.HasPrefix(line, "## Notes") {
			inNotes = true
			continue
		}
		if !inNotes {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "### ") {
			if currentNote != nil {
				currentNote.Content = strings.TrimSpace(currentNote.Content)
				notes = append(notes, *currentNote)
			}
			date := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			currentNote = &Note{Date: date}
			continue
		}
		if currentNote != nil {
			currentNote.Content += line + "\n"
		}
	}
	if currentNote != nil {
		currentNote.Content = strings.TrimSpace(currentNote.Content)
		notes = append(notes, *currentNote)
	}
	return notes, nil
}

// BumpVersion increments the document version and updates the date.
func BumpVersion(prdPath string) error {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return err
	}
	return writePRD(prdPath, string(content))
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
