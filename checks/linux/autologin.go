// Package linux provides checks for Linux systems.
package checks

import (
	"strings"

	"github.com/ParetoSecurity/agent/shared"
)

// Autologin checks for autologin misconfiguration.
type Autologin struct {
	passed bool
	status string
}

// Name returns the name of the check
func (f *Autologin) Name() string {
	return "Automatic login is disabled"
}

// Run executes the check
func (f *Autologin) Run() error {
	f.passed = true

	// Check KDE (SDDM) autologin
	sddmFiles, _ := filepathGlob("/etc/sddm.conf.d/*.conf")
	for _, file := range sddmFiles {
		content, err := shared.ReadFile(file)
		if err == nil {
			if strings.Contains(string(content), "Autologin=true") {
				f.passed = false
				f.status = "Autologin=true in SDDM is enabled"
				return nil
			}
		}
	}

	// Check main SDDM config
	if content, err := shared.ReadFile("/etc/sddm.conf"); err == nil {
		if strings.Contains(string(content), "Autologin=true") {
			f.passed = false
			f.status = "Autologin=true in SDDM is enabled"
			return nil
		}
	}

	// Check GNOME (GDM) autologin
	gdmPaths := []string{"/etc/gdm3/custom.conf", "/etc/gdm/custom.conf"}
	for _, path := range gdmPaths {
		if content, err := shared.ReadFile(path); err == nil {
			if strings.Contains(string(content), "AutomaticLoginEnable=true") {
				f.passed = false
				f.status = "AutomaticLoginEnable=true in GDM is enabled"
				return nil
			}
		}
	}

	// Check GNOME (GDM) autologin using dconf
	output, err := shared.RunCommand("dconf", "read", "/org/gnome/login-screen/enable-automatic-login")
	if err == nil && strings.TrimSpace(string(output)) == "true" {
		f.passed = false
		f.status = "Automatic login is enabled in GNOME"
		return nil
	}

	return nil
}

// Passed returns the status of the check
func (f *Autologin) Passed() bool {
	return f.passed
}

// IsRunnable returns whether Autologin is runnable.
func (f *Autologin) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *Autologin) UUID() string {
	return "f962c423-fdf5-428a-a57a-816abc9b253e"
}

// PassedMessage returns the message to return if the check passed
func (f *Autologin) PassedMessage() string {
	return "Automatic login is off"
}

// FailedMessage returns the message to return if the check failed
func (f *Autologin) FailedMessage() string {
	return "Automatic login is on"
}

// RequiresRoot returns whether the check requires root access
func (f *Autologin) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *Autologin) Status() string {
	if !f.Passed() {
		return f.status
	}
	return f.PassedMessage()
}
