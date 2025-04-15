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
	osVersion = Sanitize(fmt.Sprintf("%s %s", osVersion, device.OSVersion))

	if runtime.GOOS == "windows" {
		productName, err := RunCommand("powershell", "-Command", `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").ProductName`)
		if err != nil {
			log.WithError(err).Warn("Failed to get Windows product name")
		}
		displayVersion, err := RunCommand("powershell", "-Command", `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").DisplayVersion`)
		if err != nil {
			log.WithError(err).Warn("Failed to get Windows version")
		}
		if lo.IsNotEmpty(productName) && lo.IsNotEmpty(displayVersion) {
			osVersion = Sanitize(strings.TrimSpace(productName + " " + displayVersion))
		}
	}

	return ReportingDevice{
		MachineUUID: device.UUID,
		MachineName: Sanitize(device.Hostname),
		Auth:        Config.AuthToken,
		OSVersion:   osVersion,
		ModelName: func() string {
			modelName, err := SystemDevice()
			if err != nil {
				return "Unknown"
			}

			return Sanitize(modelName)
		}(),
		ModelSerial: func() string {
			serial, err := SystemSerial()
			if err != nil {
				return "Unknown"
			}

			return serial
		}(),
	}
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
			UUID:      "test-uuid",
			Ticket:    "test-ticket",
		}, nil
	}

	hostInfo, err := sysinfo.Host()
	if err != nil {
		log.Warn("Failed to get process information")
		return nil, err
	}
	envInfo := hostInfo.Info()

	systemUUID, err := SystemUUID()
	if err != nil {
		log.Warn("Failed to get system UUID")
		return nil, err
	}
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
		UUID:      systemUUID,
		Ticket:    ticket.String(),
	}, nil
}
