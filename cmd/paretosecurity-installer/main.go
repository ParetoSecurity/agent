//go:build windows
// +build windows

package main

import (
	"embed"
)

//go:embed ui/dist/*
var welcomeAssets embed.FS

func main() {
	config := DefaultInstallerConfig(welcomeAssets)
	app := NewInstallerApp(config)
	app.Run()
}
