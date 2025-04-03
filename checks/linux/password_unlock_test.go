package checks

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckKDE(t *testing.T) {
	tests := []struct {
		name       string
		commandOut string
		commandErr error
		expected   bool
	}{
		{
			name:       "Autolock enabled",
			commandOut: "true\n",
			commandErr: nil,
			expected:   true,
		},
		{
			name:       "Autolock disabled",
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
					Command: "kreadconfig5",
					Args:    []string{"--file", "kscreenlockerrc", "--group", "Daemon", "--key", "Autolock"},
					Out:     tt.commandOut,
					Err:     tt.commandErr,
				},
			}

			f := &PasswordToUnlock{}
			result := f.checkKDE()
			assert.Equal(t, tt.expected, result)
		})
	}
}

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

func TestPasswordToUnlock_Run(t *testing.T) {
	tests := []struct {
		name           string
		executables    []string
		mockCommands   map[string]string
		expectedPassed bool
		expectedStatus string
	}{
		{
			name:        "GNOME lock disabled",
			executables: []string{"gsettings"},
			mockCommands: map[string]string{
				"gsettings get org.gnome.desktop.screensaver lock-enabled": "false\n",
			},
			expectedPassed: false,
			expectedStatus: "Password after sleep or screensaver is off",
		},
		{
			name:        "KDE autolock disabled",
			executables: []string{"kreadconfig5"},
			mockCommands: map[string]string{
				"kreadconfig5 --file kscreenlockerrc --group Daemon --key Autolock": "false\n",
			},
			expectedPassed: false,
			expectedStatus: "Password after sleep or screensaver is off",
		},
		{
			name:        "GNOME passing and KDE failing",
			executables: []string{"gsettings", "kreadconfig5"},
			mockCommands: map[string]string{
				"gsettings get org.gnome.desktop.screensaver lock-enabled":          "true\n",
				"kreadconfig5 --file kscreenlockerrc --group Daemon --key Autolock": "false\n",
			},
			expectedPassed: false,
			expectedStatus: "Password after sleep or screensaver is off",
		},
		{
			name:           "Neither GNOME nor KDE found",
			executables:    []string{},
			mockCommands:   map[string]string{},
			expectedPassed: false,
			expectedStatus: "Password after sleep or screensaver is off",
		},
		{
			name:        "GNOME and KDE both passing",
			executables: []string{"gsettings", "kreadconfig5"},
			mockCommands: map[string]string{
				"gsettings get org.gnome.desktop.screensaver lock-enabled":          "true\n",
				"kreadconfig5 --file kscreenlockerrc --group Daemon --key Autolock": "true\n",
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = convertCommandMapToMocks(tt.mockCommands)

			lookPathMock = func(file string) (string, error) {
				if slices.Contains(tt.executables, file) {
					return "/usr/bin/" + file, nil
				}
				return "", exec.ErrNotFound
			}

			f := &PasswordToUnlock{}
			err := f.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, f.Passed())
			assert.Equal(t, tt.expectedStatus, f.Status())
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

func TestCheckSway(t *testing.T) {
	tests := []struct {
		name          string
		homeDir       string
		fileContents  map[string]string
		expected      bool
		expectedDebug string
	}{
		{
			name:    "Sway idle lock configuration found",
			homeDir: "/home/testuser",
			fileContents: map[string]string{
				"/etc/sway/config":                           "",
				"/etc/sway/config.d/file1":                   "exec swayidle -w timeout 300 'swaylock'",
				"/home/testuser/.config/sway/config.d/file2": "",
			},
			expected:      true,
			expectedDebug: "Sway idle lock configuration found",
		},
		{
			name:    "Sway idle lock configuration found with newlines",
			homeDir: "/home/testuser",
			fileContents: map[string]string{
				"/etc/sway/config":                           "",
				"/etc/sway/config.d/file1":                   "exec swayidle \n -w timeout 300 'swaylock'",
				"/home/testuser/.config/sway/config.d/file2": "",
			},
			expected:      true,
			expectedDebug: "Sway idle lock configuration found",
		},
		{
			name:    "No swayidle configuration found",
			homeDir: "/home/testuser",
			fileContents: map[string]string{
				"/etc/sway/config":                           "",
				"/etc/sway/config.d/file1":                   "",
				"/home/testuser/.config/sway/config.d/file2": "",
			},
			expected:      false,
			expectedDebug: "Sway idle lock configuration not found",
		},
		{
			name:    "Commented swayidle configuration",
			homeDir: "/home/testuser",
			fileContents: map[string]string{
				"/etc/sway/config":                           "# exec swayidle -w timeout 300 'swaylock'",
				"/etc/sway/config.d/file1":                   "",
				"/home/testuser/.config/sway/config.d/file2": "",
			},
			expected:      false,
			expectedDebug: "Sway idle lock configuration not found",
		},
		{
			name:    "Error reading files",
			homeDir: "/home/testuser",
			fileContents: map[string]string{
				"/etc/sway/config": "",
			},
			expected:      false,
			expectedDebug: "Failed to read file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			filepathGlobMock = func(pattern string) ([]string, error) {
				var files []string
				for path := range tt.fileContents {
					if strings.HasPrefix(path, pattern) {
						files = append(files, path)
					}
				}
				return files, nil
			}

			osReadFileMock = func(filename string) ([]byte, error) {
				if content, exists := tt.fileContents[filename]; exists {
					return []byte(content), nil
				}
				return nil, os.ErrNotExist
			}

			result := checkSway()
			assert.Equal(t, tt.expected, result)
		})
	}
}
