package checks

import (
	"encoding/json"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

type DiskEncryption struct {
	passed bool
	status string
}

type bitlockerVolume struct {
	MountPoint       string `json:"MountPoint"`
	ProtectionStatus int    `json:"ProtectionStatus"`
	VolumeType       int    `json:"VolumeType"`
}

// Name returns the name of the check
func (d *DiskEncryption) Name() string {
	return "Disk encryption is enabled"
}

// Run executes the check
func (d *DiskEncryption) Run() error {
	d.passed = false

	// Query BitLocker status for all volumes
	out, err := shared.RunCommand("powershell", "-Command",
		"Get-BitLockerVolume | Select-Object MountPoint, ProtectionStatus, VolumeType | ConvertTo-Json")
	if err != nil {
		log.WithError(err).Warn("Failed to query BitLocker status")
		d.status = "Failed to query BitLocker status"
		return nil
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		d.status = "No BitLocker volumes found"
		return nil
	}

	var volumes []bitlockerVolume

	// PowerShell returns a single object (not array) when there's one volume
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &volumes); err != nil {
			log.WithError(err).Warn("Failed to parse BitLocker output")
			d.status = "Failed to parse BitLocker status"
			return nil
		}
	} else {
		var single bitlockerVolume
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			log.WithError(err).Warn("Failed to parse BitLocker output")
			d.status = "Failed to parse BitLocker status"
			return nil
		}
		volumes = append(volumes, single)
	}

	// Check that the OS volume (VolumeType 1) has protection enabled (ProtectionStatus 1)
	for _, vol := range volumes {
		if vol.VolumeType == 1 {
			if vol.ProtectionStatus == 1 {
				d.passed = true
				d.status = "BitLocker is enabled on " + vol.MountPoint
				return nil
			}
			d.status = "BitLocker is not enabled on OS volume " + vol.MountPoint
			return nil
		}
	}

	// No OS volume found, check if any volume is protected
	for _, vol := range volumes {
		if vol.ProtectionStatus == 1 {
			d.passed = true
			d.status = "BitLocker is enabled on " + vol.MountPoint
			return nil
		}
	}

	d.status = "No encrypted volumes found"
	return nil
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
	return true
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
