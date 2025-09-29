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

func (f *WindowsFirewall) checkESETFirewall() bool {
	// Check if ESET firewall service is running
	out, err := shared.RunCommand("powershell", "-Command", "Get-Service -Name 'ESET Service' -ErrorAction SilentlyContinue | Select-Object Status")
	if err != nil {
		return false
	}
	if !strings.Contains(string(out), "Running") {
		return false
	}

	// Check if ESET registry key exists (indicates ESET is installed)
	out, err = shared.RunCommand("powershell", "-Command", "Test-Path 'HKLM:\\SOFTWARE\\ESET'")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "True"
}

func (f *WindowsFirewall) Name() string {
	return "Firewall is enabled"
}

func (f *WindowsFirewall) Run() error {
	// First check if Windows Firewall is enabled
	windowsFirewallEnabled := f.checkFirewallProfile("Public") && f.checkFirewallProfile("Private")

	// If Windows Firewall is not enabled, check for ESET firewall
	if !windowsFirewallEnabled {
		esetFirewallEnabled := f.checkESETFirewall()
		if esetFirewallEnabled {
			f.passed = true
			f.status = "ESET firewall is active"
			return nil
		}
	} else {
		f.passed = true
		f.status = ""
		return nil
	}

	f.passed = false
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
	return "Firewall is active"
}
func (f *WindowsFirewall) FailedMessage() string {
	return "No firewall detected"
}
func (f *WindowsFirewall) RequiresRoot() bool {
	return false
}
func (f *WindowsFirewall) Status() string {
	if f.Passed() {
		if f.status != "" {
			return f.status
		}
		return f.PassedMessage()
	}
	if f.status != "" {
		return f.status
	}
	return f.FailedMessage()
}
