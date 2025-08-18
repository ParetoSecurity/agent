package checks

import (
	"os/exec"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

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

func TestApplicationUpdates_checkUpdates(t *testing.T) {
	tests := []struct {
		name                   string
		mocks                  []shared.RunCommandMock
		presentPackageManagers []string
		expected               struct {
			passed bool
			detail string
		}
	}{
		{
			name:                   "no package managers available",
			mocks:                  []shared.RunCommandMock{},
			presentPackageManagers: []string{},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "flatpak with updates available",
			mocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--app", "--updates", "--columns=application,version"}, Out: "com.example.App\t1.2.0\n", Err: nil},
				{Command: "flatpak", Args: []string{"list", "--app", "--columns=application,version"}, Out: "com.example.App\t1.1.0\n", Err: nil},
			},
			presentPackageManagers: []string{"flatpak"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Flatpak"},
		},
		{
			name: "flatpak with multiple updates from user example",
			mocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--app", "--updates", "--columns=application,version"}, Out: "de.swsnr.pictureoftheday\t1.7.0\ndev.zed.Zed\tv0.199.6\norg.nickvision.tubeconverter\t2025.7.2\n", Err: nil},
				{Command: "flatpak", Args: []string{"list", "--app", "--columns=application,version"}, Out: "app.zen_browser.zen\t1.14.11b\ncom.github.neithern.g4music\t4.5\ncom.mattjakeman.ExtensionManager\t0.6.3\ncom.quexten.Goldwarden\tv0.3.6\nde.swsnr.pictureoftheday\t1.7.0\ndev.zed.Zed\tv0.199.6\norg.nickvision.tubeconverter\t2025.7.2\npage.tesk.Refine\t0.5.10\n", Err: nil},
			},
			presentPackageManagers: []string{"flatpak"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Flatpak"},
		},
		{
			name: "flatpak no updates",
			mocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--app", "--updates", "--columns=application,version"}, Out: "", Err: nil},
				{Command: "flatpak", Args: []string{"list", "--app", "--columns=application,version"}, Out: "com.example.App\t1.1.0\n", Err: nil},
			},
			presentPackageManagers: []string{"flatpak"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "apt with updates",
			mocks: []shared.RunCommandMock{
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "vim/stable 8.2.0 amd64 [upgradable from: 8.1.0]\n", Err: nil},
			},
			presentPackageManagers: []string{"apt"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: APT"},
		},
		{
			name: "dnf with updates",
			mocks: []shared.RunCommandMock{
				{Command: "dnf", Args: []string{"updateinfo", "list", "--security", "--quiet"}, Out: "FEDORA-2023-security vim-enhanced security\n", Err: nil},
			},
			presentPackageManagers: []string{"dnf"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: DNF"},
		},
		{
			name: "dnf no updates",
			mocks: []shared.RunCommandMock{
				{Command: "dnf", Args: []string{"updateinfo", "list", "--security", "--quiet"}, Out: "", Err: nil},
			},
			presentPackageManagers: []string{"dnf"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "pacman with updates",
			mocks: []shared.RunCommandMock{
				{Command: "pacman", Args: []string{"-Qu"}, Out: "vim 8.2.0-1 -> 8.2.1-1\n", Err: nil},
			},
			presentPackageManagers: []string{"pacman"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Pacman"},
		},
		{
			name: "snap with updates",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "Name    Version  Rev  Publisher\ncode    1.2.3    45   microsoft\n", Err: nil},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Snap"},
		},
		{
			name: "snap up to date",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "All snaps up to date.", Err: nil},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "snap with single line output not containing up to date message",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "Some other single line output", Err: nil},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Snap"},
		},
		{
			name: "snap with empty output",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "", Err: nil},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "snap with snapd not active",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "inactive", Err: nil},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "snap with snapd service check error",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "", Err: exec.ErrNotFound},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "snap with refresh list error",
			mocks: []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-active", "snapd"}, Out: "active", Err: nil},
				{Command: "snap", Args: []string{"refresh", "--list"}, Out: "", Err: exec.ErrNotFound},
			},
			presentPackageManagers: []string{"snap"},
			expected: struct {
				passed bool
				detail string
			}{true, "All packages are up to date"},
		},
		{
			name: "multiple package managers with updates",
			mocks: []shared.RunCommandMock{
				{Command: "flatpak", Args: []string{"remote-ls", "--app", "--updates", "--columns=application,version"}, Out: "com.example.App\t1.2.0\n", Err: nil},
				{Command: "flatpak", Args: []string{"list", "--app", "--columns=application,version"}, Out: "com.example.App\t1.1.0\n", Err: nil},
				{Command: "apt", Args: []string{"list", "--upgradable"}, Out: "vim/stable 8.2.0 amd64 [upgradable from: 8.1.0]\n", Err: nil},
			},
			presentPackageManagers: []string{"flatpak", "apt"},
			expected: struct {
				passed bool
				detail string
			}{false, "Updates available for: Flatpak, APT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			shared.RunCommandMocks = tt.mocks
			lookPathMock = func(cmd string) (string, error) {
				for _, pm := range tt.presentPackageManagers {
					if pm == cmd {
						return cmd, nil
					}
				}
				return "", exec.ErrNotFound
			}

			au := &ApplicationUpdates{}
			passed, detail := au.checkUpdates()

			assert.Equal(t, tt.expected.passed, passed)
			assert.Equal(t, tt.expected.detail, detail)
		})
	}
}

func TestApplicationUpdates_parseFlatpak(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "single app",
			input:    "com.example.App\t1.2.0",
			expected: map[string]string{"com.example.App": "1.2.0"},
		},
		{
			name:  "multiple apps",
			input: "com.example.App\t1.2.0\norg.test.Another\t2.1.0",
			expected: map[string]string{
				"com.example.App":  "1.2.0",
				"org.test.Another": "2.1.0",
			},
		},
		{
			name:     "skip lines without dots",
			input:    "com.example.App\t1.2.0\nInvalid Line\norg.test.Another\t2.1.0",
			expected: map[string]string{"com.example.App": "1.2.0", "org.test.Another": "2.1.0"},
		},
		{
			name:     "skip empty lines",
			input:    "com.example.App\t1.2.0\n\n\norg.test.Another\t2.1.0",
			expected: map[string]string{"com.example.App": "1.2.0", "org.test.Another": "2.1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := &ApplicationUpdates{}
			result := au.parseFlatpak(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
