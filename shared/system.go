package shared

import (
	"os"
	"testing"

	"strings"
)

// IsRoot returns true if the current process is running with root privileges.
// When running tests, it always returns true to avoid permission-related test failures.
// For normal execution, it checks if the effective user ID is 0 (root).
func IsRoot() bool {
	if testing.Testing() {
		return true
	}
	return os.Geteuid() == 0
}

// SelfExe returns the path to the current executable.
// If the executable path cannot be determined, it returns "paretosecurity" as a fallback.
// The function also removes any "-tray" suffix from the executable path, which is used
// for Windows standalone builds where the tray version has a different executable name.
func SelfExe() string {
	exePath, err := os.Executable()
	if err != nil {
		return "paretosecurity"
	}
	// Remove the -tray suffix from the executable name (WIN, standalone)
	return strings.Replace(exePath, "-tray", "", -1) // Remove -tray from the path)
}
