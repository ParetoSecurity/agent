//go:build windows
// +build windows

package trayapp

import (
	_ "embed"

	"os"
	"path/filepath"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

//go:embed startup.ps1
var startupScript string

// runStartupScript executes the PowerShell startup script with the given action
func runStartupScript(action string) (string, error) {
	roamingDir := os.Getenv("APPDATA")
	if roamingDir == "" {
		log.Error("APPDATA environment variable not found")
		return "", os.ErrNotExist
	}

	installPath := filepath.Join(roamingDir, "ParetoSecurity")
	scriptPath := filepath.Join(installPath, "startup.ps1")

	// Ensure the install directory exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		log.WithError(err).Error("failed to create install directory")
		return "", err
	}

	// Write the startup script
	if err := os.WriteFile(scriptPath, []byte(startupScript), 0644); err != nil {
		log.WithError(err).Error("failed to write startup script")
		return "", err
	}

	// Clean up the script after execution
	defer func() {
		if err := os.Remove(scriptPath); err != nil {
			log.WithError(err).Error("failed to remove startup script")
		}
	}()

	// Execute the PowerShell script
	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-File", scriptPath,
		"-Action", action,
		"-InstallPath", installPath,
	}

	output, err := shared.RunCommand("powershell.exe", args...)
	if err != nil {
		log.WithError(err).WithField("action", action).Error("failed to execute startup script")
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// getStartupShortcutPath returns the path to the startup shortcut
func getStartupShortcutPath() string {
	roamingDir := os.Getenv("APPDATA")
	if roamingDir == "" {
		return ""
	}
	return filepath.Join(roamingDir, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "Pareto Security.lnk")
}

// IsStartupEnabled checks if the startup shortcut exists using Go file operations (fast)
func IsStartupEnabled() bool {
	shortcutPath := getStartupShortcutPath()
	if shortcutPath == "" {
		return false
	}

	_, err := os.Stat(shortcutPath)
	return err == nil
}

// EnableStartup creates a startup shortcut
func EnableStartup() error {
	log.Info("Enabling startup shortcut")

	_, err := runStartupScript("enable")
	if err != nil {
		log.WithError(err).Error("failed to enable startup")
		return err
	}

	log.Info("Startup shortcut enabled successfully")
	return nil
}

// DisableStartup removes the startup shortcut
func DisableStartup() error {
	log.Info("Disabling startup shortcut")

	_, err := runStartupScript("disable")
	if err != nil {
		log.WithError(err).Error("failed to disable startup")
		return err
	}

	log.Info("Startup shortcut disabled successfully")
	return nil
}
