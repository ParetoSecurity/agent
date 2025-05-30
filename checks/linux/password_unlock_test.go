package checks

import (
	"os/exec"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckGnome(t *testing.T) {
	tests := []struct {
		name       string
		commandOut string
		commandErr error
		expected   bool
	}{
		{
			name:       "Lock enabled",
			commandOut: "true\n",
			commandErr: nil,
			expected:   true,
		},
		{
			name:       "Lock disabled",
			commandOut: "false\n",
			commandErr: nil,
			expected:   false,
		},
		{
			name:       "Command error",
			commandOut: "",
			commandErr: assert.AnError,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = []shared.RunCommandMock{
				{
					Command: "gsettings",
					Args:    []string{"get", "org.gnome.desktop.screensaver", "lock-enabled"},
					Out:     tt.commandOut,
					Err:     tt.commandErr,
				},
			}

			f := &PasswordToUnlock{}
			result := f.checkGnome()
			assert.Equal(t, tt.expected, result)
			assert.NotEmpty(t, f.UUID())
			assert.False(t, f.RequiresRoot())
		})
	}
}

func TestPasswordToUnlock_Name(t *testing.T) {
	f := &PasswordToUnlock{}
	expectedName := "Password is required to unlock the screen"
	if f.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, f.Name())
	}
}

func TestPasswordToUnlock_Status(t *testing.T) {
	f := &PasswordToUnlock{}
	expectedStatus := "Password after sleep or screensaver is off"
	if f.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, f.Status())
	}
}

func TestPasswordToUnlock_UUID(t *testing.T) {
	f := &PasswordToUnlock{}
	expectedUUID := "37dee029-605b-4aab-96b9-5438e5aa44d8"
	if f.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, f.UUID())
	}
}

func TestPasswordToUnlock_Passed(t *testing.T) {
	f := &PasswordToUnlock{passed: true}
	expectedPassed := true
	if f.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, f.Passed())
	}
}

func TestPasswordToUnlock_FailedMessage(t *testing.T) {
	f := &PasswordToUnlock{}
	expectedFailedMessage := "Password after sleep or screensaver is off"
	if f.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, f.FailedMessage())
	}
}

func TestPasswordToUnlock_PassedMessage(t *testing.T) {
	f := &PasswordToUnlock{}
	expectedPassedMessage := "Password after sleep or screensaver is on"
	if f.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, f.PassedMessage())
	}
}

func TestPasswordToUnlock_Run(t *testing.T) {
	tests := []struct {
		name                  string
		gsettingsAvailable    bool
		kreadconfig5Available bool
		gnomeResult           bool
		kdeResult             bool
		expectedPassed        bool
	}{
		{
			name:                  "Neither environment available",
			gsettingsAvailable:    false,
			kreadconfig5Available: false,
			gnomeResult:           false,
			kdeResult:             false,
			expectedPassed:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPathMock = func(file string) (string, error) {
				return "", exec.ErrNotFound
			}
			f := &PasswordToUnlock{}
			err := f.Run()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, f.Passed())
		})
	}
}
func TestPasswordToUnlock_checkKDE5(t *testing.T) {
	// Store original functions to restore later

	tests := []struct {
		name        string
		homeDirErr  error
		homeDir     string
		readFileErr error
		fileContent string
		expected    bool
	}{
		{
			name:        "LockOnResume=false present",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: nil,
			fileContent: "[Daemon]\nLockOnResume=false\nAutolock=true\n",
			expected:    false,
		},
		{
			name:        "LockOnResume=false with other content",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: nil,
			fileContent: "SomeOtherSetting=value\nLockOnResume=false\nMoreSettings=value\n",
			expected:    false,
		},
		{
			name:        "LockOnResume=true present",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: nil,
			fileContent: "[Daemon]\nLockOnResume=true\nAutolock=true\n",
			expected:    true,
		},
		{
			name:        "No LockOnResume setting (defaults to true)",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: nil,
			fileContent: "[Daemon]\nAutolock=true\nSomeOtherSetting=value\n",
			expected:    true,
		},
		{
			name:        "Empty config file",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: nil,
			fileContent: "",
			expected:    true,
		},
		{
			name:        "UserHomeDir error",
			homeDirErr:  assert.AnError,
			homeDir:     "",
			readFileErr: nil,
			fileContent: "",
			expected:    true,
		},
		{
			name:        "ReadFile error",
			homeDirErr:  nil,
			homeDir:     "/home/user",
			readFileErr: assert.AnError,
			fileContent: "",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.UserHomeDirMock = func() (string, error) {
				return tt.homeDir, nil
			}
			// Mock shared.ReadFile
			shared.ReadFileMock = func(filename string) ([]byte, error) {
				if tt.readFileErr != nil {
					return nil, tt.readFileErr
				}
				return []byte(tt.fileContent), nil
			}

			f := &PasswordToUnlock{}
			result := f.checkKDE5()
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestPasswordToUnlock_IsRunnable(t *testing.T) {
	f := &PasswordToUnlock{}
	assert.True(t, f.IsRunnable())
}
