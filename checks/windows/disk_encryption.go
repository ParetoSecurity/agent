package checks

import (
	"fmt"
	"strings"

	"github.com/caarlos0/log"
)

type DiskEncryption struct {
	passed bool
	status string
}

var (
	bitLockerSystemDrive        = defaultSystemDrive
	bitLockerListFixedDrives    = defaultListFixedDrives
	bitLockerProtectionStatusFn = defaultGetBitLockerProtectionStatus
)

// Name returns the name of the check
func (d *DiskEncryption) Name() string {
	return "Disk encryption is enabled"
}

// Run executes the check
func (d *DiskEncryption) Run() error {
	d.passed = false

	systemDrive := strings.TrimRight(bitLockerSystemDrive(), "\\")
	if systemDrive != "" {
		status, err := bitLockerProtectionStatusFn(systemDrive)
		if err != nil {
			log.WithError(err).Warn("Failed to query BitLocker status")
			d.status = "Failed to query BitLocker status"
			return nil
		}
		if isBitLockerProtected(status) {
			d.passed = true
			d.status = "BitLocker is enabled on " + systemDrive
			return nil
		}
		d.status = fmt.Sprintf("BitLocker is not enabled on OS volume %s (status: %d %s)", systemDrive, status, bitLockerStatusDescription(status))
		return nil
	}

	drives, err := bitLockerListFixedDrives()
	if err != nil {
		log.WithError(err).Warn("Failed to query BitLocker status")
		d.status = "Failed to query BitLocker status"
		return nil
	}
	if len(drives) == 0 {
		d.status = "No BitLocker volumes found"
		return nil
	}

	lastDrive := ""
	lastStatus := 0
	for _, drive := range drives {
		status, err := bitLockerProtectionStatusFn(drive)
		if err != nil {
			log.WithError(err).Warn("Failed to query BitLocker status")
			d.status = "Failed to query BitLocker status"
			return nil
		}
		lastDrive = drive
		lastStatus = status
		if isBitLockerProtected(status) {
			d.passed = true
			d.status = "BitLocker is enabled on " + drive
			return nil
		}
	}

	if lastDrive != "" {
		d.status = fmt.Sprintf("No encrypted volumes found (last checked %s: status %d %s)", lastDrive, lastStatus, bitLockerStatusDescription(lastStatus))
		return nil
	}
	d.status = "No encrypted volumes found"
	return nil
}

func isBitLockerProtected(status int) bool {
	switch status {
	// 1 = Protection On, 3 = Unknown, 5 = Decryption in Progress
	case 1, 3, 5:
		return true
	default:
		return false
	}
}

func bitLockerStatusDescription(status int) string {
	switch status {
	case 1:
		return "(Protection On)"
	case 3:
		return "(Unknown)"
	case 5:
		return "(Decryption in Progress)"
	default:
		return "(Unknown Status)"
	}
}

// Passed returns the status of the check
func (d *DiskEncryption) Passed() bool {
	return d.passed
}

// IsRunnable returns whether the check can run
func (d *DiskEncryption) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (d *DiskEncryption) UUID() string {
	return "c2cae85c-0335-4708-a428-3a16fd407912"
}

// PassedMessage returns the message to return if the check passed
func (d *DiskEncryption) PassedMessage() string {
	return "Disk encryption is enabled"
}

// FailedMessage returns the message to return if the check failed
func (d *DiskEncryption) FailedMessage() string {
	return "Disk encryption is disabled"
}

// RequiresRoot returns whether the check requires root access
func (d *DiskEncryption) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (d *DiskEncryption) Status() string {
	if d.Passed() {
		if d.status != "" {
			return d.status
		}
		return d.PassedMessage()
	}
	if d.status != "" {
		return d.status
	}
	return d.FailedMessage()
}
