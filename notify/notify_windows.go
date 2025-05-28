package notify

import (
	"github.com/caarlos0/log"
	"github.com/kolide/toast"
)

// Toast displays a notification balloon on Windows using the Shell_NotifyIcon API.
// If the notification fails to display, an error is logged.
func Toast(message string) {
	notification := toast.Notification{
		AppID:   "Pareto Security",
		Title:   "Notification",
		Message: message,
	}
	err := notification.Push()
	if err != nil {
		log.WithError(err).Error("failed to send notification")
		return
	}
}
