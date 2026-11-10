package shared

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/caarlos0/log"
)

// CommandTimeout bounds every external command a check runs, so one command
// that never returns (a daemon that does not answer, a prompt that waits for
// input) cannot hang the whole check run. Can be overridden for testing.
var CommandTimeout = 30 * time.Second

// commandWaitDelay is how long to wait for the output pipes to close after a
// timed out command is killed. Children it started (as with `sh -c`) may hold
// them open after it has gone.
var commandWaitDelay = 5 * time.Second

// Command returns an exec.Cmd that is killed once CommandTimeout has passed.
// Call done with the error the command returned once it has finished: it
// releases the timer and turns a timeout into an error that says so.
func Command(name string, arg ...string) (cmd *exec.Cmd, done func(error) error) {
	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	cmd = exec.CommandContext(ctx, name, arg...)
	cmd.WaitDelay = commandWaitDelay
	return cmd, func(err error) error {
		defer cancel()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			line := strings.TrimSpace(name + " " + strings.Join(arg, " "))
			log.WithField("cmd", line).Warnf("command timed out after %s", CommandTimeout)
			return fmt.Errorf("%s timed out after %s: %w", line, CommandTimeout, context.DeadlineExceeded)
		}
		return err
	}
}
