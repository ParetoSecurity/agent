package checks

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

// ScreensaverTimeout checks if screen turns off or device sleeps within 20 minutes on Windows
type ScreensaverTimeout struct {
	passed bool
	status string
}

// Name returns the check name
func (s *ScreensaverTimeout) Name() string {
	return "Screensaver or screen lock shows in under 20min"
}

// PassedMessage returns the message for when the check passes
func (s *ScreensaverTimeout) PassedMessage() string {
	return "Screensaver or screen lock shows in under 20min"
}

// FailedMessage returns the message for when the check fails
func (s *ScreensaverTimeout) FailedMessage() string {
	return "Screensaver or screen lock shows in more than 20min"
}

// UUID returns the unique identifier for this check (same as macOS)
func (s *ScreensaverTimeout) UUID() string {
	return "13e4dbf1-f87f-4bd9-8a82-f62044f002f4"
}

// Status returns the detailed status message
func (s *ScreensaverTimeout) Status() string {
	if s.Passed() {
		if s.status != "" {
			return s.status
		}
		return s.PassedMessage()
	}
	if s.status != "" {
		return s.status
	}
	return s.FailedMessage()
}

// RequiresRoot returns whether this check needs elevated privileges
func (s *ScreensaverTimeout) RequiresRoot() bool {
	return false
}

// IsRunnable returns whether this check can run on the current platform
func (s *ScreensaverTimeout) IsRunnable() bool {
	return true
}

// Passed returns whether the check passed
func (s *ScreensaverTimeout) Passed() bool {
	return s.passed
}

// Run executes the check
func (s *ScreensaverTimeout) Run() error {
	// Detect if system has a battery (laptop vs desktop)
	hasBattery := s.hasBattery()

	// Get display timeout from current power scheme
	displayTimeoutAC := s.getPowerSettingSeconds("SUB_VIDEO", "VIDEOIDLE", "AC")
	displayTimeoutDC := s.getPowerSettingSeconds("SUB_VIDEO", "VIDEOIDLE", "DC")

	// Get screensaver timeout (applies to both AC and DC)
	// Returns 0 if screensaver is not enabled
	screensaverTimeout := s.getScreensaverTimeout()

	// 20 minutes in seconds
	maxTimeout := 1200

	// Check display timeouts
	displayACOK := displayTimeoutAC > 0 && displayTimeoutAC <= maxTimeout
	displayDCOK := displayTimeoutDC > 0 && displayTimeoutDC <= maxTimeout

	// Check screensaver timeout (only if enabled, i.e., > 0)
	// If screensaver is not enabled (0), it doesn't affect the check
	screensaverOK := screensaverTimeout == 0 || screensaverTimeout <= maxTimeout

	// Check based on device type
	if hasBattery {
		// Laptop: both AC and DC display timeouts must be OK, plus screensaver if enabled
		acOK := displayACOK && screensaverOK
		dcOK := displayDCOK && screensaverOK

		if acOK && dcOK {
			s.passed = true
			s.status = ""
		} else if !screensaverOK {
			s.status = "Screensaver timeout exceeds 20 minutes"
			s.passed = false
		} else if !displayACOK && !displayDCOK {
			s.status = "Screen timeout exceeds 20 minutes on both AC and battery"
			s.passed = false
		} else if !displayACOK {
			s.status = "Screen timeout exceeds 20 minutes when plugged in"
			s.passed = false
		} else {
			s.status = "Screen timeout exceeds 20 minutes on battery"
			s.passed = false
		}
	} else {
		// Desktop: only AC matters, plus screensaver if enabled
		if displayACOK && screensaverOK {
			s.passed = true
			s.status = ""
		} else if !screensaverOK {
			s.status = "Screensaver timeout exceeds 20 minutes"
			s.passed = false
		} else {
			s.status = "Screen timeout exceeds 20 minutes"
			s.passed = false
		}
	}

	return nil
}

// getScreensaverTimeout gets the screensaver timeout in seconds from registry
func (s *ScreensaverTimeout) getScreensaverTimeout() int {
	// First check if screensaver is active
	activeOut, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue")
	if err != nil {
		return 0
	}

	if strings.TrimSpace(string(activeOut)) != "1" {
		return 0 // Screensaver not enabled
	}

	// Get screensaver timeout
	timeoutOut, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue")
	if err != nil {
		return 0
	}

	timeout, err := strconv.Atoi(strings.TrimSpace(string(timeoutOut)))
	if err != nil {
		return 0
	}

	return timeout
}

// getPowerSettingSeconds queries powercfg for a power setting value in seconds
func (s *ScreensaverTimeout) getPowerSettingSeconds(subgroup, setting, powerSource string) int {
	out, err := shared.RunCommand("powercfg", "/query", "SCHEME_CURRENT", subgroup, setting)
	if err != nil {
		return 0
	}

	return s.parsePowerSettingValue(string(out), powerSource)
}

// parsePowerSettingValue extracts AC or DC power setting value from powercfg output
func (s *ScreensaverTimeout) parsePowerSettingValue(output, powerSource string) int {
	lines := strings.Split(output, "\n")

	var pattern string
	if powerSource == "AC" {
		pattern = `Current AC Power Setting Index:\s*0x([0-9a-fA-F]+)`
	} else {
		pattern = `Current DC Power Setting Index:\s*0x([0-9a-fA-F]+)`
	}

	re := regexp.MustCompile(pattern)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			value, err := strconv.ParseInt(matches[1], 16, strconv.IntSize)
			if err != nil {
				return 0
			}
			return int(value)
		}
	}

	return 0
}

// hasBattery detects if the system has a battery (laptop) or not (desktop)
func (s *ScreensaverTimeout) hasBattery() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count")
	if err != nil {
		return false
	}

	count := strings.TrimSpace(string(out))
	batteryCount, err := strconv.Atoi(count)
	if err != nil {
		return false
	}

	return batteryCount > 0
}

// ScreensaverPassword checks if password is required after sleep or screen off on Windows
type ScreensaverPassword struct {
	passed bool
	status string
}

// Name returns the check name
func (p *ScreensaverPassword) Name() string {
	return "Password after sleep or screensaver"
}

// PassedMessage returns the message for when the check passes
func (p *ScreensaverPassword) PassedMessage() string {
	return "Password after sleep or screensaver is on"
}

// FailedMessage returns the message for when the check fails
func (p *ScreensaverPassword) FailedMessage() string {
	return "Password after sleep or screensaver is off"
}

// UUID returns the unique identifier for this check (same as macOS PasswordAfterSleep)
func (p *ScreensaverPassword) UUID() string {
	return "37dee029-605b-4aab-96b9-5438e5aa44d8"
}

// Status returns the detailed status message
func (p *ScreensaverPassword) Status() string {
	if p.Passed() {
		if p.status != "" {
			return p.status
		}
		return p.PassedMessage()
	}
	if p.status != "" {
		return p.status
	}
	return p.FailedMessage()
}

// RequiresRoot returns whether this check needs elevated privileges
func (p *ScreensaverPassword) RequiresRoot() bool {
	return false
}

// IsRunnable returns whether this check can run on the current platform
func (p *ScreensaverPassword) IsRunnable() bool {
	return true
}

// Passed returns whether the check passed
func (p *ScreensaverPassword) Passed() bool {
	return p.passed
}

// Run executes the check
func (p *ScreensaverPassword) Run() error {
	// Detect if system has a battery (laptop vs desktop)
	hasBattery := p.hasBattery()

	// Check screensaver password protection (ScreenSaverIsSecure)
	screensaverSecure := p.isScreensaverSecure()

	// Try CONSOLELOCK setting from current power scheme
	// CONSOLELOCK controls "Require a password on wakeup"
	// Value: 0 = No, 1 = Yes
	consoleLockAC, consoleLockDC, consoleLockAvailable := p.getConsoleLockSettings()

	// For modern Windows where CONSOLELOCK is not exposed, check lock screen enabled
	var sleepPasswordOK bool
	if consoleLockAvailable {
		if hasBattery {
			sleepPasswordOK = consoleLockAC && consoleLockDC
		} else {
			sleepPasswordOK = consoleLockAC
		}
	} else {
		// Fallback: check if lock screen is enabled
		sleepPasswordOK = p.isLockScreenEnabled()
	}

	// Pass if BOTH sleep password AND screensaver password (if screensaver enabled) are configured
	screensaverEnabled := p.isScreensaverEnabled()

	if screensaverEnabled {
		// If screensaver is enabled, both must be secure
		if sleepPasswordOK && screensaverSecure {
			p.passed = true
			p.status = ""
		} else if !sleepPasswordOK && !screensaverSecure {
			p.status = "Password not required on wake or screensaver"
			p.passed = false
		} else if !sleepPasswordOK {
			p.status = "Password not required on wake from sleep"
			p.passed = false
		} else {
			p.status = "Password not required after screensaver"
			p.passed = false
		}
	} else {
		// If screensaver is not enabled, only check sleep password
		if sleepPasswordOK {
			p.passed = true
			p.status = ""
		} else {
			p.status = "Password not required on wake from sleep"
			p.passed = false
		}
	}

	return nil
}

// isScreensaverEnabled checks if screensaver is enabled
func (p *ScreensaverPassword) isScreensaverEnabled() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

// isScreensaverSecure checks if password is required after screensaver
func (p *ScreensaverPassword) isScreensaverSecure() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "1"
}

// getConsoleLockSettings checks if password is required on wake for AC and DC
// Returns (acEnabled, dcEnabled, settingAvailable)
func (p *ScreensaverPassword) getConsoleLockSettings() (bool, bool, bool) {
	// Query CONSOLELOCK setting from current power scheme
	// SUB_NONE is used because CONSOLELOCK is not under a specific subgroup
	out, err := shared.RunCommand("powercfg", "/query", "SCHEME_CURRENT", "SUB_NONE", "CONSOLELOCK")
	if err != nil {
		return false, false, false
	}

	output := string(out)
	// Check if the output contains the actual setting values
	// On modern Windows, CONSOLELOCK may not be exposed and returns minimal output
	if !strings.Contains(output, "Current AC Power Setting Index") {
		return false, false, false
	}

	acEnabled := p.parseConsoleLockValue(output, "AC")
	dcEnabled := p.parseConsoleLockValue(output, "DC")
	return acEnabled, dcEnabled, true
}

// isLockScreenEnabled checks if the Windows lock screen is enabled
// This is a fallback for modern Windows where CONSOLELOCK is not exposed
func (p *ScreensaverPassword) isLockScreenEnabled() bool {
	// Check DisableLockWorkstation registry key
	// Value 0 = lock enabled, Value 1 = lock disabled
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Winlogon' -Name 'DisableLockWorkstation' -ErrorAction SilentlyContinue")
	if err != nil {
		// If key doesn't exist, assume lock is enabled (default Windows behavior)
		return true
	}

	value := strings.TrimSpace(string(out))
	// 0 = lock enabled (good), 1 = lock disabled (bad)
	return value == "0"
}

// parseConsoleLockValue extracts AC or DC console lock setting from powercfg output
func (p *ScreensaverPassword) parseConsoleLockValue(output, powerSource string) bool {
	lines := strings.Split(output, "\n")

	var pattern string
	if powerSource == "AC" {
		pattern = `Current AC Power Setting Index:\s*0x([0-9a-fA-F]+)`
	} else {
		pattern = `Current DC Power Setting Index:\s*0x([0-9a-fA-F]+)`
	}

	re := regexp.MustCompile(pattern)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 2 {
			value, err := strconv.ParseInt(matches[1], 16, 64)
			if err != nil {
				return false
			}
			// 1 = password required, 0 = not required
			return value == 1
		}
	}

	return false
}

// hasBattery detects if the system has a battery (laptop) or not (desktop)
func (p *ScreensaverPassword) hasBattery() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"(Get-CimInstance -ClassName Win32_Battery | Measure-Object).Count")
	if err != nil {
		return false
	}

	count := strings.TrimSpace(string(out))
	batteryCount, err := strconv.Atoi(count)
	if err != nil {
		return false
	}

	return batteryCount > 0
}
