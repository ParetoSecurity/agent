//go:build unix

package shared

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func withShortTimeouts(t *testing.T) {
	t.Helper()
	timeout, waitDelay := CommandTimeout, commandWaitDelay
	CommandTimeout, commandWaitDelay = 100*time.Millisecond, 100*time.Millisecond
	t.Cleanup(func() { CommandTimeout, commandWaitDelay = timeout, waitDelay })
}

func TestCommand_KillsCommandThatHangs(t *testing.T) {
	withShortTimeouts(t)

	// `sh -c` leaves `sleep` holding the output pipe after sh is killed,
	// as `sh -c snap list` would with a snap that never answers.
	cmd, done := Command("sh", "-c", "sleep 10; echo never")
	start := time.Now()
	_, err := cmd.CombinedOutput()
	err = done(err)

	assert.Less(t, time.Since(start), 3*time.Second)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "got %v", err)
	assert.Contains(t, err.Error(), "sh -c sleep 10; echo never timed out after 100ms")
}

func TestCommand_KeepsResultOfCommandThatFinishes(t *testing.T) {
	withShortTimeouts(t)

	cmd, done := Command("sh", "-c", "echo hello; exit 3")
	output, err := cmd.CombinedOutput()
	err = done(err)

	assert.Equal(t, "hello\n", string(output))
	var exitErr interface{ ExitCode() int }
	assert.True(t, errors.As(err, &exitErr), "got %v", err)
	assert.Equal(t, 3, exitErr.ExitCode())
}

// The root helper is a oneshot systemd service that is killed after
// TimeoutStartSec. A root check waiting on one hung command must still answer
// in time, with room left for the rest of the check.
func TestCommandTimeout_FitsRootHelperTimeout(t *testing.T) {
	unit, err := os.ReadFile("../apt/paretosecurity.service")
	assert.NoError(t, err)
	match := regexp.MustCompile(`(?m)^TimeoutStartSec=(\d+)$`).FindSubmatch(unit)
	if assert.NotNil(t, match, "TimeoutStartSec not found in paretosecurity.service") {
		seconds, _ := strconv.Atoi(string(match[1]))
		helperTimeout := time.Duration(seconds) * time.Second
		assert.Less(t, CommandTimeout+commandWaitDelay, helperTimeout-10*time.Second)
	}
}
