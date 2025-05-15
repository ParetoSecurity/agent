package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
)

// captureOutput redirects stdout and returns what was printed.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		panic(err)
	}
	os.Stdout = old // restore stdout
	return buf.String()
}

// DummyCheck implements check.Check for testing.
type DummyCheck struct {
	name      string
	runnable  bool
	runErr    error
	passedVal bool
	statusMsg string
	uuid      string

	runCalled int32
}

func (d *DummyCheck) IsRunnable() bool { return d.runnable }
func (d *DummyCheck) Name() string     { return d.name }
func (d *DummyCheck) Run() error {
	atomic.StoreInt32(&d.runCalled, 1)
	return d.runErr
}
func (d *DummyCheck) Passed() bool          { return d.passedVal }
func (d *DummyCheck) Status() string        { return d.statusMsg }
func (d *DummyCheck) UUID() string          { return d.uuid }
func (d *DummyCheck) PassedMessage() string { return "passed" }
func (d *DummyCheck) FailedMessage() string { return "failed" }
func (d *DummyCheck) RequiresRoot() bool    { return false }

func TestCheckSuccess(t *testing.T) {

	// Create a dummy check that is runnable and passes.
	dc := &DummyCheck{
		name:      "DummyPass",
		runnable:  true,
		passedVal: true,
		statusMsg: "ok",
		uuid:      "uuid-pass",
	}
	dummyClaims := []claims.Claim{
		{Title: "Test Case", Checks: []check.Check{
			dc,
		}},
	}
	ctx := context.Background()
	Check(ctx, dummyClaims, []string{}, "")

	if atomic.LoadInt32(&dc.runCalled) != 1 {
		t.Errorf("Expected Run to be called on DummyCheck, but it wasn't")
	}
}

func TestCheckNotRunnable(t *testing.T) {

	// Create a dummy check that is not runnable.
	dc := &DummyCheck{
		name:      "DummySkip",
		runnable:  false,
		passedVal: false,
		statusMsg: "skipped",
		uuid:      "uuid-skip",
	}
	dummyClaims := []claims.Claim{
		{Title: "Test Case", Checks: []check.Check{
			dc,
		}},
	}
	ctx := context.Background()
	Check(ctx, dummyClaims, []string{}, "")
	if atomic.LoadInt32(&dc.runCalled) != 0 {
		t.Errorf("Expected Run NOT to be called on non-runnable DummyCheck, but it was")
	}
}

func TestCheckContextCanceled(t *testing.T) {

	// Create a dummy check that is runnable.
	dc := &DummyCheck{
		name:      "DummyCanceled",
		runnable:  true,
		passedVal: true,
		statusMsg: "ok",
		uuid:      "uuid-cancel",
	}
	dummyClaims := []claims.Claim{
		{Title: "Test Case", Checks: []check.Check{
			dc,
		}},
	}
	// Create a context that is already canceled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// Allow a short time for the goroutine to select context.Done.
	time.Sleep(10 * time.Millisecond)
	Check(ctx, dummyClaims, []string{}, "")

	if atomic.LoadInt32(&dc.runCalled) != 0 {
		t.Errorf("Expected Run NOT to be called when context is canceled, but it was")
	}
}

func TestPrintSchemaJSON(t *testing.T) {
	// Create a dummy check that returns known passed/failed messages.
	dc := &DummyCheck{
		name:      "DummySchema",
		runnable:  true,
		passedVal: true,
		statusMsg: "ok",
		uuid:      "uuid-schema",
	}

	// Create a claim with one check.
	testClaim := claims.Claim{
		Title:  "Test Claim",
		Checks: []check.Check{dc},
	}
	claimsTorun := []claims.Claim{testClaim}

	// Capture the output of PrintSchemaJSON.
	output := captureOutput(func() {
		PrintSchemaJSON(claimsTorun)
	})

	// Build expected schema map.
	expectedSchema := map[string]map[string][]string{
		"Test Claim": {
			"uuid-schema": {"passed", "failed"},
		},
	}
	expectedBytes, err := json.MarshalIndent(expectedSchema, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal expected schema: %v", err)
	}
	expectedOutput := string(expectedBytes)

	// Remove possible newline differences.
	output = strings.TrimSpace(output)
	expectedOutput = strings.TrimSpace(expectedOutput)

	if output != expectedOutput {
		t.Errorf("PrintSchemaJSON output mismatch.\nExpected:\n%s\nGot:\n%s", expectedOutput, output)
	}
}
func TestWrapStatusRoot(t *testing.T) {
	tests := []struct {
		name     string
		status   *CheckStatus
		check    check.Check
		err      error
		expected string
	}{
		{
			name: "passed with details",
			status: &CheckStatus{
				Passed:  true,
				Details: "test details",
			},
			check:    &DummyCheck{statusMsg: "should not use this"},
			err:      nil,
			expected: "[OK] test details",
		},
		{
			name: "passed without details",
			status: &CheckStatus{
				Passed:  true,
				Details: "",
			},
			check:    &DummyCheck{statusMsg: "check status"},
			err:      nil,
			expected: "[OK] check status",
		},
		{
			name: "failed with details",
			status: &CheckStatus{
				Passed:  false,
				Details: "test failure details",
			},
			check:    &DummyCheck{statusMsg: "should not use this"},
			err:      nil,
			expected: "[FAIL] test failure details",
		},
		{
			name: "failed without details",
			status: &CheckStatus{
				Passed:  false,
				Details: "",
			},
			check:    &DummyCheck{statusMsg: "check failure status"},
			err:      nil,
			expected: "[FAIL] check failure status",
		},
		{
			name:     "with error",
			status:   &CheckStatus{},
			check:    &DummyCheck{},
			err:      fmt.Errorf("test error"),
			expected: "[DISABLED] test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapStatusRoot(tt.status, tt.check, tt.err)
			// Strip ANSI color codes for comparison
			result = strings.ReplaceAll(strings.ReplaceAll(
				strings.ReplaceAll(result, "\033[31m", ""),
				"\033[32m", ""), "\033[0m", "")
			if result != tt.expected {
				t.Errorf("wrapStatusRoot() = %q, want %q", result, tt.expected)
			}
		})
	}
}
func TestWrapStatus(t *testing.T) {
	tests := []struct {
		name     string
		check    check.Check
		err      error
		expected string
	}{
		{
			name:     "passed check",
			check:    &DummyCheck{passedVal: true, statusMsg: "system is secure"},
			err:      nil,
			expected: "[OK] system is secure",
		},
		{
			name:     "failed check",
			check:    &DummyCheck{passedVal: false, statusMsg: "system is vulnerable"},
			err:      nil,
			expected: "[FAIL] system is vulnerable",
		},
		{
			name:     "with error",
			check:    &DummyCheck{},
			err:      fmt.Errorf("test error"),
			expected: "[DISABLED] test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapStatus(tt.check, tt.err)
			// Strip ANSI color codes for comparison
			result = strings.ReplaceAll(strings.ReplaceAll(
				strings.ReplaceAll(result, "\033[31m", ""),
				"\033[32m", ""), "\033[0m", "")
			if result != tt.expected {
				t.Errorf("wrapStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}
