package checks

import (
	"encoding/json"
	"regexp"
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
	return "Antivirus software is enabled"
}

func (d *WindowsDefender) Run() error {
	// First try PowerShell method for detailed Defender status
	out, err := shared.RunCommand("powershell", "-Command", "Get-MpComputerStatus | Select-Object RealTimeProtectionEnabled, IoavProtectionEnabled, AntispywareEnabled | ConvertTo-Json")
	if err == nil {
		// Remove BOM if present
		outStr := strings.TrimPrefix(string(out), "\xef\xbb\xbf")
		var status mpStatus
		if err := json.Unmarshal([]byte(outStr), &status); err == nil {
			if status.RealTimeProtectionEnabled && status.IoavProtectionEnabled && status.AntispywareEnabled {
				d.passed = true
				d.status = ""
				return nil
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
		}
	}

	// Fallback to wmic SecurityCenter method to detect any antivirus
	avOut, avErr := shared.RunCommand("wmic", "/namespace:\\\\root\\SecurityCenter2", "path", "AntiVirusProduct", "get", "/value")
	if avErr != nil {
		// Try SecurityCenter (older systems)
		avOut, avErr = shared.RunCommand("wmic", "/namespace:\\\\root\\SecurityCenter", "path", "AntiVirusProduct", "get", "/value")
		if avErr != nil {
			d.passed = false
			d.status = "Failed to query antivirus status"
			return nil
		}
	}

	// Parse wmic output for antivirus products
	if d.parseAntivirusOutput(string(avOut)) {
		d.passed = true
		d.status = ""
	} else {
		d.passed = false
		d.status = "No antivirus software detected"
	}
	return nil
}

func (d *WindowsDefender) parseAntivirusOutput(output string) bool {
	// Look for active antivirus products in wmic output
	lines := strings.Split(output, "\n")
	var currentProduct map[string]string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line indicates end of product block
			if currentProduct != nil && d.isAntivirusActive(currentProduct) {
				return true
			}
			currentProduct = nil
			continue
		}

		// Parse key=value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				if currentProduct == nil {
					currentProduct = make(map[string]string)
				}
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				currentProduct[key] = value
			}
		}
	}

	// Check final product if exists
	if currentProduct != nil && d.isAntivirusActive(currentProduct) {
		return true
	}

	return false
}

func (d *WindowsDefender) isAntivirusActive(product map[string]string) bool {
	// Check if product has a display name (indicates it exists)
	displayName, hasName := product["displayName"]
	if !hasName || displayName == "" {
		return false
	}

	// Check product state - varies by Windows version
	// For SecurityCenter2: productState is a hex value where certain bits indicate status
	if productState, hasState := product["productState"]; hasState && productState != "" {
		// Parse hex productState to check if antivirus is enabled
		// Bit patterns for enabled/updated antivirus typically have specific values
		// This catches most third-party antivirus and Windows Defender
		re := regexp.MustCompile(`^[0-9]+$`)
		if re.MatchString(productState) {
			// Simple heuristic: non-zero state usually indicates active antivirus
			return productState != "0"
		}
	}

	// For older SecurityCenter: check if onAccessScanningEnabled exists and is true
	if onAccess, hasOnAccess := product["onAccessScanningEnabled"]; hasOnAccess {
		return strings.ToLower(onAccess) == "true"
	}

	// If we have a display name but no clear state info, assume active
	// This is conservative but prevents false negatives
	return true
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
	return "Antivirus software is active"
}
func (d *WindowsDefender) FailedMessage() string {
	return "No antivirus software detected"
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
