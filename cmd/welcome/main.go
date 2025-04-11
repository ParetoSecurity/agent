package main

//go:generate go tool wails3 generate bindings -clean -b -d dist/assets/bindings

import (
	"embed"
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed dist/*
var welcomeAssets embed.FS

type WindowService struct{}

func (s *WindowService) GeneratePanic() {
	s.call1()
}

func (s *WindowService) call1() {
	s.call2()
}

func (s *WindowService) call2() {
	panic("oh no! something went wrong deep in my service! :(")
}

func main() {
	app := application.New(application.Options{
		Name:        "Single Instance Example",
		LogLevel:    slog.LevelDebug,
		Description: "An example of single instance functionality in Wails v3",
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
		Height:                     560,
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
