//go:build windows
// +build windows

package main

import (
	"os"
	"path/filepath"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

func main() {

	lockDir, _ := shared.UserHomeDir()
	if err := shared.OnlyInstance(filepath.Join(lockDir, ".paretosecurity-tray.lock")); err != nil {
		log.WithError(err).Fatal("An instance of ParetoSecurity tray application is already running.")
		os.Exit(1) // just in case Fatal doesn't
		return
	}

	app := NewTrayApp(nil)
	app.Run()
}
