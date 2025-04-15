package checks

import (
	"encoding/json"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

type WindowsDefender struct {
	passed bool
	status string
}

type mpStatus struct {
	RealTimeProtectionEnabled bool
	IoavProtectionEnabled     bool
	AntispywareEnabled        bool
}

func (d *WindowsDefender) Name() string {
	return "Windows Defender is enabled"
}

func (d *WindowsDefender) Run() error {
	out, err := shared.RunCommand("powershell", "-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json")
	if err != nil {
		d.passed = false
		d.status = "Failed to query Defender status"
		return nil
	}
	// Remove BOM if present
	outStr := strings.TrimPrefix(string(out), "\xef\xbb\xbf")
	var status mpStatus
	if err := json.Unmarshal([]byte(outStr), &status); err != nil {
		d.passed = false
		d.status = "Failed to parse Defender status"
		return nil
	}
	if status.RealTimeProtectionEnabled && status.IoavProtectionEnabled && status.AntispywareEnabled {
		d.passed = true
		d.status = ""
	} else {
		d.passed = false
		// Compose a status message with details
		if !status.RealTimeProtectionEnabled {
			d.status = "Defender has disabled real-time protection"
			return nil
		}
		if !status.IoavProtectionEnabled {
			d.status = "Defender has disabled tamper protection"
			return nil
		}
		if !status.AntispywareEnabled {
			d.status = "Defender is disabled"
			return nil
		}
	}
	return nil
}

func (d *WindowsDefender) Passed() bool {
	return d.passed
}
func (d *WindowsDefender) IsRunnable() bool {
	return true
}
func (d *WindowsDefender) UUID() string {
	return "2be03cd7-5cb5-4778-a01a-7ba2fb22750a"
}
func (d *WindowsDefender) PassedMessage() string {
	return "Microsoft Defender is on."
}
func (d *WindowsDefender) FailedMessage() string {
	return "Microsoft Defender is off."
}
func (d *WindowsDefender) RequiresRoot() bool {
	return false
}
func (d *WindowsDefender) Status() string {
	if d.Passed() {
		return d.PassedMessage()
	}
	if d.status != "" {
		return d.status
	}
	return d.FailedMessage()
}
