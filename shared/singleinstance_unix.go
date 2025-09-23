//go:build unix

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
			// Same PID - this process already holds the lock
			if pid == currentPID {
				return nil
			}

			// Different PID - check if process is actually running
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("another instance is already running (PID: %d)", pid)
				}
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
