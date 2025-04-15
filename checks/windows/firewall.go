package checks

import (
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

type WindowsFirewall struct {
	passed bool
	status string
}

func (f *WindowsFirewall) checkFirewallProfile(profile string) bool {
	out, err := shared.RunCommand("powershell", "-Command", "Get-NetFirewallProfile -Name '"+profile+"' | Select-Object -ExpandProperty Enabled")

	if err != nil {
		f.status = "Failed to query Windows Firewall for " + profile + " profile"
		return false
	}
	enabled := strings.TrimSpace(string(out))
	if enabled == "True" {
		return true
	}
	f.status = "Windows Firewall is not enabled for " + profile + " profile"
	return false
}

func (f *WindowsFirewall) Name() string {
	return "Windows Firewall is enabled"
}

func (f *WindowsFirewall) Run() error {
	f.passed = f.checkFirewallProfile("Public") && f.checkFirewallProfile("Private")
	return nil
}

func (f *WindowsFirewall) Passed() bool {
	return f.passed
}
func (f *WindowsFirewall) IsRunnable() bool {
	return true
}
func (f *WindowsFirewall) UUID() string {
	return "e632fdd2-b939-4aeb-9a3e-5df2d67d3110"
}
func (f *WindowsFirewall) PassedMessage() string {
	return "Windows Firewall is on"
}
func (f *WindowsFirewall) FailedMessage() string {
	return "Windows Firewall is off"
}
func (f *WindowsFirewall) RequiresRoot() bool {
	return false
}
func (f *WindowsFirewall) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	if f.status != "" {
		return f.status
	}
	return f.FailedMessage()
}
