//go:build linux || darwin
// +build linux darwin

package cmd

import (
	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var trayiconCmd = &cobra.Command{
	Use:   "trayicon",
	Short: "Display the status of the checks in the system tray",
	Run: func(cc *cobra.Command, args []string) {

		onExit := func() {
			log.Info("Exiting...")
		}

		systray.Run(trayapp.OnReady, onExit)
	},
}

func init() {

	rootCmd.AddCommand(trayiconCmd)
}
