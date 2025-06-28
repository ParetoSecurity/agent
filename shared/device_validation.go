package shared

import (
	"regexp"
	"strings"

	"github.com/caarlos0/log"
)

// Validation constraints from OpenAPI spec
const (
	maxMachineNameLength = 255
	maxModelNameLength   = 60
	maxModelSerialLength = 15
	maxOSVersionLength   = 255
	uuidLength           = 36
)

var (
	// Pattern for modelSerial: ^[a-zA-Z0-9\.!\-'"_]+$
	modelSerialPattern = regexp.MustCompile(`^[a-zA-Z0-9\.!\-'"_]+$`)

	// Pattern for macOSVersion: ^(\d+\.)?(\d+\.)?(\*|\d+)$
	macOSVersionPattern = regexp.MustCompile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)

	// Pattern for linux/windowsOSVersion: ^[a-zA-Z0-9\.!\-'" _]+$
	osVersionPattern = regexp.MustCompile(`^[a-zA-Z0-9\.!\-'" _]+$`)

	// Pattern for machineUUID: ^[a-zA-Z0-9-]+$
	uuidPattern = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// TruncateString truncates a string to the specified maximum length
// It properly handles UTF-8 characters to avoid splitting multi-byte characters
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	// Convert to runes to handle UTF-8 properly
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}

	// Truncate and convert back, but ensure we don't exceed byte length
	truncated := string(runes[:maxLength])
	for len(truncated) > maxLength {
		runes = runes[:len(runes)-1]
		truncated = string(runes)
	}

	return truncated
}

// ValidateAndPrepareDevice validates and prepares device data according to OpenAPI spec
func ValidateAndPrepareDevice(device *ReportingDevice) {
	// Validate and truncate machineName
	device.MachineName = TruncateString(device.MachineName, maxMachineNameLength)

	// Validate and truncate modelName
	device.ModelName = TruncateString(device.ModelName, maxModelNameLength)

	// Validate and prepare modelSerial
	if len(device.ModelSerial) > maxModelSerialLength {
		device.ModelSerial = TruncateString(device.ModelSerial, maxModelSerialLength)
	}
	// Ensure it matches the pattern after truncation
	if !modelSerialPattern.MatchString(device.ModelSerial) {
		// If it doesn't match after sanitization, default to Unknown
		device.ModelSerial = "Unknown"
	}

	// Validate OS version
	device.OSVersion = TruncateString(device.OSVersion, maxOSVersionLength)

	// Validate UUID length and pattern
	if len(device.MachineUUID) != uuidLength || !uuidPattern.MatchString(device.MachineUUID) {
		// This should not happen, but log it if it does
		log.Errorf("Invalid UUID format: %s", device.MachineUUID)
	}
}

// ValidateMacOSVersion checks if a macOS version string matches the required pattern
func ValidateMacOSVersion(version string) bool {
	return macOSVersionPattern.MatchString(version)
}

// ValidateOSVersion checks if a Linux/Windows OS version string matches the required pattern
func ValidateOSVersion(version string) bool {
	return osVersionPattern.MatchString(version)
}

// FormatMacOSVersion attempts to format a version string to match macOS pattern
func FormatMacOSVersion(version string) string {
	// Extract version numbers from strings like "Darwin 23.5.0" or "macOS 14.5"
	// Try to extract major.minor.patch pattern
	versionExtractor := regexp.MustCompile(`(\d+)\.(\d+)(?:\.(\d+))?`)
	matches := versionExtractor.FindStringSubmatch(version)

	if len(matches) > 0 {
		result := matches[1]
		if len(matches) > 2 && matches[2] != "" {
			result += "." + matches[2]
		}
		if len(matches) > 3 && matches[3] != "" {
			result += "." + matches[3]
		}
		return result
	}

	// If no pattern matches, try to clean the string
	cleaned := strings.TrimSpace(version)
	if macOSVersionPattern.MatchString(cleaned) {
		return cleaned
	}

	// Default to a valid version
	return "0.0.0"
}
