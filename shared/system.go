package shared

import (
	"fmt"
	"net"
	"os"
	"testing"

	"strings"

	"github.com/google/uuid"
)

// systemUUID generates a unique system identifier based on the first available
// network interface's hardware address (MAC address). It iterates through all
// network interfaces, skips loopback interfaces, and uses the first interface
// with a valid hardware address (at least 6 bytes) to generate a SHA1-based
// UUID using the hardware address as input. Returns an error if no suitable
// network interface is found.
func systemUUID() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {

		// Skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if len(iface.HardwareAddr) >= 6 {
			hwAddr := iface.HardwareAddr
			// Create a namespace UUID from hardware address
			nsUUID := uuid.NewSHA1(uuid.NameSpaceOID, hwAddr)
			return nsUUID.String(), nil
		}
	}

	return "", fmt.Errorf("no network interface found")
}

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
