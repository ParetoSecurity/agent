//go:build windows
// +build windows

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "embed"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
)

// Embed the uninstall.ps1 script
//
//go:embed uninstall.ps1
var uninstallScript string

// Embed the install.ps1 script

//go:embed install.ps1
var installScript string

type WindowService struct{}

var release struct {
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (w *WindowService) downloadApp() (string, error) {
	// Download latest release from GitHub (paretosecurity/agent, asset: paretosecurity.exe)
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	installPath := filepath.Join(roamingDir, "ParetoSecurity")
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return "", err
	}

	var exeURL string
	for _, asset := range release.Assets {
		// Example: paretosecurity_0.1.4_windows_amd64.zip
		if filepath.Ext(asset.Name) == ".zip" && strings.Contains(asset.Name, "paretosecurity") && strings.Contains(asset.Name, "windows") && strings.Contains(asset.Name, runtime.GOARCH) {
			exeURL = asset.BrowserDownloadURL
			break
		}
	}
	if exeURL == "" {
		return "", errors.New("no suitable asset found")
	}
	// Download the file
	zipPath := filepath.Join(installPath, filepath.Base(exeURL))
	err = requests.
		URL(exeURL).
		ToFile(zipPath).
		Fetch(context.Background())
	if err != nil {
		return "", err
	}

	return exeURL, nil
}

func (w *WindowService) getLatestRelease() error {
	// Get the latest release from GitHub (paretosecurity/agent)
	err := requests.
		URL("https://api.github.com/repos/paretosecurity/agent/releases/latest").
		ToJSON(&release).
		Fetch(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (w *WindowService) QuitApp() error {
	os.Exit(0)
	return nil
}

func (w *WindowService) writeUninstallScript(installDir string) error {
	uninstallPath := filepath.Join(installDir, "uninstall.ps1")
	return os.WriteFile(uninstallPath, []byte(uninstallScript), 0644)
}

func (w *WindowService) InstallApp(withStartup bool) error {

	w.getLatestRelease()
	exeURL, err := w.downloadApp()
	if err != nil {
		log.WithError(err).Error("failed to download latest release")
	}

	roamingDir, err := os.UserConfigDir()
	if err != nil {
		log.WithError(err).Error("failed to get roaming directory")
		return err
	}
	installScriptPath := filepath.Join(roamingDir, "ParetoSecurity", "install.ps1")

	// Write the install script
	err = os.WriteFile(installScriptPath, []byte(installScript), 0644)
	if err != nil {
		log.WithError(err).Error("failed to write install script")
		return err
	}

	// Execute the PowerShell script
	zipPath := filepath.Join(roamingDir, "ParetoSecurity", filepath.Base(exeURL))
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		log.WithError(err).Error("zip file does not exist")
		return err
	}
	displayVersion := runtime.Version()
	args := []string{"-ExecutionPolicy", "Bypass",
		"-File", installScriptPath,
		"-ZipPath", zipPath,
		"-DisplayVersion", displayVersion,
	}
	if withStartup {
		args = append(args, "-WithStartup")
	}

	// Run the install
	_, err = shared.RunCommand("powershell.exe", args...)
	if err != nil {
		log.WithError(err).Error("failed to execute install script")
		return err
	}

	// Remove the install script after execution
	err = os.Remove(installScriptPath)
	if err != nil {
		log.WithError(err).Error("failed to remove install script")
	}

	// Write the uninstall script
	w.writeUninstallScript(filepath.Join(roamingDir, "ParetoSecurity"))
	if err != nil {
		log.WithError(err).Error("failed to write uninstall script")
	}

	return nil
}
