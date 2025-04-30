//go:build windows
// +build windows

package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed ui/dist/*
var welcomeAssets embed.FS

func main() {

	// check for /qs or /qsp argument
	// if found, install the app and exit
	for _, arg := range os.Args[1:] {
		if arg == "/qs" || arg == "/qsp" {
			(&WindowService{}).InstallApp(true)
			os.Exit(0)
			return
		}
	}

	app := application.New(application.Options{
		Name:        "Pareto Security Installer",
		LogLevel:    slog.LevelInfo,
		Description: "Installer for Pareto Security Agent",
		Services: []application.Service{
			application.NewService(&WindowService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(welcomeAssets),
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

	app.Run()
}
