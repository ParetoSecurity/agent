//go:build toast
// +build toast

package cmd

import (
	"github.com/ParetoSecurity/agent/notify"
	"github.com/spf13/cobra"
)

var toastCMD = &cobra.Command{
	Use:   "toast",
	Short: "Display random toast messages",
	Run: func(cc *cobra.Command, args []string) {

		notify.Toast("Welcome to Pareto Security Agent!")
	},
}

func init() {

	rootCmd.AddCommand(toastCMD)
}
