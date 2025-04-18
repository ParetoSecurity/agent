package main

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type WindowService struct{}

var release struct {
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (w *WindowService) installApp(name string) error {
	// Download latest release from GitHub (paretosecurity/agent, asset: paretosecurity.exe)
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	installPath := filepath.Join(roamingDir, "ParetoSecurity")
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return err
	}

	var exeURL string
	for _, asset := range release.Assets {
		// Example: paretosecurity_0.1.4_windows_amd64.zip
		if filepath.Ext(asset.Name) == ".zip" && strings.Contains(asset.Name, name) && strings.Contains(asset.Name, "windows") && strings.Contains(asset.Name, runtime.GOARCH) {
			exeURL = asset.BrowserDownloadURL
			break
		}
	}
	if exeURL == "" {
		return errors.New("no suitable asset found")
	}
	// Download the file
	zipPath := filepath.Join(installPath, filepath.Base(exeURL))
	err = requests.
		URL(exeURL).
		ToFile(zipPath).
		Fetch(context.Background())
	if err != nil {
		return err
	}

	// Unzip the file
	if err := unzip(zipPath, installPath); err != nil {
		log.WithError(err).Error("failed to unzip")
	}
	return nil
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

func (w *WindowService) createShortcut(targetPath, shortcutPath, description string) error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	shell, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer shell.Release()

	wshell, err := shell.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer wshell.Release()

	shortcut, err := oleutil.CallMethod(wshell, "CreateShortcut", shortcutPath)
	if err != nil {
		return err
	}
	defer shortcut.ToIDispatch().Release()

	oleutil.PutProperty(shortcut.ToIDispatch(), "TargetPath", targetPath)
	oleutil.PutProperty(shortcut.ToIDispatch(), "Description", description)
	_, err = oleutil.CallMethod(shortcut.ToIDispatch(), "Save")
	if err != nil {
		return err
	}

	return nil
}

func (w *WindowService) createDesktopShortcut() error {
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	installPath := filepath.Join(roamingDir, "ParetoSecurity", "paretosecurity-tray.exe")

	desktop, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shortcutPath := filepath.Join(desktop, "Desktop", "Pareto Security.lnk")

	return w.createShortcut(installPath, shortcutPath, "Pareto Security")
}

func (w *WindowService) createStartupShortcut() error {
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	installPath := filepath.Join(roamingDir, "ParetoSecurity", "paretosecurity-tray.exe")

	startupPath := filepath.Join(roamingDir, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "Pareto Security.lnk")

	return w.createShortcut(installPath, startupPath, "Pareto Security")
}

func (w *WindowService) QuitApp() error {
	os.Exit(0)
	return nil
}

func (w *WindowService) InstallApp(withStartup bool) error {
	err := w.getLatestRelease()
	if err != nil {
		log.WithError(err).Error("failed to get latest release")
		return err
	}
	// Check if the release has assets
	if len(release.Assets) == 0 {
		return errors.New("no assets found for this release")
	}

	// Install the app
	err = w.installApp("paretosecurity")
	if err != nil {
		log.WithError(err).Error("failed to install app")
		return err
	}

	// Create desktop shortcut
	err = w.createDesktopShortcut()
	if err != nil {
		log.WithError(err).Error("failed to create desktop shortcut")
		return err
	}

	// Create startup shortcut if requested
	if withStartup {
		err = w.createStartupShortcut()
		if err != nil {
			log.WithError(err).Error("failed to create startup shortcut")
			return err
		}
	}
	// Start the app
	roamingDir, err := os.UserConfigDir()
	if err != nil {
		log.WithError(err).Error("failed to get roaming directory")
	}

	trayPath := filepath.Join(roamingDir, "ParetoSecurity", "paretosecurity-tray.exe")
	if _, err := os.StartProcess(trayPath, []string{}, &os.ProcAttr{}); err != nil {
		log.WithError(err).Error("failed to start app")
	}
	return nil
}

// unzip extracts a zip archive to a destination directory.
func unzip(src, dest string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	stat, err := r.Stat()
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(r, stat.Size())
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
