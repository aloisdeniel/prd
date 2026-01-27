package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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

	if len(args) >= 2 && args[1] == "bump" {
		return runFeatBump(featureID)
	}

	path, feature, err := prd.LoadFeature("prd", featureID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if errs := prd.Validate(*feature); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
	}

	if len(args) >= 2 && args[1] == "us" {
		return routeUS(path, feature, args[2:])
	}

	// prd feat <feature-id>
	content, _ := os.ReadFile(path)
	fmt.Print(string(content))
	return nil
}

func routeUS(path string, feature *prd.Feature, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us <subcommand|user-story-id>")
		os.Exit(1)
	}

	switch args[0] {
	case "ls":
		return runUserStoryList(*feature)
	case "next":
		return runUserStoryNext(*feature)
	case "new":
		return runUserStoryNew(path)
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
			return runUserStoryDelete(path, usID)
		case "complete":
			return runUserStoryComplete(path, usID)
		case "accept":
			if len(args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us <user-story-id> accept <ls|<criterion-number> complete>")
				os.Exit(1)
			}
			if args[2] == "ls" {
				return runAcceptList(*feature, usID)
			}
			if len(args) < 4 || args[3] != "complete" {
				fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us <user-story-id> accept <criterion-number> complete")
				os.Exit(1)
			}
			acNum, err := strconv.Atoi(args[2])
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid acceptance criterion number %q\n", args[2])
				os.Exit(1)
			}
			return runAcceptCriterion(path, usID, acNum)
		}
	}

	// prd feat <feature-id> us <user-story-id>
	content, _ := os.ReadFile(path)
	return runUserStory(content, *feature, usArg)
}

func routeNote(featureID string, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> note <ls|add>")
		os.Exit(1)
	}

	path, feature, err := prd.LoadFeature("prd", featureID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch args[0] {
	case "ls":
		return runNoteList(*feature)
	case "add":
		return runNoteAdd(path)
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
		firstLine := n.Content
		if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
			firstLine = firstLine[:idx]
		}
		fmt.Printf("%s\t%s\n", n.Date, firstLine)
	}
	return nil
}

func runFeatList() error {
	entries, err := prd.ListFeatures("prd")
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no features found")
		os.Exit(1)
	}

	for _, e := range entries {
		fmt.Printf("%s  %s  [%d/%d]\n", e.ID, e.Name, e.Completed, e.Total)

		_, feature, err := prd.LoadFeature("prd", e.ID)
		if err != nil {
			continue
		}
		for _, us := range feature.UserStories {
			status := "[ ]"
			if prd.IsStoryCompleted(us) {
				status = "[x]"
			}
			fmt.Printf("  %s US-%d  %s\n", status, us.ID, us.Name)
		}
	}
	return nil
}

func runUserStoryList(feature prd.Feature) error {
	for _, us := range feature.UserStories {
		status := "[ ]"
		if prd.IsStoryCompleted(us) {
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

	section := prd.ExtractUserStorySection(string(content), usID)
	if section == "" {
		fmt.Fprintf(os.Stderr, "could not extract markdown for US-%d\n", usID)
		os.Exit(1)
	}

	fmt.Print(section)
	return nil
}

func runFeatCurrent() error {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get current branch:", err)
		os.Exit(1)
	}
	branch := strings.TrimSpace(string(out))

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

	newID, err := prd.CreateFeature("prd", flagName, flagDescription)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(newID)
	return nil
}

func runUserStoryNext(feature prd.Feature) error {
	for _, us := range feature.UserStories {
		if !prd.IsStoryCompleted(us) {
			fmt.Println(us.ID)
			return nil
		}
	}
	fmt.Fprintln(os.Stderr, "all user stories are completed")
	os.Exit(1)
	return nil
}

func runUserStoryNew(path string) error {
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> us new --name \"...\" --description \"...\" --acceptance-criteria \"...\" --technical-considerations \"...\"")
		os.Exit(1)
	}

	if err := prd.CreateUserStory(path, flagName, flagDescription, flagAcceptanceCriteria, flagTechnicalConsiderations); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return nil
}

func runUserStoryDelete(path string, usID int) error {
	if err := prd.DeleteUserStory(path, usID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func runUserStoryComplete(path string, usID int) error {
	if err := prd.CompleteUserStory(path, usID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func runAcceptList(feature prd.Feature, usID int) error {
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
	for i, ac := range story.AcceptanceCriteria {
		status := "[ ]"
		if ac.Completed {
			status = "[x]"
		}
		fmt.Printf("%d\t%s\t%s\n", i+1, status, ac.Text)
	}
	return nil
}

func runAcceptCriterion(path string, usID, acNum int) error {
	if err := prd.CompleteAcceptanceCriterion(path, usID, acNum); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func runFeatBump(featureID string) error {
	path, _, err := prd.LoadFeature("prd", featureID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := prd.BumpVersion(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}

func runNoteAdd(path string) error {
	if flagNoteContent == "" {
		fmt.Fprintln(os.Stderr, "usage: prd feat <feature-id> note add --content \"...\"")
		os.Exit(1)
	}

	if err := prd.AddNote(path, flagNoteContent); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return nil
}
