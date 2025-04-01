package notify

import (
	"github.com/godbus/dbus/v5"
)

func Blocking(message string) (string, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	// Add signal matching
	if err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/freedesktop/Notifications"),
		dbus.WithMatchInterface("org.freedesktop.Notifications"),
		dbus.WithMatchMember("ActionInvoked"),
	); err != nil {
		return "", err
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
		return "", call.Err
	}

	var notificationId uint32
	call.Store(&notificationId)

	// Wait for action
	for signal := range signals {
		if signal.Name == "org.freedesktop.Notifications.ActionInvoked" {
			id := signal.Body[0].(uint32)
			action := signal.Body[1].(string)
			if id == notificationId {
				return action, nil
			}
		}
	}

	return "", nil
}
