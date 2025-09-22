package shared

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func OnlyInstance(lockPath string) error {
	// Try to create the lock file with O_EXCL to ensure atomicity
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err == nil {
		// Successfully created the file - we're the only instance
		defer file.Close()

		// Write current PID to file. Note: on Windows, PIDs are 32 bit unsigned,
		// on unix 64 bit so should be enough for a while.
		_, err = file.WriteString(strconv.Itoa(os.Getpid()))
		return err
	}

	// If error is not "file exists", return it
	if !os.IsExist(err) {
		return err
	}

	// File exists, read PID and check if process is running
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		// Invalid PID, remove file and try again
		os.Remove(lockPath)
		return OnlyInstance(lockPath)
	}

	// Check if process exists
	if process, err := os.FindProcess(pid); err == nil {
		if err := process.Signal(syscall.Signal(0)); err == nil {
			return fmt.Errorf("another instance is already running (PID: %d)", pid)
		}
	}

	// Process doesn't exist, remove stale lock file and try again
	os.Remove(lockPath)
	return OnlyInstance(lockPath)
}
