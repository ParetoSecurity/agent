package checks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordManagerCheck_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockFiles      map[string]bool
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "1Password present",
			mockFiles: map[string]bool{
				filepath.Join("C:\\Users\\TestUser", "AppData", "Local", "1Password", "app", "8", "1Password.exe"): true,
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Bitwarden present",
			mockFiles: map[string]bool{
				filepath.Join("C:\\Users\\TestUser", "AppData", "Local", "Programs", "Bitwarden", "Bitwarden.exe"): true,
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePass present",
			mockFiles: map[string]bool{
				filepath.Join("C:\\Program Files (x86)", "KeePass Password Safe 2", "KeePass.exe"): true,
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePassXC present",
			mockFiles: map[string]bool{
				filepath.Join("C:\\Program Files", "KeePassXC", "KeePassXC.exe"): true,
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name:           "No password manager present",
			mockFiles:      map[string]bool{},
			expectedPassed: false,
			expectedStatus: "No password manager found",
		},
	}

	for _, tt := range tests {
		os.Setenv("HOME", "C:\\Users\\TestUser")
		os.Setenv("USERPROFILE", "C:\\Users\\TestUser")
		os.Setenv("PROGRAMFILES", "C:\\Program Files")
		os.Setenv("PROGRAMFILES(X86)", "C:\\Program Files (x86)")
		t.Run(tt.name, func(t *testing.T) {
			osStatMock = tt.mockFiles
			pmc := &PasswordManagerCheck{}
			err := pmc.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, pmc.Passed())
			assert.Equal(t, tt.expectedStatus, pmc.Status())
		})
		os.Unsetenv("HOME")
		os.Unsetenv("USERPROFILE")
		os.Unsetenv("PROGRAMFILES")
		os.Unsetenv("PROGRAMFILES(X86)")
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
