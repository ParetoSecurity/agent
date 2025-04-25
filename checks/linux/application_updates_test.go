package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckUpdates(t *testing.T) {

	tests := []struct {
		name           string
		setupMocks     []shared.RunCommandMock
		expectedPassed bool
		expectedDetail string
	}{
		{
			name: "All up to date",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
		{
			name: "Updates available",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "some updates", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "upgradable, upgradable", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "some updates", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "some updates", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "some updates", Err: nil},
			},
			expectedPassed: false,
			expectedDetail: "Updates available for: Flatpak, APT, Pacman, Snap",
		},
		{
			name: "Snapd not running",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "inactive\n", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
		{
			name: "Snap reports all up to date",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "All snaps up to date.\n", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.setupMocks
			lookPathMock = func(file string) (string, error) {
				return file, nil
			}
			su := &ApplicationUpdates{}
			passed, detail := su.checkUpdates()
			assert.Equal(t, tt.expectedPassed, passed)
			assert.Equal(t, tt.expectedDetail, detail)
			assert.NotEmpty(t, su.UUID())
			assert.False(t, su.RequiresRoot())
		})
	}
}

func TestApplicationUpdates_Run(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     []shared.RunCommandMock
		expectedPassed bool
		expectedDetail string
	}{
		{
			name: "All up to date",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
		{
			name: "Updates available",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "some updates", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "upgradable, upgradable", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "some updates", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "some updates", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "some updates", Err: nil},
			},
			expectedPassed: false,
			expectedDetail: "Updates available for: Flatpak, APT, Pacman, Snap",
		},
		{
			name: "Snapd not running",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "inactive\n", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
		{
			name: "Snap reports all up to date",
			setupMocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--updates"}, Out: "", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "", Err: nil},
				{Command: "dnf", Args: []string{"check-update", "--quiet"}, Out: "", Err: nil},
				{Command: "pacman", Args: []string{"-Qu"}, Out: "", Err: nil},
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active\n", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "All snaps up to date.\n", Err: nil},
			},
			expectedPassed: true,
			expectedDetail: "All packages are up to date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.setupMocks
			lookPathMock = func(file string) (string, error) {
				return file, nil
			}
			su := &ApplicationUpdates{}
			err := su.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, su.Passed())
			assert.Equal(t, tt.expectedDetail, su.Status())
		})
	}
}

func TestApplicationUpdates_Name(t *testing.T) {
	su := &ApplicationUpdates{}
	expectedName := "Apps are up to date"
	if su.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, su.Name())
	}
}

func TestApplicationUpdates_Status(t *testing.T) {
	su := &ApplicationUpdates{}
	expectedStatus := ""
	if su.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, su.Status())
	}
}

func TestApplicationUpdates_UUID(t *testing.T) {
	su := &ApplicationUpdates{}
	expectedUUID := "7436553a-ae52-479b-937b-2ae14d15a520"
	if su.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, su.UUID())
	}
}

func TestApplicationUpdates_Passed(t *testing.T) {
	su := &ApplicationUpdates{passed: true}
	expectedPassed := true
	if su.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, su.Passed())
	}
}

func TestApplicationUpdates_FailedMessage(t *testing.T) {
	su := &ApplicationUpdates{}
	expectedFailedMessage := "Some apps are out of date"
	if su.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, su.FailedMessage())
	}
}

func TestApplicationUpdates_PassedMessage(t *testing.T) {
	su := &ApplicationUpdates{}
	expectedPassedMessage := "All apps are up to date"
	if su.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, su.PassedMessage())
	}
}

func TestApplicationUpdates_IsRunnable(t *testing.T) {
	su := &ApplicationUpdates{}
	assert.True(t, su.IsRunnable(), "ApplicationUpdates should always be runnable")
}
