//go:build windows
// +build windows

package main

import ()

func main() {
	onExit := func() {
		log.Info("Exiting...")
		os.Exit(0)
	}
	log.Info("Starting system tray application...")
	trayApp := trayapp.NewTrayApp()
	systray.Run(trayApp.OnReady, onExit)
}
