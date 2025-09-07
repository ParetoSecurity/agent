//go:build linux

package cmd

import (
	"github.com/ParetoSecurity/agent/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI for running checks",
	Long:  "Launch an interactive terminal user interface for running security checks and monitoring their status",
	Run: func(cmd *cobra.Command, args []string) {
		tui.Run()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
