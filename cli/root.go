package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "prd",
	Short: "PRD document manager",
}

func Execute() error {
	return rootCmd.Execute()
}
