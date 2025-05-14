package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestFirewall_Name(t *testing.T) {
	f := &Firewall{}
	expectedName := "Firewall is configured"
	if f.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, f.Name())
	}
}

func TestFirewall_Status(t *testing.T) {
	f := &Firewall{}
	expectedStatus := "Firewall is off"
	if f.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, f.Status())
	}
}

func TestFirewall_UUID(t *testing.T) {
	f := &Firewall{}
	expectedUUID := "2e46c89a-5461-4865-a92e-3b799c12034a"
	if f.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, f.UUID())
	}
}

func TestFirewall_Passed(t *testing.T) {
	f := &Firewall{passed: true}
	expectedPassed := true
	if f.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, f.Passed())
	}
}

func TestFirewall_FailedMessage(t *testing.T) {
	f := &Firewall{}
	expectedFailedMessage := "Firewall is off"
	if f.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, f.FailedMessage())
	}
}

func TestFirewall_PassedMessage(t *testing.T) {
	f := &Firewall{}
	expectedPassedMessage := "Firewall is on"
	if f.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, f.PassedMessage())
	}
}

func TestFirewall_IsRunnable(t *testing.T) {
	firewall := &Firewall{}
	assert.True(t, firewall.IsRunnable(), "Firewall should always be runnable")
}

func TestFirewall_RequiresRoot(t *testing.T) {
	f := &Firewall{}
	assert.True(t, f.RequiresRoot(), "Firewall check should require root access")
}

func TestCheckIptables(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name: "Iptables has rules",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
1    ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0           tcp dpt:22
2    DROP       all  --  10.0.0.0/8           0.0.0.0/0           
`,
			mockError:      nil,
			expectedResult: true,
		},
		{
			name: "Iptables has no rules",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
`,
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "Iptables command error",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedResult: false,
		},
		{
			name: "Malformed rule line",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
invalid line
`,
			mockError:      nil,
			expectedResult: false,
		},
		{
			name: "Non-numeric rule number",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
abc  ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0           
`,
			mockError:      nil,
			expectedResult: false,
		},
		{
			name: "NixOS style custom chain",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination
1    nixos-fw   all  --  anywhere             anywhere
`,
			mockError:      nil,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"iptables -L INPUT --line-numbers": tt.mockOutput,
			})
			f := &Firewall{}
			result := f.checkIptables()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCheckIptables_CommandError(t *testing.T) {
	shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
		"iptables -L INPUT --line-numbers": "",
	})
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "iptables",
			Args:    []string{"-L", "INPUT", "--line-numbers"},
			Out:     "",
			Err:     assert.AnError,
		},
	}

	f := &Firewall{}
	result := f.checkIptables()
	assert.False(t, result, "Expected checkIptables to return false when RunCommand fails")
}

func TestCheckIptables_PolicyParsing(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		expectedResult bool
	}{
		{
			name: "Policy DROP",
			mockOutput: `Chain INPUT (policy DROP)
num  target     prot opt source               destination         
`,
			expectedResult: true,
		},
		{
			name: "Policy REJECT",
			mockOutput: `Chain INPUT (policy REJECT)
num  target     prot opt source               destination         
`,
			expectedResult: true,
		},
		{
			name: "Policy ACCEPT",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
`,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"iptables -L INPUT --line-numbers": tt.mockOutput,
			})
			f := &Firewall{}
			result := f.checkIptables()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestFirewall_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedPassed bool
	}{
		{
			name: "Iptables active",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
1    ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0           tcp dpt:22
`,
			mockError:      nil,
			expectedPassed: true,
		},
		{
			name: "Iptables inactive",
			mockOutput: `Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination         
`,
			mockError:      nil,
			expectedPassed: false,
		},
		{
			name:           "Iptables command error",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"iptables -L INPUT --line-numbers": tt.mockOutput,
			})
			f := &Firewall{}
			result := f.checkIptables()
			assert.Equal(t, tt.expectedPassed, result)
		})
	}
}

func TestCheckNFTables(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name:           "NFTables configured with chain INPUT",
			mockOutput:     "table inet filter {\n\tchain INPUT {\n\t\ttype filter hook input priority 0;\n\t\tpolicy drop;\n\t}\n}",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "NFTables configured without chain INPUT",
			mockOutput:     "table inet filter {\n\tchain OUTPUT {\n\t\ttype filter hook output priority 0;\n\t\tpolicy accept;\n\t}\n}",
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "nixos NFTables configured with chain input",
			mockOutput:     "table inet filter {\n\tchain input {\n\t\ttype filter hook input priority 0;\n\t\tpolicy drop;\n\t}\n}",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "NFTables command error",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedResult: false,
		},
		{
			name:           "Empty NFTables output",
			mockOutput:     "",
			mockError:      nil,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.RunCommand
			shared.RunCommandMocks = []shared.RunCommandMock{
				{
					Command: "nft",
					Args:    []string{"list", "ruleset"},
					Out:     tt.mockOutput,
					Err:     tt.mockError,
				},
			}
			f := &Firewall{}
			result := f.checkNFTables()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
