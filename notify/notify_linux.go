package notify

import (
	"github.com/caarlos0/log"
	"github.com/godbus/dbus/v5"
)

// Toast sends a persistent desktop notification using D-Bus on Linux systems.
// It displays a notification with the title "Pareto Security" and the provided body text.
// The notification is configured to be resident (persistent) and will expire after 5 seconds.
func Toast(body string) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.WithError(err).Error("failed to connect to session bus")
		return
	}
	defer conn.Close()
	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	call := obj.Call("org.freedesktop.Notifications.Notify", 0,
		"pareto-agent",       // app_name
		uint32(0),            // replaces_id
		"dialog-information", // app_icon
		"Pareto Security",    // summary
		body,                 // body
		[]string{},           // actions
		map[string]dbus.Variant{
			"resident": dbus.MakeVariant(true), // keeps notification persistent
		},
		int32(10000), // expire_timeout (0 = no expiration)
	)

	if call.Err != nil {
		log.WithError(call.Err).Error("failed to send notification")
	}
}
