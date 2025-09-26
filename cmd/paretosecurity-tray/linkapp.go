//go:build windows
// +build windows

package main

import (
	"fmt"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// createLinkApp creates a Wails application for the device linking UI
func createLinkApp(inviteID, host string) *application.App {
	app := application.New(application.Options{
		Name:        "Pareto Security Link",
		LogLevel:    slog.LevelInfo,
		Description: "Link device to Pareto Security team",
		Services: []application.Service{
			application.NewService(&LinkService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(linkAssets),
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

	// Build the URL with query parameters for the Elm app
	windowURL := "/"
	if inviteID != "" {
		windowURL = fmt.Sprintf("/?inviteId=%s", inviteID)
		if host != "" {
			windowURL = fmt.Sprintf("%s&host=%s", windowURL, host)
		}
	}

	// Create the main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:                      "Link Device - Pareto Security",
		Width:                      360,
		Height:                     580,
		URL:                        windowURL,
		AlwaysOnTop:                true,
		DisableResize:              true,
		DefaultContextMenuDisabled: true,
		MinimiseButtonState:        application.ButtonHidden,
		MaximiseButtonState:        application.ButtonHidden,
	})

	return app
}
