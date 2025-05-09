//go:build windows
// +build windows

package main

import (
	"math/rand"
	"os"
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/caarlos0/log"
)

func main() {
	if err := shared.LoadConfig(); err != nil {
		log.WithError(err).Warn("failed to load config")
	}

	// Scheduled update command
	go func() {
		log.Info("Starting update command scheduler...")
		for {
			// This is to avoid running the update command too frequently
			// and to ensure that the update command is run at least once an hour
			// also prevent update on first run
			if time.Since(shared.GetModifiedTime()) > time.Hour && !shared.GetModifiedTime().IsZero() {
				_, err := shared.RunCommand(shared.SelfExe(), "update")
				if err != nil {
					log.WithError(err).Error("Failed to run update command")
				}
			}
			time.Sleep(time.Duration(50+rand.Intn(15)) * time.Minute)
		}
	}()

	// Scheduled check command
	go func() {
		log.Info("Starting check command scheduler...")
		for {
			// This is to avoid running the check command too frequently
			time.Sleep(time.Duration(45+rand.Intn(15)) * time.Minute)
			_, err := shared.RunCommand(shared.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("Failed to run check command")
			}
		}
	}()

	// Initialize the state file
	if shared.GetModifiedTime().IsZero() {
		log.Info("Initializing state file...") // by running the check command
		go func() {
			_, err := shared.RunCommand(shared.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("Failed to run check command")
			}
		}()
	}

	onExit := func() {
		log.Info("Exiting...")
		os.Exit(0)
	}
	log.Info("Starting system tray application...")
	systray.Run(trayapp.OnReady, onExit)
}
