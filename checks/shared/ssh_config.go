package shared

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os/exec"
	"regexp"
	"strings"

	sharedG "github.com/ParetoSecurity/agent/shared"
)

// sshProbeHost is a host that no config block matches. The ".invalid" domain never resolves.
// The random label defeats a wildcard Host block for the probe.
var sshProbeHost = func() string {
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		return "pareto-probe.invalid"
	}
	return "pareto-probe-" + hex.EncodeToString(suffix) + ".invalid"
}()

// sshSettingLine matches the "keyword value" lines of `ssh -G`. Warnings on stderr do not match.
var sshSettingLine = regexp.MustCompile(`^[a-z0-9]+ `)

// sshClientConfig reads the OpenSSH client configuration for the SSH config checks.
// Checks embed it and keep their own pass logic.
// It runs ssh without a sandbox, so a "Match exec" directive in the user's config runs its command.
type sshClientConfig struct {
	passed bool
	status string
	// LookPath replaces exec.LookPath in tests.
	LookPath func(string) (string, error)
	// RunCommand replaces the ssh invocation in tests.
	RunCommand func(string, ...string) (string, error)
}

// settings returns the client settings that ssh resolves for the probe host.
// No config block matches that host, so only global and default values apply.
// It returns an error when ssh prints no settings.
func (c *sshClientConfig) settings() (map[string]string, error) {
	output, err := c.runCommand("ssh", "-G", "-oCanonicalizeHostname=no", sshProbeHost)
	settings := sshSettings(output)
	if len(settings) == 0 {
		if err != nil {
			return nil, errors.New("could not read the SSH client configuration: " + err.Error())
		}
		return nil, errors.New("could not read the SSH client configuration")
	}
	return settings, nil
}

// sshSettings parses `ssh -G` output into a map of lowercase keyword to value.
// ssh repeats some keywords (identityfile). The last value wins here.
func sshSettings(output string) map[string]string {
	settings := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if !sshSettingLine.MatchString(line) {
			continue
		}
		key, value, _ := strings.Cut(line, " ")
		value = strings.TrimSpace(value)
		if value != "" {
			settings[key] = value
		}
	}
	return settings
}

// hasSSH reports if the system has an ssh client.
func (c *sshClientConfig) hasSSH() bool {
	_, err := c.lookPath("ssh")
	return err == nil
}

func (c *sshClientConfig) lookPath(name string) (string, error) {
	if c.LookPath != nil {
		return c.LookPath(name)
	}
	return exec.LookPath(name)
}

func (c *sshClientConfig) runCommand(name string, args ...string) (string, error) {
	if c.RunCommand != nil {
		return c.RunCommand(name, args...)
	}
	return sharedG.RunCommand(name, args...)
}

// SSHAgentForwarding makes sure that ssh does not forward the agent to every host.
// A server with a forwarded agent can use the local keys while the connection lasts.
type SSHAgentForwarding struct {
	sshClientConfig
}

// Name returns the name of the check.
func (f *SSHAgentForwarding) Name() string {
	return "SSH agent is not forwarded to all hosts"
}

// Run executes the check. Any value other than "no" ("yes", a socket path or $VAR) forwards an agent.
func (f *SSHAgentForwarding) Run() error {
	settings, err := f.settings()
	if err != nil {
		f.passed = false
		f.status = err.Error()
		return err
	}
	value, ok := settings["forwardagent"]
	if !ok {
		f.passed = false
		f.status = "ssh did not report ForwardAgent"
		return errors.New(f.status)
	}
	f.passed = value == "no"
	if f.passed {
		f.status = f.PassedMessage()
	} else {
		f.status = "ForwardAgent is set to " + value + " for all hosts"
	}
	return nil
}

// Passed returns true if the check passed.
func (f *SSHAgentForwarding) Passed() bool {
	return f.passed
}

// IsRunnable returns true if the system has an ssh client.
func (f *SSHAgentForwarding) IsRunnable() bool {
	if !f.hasSSH() {
		f.status = "ssh is not installed"
		return false
	}
	return true
}

// UUID returns the UUID of the check.
func (f *SSHAgentForwarding) UUID() string {
	return "9ff5e3d6-8d56-4c86-9767-b792a856569d"
}

// PassedMessage returns the message to return if the check passed.
func (f *SSHAgentForwarding) PassedMessage() string {
	return "SSH agent is not forwarded to all hosts"
}

// FailedMessage returns the message to return if the check failed.
func (f *SSHAgentForwarding) FailedMessage() string {
	return "SSH agent is forwarded to all hosts"
}

// RequiresRoot returns true if the check needs root access.
func (f *SSHAgentForwarding) RequiresRoot() bool {
	return false
}

// Status returns details for the current check state.
func (f *SSHAgentForwarding) Status() string {
	if f.status != "" {
		return f.status
	}
	if f.Passed() {
		return f.PassedMessage()
	}
	return f.FailedMessage()
}

// SSHHostKeyVerification makes sure that ssh verifies host keys.
// Without it, an attacker can impersonate a server and ssh shows no warning.
type SSHHostKeyVerification struct {
	sshClientConfig
}

// Name returns the name of the check.
func (h *SSHHostKeyVerification) Name() string {
	return "SSH verifies host keys"
}

// Run executes the check. ssh -G prints "false" for StrictHostKeyChecking no or off.
func (h *SSHHostKeyVerification) Run() error {
	settings, err := h.settings()
	if err != nil {
		h.passed = false
		h.status = err.Error()
		return err
	}
	value, ok := settings["stricthostkeychecking"]
	if !ok {
		h.passed = false
		h.status = "ssh did not report StrictHostKeyChecking"
		return errors.New(h.status)
	}
	h.passed = value != "false"
	if h.passed {
		h.status = h.PassedMessage()
	} else {
		h.status = "StrictHostKeyChecking is disabled for all hosts"
	}
	return nil
}

// Passed returns true if the check passed.
func (h *SSHHostKeyVerification) Passed() bool {
	return h.passed
}

// IsRunnable returns true if the system has an ssh client.
func (h *SSHHostKeyVerification) IsRunnable() bool {
	if !h.hasSSH() {
		h.status = "ssh is not installed"
		return false
	}
	return true
}

// UUID returns the UUID of the check.
func (h *SSHHostKeyVerification) UUID() string {
	return "79c330c4-5365-4070-b0bd-ffdbd32cca49"
}

// PassedMessage returns the message to return if the check passed.
func (h *SSHHostKeyVerification) PassedMessage() string {
	return "SSH verifies host keys"
}

// FailedMessage returns the message to return if the check failed.
func (h *SSHHostKeyVerification) FailedMessage() string {
	return "SSH skips host key verification"
}

// RequiresRoot returns true if the check needs root access.
func (h *SSHHostKeyVerification) RequiresRoot() bool {
	return false
}

// Status returns details for the current check state.
func (h *SSHHostKeyVerification) Status() string {
	if h.status != "" {
		return h.status
	}
	if h.Passed() {
		return h.PassedMessage()
	}
	return h.FailedMessage()
}
