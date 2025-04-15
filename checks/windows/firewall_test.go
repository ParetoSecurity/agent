package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckFirewallProfile(t *testing.T) {

	tests := []struct {
		name           string
		profile        string
		mockOutput     string
		mockError      error
		expectedResult bool
		expectedStatus string
	}{
		{
			name:           "Firewall enabled for Public profile",
			profile:        "Public",
			mockOutput:     "True",
			mockError:      nil,
			expectedResult: true,
			expectedStatus: "",
		},
		{
			name:           "Firewall disabled for Private profile",
			profile:        "Private",
			mockOutput:     "False",
			mockError:      nil,
			expectedResult: false,
			expectedStatus: "Windows Firewall is not enabled for Private profile",
		},
		{
			name:           "Error querying firewall profile",
			profile:        "Domain",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedResult: false,
			expectedStatus: "Failed to query Windows Firewall for Domain profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.RunCommand
			shared.RunCommandMocks = []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name '" + tt.profile + "' | Select-Object -ExpandProperty Enabled"},
					Out:     tt.mockOutput,
					Err:     tt.mockError,
				},
			}

			firewall := &WindowsFirewall{}
			result := firewall.checkFirewallProfile(tt.profile)

			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedStatus, firewall.status)
		})
	}
}
