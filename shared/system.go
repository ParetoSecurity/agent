package shared

import (
	"fmt"
	"net"
	"os"
	"testing"

	"strings"

	"github.com/google/uuid"
)

func SystemUUID() (string, error) {
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

func IsRoot() bool {
	if testing.Testing() {
		return true
	}
	return os.Geteuid() == 0
}

func SelfExe() string {
	exePath, err := os.Executable()
	if err != nil {
		return "paretosecurity"
	}
	// Remove the -tray suffix from the executable name (WIN, standalone)
	return strings.Replace(exePath, "-tray", "", -1) // Remove -tray from the path)
}
