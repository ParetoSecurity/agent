package checks

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sudoDefaults = `Matching Defaults entries for alice on host:
    env_reset, mail_badpass, secure_path=/usr/local/sbin\:/usr/local/bin\:/usr/sbin\:/usr/bin\:/sbin\:/bin,
    use_pty

Runas and Command-specific defaults for alice:
    Defaults!/usr/sbin/visudo env_keep+="SUDO_EDITOR EDITOR VISUAL"

`

// Stock Debian sudoers: `%sudo ALL=(ALL:ALL) ALL`.
const sudoStockAdmin = sudoDefaults + `User alice may run the following commands on host:
    (ALL : ALL) ALL
`

const sudoNotInSudoers = "User bob is not allowed to run sudo on host."

func sudoPrivileges(lines ...string) string {
	return sudoDefaults + "User alice may run the following commands on host:\n    " + strings.Join(lines, "\n    ")
}

func sudoFixture(output string, err error) *SudoRequiresPassword {
	return &SudoRequiresPassword{
		callerUID:  1000,
		LookPath:   func(string) (string, error) { return "/usr/bin/sudo", nil },
		LookupUser: func(int) (string, error) { return "alice", nil },
		RunCommand: func(string, ...string) (string, error) { return output, err },
	}
}

func TestSudoRequiresPassword_Metadata(t *testing.T) {
	check := SudoRequiresPassword{}

	assert.Equal(t, "sudo requires a password", check.Name())
	assert.Equal(t, "f358454b-5d9b-4c6b-b47f-13b55b17ad0e", check.UUID())
	assert.Equal(t, "sudo requires a password", check.PassedMessage())
	assert.Equal(t, "sudo runs without a password", check.FailedMessage())
	assert.True(t, check.RequiresRoot())
}

func TestEvaluateSudoListing_StockAdminPasses(t *testing.T) {
	passes, ok := evaluateSudoListing(sudoStockAdmin)
	assert.True(t, ok)
	assert.True(t, passes)
}

func TestEvaluateSudoListing_UserNotInSudoersPasses(t *testing.T) {
	for _, listing := range []string{sudoNotInSudoers, sudoDefaults + sudoNotInSudoers} {
		passes, ok := evaluateSudoListing(listing)
		assert.True(t, ok, listing)
		assert.True(t, passes, listing)
	}
}

func TestEvaluateSudoListing_NoPasswordForAllCommandsFails(t *testing.T) {
	listings := []string{
		sudoPrivileges("(ALL) NOPASSWD: ALL"),
		sudoPrivileges("(ALL : ALL) NOPASSWD: ALL"),
		sudoPrivileges("(ALL) ALL", "(root) NOPASSWD: ALL"),
		sudoPrivileges("(ALL) SETENV: NOPASSWD: ALL"),
		sudoPrivileges("(ALL) NOPASSWD: SETENV: ALL"),
		sudoPrivileges("(root) CWD=/tmp NOPASSWD: ALL"),
		// Tags carry over to the following commands of the same line
		sudoPrivileges("(root) NOPASSWD: /usr/bin/systemctl, ALL"),
	}
	for _, listing := range listings {
		passes, ok := evaluateSudoListing(listing)
		assert.True(t, ok, listing)
		assert.False(t, passes, listing)
	}
}

func TestEvaluateSudoListing_NoPasswordForSingleCommandsPasses(t *testing.T) {
	listings := []string{
		sudoPrivileges("(ALL) ALL", "(root) NOPASSWD: /usr/bin/systemctl"),
		sudoPrivileges("(root) NOPASSWD: /usr/bin/systemctl, PASSWD: ALL"),
		// Tags do not carry over to the next privilege line
		sudoPrivileges("(root) NOPASSWD: /usr/bin/systemctl", "(ALL) ALL"),
	}
	for _, listing := range listings {
		passes, ok := evaluateSudoListing(listing)
		assert.True(t, ok, listing)
		assert.True(t, passes, listing)
	}
}

func TestEvaluateSudoListing_NotAuthenticateDefaultFails(t *testing.T) {
	global := strings.Replace(sudoStockAdmin, "use_pty", "use_pty, !authenticate", 1)
	passes, ok := evaluateSudoListing(global)
	assert.True(t, ok)
	assert.False(t, passes)

	for _, specific := range []string{"Defaults>root !authenticate", "Defaults!ALL !authenticate"} {
		listing := strings.Replace(sudoStockAdmin, `Defaults!/usr/sbin/visudo env_keep+="SUDO_EDITOR EDITOR VISUAL"`, specific, 1)
		passes, ok := evaluateSudoListing(listing)
		assert.True(t, ok, specific)
		assert.False(t, passes, specific)
	}
}

func TestEvaluateSudoListing_NotAuthenticateForSingleCommandPasses(t *testing.T) {
	listing := strings.Replace(sudoStockAdmin, `Defaults!/usr/sbin/visudo env_keep+="SUDO_EDITOR EDITOR VISUAL"`, "Defaults!/usr/bin/systemctl !authenticate", 1)
	passes, ok := evaluateSudoListing(listing)
	assert.True(t, ok)
	assert.True(t, passes)
}

func TestEvaluateSudoListing_WarningsAreIgnored(t *testing.T) {
	passes, ok := evaluateSudoListing("sudo: unable to resolve host box: Name or service not known\n" + sudoPrivileges("(ALL) NOPASSWD: ALL"))
	assert.True(t, ok)
	assert.False(t, passes)
}

func TestEvaluateSudoListing_UnknownOutputIsNotAListing(t *testing.T) {
	outputs := []string{
		"",
		"No output",
		"sudo: unknown user bob",
		"sudo: a password is required",
		"sudo: command not found",
	}
	for _, output := range outputs {
		_, ok := evaluateSudoListing(output)
		assert.False(t, ok, output)
	}
}

func TestSudoRequiresPassword_Run(t *testing.T) {
	passing := sudoFixture(sudoStockAdmin, nil)
	require.NoError(t, passing.Run())
	assert.True(t, passing.Passed())
	assert.Equal(t, "sudo requires a password", passing.Status())

	outside := sudoFixture(sudoNotInSudoers, errors.New("exit status 1"))
	require.NoError(t, outside.Run())
	assert.True(t, outside.Passed())

	failing := sudoFixture(sudoPrivileges("(ALL) NOPASSWD: ALL"), nil)
	require.NoError(t, failing.Run())
	assert.False(t, failing.Passed())
	assert.Equal(t, "alice can run any command with sudo without a password", failing.Status())
}

func TestSudoRequiresPassword_RunListsTheCallingUser(t *testing.T) {
	var args []string
	check := sudoFixture("", nil)
	check.LookupUser = func(uid int) (string, error) {
		assert.Equal(t, 1000, uid)
		return "alice", nil
	}
	check.RunCommand = func(name string, a ...string) (string, error) {
		args = append([]string{name}, a...)
		return sudoStockAdmin, nil
	}

	require.NoError(t, check.Run())
	assert.Equal(t, []string{"sudo", "-n", "-l", "-U", "alice"}, args)
}

func TestSudoRequiresPassword_RunErrors(t *testing.T) {
	unknownOutput := sudoFixture("sudo: unknown user alice", errors.New("exit status 1"))
	assert.Error(t, unknownOutput.Run())
	assert.False(t, unknownOutput.Passed())
	assert.Equal(t, "sudo listing could not be evaluated", unknownOutput.Status())

	noCaller := sudoFixture(sudoStockAdmin, nil)
	noCaller.SetCallerUID(-1)
	assert.Error(t, noCaller.Run())
	assert.Equal(t, "could not resolve the calling user", noCaller.Status())

	root := sudoFixture(sudoStockAdmin, nil)
	root.SetCallerUID(0)
	assert.Error(t, root.Run())

	unknownUser := sudoFixture(sudoStockAdmin, nil)
	unknownUser.LookupUser = func(int) (string, error) { return "", errors.New("unknown userid") }
	assert.Error(t, unknownUser.Run())
	assert.Equal(t, "could not resolve the calling user", unknownUser.Status())

	optionLikeUser := sudoFixture(sudoStockAdmin, nil)
	optionLikeUser.LookupUser = func(int) (string, error) { return "-h", nil }
	assert.Error(t, optionLikeUser.Run())
}

func TestSudoRequiresPassword_IsRunnable(t *testing.T) {
	installed := sudoFixture(sudoStockAdmin, nil)
	assert.True(t, installed.IsRunnable())

	missing := SudoRequiresPassword{LookPath: func(string) (string, error) { return "", errors.New("not found") }}
	assert.False(t, missing.IsRunnable())
	assert.Equal(t, "sudo is not installed", missing.Status())
}
