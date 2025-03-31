package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestSaveConfig_Success(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set configPath to a temporary file.
	configPath = filepath.Join(tempDir, "pareto.toml")

	// Prepare a test configuration.

	Config = ParetoConfig{
		TeamID:    "team1",
		AuthToken: "token1",
	}

	// Call SaveConfig.
	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Read the written file.
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Unmarshal the file content.
	var loadedConfig ParetoConfig
	if err := toml.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}

	// Validate the saved configuration.
	if loadedConfig.TeamID != Config.TeamID {
		t.Errorf("expected TeamID %q, got %q", Config.TeamID, loadedConfig.TeamID)
	}
	if loadedConfig.AuthToken != Config.AuthToken {
		t.Errorf("expected AuthToken %q, got %q", Config.AuthToken, loadedConfig.AuthToken)
	}

}

func TestSaveConfig_Failure(t *testing.T) {
	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "config-test-failure")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set configPath to a directory to simulate a failure (os.Create should fail).
	configPath = tempDir

	// Call SaveConfig expecting an error.
	if err := SaveConfig(); err == nil {
		t.Errorf("expected error when configPath is a directory, got nil")
	}
}

func TestIsCheckDisabled(t *testing.T) {
	tests := []struct {
		name           string
		disabledChecks []string
		checkUUID      string
		expected       bool
	}{
		{
			name:           "Check is disabled",
			disabledChecks: []string{"uuid1", "uuid2"},
			checkUUID:      "uuid1",
			expected:       true,
		},
		{
			name:           "Check is not disabled",
			disabledChecks: []string{"uuid1", "uuid2"},
			checkUUID:      "uuid3",
			expected:       false,
		},
		{
			name:           "No checks are disabled",
			disabledChecks: []string{},
			checkUUID:      "uuid1",
			expected:       false,
		},
		{
			name:           "Empty check UUID",
			disabledChecks: []string{"uuid1", "uuid2", ""},
			checkUUID:      "",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config.DisableChecks = tt.disabledChecks
			actual := IsCheckDisabled(tt.checkUUID)
			if actual != tt.expected {
				t.Errorf("IsCheckDisabled() = %v, want %v", actual, tt.expected)
			}
		})
	}
}
