package team

import (
	"os"
	"sync/atomic"
	"testing"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/h2non/gock"
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

	shared.StatePath = t.TempDir() + "/test_TestNowReportCounts.toml"
	defer func() {
		os.Remove(shared.StatePath)
	}()

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
	shared.UpdateLastState(shared.LastState{
		UUID:    c1.UUID(),
		Name:    c1.Name(),
		Passed:  c1.Passed(),
		Details: c1.Status(),
	})
	c2 := dummyCheck{
		name:      "c2",
		runnable:  true,
		runErr:    nil,
		passedVal: false,
		statusMsg: "fail",
		uuid:      "check2",
		runCalled: 0,
	}
	shared.UpdateLastState(shared.LastState{
		UUID:    c2.UUID(),
		Name:    c2.Name(),
		Passed:  c2.Passed(),
		Details: c2.Status(),
	})
	c3 := dummyCheck{
		name:      "c3",
		runnable:  false,
		runErr:    nil,
		passedVal: false,
		statusMsg: "off",
		uuid:      "check3",
		runCalled: 0,
	}
	shared.UpdateLastState(shared.LastState{
		UUID:    c3.UUID(),
		Name:    c3.Name(),
		Passed:  c3.Passed(),
		Details: c3.Status(),
	})

	dummyClaims := []claims.Claim{
		{Title: "Test Case", Checks: []check.Check{
			&c1,
			&c2,
			&c3,
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

func TestReportToTeam(t *testing.T) {
	defer gock.Off()

	shared.Config.TeamID = "testTeam"
	shared.Config.AuthToken = "testToken"

	// Test initial report (PUT request).
	gock.New(reportURL).
		Put("/api/v1/team/" + shared.Config.TeamID + "/device").
		Reply(200).
		BodyString(`{"status": "ok"}`)

	err := ReportToTeam(true)
	if err != nil {
		t.Fatalf("ReportToTeam (initial) failed: %v", err)
	}

	if !gock.IsDone() {
		t.Errorf("pending mocks: %v", gock.Pending())
	}

	gock.Clean()

	// Test subsequent report (PATCH request).
	gock.New(reportURL).
		Patch("/api/v1/team/" + shared.Config.TeamID + "/device").
		Reply(200).
		BodyString(`{"status": "ok"}`)

	err = ReportToTeam(false)
	if err != nil {
		t.Fatalf("ReportToTeam (subsequent) failed: %v", err)
	}

	if !gock.IsDone() {
		t.Errorf("pending mocks: %v", gock.Pending())
	}

	gock.Clean()

	// Test API error handling.
	gock.New(reportURL).
		Patch("/api/v1/team/" + shared.Config.TeamID + "/device").
		Reply(500).
		BodyString(`{"error": "server error"}`)

	err = ReportToTeam(false)
	if err == nil {
		t.Fatalf("ReportToTeam (API error) should have failed, but didn't")
	}

	if !gock.IsDone() {
		t.Errorf("pending mocks: %v", gock.Pending())
	}
	gock.Clean()

	// Test request failure
	gock.New(reportURL).
		Patch("/api/v1/team/" + shared.Config.TeamID + "/device").
		ReplyError(err)

	err = ReportToTeam(false)
	if err == nil {
		t.Fatalf("ReportToTeam (Request  error) should have failed, but didn't")
	}

	if !gock.IsDone() {
		t.Errorf("pending mocks: %v", gock.Pending())
	}
	gock.Clean()
}
