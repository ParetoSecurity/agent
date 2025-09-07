package tui

import (
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/runner"
	"github.com/ParetoSecurity/agent/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
)

// openBrowserForCheck opens a browser to the help page for a specific check
func openBrowserForCheck(check check.Check) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Determine the correct URL path based on OS
		arch := "check-linux"
		if runtime.GOOS == "windows" {
			arch = "check-windows"
		}

		// Get check status for error handling
		checkStatus, found, _ := shared.GetLastState(check.UUID())

		var targetURL string
		if found && checkStatus.HasError {
			targetURL = "https://paretosecurity.com/docs/linux/check-error"
		} else {
			targetURL = fmt.Sprintf("https://paretosecurity.com/%s/%s?details=%s", arch, check.UUID(), url.QueryEscape(check.Status()))
		}

		// Use the browsers package to open the URL
		if err := browser.OpenURL(targetURL); err != nil {
			// Log error but don't crash the TUI
			return nil
		}
		return nil
	})
}

// runSingleCheck creates a command to run a single security check
func runSingleCheck(claimIdx, checkIdx int, checkResult checkResult) tea.Cmd {
	return func() tea.Msg {
		result := checkResult
		result.LastRun = time.Now()

		// Skip checks that are not runnable or are disabled
		if !result.Check.IsRunnable() || shared.IsCheckDisabled(result.Check.UUID()) {
			result.Status = "Disabled"
			result.HasError = false
			if shared.IsCheckDisabled(result.Check.UUID()) {
				result.Details = "Disabled by config"
			} else {
				result.Details = result.Check.Status()
			}
			return checkCompleteMsg{claimIdx: claimIdx, checkIdx: checkIdx, result: result}
		}

		// Run check directly - this should be fast enough for most checks
		result = runCheckDirectly(result)
		return checkCompleteMsg{claimIdx: claimIdx, checkIdx: checkIdx, result: result}
	}
}

// runCheckDirectly runs a single check directly
func runCheckDirectly(result checkResult) checkResult {
	var hasError bool

	if result.Check.RequiresRoot() {
		// Run as root using the runner
		status, err := runner.RunCheckViaRoot(result.Check.UUID())
		if err != nil {
			result.Status = "Error"
			result.HasError = true
			result.Details = "Root check failed: " + err.Error()
			return result
		}
		result.Passed = status.Passed
		result.HasError = false
		result.Details = status.Details
		if result.Passed {
			result.Status = "Pass"
		} else {
			result.Status = "Fail"
		}
	} else {
		// Run check directly
		if err := result.Check.Run(); err != nil {
			hasError = true
			result.Status = "Error"
			result.HasError = true
			result.Details = "Check failed: " + err.Error()
			return result
		}

		result.Passed = result.Check.Passed()
		result.HasError = hasError
		if result.Passed {
			result.Status = "Pass"
			result.Details = result.Check.PassedMessage()
		} else {
			result.Status = "Fail"
			result.Details = result.Check.FailedMessage()
		}
	}

	// Update the last state for consistency with runner behavior
	shared.UpdateLastState(shared.LastState{
		UUID:     result.Check.UUID(),
		Name:     result.Check.Name(),
		Passed:   result.Passed,
		HasError: result.HasError,
		Details:  result.Details,
	})

	return result
}

// runAllChecks creates a command to run all security checks in batch
func runAllChecks(checks []checkResult) tea.Cmd {
	return func() tea.Msg {
		var results []checkResult

		for _, check := range checks {
			result := check
			result.LastRun = time.Now()

			// Skip checks that are not runnable or are disabled
			if !result.Check.IsRunnable() || shared.IsCheckDisabled(result.Check.UUID()) {
				result.Status = "Disabled"
				result.HasError = false
				if shared.IsCheckDisabled(result.Check.UUID()) {
					result.Details = "Disabled by config"
				} else {
					result.Details = result.Check.Status()
				}
			} else {
				// Run each check directly
				result = runCheckDirectly(result)
			}

			results = append(results, result)
		}

		// Commit all state changes at once for consistency
		if err := shared.CommitLastState(); err != nil {
			// If commit fails, don't fail the whole operation
			// Just continue with the results we have
		}

		return batchRunMsg{results: results}
	}
}
