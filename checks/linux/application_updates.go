package checks

import (
	"os/exec"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
)

type ApplicationUpdates struct {
	passed  bool
	details string
}

// Name returns the name of the check
func (f *ApplicationUpdates) Name() string {
	return "Apps are up to date"
}

func (f *ApplicationUpdates) checkUpdates() (bool, string) {
	updates := []string{}

	// Check flatpak
	if _, err := lookPath("flatpak"); err == nil {
		updatesOutput, err := shared.RunCommand("flatpak", "remote-ls", "--app", "--updates", "--columns=application,version")
		log.WithField("output", string(updatesOutput)).Debug("Flatpak updates")
		if err == nil && len(updatesOutput) > 0 {
			// Parse available updates
			updateLines := strings.Split(strings.TrimSpace(string(updatesOutput)), "\n")
			if len(updateLines) > 1 { // Skip header line
				availableUpdates := make(map[string]string)
				for _, line := range updateLines[1:] {
					if strings.TrimSpace(line) == "" {
						continue
					}
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						availableUpdates[parts[0]] = parts[1]
					}
				}

				// Get currently installed versions
				installedOutput, err := shared.RunCommand("flatpak", "list", "--app", "--columns=application,version")
				if err == nil {
					installedLines := strings.Split(strings.TrimSpace(string(installedOutput)), "\n")
					if len(installedLines) > 1 {
						for _, line := range installedLines[1:] {
							if strings.TrimSpace(line) == "" {
								continue
							}
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								appId := parts[0]
								currentVersion := parts[1]
								if updateVersion, hasUpdate := availableUpdates[appId]; hasUpdate && updateVersion != currentVersion {
									updates = append(updates, "Flatpak")
									break // Found at least one update
								}
							}
						}
					}
				}
			}
		}
	}

	// Check apt
	if _, err := lookPath("apt"); err == nil {
		output, err := shared.RunCommand("apt", "list", "--upgradable")
		log.WithField("output", string(output)).Debug("APT updates")
		if err == nil && len(output) > 0 && strings.Contains(string(output), "upgradable") {
			updates = append(updates, "APT")
		}
	}

	// Check dnf
	if _, err := lookPath("dnf"); err == nil {
		if _, err := shared.RunCommand("dnf", "check-update", "--quiet"); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 100 {
				updates = append(updates, "DNF")
			}
		}
	}

	// Check pacman
	if _, err := lookPath("pacman"); err == nil {
		output, err := shared.RunCommand("pacman", "-Qu")
		log.WithField("output", string(output)).Debug("Pacman updates")
		if err == nil && len(output) > 0 {
			updates = append(updates, "Pacman")
		}
	}

	// Check snap
	if _, err := lookPath("snap"); err == nil {
		// Check if snapd is running
		snapdStatus, err := shared.RunCommand("systemctl", "is-active", "snapd")
		if err == nil && strings.TrimSpace(string(snapdStatus)) == "active" {
			output, err := shared.RunCommand("snap", "refresh", "--list")
			log.WithField("output", string(output)).Debug("Snap updates")
			if err == nil && len(output) > 0 && !strings.Contains(string(output), "All snaps up to date.") {
				updates = append(updates, "Snap")
			}
		} else {
			log.Debug("snapd is not running, skipping snap updates check")
		}
	}

	if len(updates) == 0 {
		return true, "All packages are up to date"
	}
	updates = lo.Uniq(updates)
	return false, "Updates available for: " + strings.Join(updates, ", ")
}

// Run executes the check
func (f *ApplicationUpdates) Run() error {
	var ok bool
	ok, f.details = f.checkUpdates()
	f.passed = ok
	return nil
}

// Passed returns the status of the check
func (f *ApplicationUpdates) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *ApplicationUpdates) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *ApplicationUpdates) UUID() string {
	return "7436553a-ae52-479b-937b-2ae14d15a520"
}

// PassedMessage returns the message to return if the check passed
func (f *ApplicationUpdates) PassedMessage() string {
	return "All apps are up to date"
}

// FailedMessage returns the message to return if the check failed
func (f *ApplicationUpdates) FailedMessage() string {
	return "Some apps are out of date"
}

// RequiresRoot returns whether the check requires root access
func (f *ApplicationUpdates) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *ApplicationUpdates) Status() string {
	return f.details
}
