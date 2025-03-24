// Package shared provides SSH key algo utilities.
package shared

import (
	"crypto/rsa"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh" // Import the crypto/ssh package
)

// KeyType represents the type of SSH key.
type KeyType string

const (
	// Ed25519 is the Ed25519 key algorithm.
	Ed25519 KeyType = "ED25519"
	// Ed25519Sk is the Ed25519-SK key algorithm.
	Ed25519Sk KeyType = "ED25519-SK"
	// Ecdsa is the ECDSA key algorithm.
	Ecdsa KeyType = "ECDSA"
	// EcdsaSk is the ECDSA-SK key algorithm.
	EcdsaSk KeyType = "ECDSA-SK"
	// Dsa is the DSA key algorithm.
	Dsa KeyType = "DSA"
	// Rsa is the RSA key algorithm.
	Rsa KeyType = "RSA"
)

// SSHKeysAlgo runs the SSH keys algorithm.
type SSHKeysAlgo struct {
	passed  bool
	sshKey  string
	sshPath string
	details string
}

// Name returns the name of the check
func (f *SSHKeysAlgo) Name() string {
	return "SSH keys have sufficient algorithm strength"
}

func (f *SSHKeysAlgo) isKeyStrong(path string) bool {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	key, err := ssh.ParsePublicKey(keyBytes)
	if err != nil {
		return false
	}

	switch key.Type() {
	case "ssh-rsa":
		rsaKey, ok := key.(ssh.CryptoPublicKey).CryptoPublicKey().(*rsa.PublicKey)
		if !ok {
			return false
		}
		return rsaKey.N.BitLen() >= 2048
	case "ssh-dss":
		return false // DSS is considered weak
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		return true // ECDSA is considered strong enough
	case "ssh-ed25519":
		return true // Ed25519 is considered strong
	default:
		return false
	}
}

// Run executes the check
func (f *SSHKeysAlgo) Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	f.sshPath = filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(f.sshPath)
	if err != nil {
		return err
	}

	f.passed = true
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		pubPath := filepath.Join(f.sshPath, entry.Name())
		privPath := strings.TrimSuffix(pubPath, ".pub")

		if _, err := os.Stat(privPath); os.IsNotExist(err) {
			continue
		}

		if !f.isKeyStrong(pubPath) {
			f.passed = false
			f.sshKey = strings.TrimSuffix(entry.Name(), ".pub")
			break
		}
	}

	return nil
}

// Passed returns the status of the check
func (f *SSHKeysAlgo) Passed() bool {
	return f.passed
}

// IsRunnable returns whether SSHKeysAlgo is runnable.
func (f *SSHKeysAlgo) IsRunnable() bool {

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
				return true
			}
		}
	}
	f.details = "No private keys found in the .ssh directory"
	return false
}

// UUID returns the UUID of the check
func (f *SSHKeysAlgo) UUID() string {
	return "ef69f752-0e89-46e2-a644-310429ae5f45"
}

// PassedMessage returns the message to return if the check passed
func (f *SSHKeysAlgo) PassedMessage() string {
	return "SSH keys use strong encryption"
}

// FailedMessage returns the message to return if the check failed
func (f *SSHKeysAlgo) FailedMessage() string {
	return "SSH keys are using weak encryption"
}

// RequiresRoot returns whether the check requires root access
func (f *SSHKeysAlgo) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *SSHKeysAlgo) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}

	return "SSH key " + f.sshKey + " is using weak encryption"
}
