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

	// Set ConfigPath to a temporary file.
	ConfigPath = filepath.Join(tempDir, "pareto.toml")

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
	data, err := os.ReadFile(ConfigPath)
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

	// Set ConfigPath to a directory to simulate a failure (os.Create should fail).
	ConfigPath = tempDir

	// Call SaveConfig expecting an error.
	if err := SaveConfig(); err == nil {
		t.Errorf("expected error when ConfigPath is a directory, got nil")
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

func TestResetConfig(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set ConfigPath to a temporary file.
	ConfigPath = filepath.Join(tempDir, "pareto.toml")

	// Set some initial values in the config.
	Config = ParetoConfig{
		TeamID:        "initialTeamID",
		AuthToken:     "initialAuthToken",
		DisableChecks: []string{"uuid1"},
	}
	SaveConfig()

	// Call ResetConfig.
	ResetConfig()

	// Check that the config has been reset.
	if Config.TeamID != "" {
		t.Errorf("expected TeamID to be empty, got %q", Config.TeamID)
	}
	if Config.AuthToken != "" {
		t.Errorf("expected AuthToken to be empty, got %q", Config.AuthToken)
	}
	if len(Config.DisableChecks) != 0 {
		t.Errorf("expected DisableChecks to be empty, got %v", Config.DisableChecks)
	}

	// Read the config file to ensure it was saved.
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Unmarshal the file content.
	var loadedConfig ParetoConfig
	if err := toml.Unmarshal(data, &loadedConfig); err != nil {
		t.Fatalf("failed to decode config file: %v", err)
	}

	// Validate the saved configuration.
	if loadedConfig.TeamID != "" {
		t.Errorf("expected TeamID %q, got %q", "", loadedConfig.TeamID)
	}
	if loadedConfig.AuthToken != "" {
		t.Errorf("expected AuthToken %q, got %q", "", loadedConfig.AuthToken)
	}
	if len(loadedConfig.DisableChecks) != 0 {
		t.Errorf("expected DisableChecks to be empty, got %v", loadedConfig.DisableChecks)
	}
}

func TestEnableCheck(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set ConfigPath to a temporary file.
	ConfigPath = filepath.Join(tempDir, "pareto.toml")

	// Test cases
	tests := []struct {
		name           string
		initialConfig  ParetoConfig
		checkUUID      string
		expectedConfig ParetoConfig
		expectedError  bool
	}{
		{
			name: "Check is disabled, should be enabled",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2"},
			},
			checkUUID: "uuid1",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid2"},
			},
			expectedError: false,
		},
		{
			name: "Check is not disabled, should do nothing",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"uuid2", "uuid3"},
			},
			checkUUID: "uuid1",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid2", "uuid3"},
			},
			expectedError: false,
		},
		{
			name: "No checks are disabled, should do nothing",
			initialConfig: ParetoConfig{
				DisableChecks: []string{},
			},
			checkUUID: "uuid1",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{},
			},
			expectedError: false,
		},
		{
			name: "Empty check UUID, should do nothing",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"", "uuid2"},
			},
			checkUUID: "",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid2"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial config
			Config = tt.initialConfig
			SaveConfig()

			// Call EnableCheck
			err := EnableCheck(tt.checkUUID)

			// Check for error
			if (err != nil) != tt.expectedError {
				t.Errorf("EnableCheck() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Read the config file to ensure it was saved.
			data, err := os.ReadFile(ConfigPath)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}

			// Unmarshal the file content.
			var loadedConfig ParetoConfig
			if err := toml.Unmarshal(data, &loadedConfig); err != nil {
				t.Fatalf("failed to decode config file: %v", err)
			}

			// Validate the saved configuration.
			if len(loadedConfig.DisableChecks) != len(tt.expectedConfig.DisableChecks) {
				t.Errorf("expected DisableChecks to have length %d, got %d", len(tt.expectedConfig.DisableChecks), len(loadedConfig.DisableChecks))
			}

			for i, check := range loadedConfig.DisableChecks {
				if check != tt.expectedConfig.DisableChecks[i] {
					t.Errorf("expected DisableChecks[%d] to be %q, got %q", i, tt.expectedConfig.DisableChecks[i], check)
				}
			}
		})
	}
}

func TestDisableCheck(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set ConfigPath to a temporary file.
	ConfigPath = filepath.Join(tempDir, "pareto.toml")

	// Test cases
	tests := []struct {
		name           string
		initialConfig  ParetoConfig
		checkUUID      string
		expectedConfig ParetoConfig
		expectedError  bool
	}{
		{
			name: "Check is not disabled, should be disabled",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2"},
			},
			checkUUID: "uuid3",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2", "uuid3"},
			},
			expectedError: false,
		},
		{
			name: "Check is already disabled, should do nothing",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2"},
			},
			checkUUID: "uuid2",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2"},
			},
			expectedError: false,
		},
		{
			name: "No checks are disabled, should be disabled",
			initialConfig: ParetoConfig{
				DisableChecks: []string{},
			},
			checkUUID: "uuid1",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid1"},
			},
			expectedError: false,
		},
		{
			name: "Empty check UUID, should be disabled",
			initialConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2"},
			},
			checkUUID: "",
			expectedConfig: ParetoConfig{
				DisableChecks: []string{"uuid1", "uuid2", ""},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set initial config
			Config = tt.initialConfig
			SaveConfig()

			// Call DisableCheck
			err := DisableCheck(tt.checkUUID)

			// Check for error
			if (err != nil) != tt.expectedError {
				t.Errorf("DisableCheck() error = %v, expectedError %v", err, tt.expectedError)
			}

			// Read the config file to ensure it was saved.
			data, err := os.ReadFile(ConfigPath)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}

			// Unmarshal the file content.
			var loadedConfig ParetoConfig
			if err := toml.Unmarshal(data, &loadedConfig); err != nil {
				t.Fatalf("failed to decode config file: %v", err)
			}

			// Validate the saved configuration.
			if len(loadedConfig.DisableChecks) != len(tt.expectedConfig.DisableChecks) {
				t.Errorf("expected DisableChecks to have length %d, got %d", len(tt.expectedConfig.DisableChecks), len(loadedConfig.DisableChecks))
			}

			for i, check := range loadedConfig.DisableChecks {
				if check != tt.expectedConfig.DisableChecks[i] {
					t.Errorf("expected DisableChecks[%d] to be %q, got %q", i, tt.expectedConfig.DisableChecks[i], check)
				}
			}
		})
	}
}

func TestGetDeviceUUID(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set ConfigPath to a temporary file.
	ConfigPath = filepath.Join(tempDir, "pareto.toml")

	tests := []struct {
		name           string
		initialUUID    string
		mockSystemUUID func() (string, error)
		expectedUUID   string
		shouldContain  bool // for generated UUIDs
	}{
		{
			name:         "Returns existing UUID from config",
			initialUUID:  "existing-uuid-123",
			expectedUUID: "existing-uuid-123",
		},
		{
			name:          "Falls back to random UUID when system fails",
			initialUUID:   "",
			shouldContain: true, // UUID will be randomly generated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config
			Config = ParetoConfig{SystemUUID: tt.initialUUID}

			// Call GetDeviceUUID
			result := GetDeviceUUID()

			// Validate result
			if tt.shouldContain {
				// For generated UUIDs, just check it's not empty and saved to config
				if result == "" {
					t.Error("expected non-empty UUID, got empty string")
				}
				if Config.SystemUUID != result {
					t.Errorf("expected Config.SystemUUID to be %q, got %q", result, Config.SystemUUID)
				}
			} else {
				if result != tt.expectedUUID {
					t.Errorf("expected UUID %q, got %q", tt.expectedUUID, result)
				}
				if Config.SystemUUID != tt.expectedUUID {
					t.Errorf("expected Config.SystemUUID to be %q, got %q", tt.expectedUUID, Config.SystemUUID)
				}
			}

			// Verify config was saved to disk (except when UUID already existed)
			if tt.initialUUID == "" {
				data, err := os.ReadFile(ConfigPath)
				if err != nil {
					t.Fatalf("failed to read config file: %v", err)
				}

				var loadedConfig ParetoConfig
				if err := toml.Unmarshal(data, &loadedConfig); err != nil {
					t.Fatalf("failed to decode config file: %v", err)
				}

				if loadedConfig.SystemUUID != result {
					t.Errorf("expected saved SystemUUID to be %q, got %q", result, loadedConfig.SystemUUID)
				}
			}
		})
	}
}
