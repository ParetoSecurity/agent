package shared

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestOnlyInstance(t *testing.T) {
	t.Run("creates lock file when none exists", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		err := OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify lock file exists and contains current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Fatalf("expected PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("returns error when same process tries again", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// First call should succeed
		err := OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("first call should succeed: %v", err)
		}

		// Second call should fail
		err = OnlyInstance(lockPath)
		if err == nil {
			t.Fatal("expected error on second call")
		}

		expectedMsg := "another instance is already running"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Fatalf("expected error containing '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("removes stale lock file with invalid PID", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// Create lock file with invalid PID
		err := os.WriteFile(lockPath, []byte("invalid"), 0644)
		if err != nil {
			t.Fatalf("failed to create test lock file: %v", err)
		}

		// Should succeed and create new lock
		err = OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify lock file contains current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Fatalf("expected PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("removes stale lock file with non-existent PID", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// Create lock file with non-existent PID (99999 should be safe)
		err := os.WriteFile(lockPath, []byte("99999"), 0644)
		if err != nil {
			t.Fatalf("failed to create test lock file: %v", err)
		}

		// Should succeed and create new lock
		err = OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify lock file contains current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Fatalf("expected PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("returns error when cannot create lock file", func(t *testing.T) {
		// Try to create lock in non-existent directory
		lockPath := "/non/existent/directory/test.lock"

		err := OnlyInstance(lockPath)
		if err == nil {
			t.Fatal("expected error when cannot create lock file")
		}
	})
}
