package cli

import (
	"fmt"
	"os"

	"github.com/aloisdeniel/prd/skills"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skillCmd)
}

var skillCmd = &cobra.Command{
	Use:   "skill [file]",
	Short: "Print or save the Claude Code SKILL.md document for prd",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			if err := os.WriteFile(args[0], []byte(skills.PRDSkillDocument), 0o644); err != nil {
				return err
			}
			fmt.Println(args[0])
			return nil
		}
		fmt.Print(skills.PRDSkillDocument)
		return nil
	},
}
