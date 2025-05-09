package trayapp

import (
	"fmt"
	"time"

	"github.com/ParetoSecurity/agent/shared"
)

// lastUpdated calculates and returns a human-readable string representing the time elapsed since the last modification.
func lastUpdated() string {
	if shared.GetModifiedTime().IsZero() {
		return "never"
	}

	t := time.Since(shared.GetModifiedTime())

	switch {
	case t < time.Minute:
		return "just now"

	case t < time.Hour:
		// Less than an hour, show minutes
		minutes := int(t.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)

	case t < time.Hour*24:
		// Less than a day, show hours and minutes
		hours := int(t.Hours())
		minutes := int(t.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh ago", hours)
		}
		return fmt.Sprintf("%dh %dm ago", hours, minutes)

	case t < time.Hour*24*7:
		// Less than a week, show days and hours
		days := int(t.Hours() / 24)
		hours := int(t.Hours()) % 24
		if hours == 0 {
			return fmt.Sprintf("%dd ago", days)
		}
		return fmt.Sprintf("%dd %dh ago", days, hours)

	default:
		// More than a week, show in weeks
		days := int(t.Hours() / 24)
		weeks := days / 7
		remainingDays := days % 7

		if remainingDays == 0 {
			return fmt.Sprintf("%dw ago", weeks)
		}
		return fmt.Sprintf("%dw %dd ago", weeks, remainingDays)
	}
}
