//go:build windows
// +build windows

package shared

import (
	"errors"
	"syscall"

	"os/exec"
	"strings"
	"testing"

	"github.com/caarlos0/log"
)

// RunCommandMock represents a mock command with its arguments, output, and error
type RunCommandMock struct {
	Command string
	Args    []string
	Out     string
	Err     error
}

// RunCommandMocks is a slice that stores mock command outputs.
var RunCommandMocks []RunCommandMock

// RunCommand executes a command with the given name and arguments, and returns
// the combined standard output and standard error as a string. If testing is
// enabled, it returns a predefined fixture instead of executing the command.
func RunCommand(name string, arg ...string) (string, error) {

	// Check if testing is enabled and enable harnessing
	if testing.Testing() {
		for _, mock := range RunCommandMocks {
			isCmd := mock.Command == name
			isArg := strings.TrimSpace(strings.Join(mock.Args, " ")) == strings.TrimSpace(strings.Join(arg, " "))
			if isCmd && isArg {
				return mock.Out, mock.Err
			}
		}
		return "", errors.New("RunCommand fixture not found: " + name + " " + strings.TrimSpace(strings.Join(arg, " ")))
	}

	cmd := exec.Command(name, arg...)

	// Hide the window and set the creation flags to prevent the command from
	// creating a new console window
	// This is important for running commands in the background without
	// displaying a console window
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	output, err := cmd.CombinedOutput()
	log.WithField("cmd", string(name+" "+strings.TrimSpace(strings.Join(arg, " ")))).WithError(err).Debug(string(output))
	return string(output), err
}
