//go:build windows
// +build windows

package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Pareto Security Agent",
	Long:  `Update the Pareto Security Agent to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		shared.UpdateApp()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
