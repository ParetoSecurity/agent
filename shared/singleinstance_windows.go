//go:build windows

package shared

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

// OnlyInstance ensures that only one instance of the application is running by using a PID lock file.
func OnlyInstance(lockPath string) error {
	currentPID := os.Getpid()

	// Check if lock file exists
	if data, err := os.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(string(data)); err == nil {
			// Same PID - we already hold the lock
			if pid == currentPID {
				return nil
			}
			// Different PID - check if process exists
			if isProcessRunning(pid) {
				return fmt.Errorf("another instance is already running (PID: %d)", pid)
			}
		}
		// Invalid PID or process not running - remove stale file
		os.Remove(lockPath)
	}

	// Create/update lock file with current PID
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(currentPID)), 0644); err != nil {
		return err
	}
	return nil
}

// isProcessRunning checks if a process with the given PID is actually running on Windows
func isProcessRunning(pid int) bool {
	const PROCESS_QUERY_INFORMATION = 0x0400

	handle, err := syscall.OpenProcess(PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	// If we can open the process, check if it's still alive
	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	// STILL_ACTIVE constant is 259
	return exitCode == 259
}
