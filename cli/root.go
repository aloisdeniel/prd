package cli

import (
	"github.com/aloisdeniel/prd/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "prd",
	Short: "PRD document manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run("prd")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
