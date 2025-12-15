package cmd

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/stretchr/testify/assert"
)

func Test_CheckCMD(t *testing.T) {
	expected := "check       Run checks"
	b := bytes.NewBufferString("")
	checkCmd.SetOut(b)
	checkCmd.Execute()
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), expected) {
		t.Fatalf("expected \"%s\" got \"%s\"", expected, string(out))
	}
}

func Test_runCheckCommand_Success(t *testing.T) {
	var (
		runnerCheckCalled  bool
		reportToTeamCalled bool
		logWarnCalled      bool
	)

	config := &CheckConfig{
		IsRoot:          func() bool { return false },
		IsLinked:        func() bool { return true },
		AllChecksPassed: func() bool { return true },
		GetFailedChecks: func() []shared.LastState { return []shared.LastState{} },
		ReportToTeam: func(bool) error {
			reportToTeamCalled = true
			return nil
		},
		RunnerCheck: func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {
			runnerCheckCalled = true
		},
		LogFatal: func(msg string) {
			t.Errorf("LogFatal should not be called, got: %s", msg)
		},
		LogWarn: func(msg string) {
			logWarnCalled = true
		},
		LogErrorf: func(format string, args ...interface{}) {
			t.Errorf("LogErrorf should not be called")
		},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.True(t, runnerCheckCalled)
	assert.True(t, reportToTeamCalled)
	assert.False(t, logWarnCalled) // Not root, so no warning
}

func Test_runCheckCommand_IsRoot(t *testing.T) {
	var logWarnCalled bool

	config := &CheckConfig{
		IsRoot:          func() bool { return true },
		IsLinked:        func() bool { return false },
		AllChecksPassed: func() bool { return true },
		GetFailedChecks: func() []shared.LastState { return []shared.LastState{} },
		ReportToTeam:    func(bool) error { return nil },
		RunnerCheck:     func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {},
		LogFatal: func(msg string) {
			t.Errorf("LogFatal should not be called")
		},
		LogWarn: func(msg string) {
			logWarnCalled = true
			assert.Equal(t, "Please run this command as a normal user, as it won't report all checks correctly.", msg)
		},
		LogErrorf: func(format string, args ...interface{}) {},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.True(t, logWarnCalled)
}

func Test_runCheckCommand_ChecksFailed_Verbose(t *testing.T) {
	var logFatalCalled bool
	var logErrorfCalled bool

	// Set verbose to true for this test
	oldVerbose := verbose
	verbose = true
	defer func() { verbose = oldVerbose }()

	config := &CheckConfig{
		IsRoot:          func() bool { return false },
		IsLinked:        func() bool { return false },
		AllChecksPassed: func() bool { return false },
		GetFailedChecks: func() []shared.LastState {
			return []shared.LastState{
				{Name: "Test Check", UUID: "test-uuid"},
			}
		},
		ReportToTeam: func(bool) error { return nil },
		RunnerCheck:  func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {},
		LogFatal: func(msg string) {
			logFatalCalled = true
			assert.Equal(t, "You can use `paretosecurity check --verbose` to get a detailed report.", msg)
		},
		LogWarn: func(msg string) {},
		LogErrorf: func(format string, args ...interface{}) {
			logErrorfCalled = true
			assert.Equal(t, "Failed check: %s (UUID: %s)", format)
		},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.False(t, logFatalCalled) // verbose is true, so LogFatal should not be called
	assert.True(t, logErrorfCalled)
}

func Test_runCheckCommand_ChecksFailed_NotVerbose(t *testing.T) {
	var logFatalCalled bool
	var logErrorfCalled bool

	// Ensure verbose is false for this test
	oldVerbose := verbose
	verbose = false
	defer func() { verbose = oldVerbose }()

	config := &CheckConfig{
		IsRoot:          func() bool { return false },
		IsLinked:        func() bool { return false },
		AllChecksPassed: func() bool { return false },
		GetFailedChecks: func() []shared.LastState {
			return []shared.LastState{
				{Name: "Test Check", UUID: "test-uuid"},
			}
		},
		ReportToTeam: func(bool) error { return nil },
		RunnerCheck:  func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {},
		LogFatal: func(msg string) {
			logFatalCalled = true
			assert.Equal(t, "You can use `paretosecurity check --verbose` to get a detailed report.", msg)
		},
		LogWarn: func(msg string) {},
		LogErrorf: func(format string, args ...interface{}) {
			logErrorfCalled = true
		},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.True(t, logFatalCalled)   // suggests using --verbose when not in verbose mode
	assert.False(t, logErrorfCalled) // failed checks are not logged without verbose
}

func Test_runCheckCommand_NotLinked(t *testing.T) {
	var reportToTeamCalled bool

	config := &CheckConfig{
		IsRoot:          func() bool { return false },
		IsLinked:        func() bool { return false },
		AllChecksPassed: func() bool { return true },
		GetFailedChecks: func() []shared.LastState { return []shared.LastState{} },
		ReportToTeam: func(bool) error {
			reportToTeamCalled = true
			return nil
		},
		RunnerCheck: func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {},
		LogFatal: func(msg string) {
			t.Errorf("LogFatal should not be called")
		},
		LogWarn:   func(msg string) {},
		LogErrorf: func(format string, args ...interface{}) {},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.False(t, reportToTeamCalled) // Not linked, so no report
}

func Test_runCheckCommand_Timeout(t *testing.T) {
	// Override the global timeout for this test
	originalTimeout := shared.CheckTimeout
	shared.CheckTimeout = 100 * time.Millisecond
	defer func() { shared.CheckTimeout = originalTimeout }()

	var logFatalCalled bool

	config := &CheckConfig{
		IsRoot:          func() bool { return false },
		IsLinked:        func() bool { return false },
		AllChecksPassed: func() bool { return true },
		GetFailedChecks: func() []shared.LastState { return []shared.LastState{} },
		ReportToTeam:    func(bool) error { return nil },
		RunnerCheck: func(ctx context.Context, claims []claims.Claim, skipUUIDs []string, onlyUUID string) {
			// Simulate long running check that respects context cancellation
			select {
			case <-time.After(1 * time.Second): // Longer than the 100ms timeout
				// This should never be reached due to context timeout
			case <-ctx.Done():
				// Context was cancelled due to timeout, return immediately
				return
			}
		},
		LogFatal: func(msg string) {
			logFatalCalled = true
			assert.Equal(t, "Check run timed out", msg)
		},
		LogWarn:   func(msg string) {},
		LogErrorf: func(format string, args ...interface{}) {},
		LogWithError: func(err error) *log.Entry {
			return log.WithError(err)
		},
	}

	runCheckCommand(config, []string{}, "")

	assert.True(t, logFatalCalled)
}

func Test_DefaultCheckConfig(t *testing.T) {
	config := DefaultCheckConfig()

	assert.NotNil(t, config.IsRoot)
	assert.NotNil(t, config.IsLinked)
	assert.NotNil(t, config.AllChecksPassed)
	assert.NotNil(t, config.GetFailedChecks)
	assert.NotNil(t, config.ReportToTeam)
	assert.NotNil(t, config.RunnerCheck)
	assert.NotNil(t, config.LogFatal)
	assert.NotNil(t, config.LogWarn)
	assert.NotNil(t, config.LogErrorf)
	assert.NotNil(t, config.LogWithError)
}
