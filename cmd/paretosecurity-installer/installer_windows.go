//go:build windows
// +build windows

package main

import (
	"os"

	_ "embed"

	"github.com/ParetoSecurity/agent/shared"
)

type WindowService struct{}

func (w *WindowService) QuitApp() error {
	os.Exit(0)
	return nil
}

func (w *WindowService) InstallApp(withStartup bool) error {
	return shared.InstallApp(withStartup)
}
