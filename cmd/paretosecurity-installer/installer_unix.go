//go:build !windows

package main

import (
	"os"
)

type WindowService struct {
}

func (w *WindowService) QuitApp() error {
	os.Exit(0)
	return nil
}

func (w *WindowService) InstallApp(withStartup bool) error {
	return nil
}
