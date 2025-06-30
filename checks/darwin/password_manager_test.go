package checks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInstalledApplications(t *testing.T) {

	tests := []struct {
		name     string
		appNames []string
		expected bool
	}{
		{
			name:     "Password manager not present",
			appNames: []string{"NonExistentApp.app"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkInstalledApplications(tt.appNames)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPasswordManagerCheck_Name(t *testing.T) {
	pmc := &PasswordManagerCheck{}
	expectedName := "Password Manager Presence"
	if pmc.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, pmc.Name())
	}
}

func TestPasswordManagerCheck_Status(t *testing.T) {
	pmc := &PasswordManagerCheck{}
	expectedStatus := "No password manager found"
	if pmc.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, pmc.Status())
	}
}

func TestPasswordManagerCheck_UUID(t *testing.T) {
	pmc := &PasswordManagerCheck{}
	expectedUUID := "f962c423-fdf5-428a-a57a-827abc9b253e"
	if pmc.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, pmc.UUID())
	}
}

func TestPasswordManagerCheck_Passed(t *testing.T) {
	pmc := &PasswordManagerCheck{passed: true}
	expectedPassed := true
	if pmc.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, pmc.Passed())
	}
}

func TestPasswordManagerCheck_FailedMessage(t *testing.T) {
	pmc := &PasswordManagerCheck{}
	expectedFailedMessage := "No password manager found"
	if pmc.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, pmc.FailedMessage())
	}
}

func TestPasswordManagerCheck_PassedMessage(t *testing.T) {
	pmc := &PasswordManagerCheck{}
	expectedPassedMessage := "Password manager is present"
	if pmc.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, pmc.PassedMessage())
	}
}

func TestCheckForBrowserExtensions(t *testing.T) {
	// save original HOME so we can restore it
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name         string
		setUpDirs    map[string][]string // map of relative dir (from HOME) to list of entry names to create
		expectedBool bool
	}{
		{
			name: "Found extension in Chrome",
			setUpDirs: map[string][]string{
				filepath.Join("Library", "Application Support", "Google", "Chrome", "Default", "Extensions"): {"aeblfdkhhhdcdjpifhhbdiojplfjncoa"},
			},
			expectedBool: true,
		},
		{
			name: "Found extension in Firefox",
			setUpDirs: map[string][]string{
				filepath.Join("Library", "Application Support", "Firefox", "Profiles"): {"eiaeiblijfjekdanodkjadfinkhbfgcd"},
			},
			expectedBool: true,
		},
		{
			name: "Directories exist but no matching extension",
			setUpDirs: map[string][]string{
				filepath.Join("Library", "Application Support", "Microsoft Edge", "Default", "Extensions"):                 {"randomfolder"},
				filepath.Join("Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Extensions"): {"anotherfolder"},
			},
			expectedBool: false,
		},
		{
			name:         "No directories exist",
			setUpDirs:    map[string][]string{},
			expectedBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary HOME directory
			tempHome := t.TempDir()
			os.Setenv("HOME", tempHome)

			// Set up directories and entries as specified by the test case.
			for relDir, entries := range tt.setUpDirs {
				dirPath := filepath.Join(tempHome, relDir)
				err := os.MkdirAll(dirPath, 0755)
				assert.NoError(t, err)
				// Create subdirectories for each entry
				for _, entryName := range entries {
					entryPath := filepath.Join(dirPath, entryName)
					err := os.Mkdir(entryPath, 0755)
					assert.NoError(t, err)
				}
			}
			// Run the function and check the result
			result := checkForBrowserExtensions()
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}
