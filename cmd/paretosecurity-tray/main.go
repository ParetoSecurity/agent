// Package main provides the entry point for the application.
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
	onExit := func() {
		log.Info("Exiting...")
	}

	go func() {
		for {
			_, err := shared.RunCommand(shared.SelfExe(), "update")
			if err != nil {
				log.WithError(err).Error("Failed to run update command")
			}
			time.Sleep(time.Duration(50+rand.Intn(15)) * time.Minute)
		}
	}()

	go func() {
		for {
			_, err := shared.RunCommand(shared.SelfExe(), "check")
			if err != nil {
				log.WithError(err).Error("Failed to run check command")
			}
			time.Sleep(time.Duration(45+rand.Intn(15)) * time.Minute)
		}
	}()

	systray.Run(trayapp.OnReady, onExit)
}
