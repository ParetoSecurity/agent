package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestAutologin_Run(t *testing.T) {
	tests := []struct {
		name             string
		mockFiles        map[string]string
		mockFilepathGlob map[string][]string
		mockCommand      string
		mockCommandOut   string
		expectedPassed   bool
		expectedStatus   string
	}{
		{
			name: "SDDM autologin enabled in conf.d",
			mockFiles: map[string]string{
				"/etc/sddm.conf.d/test.conf": "Autologin=true",
				"/etc/sddm.conf":             "Autologin=true",
			},
			expectedPassed: false,
			expectedStatus: "Autologin=true in SDDM is enabled",
		},
		{
			name: "SDDM autologin enabled in main config",
			mockFiles: map[string]string{
				"/etc/sddm.conf": "Autologin=true",
			},
			expectedPassed: false,
			expectedStatus: "Autologin=true in SDDM is enabled",
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
				"/etc/sddm.conf.d/10-test.conf": "Autologin=true",
				"/etc/sddm.conf.d/20-test.conf": "Autologin=true",
			},
			mockFilepathGlob: map[string][]string{
				"/etc/sddm.conf.d/*.conf": {"/etc/sddm.conf.d/10-test.conf", "/etc/sddm.conf.d/20-test.conf"},
			},
			expectedPassed: false,
			expectedStatus: "Autologin=true in SDDM is enabled",
		},
		{
			name:           "No autologin enabled",
			expectedPassed: true,
			expectedStatus: "Automatic login is off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock shared.ReadFile
			shared.ReadFileMocks = tt.mockFiles

			// Mock filepath.Glob
			filepathGlobMock = func(pattern string) ([]string, error) {
				if tt.mockFilepathGlob != nil {
					if files, ok := tt.mockFilepathGlob[pattern]; ok {
						return files, nil
					}
				}
				return nil, nil
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
