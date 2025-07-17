//go:build windows
// +build windows

package main

import (
	"github.com/allan-simon/go-singleinstance"
	"github.com/caarlos0/log"
)

func main() {
	lockFile, err := singleinstance.CreateLockFile("paretosecurity-tray.lock")
	if err != nil {
		log.WithError(err).Fatal("An instance of ParetoSecurity tray application is already running.")
		return
	}
	defer lockFile.Close()

	app := NewTrayApp(nil)
	app.Run()
}
