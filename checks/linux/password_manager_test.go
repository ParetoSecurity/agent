package checks

import (
	"os"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestPasswordManagerCheck_Run_Linux(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   map[string]string
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "1Password found in PATH",
			mockCommands: map[string]string{
				"which 1password": "/usr/bin/1password",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Bitwarden found in PATH",
			mockCommands: map[string]string{
				"which bitwarden": "/usr/bin/bitwarden",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Dashlane found in PATH",
			mockCommands: map[string]string{
				"which dashlane": "/usr/bin/dashlane",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePassX found in PATH",
			mockCommands: map[string]string{
				"which keepassx": "/usr/bin/keepassx",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePassXC found in PATH",
			mockCommands: map[string]string{
				"which keepassxc": "/usr/bin/keepassxc",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "gnome-keyring found in PATH",
			mockCommands: map[string]string{
				"which gnome-keyring": "/usr/bin/gnome-keyring",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "gnome-keyring present via apt",
			mockCommands: map[string]string{
				"which 1password":     "not found",
				"which bitwarden":     "not found",
				"which dashlane":      "not found",
				"which keepassx":      "not found",
				"which keepassxc":     "not found",
				"which gnome-keyring": "not found",
				"which dpkg":          "/usr/bin/dpkg",
				"sh -c dpkg -l":       "ii  gnome-keyring  1.0  all  Password manager",
				"which snap":          "not found",
				"which yum":           "not found",
				"which flatpak":       "not found",
				"which pacman":        "not found",
				"which nix-store":     "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "1Password present via apt",
			mockCommands: map[string]string{
				"which 1password": "not found",
				"which bitwarden": "not found",
				"which dashlane":  "not found",
				"which keepassx":  "not found",
				"which keepassxc": "not found",
				"which dpkg":      "/usr/bin/dpkg",
				"sh -c dpkg -l":   "ii  1password  1.0  all  Password manager",
				"which snap":      "not found",
				"which yum":       "not found",
				"which flatpak":   "not found",
				"which pacman":    "not found",
				"which nix-store": "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Bitwarden present via snap",
			mockCommands: map[string]string{
				"which bitwarden": "not found",
				"which 1password": "not found",
				"which dashlane":  "not found",
				"which keepassx":  "not found",
				"which keepassxc": "not found",
				"which dpkg":      "not found",
				"which snap":      "/usr/bin/snap",
				"sh -c snap list": "bitwarden  1.0  stable  password manager",
				"which yum":       "not found",
				"which flatpak":   "not found",
				"which pacman":    "not found",
				"which nix-store": "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Dashlane present via yum",
			mockCommands: map[string]string{
				"which 1password":          "not found",
				"which bitwarden":          "not found",
				"which dashlane":           "not found",
				"which keepassx":           "not found",
				"which keepassxc":          "not found",
				"which dpkg":               "not found",
				"which snap":               "not found",
				"which yum":                "/usr/bin/yum",
				"sh -c yum list installed": "dashlane  1.0  installed  password manager",
				"which flatpak":            "not found",
				"which pacman":             "not found",
				"which nix-store":          "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePassX present via flatpak",
			mockCommands: map[string]string{
				"which 1password":    "not found",
				"which bitwarden":    "not found",
				"which dashlane":     "not found",
				"which keepassx":     "not found",
				"which keepassxc":    "not found",
				"which dpkg":         "not found",
				"which snap":         "not found",
				"which yum":          "not found",
				"which flatpak":      "/usr/bin/flatpak",
				"sh -c flatpak list": "keepassx  1.0  stable  password manager",
				"which pacman":       "not found",
				"which nix-store":    "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "KeePassXC present via pacman",
			mockCommands: map[string]string{
				"which 1password": "not found",
				"which bitwarden": "not found",
				"which dashlane":  "not found",
				"which keepassx":  "not found",
				"which keepassxc": "not found",
				"which dpkg":      "not found",
				"which snap":      "not found",
				"which yum":       "not found",
				"which flatpak":   "not found",
				"which pacman":    "/usr/bin/pacman",
				"sh -c pacman -Q": "keepassxc 1.0-1",
				"which nix-store": "not found",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "Bitwarden present via nix",
			mockCommands: map[string]string{
				"which 1password": "not found",
				"which bitwarden": "not found",
				"which dashlane":  "not found",
				"which keepassx":  "not found",
				"which keepassxc": "not found",
				"which dpkg":      "not found",
				"which snap":      "not found",
				"which yum":       "not found",
				"which flatpak":   "not found",
				"which pacman":    "not found",
				"which nix-store": "/run/current-system/sw/bin/nix-store",
				"sh -c if [ -e ~/.nix-profile ]; then nix-store -q --requisites /run/current-system ~/.nix-profile; else nix-store -q --requisites /run/current-system; fi": "  /nix/store/m2hv2xvfrq51pd8ya3blqk1xxbazpp5p-bitwarden-desktop-2025.1.1",
			},
			expectedPassed: true,
			expectedStatus: "Password manager is present",
		},
		{
			name: "No password manager present",
			mockCommands: map[string]string{
				"which 1password":          "not found",
				"which bitwarden":          "not found",
				"which dashlane":           "not found",
				"which keepassx":           "not found",
				"which keepassxc":          "not found",
				"which dpkg":               "/usr/bin/dpkg",
				"sh -c dpkg -l":            "no password managers here",
				"which snap":               "/usr/bin/snap",
				"sh -c snap list":          "no password managers here",
				"which yum":                "/usr/bin/yum",
				"sh -c yum list installed": "no password managers here",
				"which flatpak":            "/usr/bin/flatpak",
				"sh -c flatpak list":       "no password managers here",
				"which pacman":             "/usr/bin/pacman",
				"sh -c pacman -Q":          "no password managers here",
				"which nix-store":          "/usr/bin/nix-store",
				"sh -c if [ -e ~/.nix-profile ]; then nix-store -q --requisites /run/current-system ~/.nix-profile; else nix-store -q --requisites /run/current-system; fi": "no password managers here",
			},
			expectedPassed: false,
			expectedStatus: "No password manager found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.RunCommand
			shared.SetCache("pkg_apt", "", 0)
			shared.SetCache("pkg_snap", "", 0)
			shared.SetCache("pkg_yum", "", 0)
			shared.SetCache("pkg_flatpak", "", 0)
			shared.SetCache("pkg_pacman", "", 0)
			shared.SetCache("pkg_nix", "", 0)

			shared.RunCommandMocks = convertCommandMapToMocks(tt.mockCommands)

			pmc := &PasswordManagerCheck{}
			status := pmc.isManagerInstalled()
			assert.Equal(t, tt.expectedPassed, status)
		})
	}
}

func TestPasswordManagerCheck_Run_BrowserExtensions(t *testing.T) {
	tests := []struct {
		name           string
		mockFileSystem []string
		mockCommands   map[string]string
		expectedPassed bool
	}{
		{
			name: "1Password extension present in Chrome",
			mockFileSystem: []string{
				"/home/user/.config/google-chrome/Default/Extensions/aeblfdkhhhdcdjpifhhbdiojplfjncoa",
			},
			mockCommands: map[string]string{
				"which 1password":          "not found",
				"which bitwarden":          "not found",
				"which dashlane":           "not found",
				"which keepassx":           "not found",
				"which keepassxc":          "not found",
				"which dpkg":               "/usr/bin/dpkg",
				"sh -c dpkg -l":            "no password managers here",
				"which snap":               "/usr/bin/snap",
				"sh -c snap list":          "no password managers here",
				"which yum":                "/usr/bin/yum",
				"sh -c yum list installed": "no password managers here",
				"which flatpak":            "/usr/bin/flatpak",
				"sh -c flatpak list":       "no password managers here",
				"which pacman":             "/usr/bin/pacman",
				"sh -c pacman -Q":          "no password managers here",
				"which nix-store":          "/usr/bin/nix-store",
				"sh -c if [ -e ~/.nix-profile ]; then nix-store -q --requisites /run/current-system ~/.nix-profile; else nix-store -q --requisites /run/current-system; fi": "no password managers here",
			},
			expectedPassed: true,
		},
		{
			name:           "No password manager extensions present",
			mockFileSystem: []string{},
			mockCommands: map[string]string{
				"which 1password":          "not found",
				"which bitwarden":          "not found",
				"which dashlane":           "not found",
				"which keepassx":           "not found",
				"which keepassxc":          "not found",
				"which dpkg":               "/usr/bin/dpkg",
				"sh -c dpkg -l":            "no password managers here",
				"which snap":               "/usr/bin/snap",
				"sh -c snap list":          "no password managers here",
				"which yum":                "/usr/bin/yum",
				"sh -c yum list installed": "no password managers here",
				"which flatpak":            "/usr/bin/flatpak",
				"sh -c flatpak list":       "no password managers here",
				"which pacman":             "/usr/bin/pacman",
				"sh -c pacman -Q":          "no password managers here",
				"which nix-store":          "/usr/bin/nix-store",
				"sh -c if [ -e ~/.nix-profile ]; then nix-store -q --requisites /run/current-system ~/.nix-profile; else nix-store -q --requisites /run/current-system; fi": "no password managers here",
			},
			expectedPassed: false,
		},
	}

	for _, tt := range tests {
		os.Setenv("HOME", "/home/user")
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.RunCommand
			shared.SetCache("pkg_apt", "", 0)
			shared.SetCache("pkg_snap", "", 0)
			shared.SetCache("pkg_yum", "", 0)
			shared.SetCache("pkg_flatpak", "", 0)
			shared.SetCache("pkg_pacman", "", 0)
			shared.SetCache("pkg_nix", "", 0)
			shared.RunCommandMocks = convertCommandMapToMocks(tt.mockCommands)

			// Mock os.ReadDir
			osReadDirMock = func(_ string) ([]os.DirEntry, error) {
				return lo.Map(tt.mockFileSystem, func(name string, _ int) os.DirEntry {
					return &mockDirEntry{name: name}
				}), nil
			}

			pmc := &PasswordManagerCheck{}
			err := pmc.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, pmc.Passed())
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
