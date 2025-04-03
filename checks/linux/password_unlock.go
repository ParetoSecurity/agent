package checks

import (
	"os"
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

func (f *PasswordToUnlock) checkKDE() bool {
	out, err := shared.RunCommand("kreadconfig5", "--file", "kscreenlockerrc", "--group", "Daemon", "--key", "Autolock")
	if err != nil {
		log.WithError(err).Debug("Failed to check KDE screenlocker settings")
		return false
	}
	result := strings.TrimSpace(string(out)) == "true"
	log.WithField("setting", out).WithField("passed", result).Debug("KDE screenlocker check")
	return result
}

func checkSway() bool {
	paths := []string{"/etc/sway/config"}
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Debug("Failed to get home directory")
		return false
	}

	for _, path := range []string{"/etc/sway/config.d", homedir + "/.config/sway/config.d", homedir + "/.config/sway/config"} {
		files, err := filepathGlob(path)
		if err != nil {
			log.WithError(err).Debugf("Failed to read files from path: %s", path)
			continue
		}
		if len(files) == 0 {
			log.WithField("path", path).Debug("No files found in path")
			continue
		}
		paths = append(paths, files...)
	}

	for _, file := range paths {
		content, err := osReadFile(file)
		if err != nil {
			log.WithError(err).Debugf("Failed to read file: %s", file)
			continue
		}

		lines := strings.Split(string(content), "\\")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "exec swayidle") && !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "swaylock") {
				log.WithField("file", file).Debug("Sway idle lock configuration found")
				return true
			}
		}
	}

	log.Debug("Sway idle lock configuration not found")
	return false
}

// Run executes the check
func (f *PasswordToUnlock) Run() error {
	anyCheckPerformed := false
	allChecksPassed := true

	// Check if running GNOME
	if _, err := lookPath("gsettings"); err == nil {
		anyCheckPerformed = true
		allChecksPassed = allChecksPassed && f.checkGnome()
	} else {
		log.Debug("GNOME environment not detected for screensaver lock check")
	}

	// Check if running KDE
	if _, err := lookPath("kreadconfig5"); err == nil {
		anyCheckPerformed = true
		allChecksPassed = allChecksPassed && f.checkKDE()
	} else {
		log.Debug("KDE environment not detected for screensaver lock check")
	}

	// Check if running Sway
	if _, err := lookPath("sway"); err == nil {
		anyCheckPerformed = true
		allChecksPassed = allChecksPassed && checkSway()
	} else {
		log.Debug("Sway environment not detected for screensaver lock check")
	}

	// Performed at least one check and all performed checks passed
	f.passed = anyCheckPerformed && allChecksPassed
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
