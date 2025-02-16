package team

import (
	"sync/atomic"
	"testing"

	"github.com/ParetoSecurity/pareto-core/check"
	"github.com/ParetoSecurity/pareto-core/claims"
)

// DummyCheck implements check.Check for testing.
type dummyCheck struct {
	name      string
	runnable  bool
	runErr    error
	passedVal bool
	statusMsg string
	uuid      string

	runCalled int32
}

func (d *dummyCheck) IsRunnable() bool { return d.runnable }
func (d *dummyCheck) Name() string     { return d.name }
func (d *dummyCheck) Run() error {
	atomic.StoreInt32(&d.runCalled, 1)
	return d.runErr
}
func (d *dummyCheck) Passed() bool          { return d.passedVal }
func (d *dummyCheck) Status() string        { return d.statusMsg }
func (d *dummyCheck) UUID() string          { return d.uuid }
func (d *dummyCheck) PassedMessage() string { return "passed" }
func (d *dummyCheck) FailedMessage() string { return "failed" }
func (d *dummyCheck) RequiresRoot() bool    { return false }

func TestNowReportEmpty(t *testing.T) {
	// Test with no claims.
	report := NowReport([]claims.Claim{})
	if report.PassedCount != 0 || report.FailedCount != 0 || report.DisabledCount != 0 {
		t.Errorf("Expected all counts to be 0, got: pass=%d, fail=%d, disabled=%d",
			report.PassedCount, report.FailedCount, report.DisabledCount)
	}
	if len(report.State) != 0 {
		t.Errorf("Expected State to be empty, got length %d", len(report.State))
	}
}

func TestNowReportCounts(t *testing.T) {
	// Prepare one claim with three checks:
	// c1 -> runnable and passed, c2 -> runnable but failed, c3 -> disabled.
	c1 := dummyCheck{
		name:      "c1",
		runnable:  true,
		runErr:    nil,
		passedVal: true,
		statusMsg: "pass",
		uuid:      "check1",
		runCalled: 0,
	}
	c2 := dummyCheck{
		name:      "c2",
		runnable:  true,
		runErr:    nil,
		passedVal: false,
		statusMsg: "fail",
		uuid:      "check2",
		runCalled: 0,
	}
	c3 := dummyCheck{
		name:      "c3",
		runnable:  false,
		runErr:    nil,
		passedVal: false,
		statusMsg: "off",
		uuid:      "check3",
		runCalled: 0,
	}

	dummyClaims := []claims.Claim{
		{Title: "Test Case", Checks: []check.Check{
			check.Register(&c1),
			check.Register(&c2),
			check.Register(&c3),
		}},
	}
	report := NowReport(dummyClaims)

	if report.PassedCount != 1 {
		t.Errorf("Expected PassedCount = 1, got %d", report.PassedCount)
	}
	if report.FailedCount != 1 {
		t.Errorf("Expected FailedCount = 1, got %d", report.FailedCount)
	}
	if report.DisabledCount != 1 {
		t.Errorf("Expected DisabledCount = 1, got %d", report.DisabledCount)
	}

	if state, ok := report.State["check1"]; !ok || state != "pass" {
		t.Errorf("Expected check1 state = pass, got %s", state)
	}
	if state, ok := report.State["check2"]; !ok || state != "fail" {
		t.Errorf("Expected check2 state = fail, got %s", state)
	}
	if state, ok := report.State["check3"]; !ok || state != "off" {
		t.Errorf("Expected check3 state = off, got %s", state)
	}

	// The SignificantChange should be a valid hex string of length 64.
	if len(report.SignificantChange) != 64 {
		t.Errorf("Expected SignificantChange to have length 64, got %d", len(report.SignificantChange))
	}

}
