//go:build windows
// +build windows

package checks

import (
	"golang.org/x/sys/windows/registry"
)

type AutomaticUpdatesCheck struct {
	passed bool
	status string
}

func (a *AutomaticUpdatesCheck) Name() string {
	return "Automatic Updates are enabled"
}

func (a *AutomaticUpdatesCheck) Run() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\\Policies\\Microsoft\\Windows\\WindowsUpdate\\AU`, registry.QUERY_VALUE)
	if err != nil {
		// Key missing = policy not set, so fail
		a.passed = false
		return nil
	}
	defer key.Close()
	val, _, err := key.GetIntegerValue("NoAutoUpdate")
	if err != nil {
		// Value missing = updates enabled
		a.passed = true
		return nil
	}
	a.passed = (val == 0)
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
