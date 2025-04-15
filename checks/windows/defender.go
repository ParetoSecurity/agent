//go:build windows
// +build windows

package checks

import (
	"golang.org/x/sys/windows/registry"
)

type DefenderEnabledCheck struct {
	passed bool
	status string
}

func (d *DefenderEnabledCheck) Name() string {
	return "Windows Defender is enabled"
}

func (d *DefenderEnabledCheck) Run() error {
	// Check DisableAntiSpyware
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\\Policies\\Microsoft\\Windows Defender`, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		val, _, err := key.GetIntegerValue("DisableAntiSpyware")
		if err == nil && val != 0 {
			d.passed = false
			d.status = "Windows Defender is disabled"
			return nil
		}
	}
	// Check DisableRealtimeMonitoring
	key2, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\\Policies\\Microsoft\\Windows Defender\\Real-TimeProtection`, registry.QUERY_VALUE)
	if err == nil {
		defer key2.Close()
		val, _, err := key2.GetIntegerValue("DisableRealtimeMonitoring")
		if err == nil && val != 0 {
			d.status = "Windows Defender real-time protection is disabled"
			d.passed = false
			return nil
		}
	}
	// If neither disables are set to 1, Defender is enabled
	d.passed = true
	return nil
}

func (d *DefenderEnabledCheck) Passed() bool {
	return d.passed
}
func (d *DefenderEnabledCheck) IsRunnable() bool {
	return true
}
func (d *DefenderEnabledCheck) UUID() string {
	return "defender-ensure-enabled-1"
}
func (d *DefenderEnabledCheck) PassedMessage() string {
	return "Microsoft Defender Antivirus is enabled and real-time protection is on."
}
func (d *DefenderEnabledCheck) FailedMessage() string {
	return "Microsoft Defender Antivirus is disabled or real-time protection is off."
}
func (d *DefenderEnabledCheck) RequiresRoot() bool {
	return false
}
func (d *DefenderEnabledCheck) Status() string {
	if d.Passed() {
		return d.PassedMessage()
	}
	if d.status != "" {
		return d.status
	}
	return d.FailedMessage()
}
