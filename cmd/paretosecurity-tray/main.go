//go:build windows
// +build windows

package main

import ()

func main() {
	app := NewTrayApp(nil)
	app.Run()
}
