package checks

import (
	"strings"

	"github.com/caarlos0/log"

	"github.com/ParetoSecurity/agent/shared"
)

type SSHConfigCheck struct {
	passed bool
	status string
}

func (s *SSHConfigCheck) Name() string {
	return "SSH Server Configuration is Secure"
}

func (s *SSHConfigCheck) PassedMessage() string {
	return "SSH configuration is secure."
}

func (s *SSHConfigCheck) FailedMessage() string {
	return "SSH configuration is not secure."
}

func (s *SSHConfigCheck) Run() error {
	s.passed = true

	//run sshd -T to get the sshd config
	configRaw, err := shared.RunCommand("sshd", "-T")
	log.WithField("check", s.Name()).Debugf("sshd -T output: %s", configRaw)
	config := strings.ToLower(string(configRaw))
	if err != nil {
		s.passed = false
		s.status = "Failed to get sshd config"
	}

	if strings.Contains(config, "passwordauthentication yes") {
		s.passed = false
		s.status = "PasswordAuthentication is enabled"
	}
	if strings.Contains(config, "permitrootlogin yes") {
		s.passed = false
		s.status = "Root login is enabled"
	}
	if strings.Contains(config, "permitemptypasswords yes") {
		s.passed = false
		s.status = "Empty passwords are allowed"
	}

	return nil
}

func (s *SSHConfigCheck) Passed() bool {
	return s.passed
}

func (s *SSHConfigCheck) IsRunnable() bool {

	// Check if sshd service is running via systemd
	sshdStatus, _ := shared.RunCommand("systemctl", "is-active", "sshd")
	if strings.TrimSpace(string(sshdStatus)) != "inactive" {
		return true
	}

	// Check if ssh service is running via systemd
	sshStatus, _ := shared.RunCommand("systemctl", "is-active", "ssh")
	if strings.TrimSpace(string(sshStatus)) != "inactive" {
		return true
	}
	// Check if ssh socket service is enabled via systemd
	sshSocketStatus, _ := shared.RunCommand("systemctl", "is-enabled", "sshd.socket")
	if strings.TrimSpace(string(sshSocketStatus)) == "enabled" {
		return true
	}

	// Check if ssh socket service is enabled via systemd
	sshSocketStatus, _ = shared.RunCommand("systemctl", "is-enabled", "ssh.socket")
	if strings.TrimSpace(string(sshSocketStatus)) == "enabled" {
		return true
	}

	s.status = "sshd or ssh service is not running"
	return false
}

func (s *SSHConfigCheck) UUID() string {
	return "da4edd80-6af0-4fb3-9fc7-f9a0e9d07f3b"
}

func (s *SSHConfigCheck) Status() string {
	if s.Passed() {
		return s.PassedMessage()
	}
	return s.status
}

func (s *SSHConfigCheck) RequiresRoot() bool {
	return true
}
