package checks

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

// Sample powercfg output for testing
const powercfgVideoIdleOutput = `
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)
  GUID Alias: SCHEME_BALANCED
  Subgroup GUID: 7516b95f-f776-4464-8c53-06167f40cc99  (Display)
    GUID Alias: SUB_VIDEO
    Power Setting GUID: 3c0bc021-c8a8-4e07-a973-6b14cbcb2b7e  (Turn off display after)
      GUID Alias: VIDEOIDLE
      Minimum Possible Setting: 0x00000000
      Maximum Possible Setting: 0xffffffff
      Possible Settings increment: 0x00000001
      Possible Settings units: Seconds
    Current AC Power Setting Index: 0x00000708
    Current DC Power Setting Index: 0x00000384
`

// Slovenian localized output
const powercfgVideoIdleSlovenian = `
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Uravnoteženo)
  GUID Alias: SCHEME_BALANCED
  Subgroup GUID: 7516b95f-f776-4464-8c53-06167f40cc99  (Zaslon)
    GUID Alias: SUB_VIDEO
    Power Setting GUID: 3c0bc021-c8a8-4e07-a973-6b14cbcb2b7e  (Izklop zaslona po)
      GUID Alias: VIDEOIDLE
      Minimum Possible Setting: 0x00000000
      Maximum Possible Setting: 0xffffffff
      Possible Settings increment: 0x00000001
      Possible Settings units: sekunde
    Current AC Power Setting Index: 0x00000000
    Current DC Power Setting Index: 0x00000e10
`

const powercfgConsoleLockEnabled = `
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)
  GUID Alias: SCHEME_BALANCED
  Subgroup GUID: fea3413e-7e05-4911-9a71-700331f1c294  (No subgroup)
    GUID Alias: SUB_NONE
    Power Setting GUID: 0e796bdb-100d-47d6-a2d5-f7d2daa51f51  (Require a password on wakeup)
      GUID Alias: CONSOLELOCK
      Minimum Possible Setting: 0x00000000
      Maximum Possible Setting: 0x00000001
      Possible Settings increment: 0x00000001
      Possible Settings units:
      Current AC Power Setting Index: 0x00000001
      Current DC Power Setting Index: 0x00000001
`

const powercfgConsoleLockDisabled = `
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)
  GUID Alias: SCHEME_BALANCED
  Subgroup GUID: fea3413e-7e05-4911-9a71-700331f1c294  (No subgroup)
    GUID Alias: SUB_NONE
    Power Setting GUID: 0e796bdb-100d-47d6-a2d5-f7d2daa51f51  (Require a password on wakeup)
      GUID Alias: CONSOLELOCK
      Minimum Possible Setting: 0x00000000
      Maximum Possible Setting: 0x00000001
      Possible Settings increment: 0x00000001
      Possible Settings units:
      Current AC Power Setting Index: 0x00000000
      Current DC Power Setting Index: 0x00000000
`

func TestScreensaverTimeout_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Desktop with display timeout under 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000258
Current DC Power Setting Index: 0x00000258`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: true, // 0x258 = 600 seconds = 10 minutes
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Desktop with display timeout exactly 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x000004b0
Current DC Power Setting Index: 0x000004b0`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: true, // 0x4b0 = 1200 seconds = 20 minutes
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Desktop with display timeout over 20 minutes, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000708
Current DC Power Setting Index: 0x00000708`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false, // 0x708 = 1800 seconds = 30 minutes, no screensaver
			expectedStatus: "Screen timeout exceeds 20 minutes",
		},
		{
			name: "Desktop with display set to Never, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000000
Current DC Power Setting Index: 0x00000000`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false, // Display set to Never, no screensaver
			expectedStatus: "Screen timeout exceeds 20 minutes",
		},
		{
			name: "Laptop with AC timeout over 20 minutes, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000708
Current DC Power Setting Index: 0x000004b0`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false, // AC: 30 min, DC: 20 min, no screensaver
			expectedStatus: "Screen timeout exceeds 20 minutes when plugged in",
		},
		{
			name: "Laptop with DC timeout over 20 minutes, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x000004b0
Current DC Power Setting Index: 0x00000708`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false, // AC: 20 min, DC: 30 min, no screensaver
			expectedStatus: "Screen timeout exceeds 20 minutes on battery",
		},
		{
			name: "Laptop with both AC and DC over 20 minutes, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000708
Current DC Power Setting Index: 0x00000708`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Screen timeout exceeds 20 minutes on both AC and battery",
		},
		{
			name: "Laptop with both timeouts under 20 minutes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000258
Current DC Power Setting Index: 0x0000012c`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: true, // AC: 10 min, DC: 5 min
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Laptop with localized Windows (Slovenian) - AC Never, DC 60 min, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out:     powercfgVideoIdleSlovenian,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false, // AC: Never (0), DC: 60 min (0xe10 = 3600), screensaver disabled
			expectedStatus: "Screen timeout exceeds 20 minutes on both AC and battery",
		},
		{
			name: "Desktop with display Never and screensaver at 10 min - fails (display must be set)",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000000
Current DC Power Setting Index: 0x00000000`,
					Err: nil,
				},
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
			expectedPassed: false, // Display Never - both display AND screensaver must be OK
			expectedStatus: "Screen timeout exceeds 20 minutes",
		},
		{
			name: "Desktop with display 10 min and screensaver at 15 min - passes",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000258
Current DC Power Setting Index: 0x00000258`,
					Err: nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue"},
					Out:     "900",
					Err:     nil,
				},
			},
			expectedPassed: true, // Both display (10 min) and screensaver (15 min) under 20 min
			expectedStatus: "Screensaver or screen lock shows in under 20min",
		},
		{
			name: "Desktop with display 10 min but screensaver over 20 min - fails",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000258
Current DC Power Setting Index: 0x00000258`,
					Err: nil,
				},
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
			expectedPassed: false, // Display OK but screensaver (30 min) over 20 min
			expectedStatus: "Screensaver timeout exceeds 20 minutes",
		},
		{
			name: "Desktop with both display and screensaver over 20 min",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_VIDEO", "VIDEOIDLE"},
					Out: `Current AC Power Setting Index: 0x00000708
Current DC Power Setting Index: 0x00000708`,
					Err: nil,
				},
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
			expectedPassed: false, // Both display (30 min) and screensaver (30 min) over 20 min
			expectedStatus: "Screensaver timeout exceeds 20 minutes",
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

func TestScreensaverTimeout_ParsePowerSettingValue(t *testing.T) {
	check := &ScreensaverTimeout{}

	// Test parsing AC value
	acValue := check.parsePowerSettingValue(powercfgVideoIdleOutput, "AC")
	assert.Equal(t, 1800, acValue) // 0x708 = 1800

	// Test parsing DC value
	dcValue := check.parsePowerSettingValue(powercfgVideoIdleOutput, "DC")
	assert.Equal(t, 900, dcValue) // 0x384 = 900

	// Test parsing localized (Slovenian) output
	acValueSlovenian := check.parsePowerSettingValue(powercfgVideoIdleSlovenian, "AC")
	assert.Equal(t, 0, acValueSlovenian) // 0x00000000 = Never

	dcValueSlovenian := check.parsePowerSettingValue(powercfgVideoIdleSlovenian, "DC")
	assert.Equal(t, 3600, dcValueSlovenian) // 0xe10 = 3600 (60 min)
}

// Minimal powercfg output for modern Windows where CONSOLELOCK is not exposed
const powercfgConsoleLockNotAvailable = `
Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)
  GUID Alias: SCHEME_BALANCED
`

func TestScreensaverPassword_Run(t *testing.T) {
	tests := []struct {
		name           string
		mockCommands   []shared.RunCommandMock
		expectedPassed bool
		expectedStatus string
	}{
		{
			name: "Desktop - sleep password OK, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockEnabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "Desktop - sleep password OK, screensaver enabled and secure",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockEnabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: true,
			expectedStatus: "Password after sleep or screensaver is on",
		},
		{
			name: "Desktop - sleep password OK, screensaver enabled but NOT secure",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockEnabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password not required after screensaver",
		},
		{
			name: "Desktop - sleep password NOT OK, screensaver enabled and secure",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockDisabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password not required on wake from sleep",
		},
		{
			name: "Desktop - neither sleep nor screensaver password enabled",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockDisabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "1",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password not required on wake or screensaver",
		},
		{
			name: "Desktop - sleep password NOT OK, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockDisabled,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
			},
			expectedPassed: false,
			expectedStatus: "Password not required on wake from sleep",
		},
		{
			name: "Modern Windows - CONSOLELOCK not available, lock screen enabled, no screensaver",
			mockCommands: []shared.RunCommandMock{
				{
					Command: "powershell",
					Args:    []string{"-Command", "(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powercfg",
					Args:    []string{"/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK"},
					Out:     powercfgConsoleLockNotAvailable,
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Winlogon' -Name 'DisableLockWorkstation' -ErrorAction SilentlyContinue"},
					Out:     "0",
					Err:     nil,
				},
				{
					Command: "powershell",
					Args:    []string{"-Command", "Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue"},
					Out:     "0",
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

	assert.Equal(t, "Password after sleep or screensaver", check.Name())
	assert.Equal(t, "Password after sleep or screensaver is on", check.PassedMessage())
	assert.Equal(t, "Password after sleep or screensaver is off", check.FailedMessage())
	assert.Equal(t, "37dee029-605b-4aab-96b9-5438e5aa44d8", check.UUID())
	assert.False(t, check.RequiresRoot())
	assert.True(t, check.IsRunnable())
}

func TestScreensaverPassword_ParseConsoleLockValue(t *testing.T) {
	check := &ScreensaverPassword{}

	// Test enabled
	assert.True(t, check.parseConsoleLockValue(powercfgConsoleLockEnabled, "AC"))
	assert.True(t, check.parseConsoleLockValue(powercfgConsoleLockEnabled, "DC"))

	// Test disabled
	assert.False(t, check.parseConsoleLockValue(powercfgConsoleLockDisabled, "AC"))
	assert.False(t, check.parseConsoleLockValue(powercfgConsoleLockDisabled, "DC"))

	// Test empty output
	assert.False(t, check.parseConsoleLockValue("", "AC"))

	// Test minimal output (CONSOLELOCK not available)
	assert.False(t, check.parseConsoleLockValue(powercfgConsoleLockNotAvailable, "AC"))
	assert.False(t, check.parseConsoleLockValue(powercfgConsoleLockNotAvailable, "DC"))
}
