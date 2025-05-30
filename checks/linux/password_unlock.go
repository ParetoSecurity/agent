package checks

import (
	"path/filepath"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

// PasswordToUnlock represents a check to ensure that a password is required to unlock the screen.
type PasswordToUnlock struct {
	passed bool
}

// Name returns the name of the check
func (f *PasswordToUnlock) Name() string {
	return "Password is required to unlock the screen"
}

func (f *PasswordToUnlock) checkGnome() bool {
	out, err := shared.RunCommand("gsettings", "get", "org.gnome.desktop.screensaver", "lock-enabled")
	if err != nil {
		log.WithError(err).Debug("Failed to check GNOME screensaver settings")
		return false
	}
	result := strings.TrimSpace(string(out)) == "true"
	log.WithField("setting", out).WithField("passed", result).Debug("GNOME screensaver lock check")
	return result
}

func (f *PasswordToUnlock) checkKDE5() bool {
	// First try reading config file directly
	homeDir, err := shared.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "kscreenlockerrc")
		if content, err := shared.ReadFile(configPath); err == nil {
			configStr := string(content)
			// Check if LockOnResume=false is present
			if strings.Contains(configStr, "LockOnResume=false") {
				log.WithField("config_file", configPath).Debug("Found LockOnResume=false in KDE config")
				return false
			}
			// If LockOnResume=true is explicitly set or not present (defaults to true)
			log.WithField("config_file", configPath).Debug("KDE config allows screen locking")
			return true
		}
	}
	return false
}

// Run executes the check
func (f *PasswordToUnlock) Run() error {

	// Check if running GNOME
	if _, err := lookPath("gsettings"); err == nil {
		f.passed = f.checkGnome()
	} else {
		log.Info("GNOME environment not detected for screensaver lock check")
	}

	// Check if running KDE
	if _, err := lookPath("kreadconfig5"); err == nil {
		f.passed = f.checkKDE5()
	} else {
		log.Debug("KDE environment(5) not detected for screensaver lock check")
	}

	return nil
}

// Passed returns the status of the check
func (f *PasswordToUnlock) Passed() bool {
	return f.passed
}

// IsRunnable returns whether the check can run
func (f *PasswordToUnlock) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *PasswordToUnlock) UUID() string {
	return "37dee029-605b-4aab-96b9-5438e5aa44d8"
}

// PassedMessage returns the message to return if the check passed
func (f *PasswordToUnlock) PassedMessage() string {
	return "Password after sleep or screensaver is on"
}

// FailedMessage returns the message to return if the check failed
func (f *PasswordToUnlock) FailedMessage() string {
	return "Password after sleep or screensaver is off"
}

// RequiresRoot returns whether the check requires root access
func (f *PasswordToUnlock) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *PasswordToUnlock) Status() string {
	if f.Passed() {
		return f.PassedMessage()
	}
	return f.FailedMessage()
}
