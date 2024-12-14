package shared

import (
	"fmt"
	"net"
	"os"

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

func SystemDevice() (string, error) {
	content, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_name")
	if err != nil {
		return "", err
	}

	deviceName := strings.TrimSpace(string(content))
	if deviceName == "" {
		return "", fmt.Errorf("unable to retrieve device name")
	}

	return deviceName, nil
}

func SystemSerial() (string, error) {
	content, err := os.ReadFile("/sys/devices/virtual/dmi/id/product_serial")
	if err != nil {
		return "", err
	}
	serialNumber := strings.TrimSpace(string(content))
	if serialNumber == "" {
		return "", fmt.Errorf("unable to retrieve serial number")
	}

	return serialNumber, nil
}

func IsRoot() bool {
	return os.Geteuid() == 0
}

func SelfExe() string {
	exePath, err := os.Executable()
	if err != nil {
		return "paretosecurity"
	}
	return exePath
}
