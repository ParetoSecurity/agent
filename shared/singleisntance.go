package shared

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func OnlyInstance(lockPath string) error {
	// Check if lock file exists
	if _, err := os.Stat(lockPath); err == nil {
		// File exists, read PID and check if process is running
		data, err := os.ReadFile(lockPath)
		if err != nil {
			return err
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			// Invalid PID, remove file and continue
			os.Remove(lockPath)
		} else {
			// Check if process exists
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("another instance is already running (PID: %d)", pid)
				}
			}
			// Process doesn't exist, remove stale lock file
			os.Remove(lockPath)
		}
	}

	// Create/write PID to lock file
	file, err := os.Create(lockPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write current PID to file. Note: on Windows, PIDs are 32 bit unsigned,
	// on unix 64 bit so should be enough for a while.
	_, err = file.WriteString(strconv.Itoa(os.Getpid()))
	return err
}
