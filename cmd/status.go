package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the checks",
	Run: func(cmd *cobra.Command, args []string) {
		shared.PrintStates()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
