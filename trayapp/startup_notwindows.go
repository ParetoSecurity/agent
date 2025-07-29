//go:build !windows && !linux
// +build !windows,!linux

package trayapp

// IsStartupEnabled is a stub for non-Windows, non-Linux platforms
func IsStartupEnabled() bool {
	return false
}

// EnableStartup is a stub for non-Windows, non-Linux platforms
func EnableStartup() error {
	return nil
}

// DisableStartup is a stub for non-Windows, non-Linux platforms
func DisableStartup() error {
	return nil
}
