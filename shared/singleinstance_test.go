package shared

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestOnlyInstance(t *testing.T) {
	t.Run("first instance succeeds", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		err := OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify lock file was created with current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Errorf("expected PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("second instance with same PID succeeds", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// First call
		err := OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("first call should succeed: %v", err)
		}

		// Second call with same process should succeed
		err = OnlyInstance(lockPath)
		if err != nil {
			t.Errorf("second call with same PID should succeed: %v", err)
		}
	})

	t.Run("stale lock file is removed", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// Create stale lock file with non-existent PID
		stalePID := "999999"
		err := os.WriteFile(lockPath, []byte(stalePID), 0644)
		if err != nil {
			t.Fatalf("failed to create stale lock file: %v", err)
		}

		err = OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("should succeed with stale lock file: %v", err)
		}

		// Verify new lock file has current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Errorf("expected current PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("invalid lock file content is handled", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// Create lock file with invalid content
		err := os.WriteFile(lockPath, []byte("invalid-pid"), 0644)
		if err != nil {
			t.Fatalf("failed to create invalid lock file: %v", err)
		}

		err = OnlyInstance(lockPath)
		if err != nil {
			t.Fatalf("should succeed with invalid lock file: %v", err)
		}

		// Verify new lock file has current PID
		data, err := os.ReadFile(lockPath)
		if err != nil {
			t.Fatalf("lock file should exist: %v", err)
		}

		pid, err := strconv.Atoi(string(data))
		if err != nil {
			t.Fatalf("lock file should contain valid PID: %v", err)
		}

		if pid != os.Getpid() {
			t.Errorf("expected current PID %d, got %d", os.Getpid(), pid)
		}
	})

	t.Run("write permission error", func(t *testing.T) {
		// Use a path that doesn't exist and can't be created
		lockPath := "/non-existent-dir/test.lock"

		err := OnlyInstance(lockPath)
		if err == nil {
			t.Error("expected error when writing to invalid path")
		}
	})

	t.Run("existing valid process returns error", func(t *testing.T) {
		tempDir := t.TempDir()
		lockPath := filepath.Join(tempDir, "test.lock")

		// Write current PID to simulate another instance
		currentPID := os.Getpid()
		err := os.WriteFile(lockPath, []byte(strconv.Itoa(currentPID)), 0644)
		if err != nil {
			t.Fatalf("failed to create lock file: %v", err)
		}

		// This should succeed since it's the same process
		err = OnlyInstance(lockPath)
		if err != nil {
			t.Errorf("same process should succeed: %v", err)
		}
	})
}
