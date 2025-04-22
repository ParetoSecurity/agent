package trayapp

import (
	"log"
	"sync"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	themeChangeOnce sync.Once
)

// SubscribeToThemeChanges starts monitoring theme changes and sends updates to the provided channel.
func SubscribeToThemeChanges(themeChangeChan chan<- bool) {
	themeChangeOnce.Do(func() {
		go func() {
			for {
				// Open the registry key for monitoring
				key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.NOTIFY)
				if err != nil {
					log.Printf("Failed to open registry key: %v", err)
					time.Sleep(10 * time.Second) // Retry after delay
					continue
				}

				// Create an event for change notifications
				event, err := windows.CreateEvent(nil, 0, 0, nil)
				if err != nil {
					log.Printf("Failed to create event: %v", err)
					key.Close()
					time.Sleep(10 * time.Second)
					continue
				}
				err = windows.RegNotifyChangeKeyValue(windows.Handle(key), true, windows.REG_NOTIFY_CHANGE_LAST_SET, event, false)
				key.Close() // Close the key after setting up notification

				// Wait for the event to be triggered
				windows.WaitForSingleObject(event, windows.INFINITE)
				windows.CloseHandle(event)

				// Notify the channel about the theme change
				select {
				case themeChangeChan <- IsDarkTheme():
				default:
					log.Println("Theme change notification dropped due to no listener")
				}

				time.Sleep(10 * time.Second) // Retry after delay
			}
		}()
	})
}

func IsDarkTheme() bool {
	// Equivalent to: (Get-ItemProperty -Path "HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize").AppsUseLightTheme -eq 0
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	val, _, err := key.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		return false
	}
	return val == 0
}
