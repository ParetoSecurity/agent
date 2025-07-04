//go:build windows
// +build windows

package main

func main() {
	app := NewTrayApp(nil)
	app.Run()
}
