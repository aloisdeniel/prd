package cli

import (
	"github.com/aloisdeniel/prd/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tuiCmd)
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run("prd")
	},
}
