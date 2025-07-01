package checks

import (
	"errors"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestWindowsDefender_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Defender all protections enabled",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":true,"AntispywareEnabled":true}`,
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus software is active",
		},
		{
			name: "Defender real-time protection disabled",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     `{"RealTimeProtectionEnabled":false,"IoavProtectionEnabled":true,"AntispywareEnabled":true}`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Defender has disabled real-time protection",
		},
		{
			name: "Defender tamper protection disabled",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":false,"AntispywareEnabled":true}`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Defender has disabled tamper protection",
		},
		{
			name: "Defender antispyware disabled",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":true,"AntispywareEnabled":false}`,
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Defender is disabled",
		},
		{
			name: "Fallback to wmic SecurityCenter2 with active antivirus",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     "",
					Err:     errors.New("powershell failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "displayName=Norton Security\nproductState=266240\n\n",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus software is active",
		},
		{
			name: "Fallback to wmic SecurityCenter (older systems)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     "",
					Err:     errors.New("powershell failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "",
					Err:     errors.New("SecurityCenter2 failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "displayName=McAfee VirusScan\nonAccessScanningEnabled=TRUE\n\n",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus software is active",
		},
		{
			name: "No antivirus detected via wmic",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     "",
					Err:     errors.New("powershell failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "No antivirus software detected",
		},
		{
			name: "All methods fail",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     "",
					Err:     errors.New("powershell failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "",
					Err:     errors.New("SecurityCenter2 failed"),
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "",
					Err:     errors.New("SecurityCenter failed"),
				},
			},
			expectedPassed: false,
			expectedStatus: "Failed to query antivirus status",
		},
		{
			name: "PowerShell invalid JSON, fallback to wmic success",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
					Out:     `invalid-json`,
					Err:     nil,
				},
				{
					Command: "wmic",
					Args:    []string{"/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value"},
					Out:     "displayName=Kaspersky Internet Security\nproductState=397312\n\n",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Antivirus software is active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands

			check := &WindowsDefender{}
			err := check.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, check.Passed())
			assert.Equal(t, tt.expectedStatus, check.Status())
		})
	}
}
