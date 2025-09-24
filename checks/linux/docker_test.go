package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestDockerAccess_Run(t *testing.T) {
	tests := []struct {
		name           string
		versionOutput  string
		infoOutput     string
		expectedPassed bool
		expectedStatus string
	}{
		{
			name:           "Docker not installed",
			versionOutput:  "",
			infoOutput:     "",
			expectedPassed: true,
			expectedStatus: "Docker is not installed",
		},
		{
			name:           "Docker installed but no daemon access",
			versionOutput:  "Docker Version 20.10.7\nCannot connect to the Docker daemon",
			infoOutput:     "",
			expectedPassed: true,
			expectedStatus: "No access to Docker daemon with the current user",
		},
		{
			name:           "Docker info command fails",
			versionOutput:  "Docker Version 20.10.7",
			infoOutput:     "",
			expectedPassed: false,
			expectedStatus: "Failed to get Docker info",
		},
		{
			name:           "Docker not running in rootless mode",
			versionOutput:  "Docker Version 20.10.7",
			infoOutput:     "seccomp",
			expectedPassed: false,
			expectedStatus: "Docker is not running in rootless mode",
		},
		{
			name:           "Docker running in rootless mode",
			versionOutput:  "Docker Version 20.10.7",
			infoOutput:     "rootless",
			expectedPassed: true,
			expectedStatus: "Docker is running in rootless mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				"docker version": tt.versionOutput,
				"docker info --format {{.SecurityOptions}}": tt.infoOutput,
			})
			dockerAccess := &DockerAccess{}
			err := dockerAccess.Run()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, dockerAccess.passed)
			assert.Equal(t, tt.expectedStatus, dockerAccess.status)
			assert.NotEmpty(t, dockerAccess.UUID())
			assert.False(t, dockerAccess.RequiresRoot())
		})
	}
}

func TestDockerAccess_IsRunnable(t *testing.T) {
	// IsRunnable should always return true now
	dockerAccess := &DockerAccess{}
	result := dockerAccess.IsRunnable()
	assert.True(t, result, "IsRunnable should always return true")
}

func TestDockerAccess_Name(t *testing.T) {
	dockerAccess := &DockerAccess{}
	expectedName := "Access to Docker is restricted"
	if dockerAccess.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, dockerAccess.Name())
	}
}

func TestDockerAccess_Status(t *testing.T) {
	dockerAccess := &DockerAccess{}
	expectedStatus := ""
	if dockerAccess.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, dockerAccess.Status())
	}
}

func TestDockerAccess_UUID(t *testing.T) {
	dockerAccess := &DockerAccess{}
	expectedUUID := "25443ceb-c1ec-408c-b4f3-2328ea0c84e1"
	if dockerAccess.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, dockerAccess.UUID())
	}
}

func TestDockerAccess_Passed(t *testing.T) {
	dockerAccess := &DockerAccess{passed: true}
	expectedPassed := true
	if dockerAccess.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, dockerAccess.Passed())
	}
}

func TestDockerAccess_FailedMessage(t *testing.T) {
	dockerAccess := &DockerAccess{}
	expectedFailedMessage := "Docker is not running in rootless mode"
	if dockerAccess.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, dockerAccess.FailedMessage())
	}
}

func TestDockerAccess_PassedMessage(t *testing.T) {
	dockerAccess := &DockerAccess{}
	expectedPassedMessage := "Docker is running in rootless mode"
	if dockerAccess.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, dockerAccess.PassedMessage())
	}
}

func TestDockerAccess_DeprecatedDockerPackage(t *testing.T) {
	// Mock shared.RunCommand
	shared.RunCommandMocks = []shared.RunCommandMock{
		// Mock "docker version" to show Docker is installed
		{Command: "docker", Args: []string{"version"}, Out: "Docker Version 20.10.7", Err: nil},
		// Mock "which dpkg-query" to succeed
		{Command: "which", Args: []string{"dpkg-query"}, Out: "/usr/bin/dpkg-query", Err: nil},
		// Mock "dpkg-query -W -f='${Package}' docker.io" to return a deprecated package
		{Command: "dpkg-query", Args: []string{"-W", "-f='${Package}'", "docker.io"}, Out: "docker.io", Err: nil},
	}

	dockerAccess := &DockerAccess{}
	err := dockerAccess.Run()

	// Assertions
	assert.NoError(t, err)
	assert.False(t, dockerAccess.Passed())
	assert.Equal(t, "Deprecated docker.io package installed via apt", dockerAccess.Status())
}
