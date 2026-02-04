package shared

import (
	"fmt"
	"time"

	sharedG "github.com/ParetoSecurity/agent/shared"
)

// TeamReportSentCheck verifies that team reports have been sent recently.
type TeamReportSentCheck struct {
	passed  bool
	details string
}

// Name returns the name of the check.
func (t *TeamReportSentCheck) Name() string {
	return "Pareto Cloud is receiving reports"
}

// Run executes the check.
func (t *TeamReportSentCheck) Run() error {
	lastSuccess := sharedG.Config.LastTeamReportSuccess
	if lastSuccess == 0 {
		t.passed = false
		t.details = "No successful report has been sent"
		return nil
	}

	lastTime := time.UnixMilli(lastSuccess)
	threshold := 25 * time.Hour
	t.passed = time.Since(lastTime) <= threshold
	t.details = fmt.Sprintf("Last successful report: %s", formatRelativeTime(lastTime))
	return nil
}

// Passed returns the status of the check.
func (t *TeamReportSentCheck) Passed() bool {
	return t.passed
}

// IsRunnable returns whether the check can run.
func (t *TeamReportSentCheck) IsRunnable() bool {
	return sharedG.IsLinked()
}

// UUID returns the UUID of the check.
func (t *TeamReportSentCheck) UUID() string {
	return "e29dfff7-afe3-4800-8919-74be2d74c3be"
}

// PassedMessage returns the message to return if the check passed.
func (t *TeamReportSentCheck) PassedMessage() string {
	return "Pareto Cloud is receiving reports"
}

// FailedMessage returns the message to return if the check failed.
func (t *TeamReportSentCheck) FailedMessage() string {
	return "Pareto Cloud is not receiving reports"
}

// RequiresRoot returns whether the check requires root access.
func (t *TeamReportSentCheck) RequiresRoot() bool {
	return false
}

// Status returns the status of the check.
func (t *TeamReportSentCheck) Status() string {
	if t.passed {
		if t.details != "" {
			return t.details
		}
		return t.PassedMessage()
	}
	if t.details != "" {
		return t.details
	}
	return t.FailedMessage()
}

func formatRelativeTime(at time.Time) string {
	now := time.Now()
	if at.After(now) {
		return "just now"
	}
	diff := now.Sub(at)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
