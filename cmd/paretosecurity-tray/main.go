//go:build windows
// +build windows

package main

import (
	"math/rand"
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
	onExit := func() {
		log.Info("Exiting...")
	}

	go func() {
		for {
			// This is to avoid running the update command too frequently
			// and to ensure that the update command is run at least once an hour
			if time.Since(shared.GetModifiedTime()) > time.Hour {
				_, err := shared.RunCommand(shared.SelfExe(), "update")
				if err != nil {
					log.WithError(err).Error("Failed to run update command")
				}
			}
			time.Sleep(time.Duration(50+rand.Intn(15)) * time.Minute)
		}
	}()

	go func() {
		for {
			// This is to avoid running the check command too frequently
			time.Sleep(time.Duration(45+rand.Intn(15)) * time.Minute)
			_, err := shared.RunCommand(shared.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("Failed to run check command")
			}
		}
	}()

	systray.Run(trayapp.OnReady, onExit)
}
