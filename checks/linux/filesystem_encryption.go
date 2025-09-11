package checks

import (
	"bufio"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

type EncryptingFS struct {
	passed bool
}

// Name returns the name of the check
func (f *EncryptingFS) Name() string {
	return "Filesystem encryption is enabled"
}

// Passed returns the status of the check
func (f *EncryptingFS) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *EncryptingFS) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *EncryptingFS) UUID() string {
	return "21830a4e-84f1-48fe-9c5b-beab436b2cdb"
}

// PassedMessage returns the message to return if the check passed
func (f *EncryptingFS) PassedMessage() string {
	return "Block device encryption is enabled"
}

// FailedMessage returns the message to return if the check failed
func (f *EncryptingFS) FailedMessage() string {
	return "Block device encryption is disabled"
}

// RequiresRoot returns whether the check requires root access
func (f *EncryptingFS) RequiresRoot() bool {
	return true
}

// Run executes the check
func (f *EncryptingFS) Run() error {
	f.passed = false

	// Check if the system is using LUKS
	if maybeCryptoViaLuks() {
		f.passed = true
		return nil
	}
	// Check if the system is using kernel parameters for encryption
	if maybeCryptoViaKernel() {
		f.passed = true
		return nil
	}
	// Check if the system is using ZFS encryption
	if maybeCryptoViaZFS() {
		f.passed = true
		return nil
	}

	return nil
}

// Status returns the status of the check
func (f *EncryptingFS) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	return f.FailedMessage()
}

func maybeCryptoViaLuks() bool {
	// Check if the system is using LUKS
	lsblk, err := shared.RunCommand("lsblk", "-o", "TYPE,MOUNTPOINT")
	if err != nil {
		log.WithError(err).Warn("Failed to run lsblk command")
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(lsblk))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "crypt") {
			log.WithField("line", line).Debug("LUKS encryption detected")
			return true
		}
	}
	log.WithField("output", lsblk).Warn("Failed to scan lsblk output")
	return false
}

func maybeCryptoViaKernel() bool {
	// Read kernel parameters to check if root is booted via crypt
	cmdline, err := shared.ReadFile("/proc/cmdline")
	if err != nil {
		log.WithError(err).Warn("Failed to read /proc/cmdline")
	}

	params := strings.Fields(string(cmdline))
	for _, param := range params {
		if strings.HasPrefix(param, "cryptdevice=") {
			parts := strings.Split(param, ":")
			if len(parts) == 3 && parts[2] == "root" {
				log.WithField("param", param).Debug("Kernel crypto parameters detected")
				return true
			}
		}
	}
	return false
}

func maybeCryptoViaZFS() bool {
	out, err := shared.RunCommand("zfs", "get", "-H", "-o", "name,value", "encryption")
	if err != nil {
		log.WithError(err).Debug("Failed to run zfs get encryption")
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 {
			enc := fields[1]
			if enc != "off" && enc != "-" {
				log.WithField("dataset", fields[0]).WithField("enc", enc).Debug("ZFS encryption detected")
				return true
			}
		}
	}
	log.WithField("output", out).Debug("No encrypted ZFS pools found")
	return false
}
