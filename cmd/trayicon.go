//go:build linux || darwin
// +build linux darwin

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/allan-simon/go-singleinstance"
	"github.com/caarlos0/log"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

var trayiconCmd = &cobra.Command{
	Use:   "trayicon",
	Short: "Display the status of the checks in the system tray",
	Run: func(cc *cobra.Command, args []string) {

		lockFile, err := singleinstance.CreateLockFile("paretosecurity-tray.lock")
		if err != nil {
			log.WithError(err).Fatal("An instance of ParetoSecurity tray application is already running.")
			return
		}
		defer lockFile.Close()

		onExit := func() {
			log.Info("Exiting...")
		}

		trayApp := trayapp.NewTrayApp()

		// On Linux, handle potential systray registration failure
		if runtime.GOOS == "linux" {
			// Check if the desktop environment supports StatusNotifierWatcher
			if !checkStatusNotifierSupport() {
				handleSystrayError()
				return
			}
		}

		systray.Run(trayApp.OnReady, onExit)
	},
}

func checkStatusNotifierSupport() bool {
	// First try using dbus-send command
	output, err := shared.RunCommand("dbus-send", "--session", "--dest=org.freedesktop.DBus", "--type=method_call", "--print-reply", "/org/freedesktop/DBus", "org.freedesktop.DBus.ListNames")
	if err == nil {
		// Check if StatusNotifierWatcher is in the list of available services
		if strings.Contains(output, "org.kde.StatusNotifierWatcher") || strings.Contains(output, "org.freedesktop.StatusNotifierWatcher") {
			return true
		}
		return false
	}

	// Fallback to native D-Bus package if dbus-send fails
	log.WithError(err).Debug("dbus-send failed, trying native D-Bus")
	conn, err := dbus.SessionBus()
	if err != nil {
		log.WithError(err).Debug("Failed to connect to D-Bus session bus")
		return false
	}
	defer conn.Close()

	// Get list of names on the bus
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	call := obj.Call("org.freedesktop.DBus.ListNames", 0)
	if call.Err != nil {
		log.WithError(call.Err).Debug("Failed to list D-Bus names")
		return false
	}

	var names []string
	err = call.Store(&names)
	if err != nil {
		log.WithError(err).Debug("Failed to parse D-Bus names")
		return false
	}

	// Check if StatusNotifierWatcher is available
	for _, name := range names {
		if name == "org.kde.StatusNotifierWatcher" || name == "org.freedesktop.StatusNotifierWatcher" {
			return true
		}
	}

	return false
}

func handleSystrayError() {
	errorMsg := `System tray error: StatusNotifierWatcher not found.

This usually means your desktop environment doesn't support the modern system tray protocol.

To fix this issue, you can:
1. Install the gnome-shell-extension-appindicator (already recommended in the package)
2. Install snixembed for compatibility with older desktop environments
3. For NixOS users: Enable services.status-notifier-watcher in Home Manager
4. For Wayland users: Use waybar with tray support enabled
5. Check the documentation for more solutions

Opening documentation in your browser...`

	log.Error(errorMsg)
	fmt.Fprintln(os.Stderr, errorMsg)
	os.Exit(1)
}

func init() {

	rootCmd.AddCommand(trayiconCmd)
}
