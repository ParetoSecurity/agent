package shared

import (
	"runtime"
	"testing"
)

func TestCurrentReportingDevice(t *testing.T) {
	// Ensure Config.AuthToken is cleared by default.
	Config.AuthToken = ""

	// determine expected OSVersion based on runtime
	var expectedOSVersion string
	switch runtime.GOOS {
	case "darwin":
		// macOS version format: ^(\d+\.)?(\d+\.)?(\*|\d+)
		expectedOSVersion = "0.0.0" // FormatMacOSVersion returns this for non-numeric versions
	case "windows":
		// Windows preserves spaces
		expectedOSVersion = "test-os test-os-version test-os-version"
	default:
		// Linux preserves spaces
		expectedOSVersion = "test-os test-os-version"
	}

	t.Run("successful device info with working SystemDevice and SystemSerial", func(t *testing.T) {

		rd := CurrentReportingDevice()

		if rd.MachineUUID != "12345678-1234-1234-1234-123456789012" {
			t.Errorf("Expected MachineUUID %q, got %q", "12345678-1234-1234-1234-123456789012", rd.MachineUUID)
		}
		if rd.MachineName != "test-hostname" {
			t.Errorf("Expected MachineName %q, got %q", "test-hostname", rd.MachineName)
		}
		if rd.OSVersion != expectedOSVersion {
			t.Errorf("Expected OSVersion %q, got %q", expectedOSVersion, rd.OSVersion)
		}
		if rd.ModelName != "Unknown" {
			t.Errorf("Expected ModelName %q, got %q", "Unknown", rd.ModelName)
		}
		if rd.ModelSerial != "Unknown" {
			t.Errorf("Expected ModelSerial %q, got %q", "Unknown", rd.ModelSerial)
		}
	})

	t.Run("SystemDevice error returns Unknown model name", func(t *testing.T) {

		rd := CurrentReportingDevice()

		if rd.ModelName != "Unknown" {
			t.Errorf("Expected ModelName to be \"Unknown\" on error, got %q", rd.ModelName)
		}
		if rd.ModelSerial != "Unknown" {
			t.Errorf("Expected ModelSerial %q, got %q", "Unknown", rd.ModelSerial)
		}
	})

	t.Run("SystemSerial error returns Unknown serial", func(t *testing.T) {

		rd := CurrentReportingDevice()

		if rd.ModelSerial != "Unknown" {
			t.Errorf("Expected ModelSerial to be \"Unknown\" on error, got %q", rd.ModelSerial)
		}
		if rd.ModelName != "Unknown" {
			t.Errorf("Expected ModelName %q, got %q", "Unknown", rd.ModelName)
		}
	})
}
