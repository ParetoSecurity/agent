package shared

import (
	"testing"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, 世界!", "Hello__!"},
		{"123 ABC abc", "123ABCabc"},
		{"Special chars: @#$%^&*()", "Specialchars"},
		{"Mixed: 你好, 世界! 123", "Mixed____!123"},
		{"borxed\r\n", "borxed"},
		{"To Be Filled By O.E.M.", "ToBeFilledByO.E.M."},
		{"Valid-Serial_123", "Valid-Serial_123"},
		{"Serial.with!quotes'\"", "Serial.with!quotes'\""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := Sanitize(test.input)
			if result != test.expected {
				t.Errorf("Sanitize(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}

func TestSanitizeWithSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ubuntu 22.04 LTS", "Ubuntu 22.04 LTS"},
		{"macOS 14.0 Sonoma", "macOS 14.0 Sonoma"},
		{"Windows 11 Pro", "Windows 11 Pro"},
		{"CentOS Linux 8 (Core)", "CentOS Linux 8 Core"},
		{"Debian GNU/Linux 12", "Debian GNULinux 12"},
		{"OS\r\nVersion\t2023", "OSVersion2023"},
		{"Arch Linux 世界", "Arch Linux __"},
		{"Red Hat Enterprise Linux 9.2", "Red Hat Enterprise Linux 9.2"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := SanitizeWithSpaces(test.input)
			if result != test.expected {
				t.Errorf("SanitizeWithSpaces(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}
