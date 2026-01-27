package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aloisdeniel/prd/prd"
	"github.com/spf13/cobra"
)

var (
	flagName                    string
	flagDescription             string
	flagAcceptanceCriteria      []string
	flagTechnicalConsiderations string
	flagNoteContent             string
)

func init() {
	featCmd.Flags().StringVar(&flagName, "name", "", "Feature or user story name")
	featCmd.Flags().StringVar(&flagDescription, "description", "", "Feature or user story description")
	featCmd.Flags().StringSliceVar(&flagAcceptanceCriteria, "acceptance-criteria", nil, "Acceptance criteria for a user story")
	featCmd.Flags().StringVar(&flagTechnicalConsiderations, "technical-considerations", "", "Technical considerations for a user story")
	featCmd.Flags().StringVar(&flagNoteContent, "content", "", "Note content")
	rootCmd.AddCommand(featCmd)
}

var featCmd = &cobra.Command{
	Use:   "feat",
	Short: "Manage feature PRDs",
	Args:  cobra.ArbitraryArgs,
	RunE:  runFeat,
}

func runFeat(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	switch args[0] {
	case "ls":
		return runFeatList()
	case "current":
		return runFeatCurrent()
	case "start":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: prd feat start <feature-id>")
			os.Exit(1)
		}
		return runFeatStart(args[1])
	case "new":
		return runFeatNew()
	}

	// args[0] is a feature-id
	featureID := args[0]

	if len(args) >= 2 && args[1] == "note" {
		return routeNote(featureID, args[2:])
	}

	path, content, feature, ok := loadFeature(featureID)
	if !ok {
		return nil
	}

	if errs := prd.Validate(feature); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
	}

	if len(args) >= 2 && args[1] == "us" {
		return routeUS(path, content, feature, args[2:])
	}

	// prd feat <feature-id>
	fmt.Print(string(content))
	return nil
}

func routeUS(path string, content []byte, feature prd.Feature, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us <subcommand|user-story-id>")
		os.Exit(1)
	}

	switch args[0] {
	case "ls":
		return runUserStoryList(feature)
	case "next":
		return runUserStoryNext(feature)
	case "new":
		return runUserStoryNew(path, string(content), feature)
	}

	// args[0] is a user-story-id
	usArg := args[0]
	usID, err := strconv.Atoi(usArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid user story ID %q: must be a number\n", usArg)
		os.Exit(1)
	}

	if len(args) >= 2 {
		switch args[1] {
		case "delete":
			return runUserStoryDelete(path, string(content), usID)
		case "complete":
			return runUserStoryComplete(path, string(content), usID)
		case "accept":
			if len(args) < 4 || args[3] != "complete" {
				fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us <user-story-id> accept <criterion-number> complete")
				os.Exit(1)
			}
			acNum, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid acceptance criterion number %q\n", args[2])
				os.Exit(1)
			}
			return runAcceptCriterion(path, string(content), usID, acNum)
		}
	}

	// prd feat <feature-id> us <user-story-id>
	return runUserStory(content, feature, usArg)
}

func routeNote(featureID string, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> note <ls|add>")
		os.Exit(1)
	}

	path, content, feature, ok := loadFeature(featureID)
	if !ok {
		return nil
	}

	switch args[0] {
	case "ls":
		return runNoteList(feature)
	case "add":
		return runNoteAdd(path, string(content))
	default:
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> note <ls|add>")
		os.Exit(1)
	}
	return nil
}

func runNoteList(feature prd.Feature) error {
	if len(feature.Notes) == 0 {
		fmt.Fprintln(os.Stderr, "no notes found")
		os.Exit(1)
	}
	for _, n := range feature.Notes {
		// Print date and first line of content
		firstLine := n.Content
		if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
			firstLine = firstLine[:idx]
		}
		fmt.Printf("%s\t%s\n", n.Date, firstLine)
	}
	return nil
}

// --- Existing commands ---

func runFeatList() error {
	matches, err := filepath.Glob(filepath.Join("prd", "*", "prd.md"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "no features found")
		os.Exit(1)
	}

	sort.Strings(matches)

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
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
			continue
		}

		feature, err := prd.Parse(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing %s: %v\n", path, err)
			continue
		}

		displayName := feature.Name
		if displayName == "" {
			displayName = name
		}

		completed := 0
		for _, us := range feature.UserStories {
			if isStoryCompleted(us) {
				completed++
			}
		}
		total := len(feature.UserStories)

		fmt.Printf("%s\t%s\t[%d/%d]\n", id, displayName, completed, total)
	}
	return nil
}

func runUserStoryList(feature prd.Feature) error {
	for _, us := range feature.UserStories {
		status := "[ ]"
		if isStoryCompleted(us) {
			status = "[x]"
		}
		fmt.Printf("US-%d\t%s\t%s\n", us.ID, us.Name, status)
	}
	return nil
}

func runUserStory(content []byte, feature prd.Feature, usArg string) error {
	usID, err := strconv.Atoi(usArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid user story ID %q: must be a number\n", usArg)
		os.Exit(1)
	}

	var story *prd.UserStory
	for i := range feature.UserStories {
		if feature.UserStories[i].ID == usID {
			story = &feature.UserStories[i]
			break
		}
	}
	if story == nil {
		fmt.Fprintf(os.Stderr, "user story US-%d not found\n", usID)
		os.Exit(1)
	}

	section := extractUserStorySection(string(content), usID)
	if section == "" {
		fmt.Fprintf(os.Stderr, "could not extract markdown for US-%d\n", usID)
		os.Exit(1)
	}

	fmt.Print(section)
	return nil
}

// --- New commands ---

func runFeatCurrent() error {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get current branch:", err)
		os.Exit(1)
	}
	branch := strings.TrimSpace(string(out))

	// Match branch names like feat/001-feature-name or feature/001-name
	re := regexp.MustCompile(`^feat(?:ure)?/(\d+)`)
	m := re.FindStringSubmatch(branch)
	if m == nil {
		fmt.Fprintf(os.Stderr, "current branch %q is not a feature branch\n", branch)
		os.Exit(1)
	}

	fmt.Println(m[1])
	return nil
}

func runFeatStart(featureID string) error {
	// Verify the feature exists
	matches, err := filepath.Glob(filepath.Join("prd", featureID+"-*"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no PRD found for feature ID %q\n", featureID)
		os.Exit(1)
	}

	dirName := filepath.Base(matches[0])
	branch := "feat/" + dirName

	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	return nil
}

func runFeatNew() error {
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "usage: prd feat new --name \"...\" --description \"...\"")
		os.Exit(1)
	}

	// Find next available ID
	matches, _ := filepath.Glob(filepath.Join("prd", "*"))
	maxID := 0
	for _, m := range matches {
		base := filepath.Base(m)
		parts := strings.SplitN(base, "-", 2)
		if id, err := strconv.Atoi(parts[0]); err == nil && id > maxID {
			maxID = id
		}
	}
	newID := fmt.Sprintf("%03d", maxID+1)

	slug := slugify(flagName)
	dirName := newID + "-" + slug
	dirPath := filepath.Join("prd", dirName)
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory: %v\n", err)
		os.Exit(1)
	}

	desc := flagDescription
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
`, flagName, desc)

	prdPath := filepath.Join(dirPath, "prd.md")
	if err := os.WriteFile(prdPath, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(newID)
	return nil
}

func runUserStoryNext(feature prd.Feature) error {
	for _, us := range feature.UserStories {
		if !isStoryCompleted(us) {
			fmt.Println(us.ID)
			return nil
		}
	}
	fmt.Fprintln(os.Stderr, "all user stories are completed")
	os.Exit(1)
	return nil
}

func runUserStoryNew(path, content string, feature prd.Feature) error {
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us new --name \"...\" --description \"...\" --acceptance-criteria \"...\" --technical-considerations \"...\"")
		os.Exit(1)
	}

	// Determine next US ID
	nextID := 1
	for _, us := range feature.UserStories {
		if us.ID >= nextID {
			nextID = us.ID + 1
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### US-%d | %s\n", nextID, flagName))
	if flagDescription != "" {
		sb.WriteString("\n" + flagDescription + "\n")
	}
	sb.WriteString("\n#### Acceptance Criteria\n\n")
	if len(flagAcceptanceCriteria) > 0 {
		for _, ac := range flagAcceptanceCriteria {
			sb.WriteString("- [ ] " + ac + "\n")
		}
	} else {
		sb.WriteString("- [ ] TODO\n")
	}
	if flagTechnicalConsiderations != "" {
		sb.WriteString("\n#### Technical Considerations\n\n" + flagTechnicalConsiderations + "\n")
	}

	newSection := sb.String()

	// Insert before the next ## heading after ## User Stories
	lines := splitLines(content)
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
		// Insert blank line + new section before the next ## heading
		for i := 0; i < insertIdx; i++ {
			result.WriteString(lines[i] + "\n")
		}
		result.WriteString(newSection + "\n")
		for i := insertIdx; i < len(lines); i++ {
			result.WriteString(lines[i] + "\n")
		}
	} else {
		// Append at end of file
		result.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n" + newSection)
	}

	if err := os.WriteFile(path, []byte(result.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(nextID)
	return nil
}

func runUserStoryDelete(path, content string, usID int) error {
	section := extractUserStorySectionRange(content, usID)
	if section.start < 0 {
		fmt.Fprintf(os.Stderr, "user story US-%d not found\n", usID)
		os.Exit(1)
	}

	lines := splitLines(content)

	// Also remove trailing blank lines after the section
	end := section.end
	for end < len(lines) && lines[end] == "" {
		end++
	}

	var result strings.Builder
	for i := 0; i < section.start; i++ {
		result.WriteString(lines[i] + "\n")
	}
	for i := end; i < len(lines); i++ {
		result.WriteString(lines[i] + "\n")
	}

	if err := os.WriteFile(path, []byte(result.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func runUserStoryComplete(path, content string, usID int) error {
	lines := splitLines(content)
	section := extractUserStorySectionRange(content, usID)
	if section.start < 0 {
		fmt.Fprintf(os.Stderr, "user story US-%d not found\n", usID)
		os.Exit(1)
	}

	for i := section.start; i < section.end; i++ {
		lines[i] = strings.Replace(lines[i], "- [ ] ", "- [x] ", 1)
	}

	var result strings.Builder
	for i, line := range lines {
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	// Preserve trailing newline if original had one
	if strings.HasSuffix(content, "\n") && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(result.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func runAcceptCriterion(path, content string, usID, acNum int) error {
	lines := splitLines(content)
	section := extractUserStorySectionRange(content, usID)
	if section.start < 0 {
		fmt.Fprintf(os.Stderr, "user story US-%d not found\n", usID)
		os.Exit(1)
	}

	// Find the Nth unchecked criterion (1-based) within the section
	count := 0
	found := false
	for i := section.start; i < section.end; i++ {
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
		fmt.Fprintf(os.Stderr, "acceptance criterion %d not found in US-%d\n", acNum, usID)
		os.Exit(1)
	}

	var result strings.Builder
	for i, line := range lines {
		result.WriteString(line)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	if strings.HasSuffix(content, "\n") && !strings.HasSuffix(result.String(), "\n") {
		result.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(result.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}
	return nil
}

func runNoteAdd(path, content string) error {
	if flagNoteContent == "" {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> note add --content \"...\"")
		os.Exit(1)
	}

	date := time.Now().Format("2006-01-02")
	noteBlock := fmt.Sprintf("### %s\n\n%s\n", date, flagNoteContent)

	lines := splitLines(content)

	// Find ## Notes section and insert after it
	insertIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "## Notes") {
			// Insert after this heading (and any blank line following)
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
		// No Notes section exists — add one before the --- footer or at the end
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
			result.WriteString(content)
			if !strings.HasSuffix(content, "\n") {
				result.WriteString("\n")
			}
			result.WriteString("\n## Notes\n\n" + noteBlock)
		}
	}

	if err := os.WriteFile(path, []byte(result.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file: %v\n", err)
		os.Exit(1)
	}
	return nil
}

// --- Helpers ---

func isStoryCompleted(us prd.UserStory) bool {
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

func loadFeature(featureID string) (path string, content []byte, feature prd.Feature, ok bool) {
	matches, err := filepath.Glob(filepath.Join("prd", featureID+"-*", "prd.md"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob error: %v\n", err)
		os.Exit(1)
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "no PRD found for feature ID %q\n", featureID)
		os.Exit(1)
	}

	path = matches[0]
	content, err = os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
		os.Exit(1)
	}

	feature, err = prd.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	return path, content, feature, true
}

type lineRange struct {
	start, end int // line indices [start, end)
}

func extractUserStorySectionRange(content string, usID int) lineRange {
	lines := splitLines(content)

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
		return lineRange{-1, -1}
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if isHeadingLevel(lines[i], 3) {
			end = i
			break
		}
	}

	// Trim trailing blank lines
	for end > start && lines[end-1] == "" {
		end--
	}

	return lineRange{start, end}
}

func extractUserStorySection(content string, usID int) string {
	r := extractUserStorySectionRange(content, usID)
	if r.start < 0 {
		return ""
	}
	lines := splitLines(content)
	var result strings.Builder
	for i := r.start; i < r.end; i++ {
		result.WriteString(lines[i] + "\n")
	}
	return result.String()
}

func splitLines(s string) []string {
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

func isHeadingLevel(line string, level int) bool {
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

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
