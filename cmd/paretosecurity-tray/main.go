// Package main provides the entry point for the application.
package main

import (
	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/caarlos0/log"
)

func main() {
	onExit := func() {
		log.Info("Exiting...")
	}

	systray.Run(trayapp.OnReady, onExit)
}
