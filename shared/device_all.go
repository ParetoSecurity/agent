package shared

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/caarlos0/log"
	"github.com/elastic/go-sysinfo"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

func CurrentReportingDevice() ReportingDevice {
	device, err := NewLinkingDevice()
	if err != nil {
		log.WithError(err).Fatal("Failed to get device information")
	}

	osVersion := device.OS

	// Format OS version based on platform requirements
	switch runtime.GOOS {
	case "darwin":
		// macOS needs version in format: ^(\d+\.)?(\d+\.)?(\*|\d+)
		osVersion = FormatMacOSVersion(device.OSVersion)
	case "windows":
		productName, err := RunCommand("powershell", "-Command", `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").ProductName`)
		if err != nil {
			log.WithError(err).Warn("Failed to get Windows product name")
		}
		displayVersion, err := RunCommand("powershell", "-Command", `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").DisplayVersion`)
		if err != nil {
			log.WithError(err).Warn("Failed to get Windows version")
		}
		if lo.IsNotEmpty(productName) && lo.IsNotEmpty(displayVersion) {
			osVersion = SanitizeWithSpaces(strings.TrimSpace(productName + " " + displayVersion))
		} else {
			osVersion = SanitizeWithSpaces(fmt.Sprintf("%s %s", device.OS, device.OSVersion))
		}
	default:
		// Linux and others
		osVersion = SanitizeWithSpaces(fmt.Sprintf("%s %s", device.OS, device.OSVersion))
	}

	rd := ReportingDevice{
		MachineUUID: device.UUID,
		MachineName: Sanitize(device.Hostname),
		Auth:        Config.AuthToken,
		OSVersion:   osVersion,
		ModelName: func() string {
			modelName, err := SystemDevice()
			if err != nil || modelName == "" {
				return "Unknown"
			}

			return Sanitize(modelName)
		}(),
		ModelSerial: func() string {
			serial, err := SystemSerial()
			if err != nil || serial == "" {
				return "Unknown"
			}

			// Handle common placeholder values
			placeholders := []string{
				"To Be Filled By O.E.M.",
				"To be filled by O.E.M.",
				"0123456789",
				"Default string",
				"System Serial Number",
				"Not Specified",
				"Not Available",
				"N/A",
				"None",
				".",
				"-",
				"000000000000",
			}

			for _, placeholder := range placeholders {
				if strings.EqualFold(serial, placeholder) || strings.TrimSpace(serial) == placeholder {
					return "Unknown"
				}
			}

			sanitized := Sanitize(serial)
			// If sanitization results in empty string or just spaces, return Unknown
			if sanitized == "" {
				return "Unknown"
			}

			return sanitized
		}(),
	}

	// Apply OpenAPI spec validation and constraints
	ValidateAndPrepareDevice(&rd)

	return rd
}

type LinkingDevice struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	OSVersion string `json:"osVersion"`
	Kernel    string `json:"kernel"`
	UUID      string `json:"uuid"`
	Ticket    string `json:"ticket"`
	Version   string `json:"version"`
}

// NewLinkingDevice creates a new instance of LinkingDevice with system information.
// It retrieves the system UUID and device ticket, and populates the LinkingDevice struct
// with the hostname, OS name, OS version, kernel version, UUID, and ticket.
// Returns a pointer to the LinkingDevice and an error if any occurs during the process.
func NewLinkingDevice() (*LinkingDevice, error) {

	if testing.Testing() {
		return &LinkingDevice{
			Hostname:  "test-hostname",
			OS:        "test-os",
			OSVersion: "test-os-version",
			Kernel:    "test-kernel",
			UUID:      "12345678-1234-1234-1234-123456789012",
			Ticket:    "test-ticket",
		}, nil
	}

	hostInfo, err := sysinfo.Host()
	if err != nil {
		log.Warn("Failed to get process information")
		return nil, err
	}
	envInfo := hostInfo.Info()

	ticket, err := uuid.NewRandom()
	if err != nil {
		log.Warn("Failed to generate ticket")
		return nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn("Failed to get hostname")
		return nil, err
	}

	return &LinkingDevice{
		Hostname:  hostname,
		OS:        envInfo.OS.Name,
		OSVersion: envInfo.OS.Version,
		Kernel:    envInfo.OS.Build,
		UUID:      GetDeviceUUID(),
		Ticket:    ticket.String(),
	}, nil
}
