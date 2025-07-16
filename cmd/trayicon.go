//go:build linux || darwin
// +build linux darwin

package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/trayapp"
	"github.com/caarlos0/log"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var trayiconCmd = &cobra.Command{
	Use:   "trayicon",
	Short: "Display the status of the checks in the system tray",
	Run: func(cc *cobra.Command, args []string) {

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

			// Set up a channel to capture potential panic or error
			done := make(chan bool, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok && strings.Contains(err.Error(), "StatusNotifierWatcher") {
							handleSystrayError()
						}
					}
					done <- true
				}()

				systray.Run(trayApp.OnReady, onExit)
			}()

			// Wait for a short time to see if systray starts successfully
			select {
			case <-done:
				// If we get here quickly, it might be due to an error
				time.Sleep(100 * time.Millisecond)
			case <-time.After(2 * time.Second):
				// Systray seems to have started successfully
				<-done
			}
		} else {
			systray.Run(trayApp.OnReady, onExit)
		}
	},
}

func checkStatusNotifierSupport() bool {
	// Check if StatusNotifierWatcher is available on the D-Bus session bus
	output, err := shared.RunCommand("dbus-send", "--session", "--dest=org.freedesktop.DBus", "--type=method_call", "--print-reply", "/org/freedesktop/DBus", "org.freedesktop.DBus.ListNames")
	if err != nil {
		log.WithError(err).Debug("Failed to check D-Bus services")
		return true // Assume support if we can't check
	}

	// Check if StatusNotifierWatcher is in the list of available services
	if strings.Contains(output, "org.kde.StatusNotifierWatcher") {
		return true
	}

	// Also check for alternative implementations
	if strings.Contains(output, "org.freedesktop.StatusNotifierWatcher") {
		return true
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

	// Try to open browser with documentation
	docURL := "https://paretosecurity.com/docs/linux/trayicon"
	if err := browser.OpenURL(docURL); err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed to open browser. Please visit: %s\n", docURL)
	}

	os.Exit(1)
}

func init() {

	rootCmd.AddCommand(trayiconCmd)
}
