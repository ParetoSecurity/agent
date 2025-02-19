package shared

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestCurrentReportingDevice(t *testing.T) {
	// Ensure Config.AuthToken is cleared by default.
	Config.AuthToken = ""

	// determine expected OSVersion based on runtime
	expectedOSVersion := fmt.Sprintf("%s %s", "test-os", "test-os-version")
	if runtime.GOOS == "windows" {
		// additional formatting on windows
		expectedOSVersion = fmt.Sprintf("%s %s", expectedOSVersion, "test-os-version")
	}

	t.Run("successful device info with working SystemDevice and SystemSerial", func(t *testing.T) {

		rd := CurrentReportingDevice()

		if rd.MachineUUID != "test-uuid" {
			t.Errorf("Expected MachineUUID %q, got %q", "test-uuid", rd.MachineUUID)
		}
		if rd.MachineName != "test-hostname" {
			t.Errorf("Expected MachineName %q, got %q", "test-hostname", rd.MachineName)
		}
		if rd.Auth != "" {
			t.Errorf("Expected empty Auth, got %q", rd.Auth)
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

	t.Run("with valid auth token", func(t *testing.T) {
		// Prepare a dummy JWT-like token.
		payload := map[string]interface{}{
			"sub":    "dummy",
			"teamID": "dummy",
			"role":   "dummy",
			"iat":    1,
			"token":  "authValue",
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
		// simple dummy header and signature parts.
		dummyToken := strings.Join([]string{"header", encodedPayload, "signature"}, ".")
		Config.AuthToken = dummyToken

		rd := CurrentReportingDevice()
		if rd.Auth != "authValue" {
			t.Errorf("Expected Auth %q, got %q", "authValue", rd.Auth)
		}
	})
}
