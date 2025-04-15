//go:build windows
// +build windows

package checks

import (
	"golang.org/x/sys/windows/registry"
)

// Firewall checks if Windows Firewall is enabled for Public and Private profiles.
type Firewall struct {
	passed bool
	status string
}

var checkFirewallProfile = func(profile string) bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\\Policies\\Microsoft\\WindowsFirewall\\`+profile, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	val, _, err := key.GetIntegerValue("EnableFirewall")
	return err == nil && val == 1
}

func (f *Firewall) Name() string {
	return "Windows Firewal is enabled"
}

func (f *Firewall) Run() error {
	f.passed = checkFirewallProfile("PublicProfile") && checkFirewallProfile("PrivateProfile")
	return nil
}

func (f *Firewall) Passed() bool {
	return f.passed
}
func (f *Firewall) IsRunnable() bool {
	return true
}
func (f *Firewall) UUID() string {
	return "e632fdd2-b939-4aeb-9a3e-5df2d67d3110"
}
func (f *Firewall) PassedMessage() string {
	return "Windows Firewall is on"
}
func (f *Firewall) FailedMessage() string {
	return "Windows Firewall is off"
}
func (f *Firewall) RequiresRoot() bool {
	return false
}
func (f *Firewall) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	if f.status != "" {
		return f.status
	}
	return f.FailedMessage()
}
