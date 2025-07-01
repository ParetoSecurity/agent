package checks

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

type WindowsDefender struct {
	passed bool
	status string
}

type AntivirusProduct struct {
	DisplayName              string `json:"displayName"`
	InstanceGuid             string `json:"instanceGuid"`
	PathToSignedProductExe   string `json:"pathToSignedProductExe"`
	PathToSignedReportingExe string `json:"pathToSignedReportingExe"`
	ProductState             string `json:"productState"`
	Timestamp                string `json:"timestamp"`
}

type EDRProduct struct {
	Name         string
	Processes    []string
	Services     []string
	RegistryKeys []string
	InstallPaths []string
}

func (d *WindowsDefender) Name() string {
	return "Antivirus software is enabled"
}

func (d *WindowsDefender) getEDRProducts() []EDRProduct {
	return []EDRProduct{
		{
			Name:         "CrowdStrike Falcon",
			Processes:    []string{"CSFalconService", "CSFalconContainer"},
			Services:     []string{"CSAgent", "CSFalconService"},
			RegistryKeys: []string{"HKLM:\\SYSTEM\\CrowdStrike"},
			InstallPaths: []string{"%ProgramFiles%\\CrowdStrike"},
		},
		{
			Name:         "SentinelOne",
			Processes:    []string{"SentinelAgent", "SentinelServiceHost", "SentinelUI"},
			Services:     []string{"SentinelAgent", "LogProcessorService", "SentinelStaticEngine"},
			RegistryKeys: []string{"HKLM:\\SOFTWARE\\SentinelOne"},
			InstallPaths: []string{"%ProgramFiles%\\SentinelOne", "%ProgramData%\\Sentinel"},
		},
		{
			Name:         "Carbon Black",
			Processes:    []string{"cb", "cbcomms", "cbstream", "RepMgr"},
			Services:     []string{"CbDefense", "CarbonBlack", "RepMgr"},
			RegistryKeys: []string{"HKLM:\\SOFTWARE\\CarbonBlack"},
			InstallPaths: []string{"%ProgramFiles%\\Confer"},
		},
		{
			Name:         "Symantec Endpoint Protection",
			Processes:    []string{"Smc", "SmcGui", "ccSvcHst", "Rtvscan"},
			Services:     []string{"Symantec Endpoint Protection", "SepMasterService", "ccSetMgr"},
			RegistryKeys: []string{"HKLM:\\SOFTWARE\\Symantec\\Symantec Endpoint Protection"},
			InstallPaths: []string{"%ProgramFiles%\\Symantec\\Symantec Endpoint Protection"},
		},
	}
}

func (d *WindowsDefender) Run() error {
	// Use Get-CimInstance to query antivirus products from SecurityCenter2
	out, err := shared.RunCommand("powershell", "-Command", "Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | ConvertTo-Json")
	if err != nil {
		d.passed = false
		d.status = "Failed to query antivirus status"
		return nil
	}

	// Remove BOM if present
	outStr := strings.TrimPrefix(string(out), "\xef\xbb\xbf")
	outStr = strings.TrimSpace(outStr)

	if outStr == "" {
		d.passed = false
		d.status = "No antivirus software detected"
		return nil
	}

	// Parse JSON output - could be single object or array
	var products []AntivirusProduct
	if strings.HasPrefix(outStr, "[") {
		// Multiple products
		if err := json.Unmarshal([]byte(outStr), &products); err != nil {
			d.passed = false
			d.status = "Failed to parse antivirus data"
			return nil
		}
	} else {
		// Single product
		var product AntivirusProduct
		if err := json.Unmarshal([]byte(outStr), &product); err != nil {
			d.passed = false
			d.status = "Failed to parse antivirus data"
			return nil
		}
		products = []AntivirusProduct{product}
	}

	// Check if any antivirus product is active
	for _, product := range products {
		if d.isAntivirusActive(product) {
			d.passed = true
			d.status = ""
			return nil
		}
	}

	d.passed = false
	d.status = "No active antivirus software detected"
	return nil
}

func (d *WindowsDefender) isAntivirusActive(product AntivirusProduct) bool {
	// Check if product has a display name (indicates it exists)
	if product.DisplayName == "" {
		return false
	}

	// Check product state - SecurityCenter2 productState is a decimal value where bits indicate status
	if product.ProductState != "" {
		// Parse decimal productState to check if antivirus is enabled
		// ProductState bit analysis:
		// Bit 13 (0x1000): Real-time protection enabled when set
		// Bit 5 (0x10): Virus definitions outdated when set
		// Common values:
		// 266240 (0x41000): Up to date, Enabled, Real-time ON, Definitions Current (AVG Antivirus)
		// 266256 (0x41010): Out of date, Enabled, Real-time ON, Definitions Outdated
		// 262144 (0x40000): Up to date, Disabled, Real-time OFF, Definitions Current
		// 393472 (0x60100): Up to date, Disabled, Real-time OFF, Definitions Current (Windows Defender disabled)
		// 397568 (0x61100): Up to date, Enabled, Real-time ON, Definitions Current (Windows Defender enabled)
		if state, err := strconv.ParseUint(product.ProductState, 10, 32); err == nil {
			// Check if real-time protection is enabled (bit 13)
			realTimeEnabled := (state & 0x1000) != 0

			// Antivirus is considered active if real-time protection is enabled
			// Note: Bit 5 (0x10) indicates if definitions are outdated, but we don't fail
			// the check for outdated definitions as real-time protection can still work
			return realTimeEnabled
		}
	}

	// If we have a display name but can't parse state, assume active
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
