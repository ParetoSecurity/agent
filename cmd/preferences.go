package cmd

import (
	"github.com/ParetoSecurity/agent/cmd/ui"
	"github.com/spf13/cobra"
)

var preferencesUICmd = &cobra.Command{
	Use:   "preferences",
	Short: "Display the preferences dialog",
	Run: func(cc *cobra.Command, args []string) {
		ui.CreatePreferencesWindow()
	},
}

func init() {
	rootCmd.AddCommand(preferencesUICmd)
}
