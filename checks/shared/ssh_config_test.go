package shared

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Trimmed from `ssh -G -oCanonicalizeHostname=no pareto-probe.invalid` on a default OpenSSH client,
// with a stderr warning merged in the way shared.RunCommand returns it.
const defaultSSHConfig = `Pseudo-terminal will not be allocated because stdin is not a terminal.` + "\r" + `
host pareto-probe-1f2c170c.invalid
user admin
hostname pareto-probe-1f2c170c.invalid
port 22
addressfamily any
batchmode no
canonicalizehostname false
checkhostip no
stricthostkeychecking ask
tcpkeepalive yes
updatehostkeys true
identityfile ~/.ssh/id_rsa
identityfile ~/.ssh/id_ed25519
userknownhostsfile /home/admin/.ssh/known_hosts /home/admin/.ssh/known_hosts2
addkeystoagent false
forwardagent no
connecttimeout none
escapechar ~
`

func sshConfigFixture(output string, err error) sshClientConfig {
	return sshClientConfig{
		LookPath:   func(string) (string, error) { return "/usr/bin/ssh", nil },
		RunCommand: func(string, ...string) (string, error) { return output, err },
	}
}

func withSetting(key string, value string) string {
	var lines []string
	for _, line := range strings.Split(defaultSSHConfig, "\n") {
		if strings.HasPrefix(line, key+" ") {
			line = key + " " + value
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func TestSSHConfigChecks_Metadata(t *testing.T) {
	forwarding := SSHAgentForwarding{}
	assert.Equal(t, "SSH agent is not forwarded to all hosts", forwarding.Name())
	assert.Equal(t, "9ff5e3d6-8d56-4c86-9767-b792a856569d", forwarding.UUID())
	assert.Equal(t, "SSH agent is not forwarded to all hosts", forwarding.PassedMessage())
	assert.Equal(t, "SSH agent is forwarded to all hosts", forwarding.FailedMessage())
	assert.False(t, forwarding.RequiresRoot())

	hostKeys := SSHHostKeyVerification{}
	assert.Equal(t, "SSH verifies host keys", hostKeys.Name())
	assert.Equal(t, "79c330c4-5365-4070-b0bd-ffdbd32cca49", hostKeys.UUID())
	assert.Equal(t, "SSH verifies host keys", hostKeys.PassedMessage())
	assert.Equal(t, "SSH skips host key verification", hostKeys.FailedMessage())
	assert.False(t, hostKeys.RequiresRoot())
}

func TestSSHConfigChecks_RunPassesWithDefaults(t *testing.T) {
	forwarding := SSHAgentForwarding{sshConfigFixture(defaultSSHConfig, nil)}
	require.NoError(t, forwarding.Run())
	assert.True(t, forwarding.Passed())
	assert.Equal(t, "SSH agent is not forwarded to all hosts", forwarding.Status())

	hostKeys := SSHHostKeyVerification{sshConfigFixture(defaultSSHConfig, nil)}
	require.NoError(t, hostKeys.Run())
	assert.True(t, hostKeys.Passed())
	assert.Equal(t, "SSH verifies host keys", hostKeys.Status())
}

func TestSSHAgentForwarding_RunFailsWhenForwarded(t *testing.T) {
	for _, value := range []string{"yes", "/run/user/1000/agent.sock", "$SSH_AUTH_SOCK"} {
		check := SSHAgentForwarding{sshConfigFixture(withSetting("forwardagent", value), nil)}
		require.NoError(t, check.Run(), value)
		assert.False(t, check.Passed(), value)
		assert.Equal(t, "ForwardAgent is set to "+value+" for all hosts", check.Status(), value)
	}
}

func TestSSHHostKeyVerification_Run(t *testing.T) {
	check := SSHHostKeyVerification{sshConfigFixture(withSetting("stricthostkeychecking", "false"), nil)}
	require.NoError(t, check.Run())
	assert.False(t, check.Passed())
	assert.Equal(t, "StrictHostKeyChecking is disabled for all hosts", check.Status())

	for _, value := range []string{"true", "ask", "accept-new"} {
		check := SSHHostKeyVerification{sshConfigFixture(withSetting("stricthostkeychecking", value), nil)}
		require.NoError(t, check.Run(), value)
		assert.True(t, check.Passed(), value)
	}
}

func TestSSHConfigChecks_SimilarKeysAreIgnored(t *testing.T) {
	output := defaultSSHConfig + "xforwardagent yes\nforwardagentx yes\nxstricthostkeychecking false\n"

	forwarding := SSHAgentForwarding{sshConfigFixture(output, nil)}
	require.NoError(t, forwarding.Run())
	assert.True(t, forwarding.Passed())

	hostKeys := SSHHostKeyVerification{sshConfigFixture(output, nil)}
	require.NoError(t, hostKeys.Run())
	assert.True(t, hostKeys.Passed())
}

func TestSSHConfigChecks_CRLFLineEndings(t *testing.T) {
	output := strings.ReplaceAll(withSetting("forwardagent", "yes"), "\n", "\r\n")

	forwarding := SSHAgentForwarding{sshConfigFixture(output, nil)}
	require.NoError(t, forwarding.Run())
	assert.False(t, forwarding.Passed())
	assert.Equal(t, "ForwardAgent is set to yes for all hosts", forwarding.Status())
}

func TestSSHConfigChecks_ErrorWhenSSHPrintsNoSettings(t *testing.T) {
	outputs := map[string]error{
		"":                       nil,
		"ssh: command not found": errors.New("exit status 127"),
		"/home/admin/.ssh/config: line 3: Bad configuration option: forwardagnt\r\n": errors.New("exit status 255"),
	}
	for output, err := range outputs {
		forwarding := SSHAgentForwarding{sshConfigFixture(output, err)}
		assert.Error(t, forwarding.Run(), output)
		assert.False(t, forwarding.Passed(), output)
		assert.Contains(t, forwarding.Status(), "could not read the SSH client configuration", output)

		hostKeys := SSHHostKeyVerification{sshConfigFixture(output, err)}
		assert.Error(t, hostKeys.Run(), output)
		assert.False(t, hostKeys.Passed(), output)
	}
}

func TestSSHConfigChecks_ErrorWhenSettingIsMissing(t *testing.T) {
	output := "host pareto-probe.invalid\nport 22\n"

	forwarding := SSHAgentForwarding{sshConfigFixture(output, nil)}
	assert.Error(t, forwarding.Run())
	assert.Equal(t, "ssh did not report ForwardAgent", forwarding.Status())

	hostKeys := SSHHostKeyVerification{sshConfigFixture(output, nil)}
	assert.Error(t, hostKeys.Run())
	assert.Equal(t, "ssh did not report StrictHostKeyChecking", hostKeys.Status())
}

func TestSSHConfigChecks_IsRunnable(t *testing.T) {
	installed := SSHAgentForwarding{sshConfigFixture(defaultSSHConfig, nil)}
	assert.True(t, installed.IsRunnable())

	missing := SSHHostKeyVerification{sshClientConfig{
		LookPath: func(string) (string, error) { return "", errors.New("not found") },
	}}
	assert.False(t, missing.IsRunnable())
	assert.Equal(t, "ssh is not installed", missing.Status())
}

func TestSSHConfigChecks_ProbeHost(t *testing.T) {
	var args []string
	check := SSHAgentForwarding{sshClientConfig{
		RunCommand: func(name string, a ...string) (string, error) {
			args = append([]string{name}, a...)
			return defaultSSHConfig, nil
		},
	}}
	require.NoError(t, check.Run())

	require.Len(t, args, 4)
	assert.Equal(t, []string{"ssh", "-G", "-oCanonicalizeHostname=no"}, args[:3])
	assert.True(t, strings.HasPrefix(args[3], "pareto-probe-"), args[3])
	assert.True(t, strings.HasSuffix(args[3], ".invalid"), args[3])
}
