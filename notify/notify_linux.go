package notify

import (
	"github.com/caarlos0/log"
	"github.com/godbus/dbus/v5"
)

func Blocking(message string) {
	conn, err := dbus.SessionBus()
	if err != nil {
		log.WithError(err).Error("Failed to connect to session bus")
		return
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	// Add signal matching
	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/Notifications"),
		dbus.WithMatchInterface("org.freedesktop.Notifications"),
		dbus.WithMatchMember("ActionInvoked"),
	); err != nil {
		log.WithError(err).Error("Failed to add signal match")
		return
	}

	// Create a channel to receive the signal
	signals := make(chan *dbus.Signal, 1)
	conn.Signal(signals)

	// Send notification with an action button
	call := obj.Call("org.freedesktop.Notifications.Notify", 0,
		"ParetoSecurity",          // Application name
		uint32(0),                 // Replace ID
		"dialog-information",      // Icon (system dialog icon)
		"Pareto Security",         // Summary
		message,                   // Body
		[]string{"default", "OK"}, // Actions (default is the action id, OK is the label)
		map[string]interface{}{
			"urgency": byte(2), // Critical urgency
		},
		int32(-1)) // Timeout (-1 means no timeout)

	if call.Err != nil {
		log.WithError(call.Err).Error("Failed to send notification")
		return
	}

	var notificationId uint32
	call.Store(&notificationId)

	// Wait for action
	for signal := range signals {
		if signal.Name == "org.freedesktop.Notifications.ActionInvoked" {
			id := signal.Body[0].(uint32)
			action := signal.Body[1].(string)
			if id == notificationId {
				log.Infof("Action invoked: %s", action)
				return
			}
		}
	}

	return
}
