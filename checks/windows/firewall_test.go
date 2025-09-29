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

func TestWindowsFirewall_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Windows Firewall enabled for both Public and Private",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "True",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "True",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Firewall is active",
		},
		{
			name: "Windows Firewall disabled, ESET firewall active",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'ESET Service' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "Status\n------\nRunning",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Test-Path 'HKLM:\\SOFTWARE\\ESET'"},
					Out:     "True",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "ESET firewall is active",
		},
		{
			name: "Windows Firewall disabled, ESET service not running",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'ESET Service' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "Status\n------\nStopped",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Windows Firewall is not enabled for Public profile",
		},
		{
			name: "Windows Firewall disabled, ESET not installed",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'ESET Service' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expectedPassed: false,
			expectedStatus: "Windows Firewall is not enabled for Public profile",
		},
		{
			name: "Windows Firewall enabled for Public only (still fails)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "True",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Windows Firewall is not enabled for Private profile",
		},
		{
			name: "ESET service running but registry key missing",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Public' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-NetFirewallProfile -Name 'Private' | Select-Object -ExpandProperty Enabled"},
					Out:     "False",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-Service -Name 'ESET Service' -ErrorAction SilentlyContinue | Select-Object Status"},
					Out:     "Status\n------\nRunning",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Test-Path 'HKLM:\\SOFTWARE\\ESET'"},
					Out:     "False",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Windows Firewall is not enabled for Public profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands

			firewall := &WindowsFirewall{}
			err := firewall.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, firewall.Passed())
			assert.Equal(t, tt.expectedStatus, firewall.Status())
		})
	}
}
