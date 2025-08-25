package shared

import (
	"os"
	"path/filepath"
	"strings"

	sharedG "github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"golang.org/x/crypto/ssh"
)

type SSHKeys struct {
	passed     bool
	failedKeys []string
	details    string
}

// Name returns the name of the check
func (f *SSHKeys) Name() string {
	return "SSH keys have password protection"
}

// checks if private key has password protection
func (f *SSHKeys) hasPassword(privateKeyPath string) bool {
	keyBytes, err := sharedG.ReadFile(privateKeyPath)
	if err != nil {
		return true // assume secure if can't read
	}

	_, err = ssh.ParsePrivateKey(keyBytes)
	return err != nil // if error occurs, key likely has password or it's FIDO2 managed key
}

// Run executes the check
func (f *SSHKeys) Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	sshDir := filepath.Join(home, ".ssh")

	files, err := os.ReadDir(sshDir)
	if err != nil {
		f.passed = true
		return nil
	}

	f.passed = true
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			privateKeyPath := filepath.Join(sshDir, strings.TrimSuffix(file.Name(), ".pub"))
			if _, err := os.Stat(privateKeyPath); err == nil {
				if !f.hasPassword(privateKeyPath) {
					f.passed = false
					f.failedKeys = append(f.failedKeys, file.Name())
				}
			}
		}
	}

	return nil
}

// Passed returns the status of the check
func (f *SSHKeys) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *SSHKeys) IsRunnable() bool {
	f.details = "No private keys found in ~/.ssh directory"
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	sshPath := filepath.Join(home, ".ssh")
	if _, err := os.Stat(sshPath); os.IsNotExist(err) {
		return false
	}

	//check if there are any private keys in the .ssh directory
	files, err := os.ReadDir(sshPath)
	if err != nil {
		return false
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			privateKeyPath := filepath.Join(sshPath, strings.TrimSuffix(file.Name(), ".pub"))
			if _, err := os.Stat(privateKeyPath); err == nil {
				f.details = "Found private key: " + file.Name()
				log.WithField("file", file.Name()).Debug("Found private key")
				return true
			}
		}
	}

	return false

}

// UUID returns the UUID of the check
func (f *SSHKeys) UUID() string {
	return "b6aaec0f-d76c-429e-aecf-edab7f1ac400"
}

// PassedMessage returns the message to return if the check passed
func (f *SSHKeys) PassedMessage() string {
	return "SSH keys are password protected"
}

// FailedMessage returns the message to return if the check failed
func (f *SSHKeys) FailedMessage() string {
	return "SSH keys are not using password"
}

// RequiresRoot returns whether the check requires root access
func (f *SSHKeys) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *SSHKeys) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	if f.details != "" {
		return f.details
	}
	return "Found unprotected SSH key(s): " + strings.Join(f.failedKeys, ", ")
}
