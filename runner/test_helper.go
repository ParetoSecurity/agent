package runner

import (
	"os"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
)

// setupTestStatePath creates a temporary directory and sets the shared.StatePath
// to a file in that directory for testing.
func setupTestStatePath(t *testing.T) func() {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "paretosecurity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Save the original state path
	origStatePath := shared.StatePath

	// Set the state path to a file in the temp directory
	shared.StatePath = tmpDir + "/.paretosecurity.state"

	// Return a cleanup function
	return func() {
		// Restore the original state path
		shared.StatePath = origStatePath

		// Remove the temp directory
		os.RemoveAll(tmpDir)
	}
}
