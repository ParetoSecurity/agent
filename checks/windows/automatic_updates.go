package checks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ParetoSecurity/agent/shared"
)

type AutomaticUpdatesCheck struct {
	passed bool
	status string
}

type autoUpdateSettings struct {
	NotificationLevel int `json:"NotificationLevel"`
}

func (a *AutomaticUpdatesCheck) Name() string {
	return "Automatic Updates are enabled"
}

func (a *AutomaticUpdatesCheck) Run() error {
	out, err := shared.RunCommand("powershell", "-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json")

	if err != nil {
		a.passed = false
		a.status = "Failed to query update settings"
		return nil
	}
	var settings autoUpdateSettings
	if err := json.Unmarshal([]byte(out), &settings); err != nil {
		a.passed = false
		a.status = "Failed to parse update settings"
		return nil
	}
	// NotificationLevel 1 = Never check for updates, 2 = Notify before download, 3 = Notify before install, 4 = Scheduled install
	if settings.NotificationLevel == 1 {
		a.status = "Automatic Updates are disabled"
		a.passed = false
		return nil
	}

	// Check if updates are paused
	psCmd := `try { Get-ItemPropertyValue -Path "HKLM:\SOFTWARE\Microsoft\WindowsUpdate\UX\Settings" -Name "PauseUpdatesExpiryTime" } catch { 0 }`
	pauseOut, pauseErr := shared.RunCommand("powershell", "-Command", psCmd)
	if pauseErr == nil && len(pauseOut) > 0 {
		// Parse output as int64 (epoch seconds)
		var expiry int64
		_, scanErr := fmt.Sscanf(pauseOut, "%d", &expiry)
		if scanErr == nil && expiry > 0 {
			now := time.Now().Unix()
			if expiry > now {
				a.passed = false
				a.status = "Updates are paused"
				return nil
			}
		}
	}

	a.passed = true
	return nil
}

func (a *AutomaticUpdatesCheck) Passed() bool {
	return a.passed
}
func (a *AutomaticUpdatesCheck) IsRunnable() bool {
	return true
}
func (a *AutomaticUpdatesCheck) UUID() string {
	return "28d98536-a93a-4092-845a-92ec081cc82a"
}
func (a *AutomaticUpdatesCheck) PassedMessage() string {
	return "Automatic Updates are on"
}
func (a *AutomaticUpdatesCheck) FailedMessage() string {
	return "Automatic Updates are off/paused"
}
func (a *AutomaticUpdatesCheck) RequiresRoot() bool {
	return false
}
func (a *AutomaticUpdatesCheck) Status() string {
	if a.Passed() {
		return a.PassedMessage()
	}
	if a.status != "" {
		return a.status
	}
	return a.FailedMessage()
}
