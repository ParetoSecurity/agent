package checks

import (
	"os"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestAutologin_Run(t *testing.T) {
	tests := []struct {
		name             string
		mockFiles        map[string]string
		mockFilepathGlob map[string][]string
		mockOsStat       map[string]bool
		mockCommand      string
		mockCommandOut   string
		expectedPassed   bool
		expectedStatus   string
	}{
		{
			name: "SDDM autologin enabled in conf.d",
			mockFiles: map[string]string{
				"/etc/sddm.conf.d/test.conf": "[Autologin]\nUser=alice",
			},
			mockFilepathGlob: map[string][]string{
				"/etc/sddm.conf.d/*.conf": {"/etc/sddm.conf.d/test.conf"},
			},
			expectedPassed: false,
			expectedStatus: "SDDM autologin user is configured",
		},
		{
			name: "SDDM autologin enabled in main config",
			mockFiles: map[string]string{
				"/etc/sddm.conf": "[Autologin]\nUser=bob",
			},
			expectedPassed: false,
			expectedStatus: "SDDM autologin user is configured",
		},
		{
			name: "GDM autologin enabled in custom.conf",
			mockFiles: map[string]string{
				"/etc/gdm3/custom.conf": "AutomaticLoginEnable=true",
			},
			expectedPassed: false,
			expectedStatus: "AutomaticLoginEnable=true in GDM is enabled",
		},
		{
			name: "GDM autologin enabled in custom.conf (alternative path)",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "AutomaticLoginEnable=true",
			},
			expectedPassed: false,
			expectedStatus: "AutomaticLoginEnable=true in GDM is enabled",
		},
		{
			name:           "GDM autologin enabled in dconf",
			mockCommand:    "dconf read /org/gnome/login-screen/enable-automatic-login",
			mockCommandOut: "true",
			expectedPassed: false,
			expectedStatus: "Automatic login is enabled in GNOME",
		},
		{
			name: "Multiple SDDM configs with autologin enabled",
			mockFiles: map[string]string{
				"/etc/sddm.conf.d/10-test.conf": "[Autologin]\nUser=charlie",
				"/etc/sddm.conf.d/20-test.conf": "[Autologin]\nSession=plasma",
			},
			mockFilepathGlob: map[string][]string{
				"/etc/sddm.conf.d/*.conf": {"/etc/sddm.conf.d/10-test.conf", "/etc/sddm.conf.d/20-test.conf"},
			},
			expectedPassed: false,
			expectedStatus: "SDDM autologin user is configured",
		},
		{
			name:           "No autologin enabled",
			expectedPassed: true,
			expectedStatus: "Automatic login is off",
		},
		{
			name: "GDM timed login enabled (NixOS style)",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "TimedLoginEnable=true",
			},
			expectedPassed: false,
			expectedStatus: "TimedLoginEnable=true in GDM is enabled",
		},
		{
			name: "GDM timed login user configured",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "[daemon]\nTimedLogin=foo",
			},
			expectedPassed: false,
			expectedStatus: "TimedLogin user is configured in GDM",
		},
		{
			name: "GDM timed login delay with non-zero value",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "[daemon]\nTimedLoginDelay=30",
			},
			expectedPassed: false,
			expectedStatus: "TimedLoginDelay is configured in GDM",
		},
		{
			name: "GDM timed login delay with zero value (should also fail)",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "[daemon]\nTimedLoginDelay=0",
			},
			expectedPassed: false,
			expectedStatus: "TimedLoginDelay is configured in GDM",
		},
		{
			name: "GDM complete timed login configuration",
			mockFiles: map[string]string{
				"/etc/gdm/custom.conf": "[daemon]\nTimedLoginEnable=true\nTimedLogin=alice\nTimedLoginDelay=5",
			},
			expectedPassed: false,
			expectedStatus: "TimedLoginEnable=true in GDM is enabled", // First match wins
		},
		{
			name:           "NixOS getty autologin marker file exists",
			mockOsStat:     map[string]bool{"/run/agetty.autologged": true},
			expectedPassed: false,
			expectedStatus: "Getty autologin detected (NixOS /run/agetty.autologged exists)",
		},
		{
			name: "Getty autologin in systemd override",
			mockFiles: map[string]string{
				"/etc/systemd/system/getty@.service.d/overrides.conf": "ExecStart=-/sbin/agetty --autologin root",
			},
			mockFilepathGlob: map[string][]string{
				"/etc/systemd/system/getty@*.service.d/*.conf": {"/etc/systemd/system/getty@.service.d/overrides.conf"},
			},
			expectedPassed: false,
			expectedStatus: "Getty autologin detected in systemd service override",
		},
		{
			name: "LightDM autologin enabled",
			mockFiles: map[string]string{
				"/etc/lightdm/lightdm.conf": "[Seat:*]\nautologin-user=alice",
			},
			expectedPassed: false,
			expectedStatus: "LightDM autologin user is configured",
		},
		{
			name: "LightDM autologin commented out",
			mockFiles: map[string]string{
				"/etc/lightdm/lightdm.conf": "[Seat:*]\n#autologin-user=alice",
			},
			expectedPassed: true,
			expectedStatus: "Automatic login is off",
		},
		{
			name: "SDDM autologin user configured",
			mockFiles: map[string]string{
				"/etc/sddm.conf": "[Autologin]\nUser=testuser\nSession=plasma",
			},
			expectedPassed: false,
			expectedStatus: "SDDM autologin user is configured",
		},
		{
			name: "SDDM autologin session only",
			mockFiles: map[string]string{
				"/etc/sddm.conf": "[Autologin]\nSession=plasma",
			},
			expectedPassed: false,
			expectedStatus: "SDDM autologin session is configured",
		},
		{
			name: "LightDM guest autologin",
			mockFiles: map[string]string{
				"/etc/lightdm/lightdm.conf": "[Seat:*]\nautologin-guest=true",
			},
			expectedPassed: false,
			expectedStatus: "LightDM guest autologin is enabled",
		},
		{
			name: "LightDM autologin session configured",
			mockFiles: map[string]string{
				"/etc/lightdm/lightdm.conf": "[Seat:*]\nautologin-session=xfce",
			},
			expectedPassed: false,
			expectedStatus: "LightDM autologin session is configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.ReadFile
			shared.ReadFileMock = func(name string) ([]byte, error) {
				if content, ok := tt.mockFiles[name]; ok {
					return []byte(content), nil
				}
				return nil, nil // Return nil if file not found
			}

			// Mock filepath.Glob
			filepathGlobMock = func(pattern string) ([]string, error) {
				if tt.mockFilepathGlob != nil {
					if files, ok := tt.mockFilepathGlob[pattern]; ok {
						return files, nil
					}
				}
				return nil, nil
			}

			// Mock os.Stat
			osStatMock = func(file string) (os.FileInfo, error) {
				if tt.mockOsStat != nil {
					if exists, ok := tt.mockOsStat[file]; ok && exists {
						return nil, nil // File exists
					}
				}
				return nil, os.ErrNotExist // File doesn't exist
			}

			// Mock shared.RunCommand
			shared.RunCommandMocks = convertCommandMapToMocks(map[string]string{
				tt.mockCommand: tt.mockCommandOut,
			})

			a := &Autologin{}
			err := a.Run()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, a.Passed())
			assert.Equal(t, tt.expectedStatus, a.Status())
			assert.NotEmpty(t, a.UUID())
			assert.False(t, a.RequiresRoot())
		})
	}
}

func TestAutologin_Name(t *testing.T) {
	a := &Autologin{}
	expectedName := "Automatic login is disabled"
	if a.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, a.Name())
	}
}

func TestAutologin_Status(t *testing.T) {
	a := &Autologin{}
	expectedStatus := ""
	if a.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, a.Status())
	}
}

func TestAutologin_UUID(t *testing.T) {
	a := &Autologin{}
	expectedUUID := "f962c423-fdf5-428a-a57a-816abc9b253e"
	if a.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, a.UUID())
	}
}

func TestAutologin_Passed(t *testing.T) {
	a := &Autologin{passed: true}
	expectedPassed := true
	if a.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, a.Passed())
	}
}

func TestAutologin_FailedMessage(t *testing.T) {
	a := &Autologin{}
	expectedFailedMessage := "Automatic login is on"
	if a.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, a.FailedMessage())
	}
}

func TestAutologin_PassedMessage(t *testing.T) {
	a := &Autologin{}
	expectedPassedMessage := "Automatic login is off"
	if a.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, a.PassedMessage())
	}
}

func TestAutologin_IsRunnable(t *testing.T) {
	a := &Autologin{}
	assert.True(t, a.IsRunnable(), "Autologin check should be runnable")
}
