package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteLogin_Run_NoOpenPorts(t *testing.T) {
	remoteLogin := &RemoteLogin{}

	// Mock checkPort to always return false
	checkPortMock = func(port int, proto string) bool {
		return false
	}

	err := remoteLogin.Run()
	assert.NoError(t, err)
	assert.True(t, remoteLogin.Passed())
	assert.Empty(t, remoteLogin.ports)
}

func TestRemoteLogin_Run_OpenPorts(t *testing.T) {
	remoteLogin := &RemoteLogin{}

	// Mock checkPort to return true for specific ports
	checkPortMock = func(port int, _ string) bool {
		return port == 22 || port == 3389
	}

	err := remoteLogin.Run()
	assert.NoError(t, err)
	assert.False(t, remoteLogin.Passed())
	assert.NotEmpty(t, remoteLogin.ports)
	assert.Contains(t, remoteLogin.ports, 22)
	assert.Contains(t, remoteLogin.ports, 3389)
	assert.NotContains(t, remoteLogin.ports, 3390)
	assert.NotContains(t, remoteLogin.ports, 5900)
	assert.NotEmpty(t, remoteLogin.UUID())
	assert.False(t, remoteLogin.RequiresRoot())
}

func TestRemoteLogin_Name(t *testing.T) {
	remoteLogin := &RemoteLogin{}
	expectedName := "Remote login is disabled"
	if remoteLogin.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, remoteLogin.Name())
	}
}

func TestRemoteLogin_Status(t *testing.T) {
	remoteLogin := &RemoteLogin{}
	expectedStatus := "Remote access services found running on ports:"
	if remoteLogin.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, remoteLogin.Status())
	}
}

func TestRemoteLogin_UUID(t *testing.T) {
	remoteLogin := &RemoteLogin{}
	expectedUUID := "4ced961d-7cfc-4e7b-8f80-195f6379446e"
	if remoteLogin.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, remoteLogin.UUID())
	}
}

func TestRemoteLogin_Passed(t *testing.T) {
	remoteLogin := &RemoteLogin{passed: true}
	expectedPassed := true
	if remoteLogin.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, remoteLogin.Passed())
	}
}

func TestRemoteLogin_FailedMessage(t *testing.T) {
	remoteLogin := &RemoteLogin{}
	expectedFailedMessage := "Remote access services found running"
	if remoteLogin.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, remoteLogin.FailedMessage())
	}
}

func TestRemoteLogin_PassedMessage(t *testing.T) {
	remoteLogin := &RemoteLogin{}
	expectedPassedMessage := "No remote access services found running"
	if remoteLogin.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, remoteLogin.PassedMessage())
	}
}
