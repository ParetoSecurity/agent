//go:build windows
// +build windows

package main

import (
	"embed"
	"log/slog"
	"os"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// InstallerConfig holds the configuration for the installer application
type InstallerConfig struct {
	Args       []string
	InstallApp func(bool) error
	Exit       func(int)
	NewApp     func(opts application.Options) *application.App
	Assets     embed.FS
}

// DefaultInstallerConfig returns the default configuration
func DefaultInstallerConfig(assets embed.FS) *InstallerConfig {
	return &InstallerConfig{
		Args:       os.Args[1:],
		InstallApp: shared.InstallApp,
		Exit:       os.Exit,
		NewApp:     application.New,
		Assets:     assets,
	}
}

// InstallerApp represents the installer application
type InstallerApp struct {
	config *InstallerConfig
}

// NewInstallerApp creates a new installer application instance
func NewInstallerApp(config *InstallerConfig) *InstallerApp {
	return &InstallerApp{config: config}
}

// Run executes the installer application logic
func (i *InstallerApp) Run() {
	// Check for silent install arguments
	if i.shouldInstallSilently() {
		err := i.config.InstallApp(true)
		if err != nil {
			slog.Error(err.Error())
		}
		i.config.Exit(0)
		return
	}

	// Create GUI application
	app := i.createGUIApp()
	i.createMainWindow(app)
	app.Run()
}

func (i *InstallerApp) shouldInstallSilently() bool {
	for _, arg := range i.config.Args {
		arg = strings.ToLower(arg)
		if arg == "/qs" || arg == "/qsp" || arg == "/s" || arg == "/q" {
			return true
		}
	}
	return false
}

func (i *InstallerApp) createGUIApp() *application.App {
	return i.config.NewApp(application.Options{
		Name:        "Pareto Security Installer",
		LogLevel:    slog.LevelInfo,
		Description: "Installer for Pareto Security Agent",
		Services: []application.Service{
			application.NewService(&WindowService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(i.config.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Windows: application.WindowsOptions{
			WndProcInterceptor:            nil,
			DisableQuitOnLastWindowClosed: false,
			WebviewUserDataPath:           "",
			WebviewBrowserPath:            "",
		},
	})
}

func (i *InstallerApp) createMainWindow(app *application.App) {
	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:                      "Welcome to Pareto Security",
		Width:                      360,
		Height:                     580,
		URL:                        "/",
		AlwaysOnTop:                true,
		DisableResize:              true,
		FullscreenButtonEnabled:    false,
		DefaultContextMenuDisabled: true,
		MinimiseButtonState:        application.ButtonHidden,
		MaximiseButtonState:        application.ButtonHidden,
		Mac: application.MacWindow{
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInsetUnified,
			InvisibleTitleBarHeight: 50,
			WindowLevel:             application.MacWindowLevelFloating,
		},
	})
}
