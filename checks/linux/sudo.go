package checks

import (
	"errors"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/caarlos0/log"
)

// SudoRequiresPassword makes sure that sudo asks the calling user for a password.
// It runs in the root helper, because only root can list the privileges of another user.
// `sudo -l -U` never authenticates, so it writes no failed try to the auth log.
type SudoRequiresPassword struct {
	passed    bool
	status    string
	callerUID int
	// LookPath replaces exec.LookPath in tests.
	LookPath func(string) (string, error)
	// LookupUser replaces the uid to user name lookup in tests.
	LookupUser func(uid int) (string, error)
	// RunCommand replaces the sudo invocation in tests.
	RunCommand func(string, ...string) (string, error)
}

// Name returns the name of the check.
func (s *SudoRequiresPassword) Name() string {
	return "sudo requires a password"
}

// SetCallerUID tells the check which user asked the root helper to run it.
func (s *SudoRequiresPassword) SetCallerUID(uid int) {
	s.callerUID = uid
}

// Run executes the check.
func (s *SudoRequiresPassword) Run() error {
	s.passed = false
	if s.callerUID <= 0 {
		s.status = "could not resolve the calling user"
		return errors.New(s.status)
	}
	name, err := s.lookupUser(s.callerUID)
	if err != nil || name == "" || strings.HasPrefix(name, "-") {
		s.status = "could not resolve the calling user"
		return errors.New(s.status)
	}

	// sudo exits non-zero for a user outside sudoers. That output still says so.
	output, _ := s.runCommand("sudo", "-n", "-l", "-U", name)
	passes, ok := evaluateSudoListing(output)
	if !ok {
		log.WithField("output", output).Debug("sudo listing could not be evaluated")
		s.status = "sudo listing could not be evaluated"
		return errors.New(s.status)
	}
	s.passed = passes
	if passes {
		s.status = s.PassedMessage()
	} else {
		s.status = name + " can run any command with sudo without a password"
	}
	return nil
}

// Passed returns true if the check passed.
func (s *SudoRequiresPassword) Passed() bool {
	return s.passed
}

// IsRunnable returns true if the system has sudo.
func (s *SudoRequiresPassword) IsRunnable() bool {
	if _, err := s.lookPath("sudo"); err != nil {
		s.status = "sudo is not installed"
		return false
	}
	return true
}

// UUID returns the UUID of the check.
func (s *SudoRequiresPassword) UUID() string {
	return "f358454b-5d9b-4c6b-b47f-13b55b17ad0e"
}

// PassedMessage returns the message to return if the check passed.
func (s *SudoRequiresPassword) PassedMessage() string {
	return "sudo requires a password"
}

// FailedMessage returns the message to return if the check failed.
func (s *SudoRequiresPassword) FailedMessage() string {
	return "sudo runs without a password"
}

// RequiresRoot returns true. Only root can list the sudo privileges of another user.
func (s *SudoRequiresPassword) RequiresRoot() bool {
	return true
}

// Status returns details for the current check state.
func (s *SudoRequiresPassword) Status() string {
	if s.status != "" {
		return s.status
	}
	if s.Passed() {
		return s.PassedMessage()
	}
	return s.FailedMessage()
}

type sudoListingSection int

const (
	sudoSectionNone sudoListingSection = iota
	sudoSectionDefaults
	sudoSectionSpecificDefaults
	sudoSectionCommands
)

var (
	// sudoTagPattern matches the tags before a command, e.g. `(ALL) SETENV: NOPASSWD: ALL`.
	sudoTagPattern = regexp.MustCompile(`^(NO)?(PASSWD|SETENV|EXEC|LOG_INPUT|LOG_OUTPUT|MAIL|FOLLOW|INTERCEPT):\s*`)
	// sudoOptionPattern matches the options before the tags, e.g. `(root) CWD=/tmp NOPASSWD: ALL`.
	sudoOptionPattern = regexp.MustCompile(`^[A-Z_]+=\S+\s+`)
)

// evaluateSudoListing reads the output of `sudo -l -U <user>`.
// The first value is false when the user can run any command without a password.
// That is `NOPASSWD: ALL` or `!authenticate` for all commands.
// It is true when sudo asks for a password, or when the user may not use sudo.
// The second value is false when the output is not a sudo listing.
func evaluateSudoListing(output string) (bool, bool) {
	section := sudoSectionNone
	sawListing := false
	noPassword := false

	for _, rawLine := range strings.Split(output, "\n") {
		rawLine = strings.TrimRight(rawLine, "\r")
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "Matching Defaults entries for "):
			section = sudoSectionDefaults
			sawListing = true
			continue
		case strings.HasPrefix(line, "Runas and Command-specific defaults for "):
			section = sudoSectionSpecificDefaults
			sawListing = true
			continue
		case strings.HasPrefix(line, "User ") && strings.Contains(line, " may run the following commands on "):
			section = sudoSectionCommands
			sawListing = true
			continue
		case strings.HasPrefix(line, "User ") && strings.Contains(line, " is not allowed to run sudo on "):
			section = sudoSectionNone
			sawListing = true
			continue
		}

		// Lines outside a section are warnings from sudo. They say nothing about the rules.
		if rawLine[0] != ' ' && rawLine[0] != '\t' {
			section = sudoSectionNone
			continue
		}

		switch section {
		case sudoSectionNone:
			continue
		case sudoSectionDefaults:
			for _, entry := range strings.Split(line, ",") {
				if strings.TrimSpace(entry) == "!authenticate" {
					return false, true
				}
			}
		case sudoSectionSpecificDefaults:
			// Runas-specific defaults and defaults for every command apply to all commands.
			appliesToAll := strings.HasPrefix(line, "Defaults>") || strings.HasPrefix(line, "Defaults!ALL ")
			if appliesToAll && strings.Contains(line, "!authenticate") {
				return false, true
			}
		case sudoSectionCommands:
			rest := line
			if strings.HasPrefix(rest, "(") {
				// A new privilege line starts with its runas spec and resets the inherited tags.
				noPassword = false
				if end := strings.Index(rest, ")"); end >= 0 {
					rest = rest[end+1:]
				}
			}
			// Commands stay in order, so NOPASSWD: carries over to the commands after it.
			for _, part := range strings.Split(rest, ",") {
				if sudoCommandRunsWithoutPassword(part, &noPassword) {
					return false, true
				}
			}
		}
	}

	return sawListing, sawListing
}

// sudoCommandRunsWithoutPassword strips options and tags from one command of a privilege line.
// It updates the inherited NOPASSWD state. It returns true if the command is `ALL` without a password.
func sudoCommandRunsWithoutPassword(part string, noPassword *bool) bool {
	command := strings.TrimSpace(part)
	for {
		if match := sudoOptionPattern.FindString(command); match != "" {
			command = command[len(match):]
			continue
		}
		if match := sudoTagPattern.FindString(command); match != "" {
			switch strings.TrimSpace(match) {
			case "NOPASSWD:":
				*noPassword = true
			case "PASSWD:":
				*noPassword = false
			}
			command = command[len(match):]
			continue
		}
		break
	}
	return *noPassword && command == "ALL"
}

func (s *SudoRequiresPassword) lookPath(name string) (string, error) {
	if s.LookPath != nil {
		return s.LookPath(name)
	}
	return exec.LookPath(name)
}

func (s *SudoRequiresPassword) lookupUser(uid int) (string, error) {
	if s.LookupUser != nil {
		return s.LookupUser(uid)
	}
	entry, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return "", err
	}
	return entry.Username, nil
}

// runCommand runs sudo with a C locale, so the listing is in the English that the parser reads.
func (s *SudoRequiresPassword) runCommand(name string, args ...string) (string, error) {
	if s.RunCommand != nil {
		return s.RunCommand(name, args...)
	}
	cmd := exec.Command(name, args...)
	cmd.Env = []string{"LC_ALL=C", "PATH=/run/wrappers/bin:/usr/bin:/bin:/usr/sbin:/sbin"}
	output, err := cmd.CombinedOutput()
	log.WithField("cmd", name+" "+strings.Join(args, " ")).WithError(err).Debug(string(output))
	return string(output), err
}
