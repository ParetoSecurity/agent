//go:build windows
// +build windows

package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWindowsVersionDetection(t *testing.T) {
	tests := []struct {
		name            string
		productName     string
		displayVersion  string
		buildNumber     string
		expectedVersion string
	}{
		{
			name:            "Windows 11 with high build number",
			productName:     "Windows 10 Pro",
			displayVersion:  "24H2",
			buildNumber:     "26100",
			expectedVersion: "Windows 11 Pro 24H2",
		},
		{
			name:            "Windows 11 with build number including dots",
			productName:     "Windows 10 Pro",
			displayVersion:  "23H2",
			buildNumber:     "26100.6584",
			expectedVersion: "Windows 11 Pro 23H2",
		},
		{
			name:            "Windows 10 with low build number",
			productName:     "Windows 10 Pro",
			displayVersion:  "22H2",
			buildNumber:     "19045",
			expectedVersion: "Windows 10 Pro 22H2",
		},
		{
			name:            "Windows 11 at minimum build threshold",
			productName:     "Windows 10 Enterprise",
			displayVersion:  "21H2",
			buildNumber:     "22000",
			expectedVersion: "Windows 11 Enterprise 21H2",
		},
		{
			name:            "Windows 10 just below threshold",
			productName:     "Windows 10 Pro",
			displayVersion:  "21H2",
			buildNumber:     "21999",
			expectedVersion: "Windows 10 Pro 21H2",
		},
		{
			name:            "Already reports as Windows 11",
			productName:     "Windows 11 Pro",
			displayVersion:  "24H2",
			buildNumber:     "26100",
			expectedVersion: "Windows 11 Pro 24H2",
		},
		{
			name:            "Windows Server should not be changed",
			productName:     "Windows Server 2022",
			displayVersion:  "21H2",
			buildNumber:     "23000",
			expectedVersion: "Windows Server 2022 21H2",
		},
		{
			name:            "Empty build number keeps original",
			productName:     "Windows 10 Pro",
			displayVersion:  "22H2",
			buildNumber:     "",
			expectedVersion: "Windows 10 Pro 22H2",
		},
		{
			name:            "Invalid build number keeps original",
			productName:     "Windows 10 Pro",
			displayVersion:  "22H2",
			buildNumber:     "invalid",
			expectedVersion: "Windows 10 Pro 22H2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock RunCommand to return our test values
			originalRunCommand := RunCommandMock
			defer func() { RunCommandMock = originalRunCommand }()

			RunCommandMock = func(name string, args ...string) (string, error) {
				if len(args) > 0 {
					switch args[1] {
					case `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").ProductName`:
						return tt.productName, nil
					case `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").DisplayVersion`:
						return tt.displayVersion, nil
					case `(Get-ItemProperty -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").CurrentBuildNumber`:
						return tt.buildNumber, nil
					}
				}
				return "", nil
			}

			device := CurrentReportingDevice()
			assert.Equal(t, tt.expectedVersion, device.OSVersion)
		})
	}
}
