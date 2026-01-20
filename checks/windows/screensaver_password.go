package checks

import (
	"strconv"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

// ScreensaverTimeout checks if screensaver or screen lock shows in under 20 minutes on Windows
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
	// Check if screensaver is active
	activeOut, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveActive' -ErrorAction SilentlyContinue")
	if err != nil {
		s.status = "Failed to query screensaver status"
		s.passed = false
		return nil
	}

	screensaverActive := strings.TrimSpace(string(activeOut))
	if screensaverActive != "1" {
		s.status = "Screensaver is not enabled"
		s.passed = false
		return nil
	}

	// Check screensaver timeout (in seconds)
	timeoutOut, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaveTimeOut' -ErrorAction SilentlyContinue")
	if err != nil {
		s.status = "Failed to query screensaver timeout"
		s.passed = false
		return nil
	}

	timeout := strings.TrimSpace(string(timeoutOut))
	timeoutSeconds, err := strconv.Atoi(timeout)
	if err != nil {
		s.status = "Failed to parse screensaver timeout"
		s.passed = false
		return nil
	}

	// Check if timeout is within 20 minutes (1200 seconds)
	if timeoutSeconds > 0 && timeoutSeconds <= 1200 {
		s.passed = true
		s.status = ""
	} else {
		s.status = "Screensaver timeout is set to more than 20 minutes"
		s.passed = false
	}

	return nil
}

// ScreensaverPassword checks if password is required after sleep or screensaver on Windows
type ScreensaverPassword struct {
	passed bool
	status string
}

// Name returns the check name
func (p *ScreensaverPassword) Name() string {
	return "Password after sleep or screensaver is on"
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
	// Check screensaver password protection
	screensaverSecure := p.checkScreensaverPassword()

	// Detect if system has a battery
	hasBattery := p.hasBattery()

	// Check password after sleep/wake
	sleepPasswordAC := p.checkSleepPasswordAC()

	// For laptops (with battery), check both AC and DC settings
	// For desktops (no battery), only check AC settings
	var sleepPasswordConfigured bool
	if hasBattery {
		sleepPasswordDC := p.checkSleepPasswordDC()
		sleepPasswordConfigured = sleepPasswordAC && sleepPasswordDC
	} else {
		// Desktop - only AC power matters
		sleepPasswordConfigured = sleepPasswordAC
	}

	// Pass if either screensaver password OR sleep password is enabled
	if screensaverSecure || sleepPasswordConfigured {
		p.passed = true
		p.status = ""
	} else {
		p.status = "Password protection is not configured for screensaver or sleep"
		p.passed = false
	}

	return nil
}

// checkScreensaverPassword checks if password is required for screensaver
func (p *ScreensaverPassword) checkScreensaverPassword() bool {
	secureOut, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKCU:\\Control Panel\\Desktop' -Name 'ScreenSaverIsSecure' -ErrorAction SilentlyContinue")
	if err != nil {
		return false
	}

	passwordRequired := strings.TrimSpace(string(secureOut))
	return passwordRequired == "1"
}

// checkSleepPasswordAC checks if password is required when waking from sleep (AC power)
func (p *ScreensaverPassword) checkSleepPasswordAC() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'ACSettingIndex' -ErrorAction SilentlyContinue")
	if err != nil {
		return false
	}

	value := strings.TrimSpace(string(out))
	return value == "1"
}

// checkSleepPasswordDC checks if password is required when waking from sleep (battery power)
func (p *ScreensaverPassword) checkSleepPasswordDC() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-ItemPropertyValue -Path 'HKLM:\\SOFTWARE\\Policies\\Microsoft\\Power\\PowerSettings\\0e796bdb-100d-47d6-a2d5-f7d2daa51f51' -Name 'DCSettingIndex' -ErrorAction SilentlyContinue")
	if err != nil {
		return false
	}

	value := strings.TrimSpace(string(out))
	return value == "1"
}

// hasBattery detects if the system has a battery (laptop) or not (desktop)
func (p *ScreensaverPassword) hasBattery() bool {
	out, err := shared.RunCommand("powershell", "-Command",
		"(Get-WmiObject -Class Win32_Battery | Measure-Object).Count")
	if err != nil {
		// If we can't detect battery, assume it's a desktop (safer default)
		return false
	}

	count := strings.TrimSpace(string(out))
	// If count > 0, system has at least one battery
	batteryCount, err := strconv.Atoi(count)
	if err != nil {
		return false
	}

	return batteryCount > 0
}
