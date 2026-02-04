package shared

import (
	"strings"
	"testing"
	"time"

	sharedG "github.com/ParetoSecurity/agent/shared"
)

func TestTeamReportSentCheck_Metadata(t *testing.T) {
	check := &TeamReportSentCheck{}
	if check.Name() == "" {
		t.Fatal("expected Name to be non-empty")
	}
	if check.UUID() == "" {
		t.Fatal("expected UUID to be non-empty")
	}
	if check.PassedMessage() == "" {
		t.Fatal("expected PassedMessage to be non-empty")
	}
	if check.FailedMessage() == "" {
		t.Fatal("expected FailedMessage to be non-empty")
	}
	if check.RequiresRoot() {
		t.Fatal("expected RequiresRoot to be false")
	}
}

func TestTeamReportSentCheck_Run_NoSuccess(t *testing.T) {
	orig := sharedG.Config
	t.Cleanup(func() { sharedG.Config = orig })

	sharedG.Config.TeamID = "team"
	sharedG.Config.AuthToken = "token"
	sharedG.Config.LastTeamReportSuccess = 0

	check := &TeamReportSentCheck{}
	if !check.IsRunnable() {
		t.Fatal("expected check to be runnable when linked")
	}
	if err := check.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if check.Passed() {
		t.Fatal("expected check to fail when no successful report exists")
	}
	if !strings.Contains(check.Status(), "No successful report") {
		t.Fatalf("unexpected status: %s", check.Status())
	}
}

func TestTeamReportSentCheck_Run_WithinThreshold(t *testing.T) {
	orig := sharedG.Config
	t.Cleanup(func() { sharedG.Config = orig })

	sharedG.Config.TeamID = "team"
	sharedG.Config.AuthToken = "token"
	sharedG.Config.LastTeamReportSuccess = time.Now().Add(-2 * time.Hour).UnixMilli()

	check := &TeamReportSentCheck{}
	if !check.IsRunnable() {
		t.Fatal("expected check to be runnable when linked")
	}
	if err := check.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !check.Passed() {
		t.Fatal("expected check to pass when report is within threshold")
	}
	if !strings.Contains(check.Status(), "Last successful report:") {
		t.Fatalf("unexpected status: %s", check.Status())
	}
}

func TestTeamReportSentCheck_Run_BeyondThreshold(t *testing.T) {
	orig := sharedG.Config
	t.Cleanup(func() { sharedG.Config = orig })

	sharedG.Config.TeamID = "team"
	sharedG.Config.AuthToken = "token"
	sharedG.Config.LastTeamReportSuccess = time.Now().Add(-26 * time.Hour).UnixMilli()

	check := &TeamReportSentCheck{}
	if !check.IsRunnable() {
		t.Fatal("expected check to be runnable when linked")
	}
	if err := check.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if check.Passed() {
		t.Fatal("expected check to fail when report is beyond threshold")
	}
	if !strings.Contains(check.Status(), "Last successful report:") {
		t.Fatalf("unexpected status: %s", check.Status())
	}
}

func TestTeamReportSentCheck_IsRunnable_NotLinked(t *testing.T) {
	orig := sharedG.Config
	t.Cleanup(func() { sharedG.Config = orig })

	sharedG.Config.TeamID = ""
	sharedG.Config.AuthToken = ""
	sharedG.Config.LastTeamReportSuccess = time.Now().UnixMilli()

	check := &TeamReportSentCheck{}
	if check.IsRunnable() {
		t.Fatal("expected check to be not runnable when not linked")
	}
}
