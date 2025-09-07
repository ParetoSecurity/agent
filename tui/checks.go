package tui

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ParetoSecurity/agent/check"
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
	return tea.Cmd(func() tea.Msg {
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

		// Run check in completely isolated subprocess to avoid terminal interference
		result = runCheckInSubprocess(result)
		return checkCompleteMsg{claimIdx: claimIdx, checkIdx: checkIdx, result: result}
	})
}

// runCheckInSubprocess runs a single check in an isolated subprocess
func runCheckInSubprocess(result checkResult) checkResult {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		result.Status = "Error"
		result.HasError = true
		result.Details = "Failed to get executable path: " + err.Error()
		return result
	}

	// Run the check command in a subprocess with all output redirected
	cmd := exec.Command(execPath, "check", "--only", result.Check.UUID())

	// Create completely isolated environment
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil

	// Set process group to isolate from terminal
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Run with timeout
	err = cmd.Run()

	if err != nil {
		result.Status = "Error"
		result.HasError = true
		result.Details = "Check failed: " + err.Error()
		if stderr.Len() > 0 {
			result.Details += " | " + strings.TrimSpace(stderr.String())
		}
		return result
	}

	// Check exit code to determine pass/fail
	if cmd.ProcessState.Success() {
		result.Status = "Pass"
		result.Passed = true
		result.HasError = false
		result.Details = result.Check.PassedMessage()
	} else {
		result.Status = "Fail"
		result.Passed = false
		result.HasError = false
		result.Details = result.Check.FailedMessage()
	}

	return result
}

// runAllChecks creates a command to run all security checks in batch
func runAllChecks(checks []checkResult) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		var results []checkResult

		// Get the current executable path
		execPath, err := os.Executable()
		if err != nil {
			// If we can't get executable path, run checks individually
			for _, check := range checks {
				result := check
				result.LastRun = time.Now()
				result.Status = "Error"
				result.HasError = true
				result.Details = "Failed to get executable path"
				results = append(results, result)
			}
			return batchRunMsg{results: results}
		}

		// Run all checks in a single subprocess to be completely isolated
		cmd := exec.Command(execPath, "check")

		// Create completely isolated environment
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Stdin = nil

		// Set process group to isolate from terminal
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}

		// Run all checks
		err = cmd.Run()

		// Update all results based on the final state
		currentStates := shared.GetLastStates()

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
			} else if state, exists := currentStates[result.Check.UUID()]; exists {
				// Use the state from the check run
				if state.HasError {
					result.Status = "Error"
					result.HasError = true
					result.Details = state.Details
				} else if state.Passed {
					result.Status = "Pass"
					result.Passed = true
					result.HasError = false
					result.Details = state.Details
				} else {
					result.Status = "Fail"
					result.Passed = false
					result.HasError = false
					result.Details = state.Details
				}
			} else {
				// Fallback if no state available
				result.Status = "Error"
				result.HasError = true
				result.Details = "No result available"
			}

			results = append(results, result)
		}

		return batchRunMsg{results: results}
	})
}
