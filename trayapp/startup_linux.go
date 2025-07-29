//go:build linux
// +build linux

package trayapp

import "github.com/ParetoSecurity/agent/systemd"

// IsStartupEnabled checks if the tray icon systemd service is enabled
func IsStartupEnabled() bool {
	return systemd.IsTrayIconEnabled()
}

// EnableStartup enables the tray icon systemd service
func EnableStartup() error {
	return systemd.EnableTrayIcon()
}

// DisableStartup disables the tray icon systemd service
func DisableStartup() error {
	return systemd.DisableTrayIcon()
}
