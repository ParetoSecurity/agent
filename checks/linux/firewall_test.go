package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckUFW(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name:           "UFW is active",
			mockOutput:     "Status: active",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "UFW is inactive",
			mockOutput:     "Status: inactive",
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "UFW command error",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"ufw status": tt.mockOutput,
			})
			f := &Firewall{}
			result := f.checkUFW()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestCheckFirewalld(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name:           "Firewalld is active",
			mockOutput:     "active",
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "Firewalld is inactive",
			mockOutput:     "inactive",
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "Firewalld command error",
			mockOutput:     "",
			mockError:      assert.AnError,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"systemctl is-active firewalld": tt.mockOutput,
			})

			f := &Firewall{}
			result := f.checkFirewalld()
			assert.Equal(t, tt.expectedResult, result)
			assert.NotEmpty(t, f.UUID())
			assert.True(t, f.RequiresRoot())
		})
	}
}

func TestFirewall_Run(t *testing.T) {
	tests := []struct {
		name                string
		mockUFWOutput       string
		mockFirewalldOutput string
		expectedPassed      bool
		expectedStatus      string
	}{
		{
			name:                "UFW is active",
			mockUFWOutput:       "Status: active",
			mockFirewalldOutput: "",
			expectedPassed:      true,
			expectedStatus:      "Firewall is on",
		},
		{
			name:                "Firewalld is active",
			mockUFWOutput:       "Status: inactive",
			mockFirewalldOutput: "active",
			expectedPassed:      true,
			expectedStatus:      "Firewall is on",
		},
		{
			name:                "Both UFW and Firewalld are inactive",
			mockUFWOutput:       "Status: inactive",
			mockFirewalldOutput: "inactive",
			expectedPassed:      false,
			expectedStatus:      "Firewall is off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"ufw status":                    tt.mockUFWOutput,
				"systemctl is-active firewalld": tt.mockFirewalldOutput,
			})

			f := &Firewall{}
			err := f.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, f.Passed())
			assert.Equal(t, tt.expectedStatus, f.Status())
		})
	}
}

func TestFirewall_Name(t *testing.T) {
	f := &Firewall{}
	expectedName := "Firewall is on"
	if f.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, f.Name())
	}
}

func TestFirewall_Status(t *testing.T) {
	f := &Firewall{}
	expectedStatus := ""
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
