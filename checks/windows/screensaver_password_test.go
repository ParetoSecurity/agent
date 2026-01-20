package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestScreensaverTimeout_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Screensaver active with timeout under 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue"},
					Out:     "600",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Screensaver active with timeout exactly 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue"},
					Out:     "1200",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Screensaver active with timeout over 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue"},
					Out:     "1800",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Screensaver timeout is set to more than 20 minutes",
		},
		{
			name: "Screensaver not active",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Screensaver is not enabled",
		},
		{
			name: "Failed to query screensaver status",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expectedPassed: false,
			expectedStatus: "Failed to query screensaver status",
		},
		{
			name: "Failed to query timeout",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue"},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expectedPassed: false,
			expectedStatus: "Failed to query screensaver timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands
			defer func() { shared.RunCommandMocks = nil }()

			check := &ScreensaverTimeout{}
			err := check.Run()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, check.Passed())
			assert.Equal(t, tt.expectedStatus, check.Status())
		})
	}
}

func TestScreensaverTimeout_Metadata(t *testing.T) {
	check := &ScreensaverTimeout{}

	assert.Equal(t, "Screensaver or screen lock shows in under 20min", check.Name())
	assert.Equal(t, "Screensaver or screen lock shows in under 20min", check.PassedMessage())
	assert.Equal(t, "Screensaver or screen lock shows in more than 20min", check.FailedMessage())
	assert.Equal(t, "13e4dbf1-f87f-4bd9-8a82-f62044f002f4", check.UUID())
	assert.False(t, check.RequiresRoot())
	assert.True(t, check.IsRunnable())
}

func TestScreensaverPassword_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Screensaver password enabled (desktop)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "Sleep password enabled (AC and DC) on laptop",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'DCSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "Sleep password enabled (AC only) on desktop",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "Both screensaver and sleep password enabled on laptop",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'DCSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "No password protection enabled on desktop",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password protection is not configured for screensaver or sleep",
		},
		{
			name: "Only AC sleep password enabled on laptop (fails - missing DC)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'DCSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password protection is not configured for screensaver or sleep",
		},
		{
			name: "All queries fail on desktop",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "",
					Err:     assert.AnError,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "",
					Err:     assert.AnError,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password protection is not configured for screensaver or sleep",
		},
		{
			name: "Battery detection fails (assumes desktop)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-WmiObject -Class Win32_Battery | Measure-Object).Count"},
					Out:     "",
					Err:     assert.AnError,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = tt.mockCommands
			defer func() { shared.RunCommandMocks = nil }()

			check := &ScreensaverPassword{}
			err := check.Run()

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPassed, check.Passed())
			assert.Equal(t, tt.expectedStatus, check.Status())
		})
	}
}

func TestScreensaverPassword_Metadata(t *testing.T) {
	check := &ScreensaverPassword{}

	assert.Equal(t, "Password after sleep or screensaver is on", check.Name())
	assert.Equal(t, "Password after sleep or screensaver is on", check.PassedMessage())
	assert.Equal(t, "Password after sleep or screensaver is off", check.FailedMessage())
	assert.Equal(t, "37dee029-605b-4aab-96b9-5438e5aa44d8", check.UUID())
	assert.False(t, check.RequiresRoot())
	assert.True(t, check.IsRunnable())
}
