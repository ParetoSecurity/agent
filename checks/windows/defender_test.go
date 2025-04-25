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
		mockCommand    shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "All protections enabled",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":true,"AntispywareEnabled":true}`,
				Err:     nil,
			},
			expectedPassed: true,
			expectedStatus: "Microsoft Defender is on",
		},
		{
			name: "Real-time protection disabled",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     `{"RealTimeProtectionEnabled":false,"IoavProtectionEnabled":true,"AntispywareEnabled":true}`,
				Err:     nil,
			},
			expectedPassed: false,
			expectedStatus: "Defender has disabled real-time protection",
		},
		{
			name: "Tamper protection disabled",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":false,"AntispywareEnabled":true}`,
				Err:     nil,
			},
			expectedPassed: false,
			expectedStatus: "Defender has disabled tamper protection",
		},
		{
			name: "Antispyware disabled",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     `{"RealTimeProtectionEnabled":true,"IoavProtectionEnabled":true,"AntispywareEnabled":false}`,
				Err:     nil,
			},
			expectedPassed: false,
			expectedStatus: "Defender is disabled",
		},
		{
			name: "Command execution error",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     "",
				Err:     errors.New("command failed"),
			},
			expectedPassed: false,
			expectedStatus: "Failed to query Defender status",
		},
		{
			name: "Invalid JSON output",
			mockCommand: shared.RunCommandMock{
				Command: "powershell",
				Args:    []string{"-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json"},
				Out:     `invalid-json`,
				Err:     nil,
			},
			expectedPassed: false,
			expectedStatus: "Failed to parse Defender status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = []shared.RunCommandMock{tt.mockCommand}

			check := &WindowsDefender{}
			err := check.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, check.Passed())
			assert.Equal(t, tt.expectedStatus, check.Status())
		})
	}
}
