package checks

import "os"

// SecureBoot checks secure boot configuration.
type SecureBoot struct {
	passed bool
	status string
}

// Name returns the name of the check
func (f *SecureBoot) Name() string {
	return "SecureBoot is enabled"
}

// Run executes the check
func (f *SecureBoot) Run() error {

	if _, err := osStat("/sys/firmware/efi/efivars/"); err != nil && os.IsNotExist(err) {
		f.passed = false
		f.status = "System is not running in UEFI mode"
		return nil
	}

	// Find and read the SecureBoot EFI variable
	pattern := "/sys/firmware/efi/efivars/SecureBoot-*"
	matches, err := filepathGlob(pattern)
	if err != nil || len(matches) == 0 {
		f.passed = false
		f.status = "Could not find SecureBoot EFI variable"
		return nil
	}

	data, err := osReadFile(matches[0])
	if err != nil {
		f.passed = false
		f.status = "Could not read SecureBoot status"
		return nil
	}

	// The SecureBoot variable has a 5-byte structure
	// First 4 bytes are the attribute flags, last byte is the value
	// Value of 1 means enabled, 0 means disabled
	if len(data) >= 5 && data[4] == 1 {
		f.passed = true
		return nil
	}
	f.passed = false

	return nil
}

// Passed returns the status of the check
func (f *SecureBoot) Passed() bool {
	return f.passed
}

// IsRunnable returns whether SecureBoot is runnable.
func (f *SecureBoot) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *SecureBoot) UUID() string {
	return "c96524f2-850b-4bb9-abc7-517051b6c14e"
}

// PassedMessage returns the message to return if the check passed
func (f *SecureBoot) PassedMessage() string {
	return "SecureBoot is enabled"
}

// FailedMessage returns the message to return if the check failed
func (f *SecureBoot) FailedMessage() string {
	return "SecureBoot is disabled"
}

// RequiresRoot returns whether the check requires root access
func (f *SecureBoot) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *SecureBoot) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	if f.status != "" {
		return f.status
	}
	return f.FailedMessage()
}
