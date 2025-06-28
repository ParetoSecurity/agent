package shared

import (
	"strings"
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{"Short string", "hello", 10, "hello"},
		{"Exact length", "hello", 5, "hello"},
		{"Long string", "hello world", 5, "hello"},
		{"Empty string", "", 5, ""},
		{"Unicode truncation bytes", "hello 世界", 7, "hello "},    // 7 bytes: "hello " + first byte of 世
		{"Unicode truncation runes", "hello 世界", 10, "hello 世界"}, // Fits within 10 bytes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q; want %q", tt.input, tt.maxLength, result, tt.expected)
			}
		})
	}
}

func TestValidateMacOSVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		// Valid versions
		{"14", true},
		{"14.5", true},
		{"14.5.0", true},
		{"10.15.7", true},
		{"11.0", true},
		{"*", true},
		{"12.*", true},

		// Invalid versions
		{"", false},
		{"14.5.0.1", false},
		{"macOS 14.5", false},
		{"14.5-beta", false},
		{"version 14", false},
		{"14.5.0 (23F79)", false},
		{"Darwin 23.5.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := ValidateMacOSVersion(tt.version)
			if result != tt.expected {
				t.Errorf("ValidateMacOSVersion(%q) = %v; want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestValidateOSVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		// Valid versions
		{"Ubuntu 22.04 LTS", true},
		{"Windows 11 Pro", true},
		{"CentOS Linux 8", true},
		{"Debian GNU-Linux 12", true},
		{"Arch Linux", true},
		{"Red Hat Enterprise Linux 9.2", true},

		// Invalid versions (contain characters not in pattern)
		{"Ubuntu 22.04 (Jammy)", false},
		{"Windows 11 [Pro]", false},
		{"Linux@2023", false},
		{"OS#Version", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := ValidateOSVersion(tt.version)
			if result != tt.expected {
				t.Errorf("ValidateOSVersion(%q) = %v; want %v", tt.version, result, tt.expected)
			}
		})
	}
}

func TestFormatMacOSVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Extract from Darwin versions
		{"Darwin 23.5.0", "23.5.0"},
		{"Darwin 22.6.0", "22.6.0"},
		{"Darwin 21.6.0", "21.6.0"},

		// Extract from macOS versions
		{"macOS 14.5", "14.5"},
		{"macOS 14.5.0", "14.5.0"},
		{"macOS 13.4.1", "13.4.1"},

		// Already valid versions
		{"14.5", "14.5"},
		{"14.5.0", "14.5.0"},
		{"14", "14"},

		// Complex strings
		{"Version 14.5.0 (Build 23F79)", "14.5.0"},
		{"Mac OS X 10.15.7", "10.15.7"},

		// Invalid inputs
		{"No version here", "0.0.0"},
		{"", "0.0.0"},
		{"macOS Sonoma", "0.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatMacOSVersion(tt.input)
			if result != tt.expected {
				t.Errorf("FormatMacOSVersion(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateAndPrepareDevice(t *testing.T) {
	tests := []struct {
		name     string
		device   ReportingDevice
		expected ReportingDevice
	}{
		{
			name: "Valid device data",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
		},
		{
			name: "Long machine name",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: strings.Repeat("a", 300),
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: strings.Repeat("a", 255),
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
		},
		{
			name: "Long model name",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   strings.Repeat("b", 100),
				ModelSerial: "ABC123",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   strings.Repeat("b", 60),
				ModelSerial: "ABC123",
			},
		},
		{
			name: "Long model serial",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABCDEFGHIJKLMNO",
			},
		},
		{
			name: "Invalid model serial characters",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "ABC@123#INVALID",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   "Ubuntu 22.04",
				ModelName:   "Dell XPS",
				ModelSerial: "Unknown",
			},
		},
		{
			name: "Long OS version",
			device: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   strings.Repeat("Ubuntu 22.04 LTS ", 20),
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
			expected: ReportingDevice{
				MachineUUID: "123e4567-e89b-12d3-a456-426614174000",
				MachineName: "test-machine",
				Auth:        "test-auth",
				OSVersion:   strings.Repeat("Ubuntu 22.04 LTS ", 20)[:255],
				ModelName:   "Dell XPS",
				ModelSerial: "ABC123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := tt.device
			ValidateAndPrepareDevice(&device)

			if device.MachineUUID != tt.expected.MachineUUID {
				t.Errorf("MachineUUID = %q; want %q", device.MachineUUID, tt.expected.MachineUUID)
			}
			if device.MachineName != tt.expected.MachineName {
				t.Errorf("MachineName = %q; want %q", device.MachineName, tt.expected.MachineName)
			}
			if device.Auth != tt.expected.Auth {
				t.Errorf("Auth = %q; want %q", device.Auth, tt.expected.Auth)
			}
			if device.OSVersion != tt.expected.OSVersion {
				t.Errorf("OSVersion = %q; want %q", device.OSVersion, tt.expected.OSVersion)
			}
			if device.ModelName != tt.expected.ModelName {
				t.Errorf("ModelName = %q; want %q", device.ModelName, tt.expected.ModelName)
			}
			if device.ModelSerial != tt.expected.ModelSerial {
				t.Errorf("ModelSerial = %q; want %q", device.ModelSerial, tt.expected.ModelSerial)
			}
		})
	}
}

func TestModelSerialPattern(t *testing.T) {
	tests := []struct {
		serial   string
		expected bool
	}{
		// Valid serials
		{"ABC123", true},
		{"A-B_C.123", true},
		{"Serial!Number", true},
		{"S'N\"123", true},
		{"12345", true},
		{"Unknown", true},

		// Invalid serials
		{"ABC 123", false},  // Contains space
		{"ABC@123", false},  // Contains @
		{"ABC#123", false},  // Contains #
		{"ABC,123", false},  // Contains comma
		{"", false},         // Empty
		{"ABC(123)", false}, // Contains parentheses
		{"ABC[123]", false}, // Contains brackets
	}

	for _, tt := range tests {
		t.Run(tt.serial, func(t *testing.T) {
			result := modelSerialPattern.MatchString(tt.serial)
			if result != tt.expected {
				t.Errorf("modelSerialPattern.MatchString(%q) = %v; want %v", tt.serial, result, tt.expected)
			}
		})
	}
}

func TestUUIDPattern(t *testing.T) {
	tests := []struct {
		uuid     string
		expected bool
	}{
		// Valid UUIDs
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE", true},

		// Invalid UUIDs
		{"123e4567-e89b-12d3-a456-42661417400", false},   // Too short
		{"123e4567-e89b-12d3-a456-4266141740000", false}, // Too long
		{"123e4567_e89b_12d3_a456_426614174000", false},  // Wrong separator
		{"123e4567-e89b-12d3-a456-426614174000", true},   // Valid hex UUID
		{"123e4567-e89b-12d3-a456-42661417400G", true},   // Valid with uppercase
		{"123e4567-e89b-12d3-a456-42661417400!", false},  // Invalid character
		{"", false}, // Empty
	}

	for _, tt := range tests {
		t.Run(tt.uuid, func(t *testing.T) {
			result := uuidPattern.MatchString(tt.uuid) && len(tt.uuid) == uuidLength
			if result != tt.expected {
				t.Errorf("UUID validation for %q = %v; want %v", tt.uuid, result, tt.expected)
			}
		})
	}
}
