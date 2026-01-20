//go:build windows

package shared

import (
	"context"
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
)

var release struct {
	Name   string `json:"name"`
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Embed the uninstall.ps1 script
//
//go:embed uninstall.ps1
var uninstallScript string

// Embed the install.ps1 script

//go:embed install.ps1
var installScript string

func getApp() (string, string, error) {
	// Download latest release from GitHub (paretosecurity/agent, asset: paretosecurity.exe)
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return "", "", err
	}
	installPath := filepath.Join(roamingDir, "ParetoSecurity")
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return "", "", err
	}

	var exeURL, version string
	for _, asset := range release.Assets {
		// Example: paretosecurity_0.1.4_windows_amd64.zip
		if filepath.Ext(asset.Name) == ".zip" &&
			strings.Contains(asset.Name, "paretosecurity") &&
			strings.Contains(asset.Name, "windows") &&
			strings.Contains(asset.Name, runtime.GOARCH) {
			exeURL = asset.BrowserDownloadURL
			version = release.Name
			break
		}
	}
	if exeURL == "" {
		return "", "", errors.New("no suitable asset found")
	}
	// Download the file
	zipPath := filepath.Join(installPath, filepath.Base(exeURL))
	err = requests.
		URL(exeURL).
		ToFile(zipPath).
		Fetch(context.Background())
	if err != nil {
		return "", "", err
	}

	return exeURL, version, nil
}

func getLatestRelease() error {

	// Create a context with a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get the latest release from GitHub (paretosecurity/agent)
	err := requests.
		URL("https://api.github.com/repos/paretosecurity/agent/releases/latest").
		Header("User-Agent", UserAgent()).
		ToJSON(&release).
		Fetch(ctx)
	if err != nil {
		return err
	}

	return nil
}

func UpdateApp() error {
	err := getLatestRelease()
	if err != nil {
		log.WithError(err).Error("failed to get latest release")
		return err
	}
	_, version, err := getApp()
	if err != nil {
		log.WithError(err).Error("failed to download latest release")
		return err
	}

	if version == Version {
		log.Info("already on the latest version")
		return nil
	}

	return InstallApp(false)
}

func InstallApp(withStartup bool) error {
	err := getLatestRelease()
	if err != nil {
		log.WithError(err).Error("failed to get latest release")
		return err
	}
	exeURL, _, err := getApp()
	log.WithField("exeURL", exeURL).Info("downloading latest release")
	if err != nil {
		log.WithError(err).Error("failed to download latest release")
		return err
	}

	roamingDir, err := os.UserConfigDir()
	if err != nil {
		log.WithError(err).Error("failed to get roaming directory")
		return err
	}
	installScriptPath := filepath.Join(roamingDir, "ParetoSecurity", "install.ps1")

	// Write the install script
	log.WithField("installScriptPath", installScriptPath).Info("writing install script")
	err = os.WriteFile(installScriptPath, []byte(installScript), 0644)
	if err != nil {
		log.WithError(err).Error("failed to write install script")
		return err
	}

	// Execute the PowerShell script
	log.WithField("zipPath", exeURL).Info("executing install script")
	zipPath := filepath.Join(roamingDir, "ParetoSecurity", filepath.Base(exeURL))
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		log.WithError(err).Error("zip file does not exist")
		return err
	}

	args := []string{
		"-ExecutionPolicy", "Bypass",
		"-File", installScriptPath,
		"-ZipPath", zipPath,
		"-DisplayVersion", Version,
	}
	if withStartup {
		args = append(args, "-WithStartup")
	}

	// Run the install
	_, err = RunCommand("powershell.exe", args...)
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
	uninstallPath := filepath.Join(roamingDir, "ParetoSecurity", "uninstall.ps1")
	err = os.WriteFile(uninstallPath, []byte(uninstallScript), 0644)
	if err != nil {
		log.WithError(err).Error("failed to write uninstall script")
	}
	return nil
}
