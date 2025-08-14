package checks

import (
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

// parseFlatpak parses the output of flatpak update commands and extracts application
func (f *ApplicationUpdates) parseFlatpak(updateLines string) (apps map[string]string) {
	apps = make(map[string]string)
	for line := range strings.Lines(updateLines) {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip lines that do not contain a dot, which indicates a version number
		if !strings.Contains(line, ".") {
			continue
		}

		// Split the line into parts, expecting at least two: application ID and version
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			apps[parts[0]] = parts[1]
		}
	}
	return
}

func (f *ApplicationUpdates) checkUpdates() (bool, string) {
	updates := []string{}

	// Check flatpak
	if _, err := lookPath("flatpak"); err == nil {
		updatesOutput, err := shared.RunCommand("flatpak", "remote-ls", "--app", "--updates", "--columns=application,version")
		if err != nil {
			log.WithError(err).Error("Failed to check flatpak updates")
			return true, "Flatpak updates check failed"
		}
		installedOutput, err := shared.RunCommand("flatpak", "list", "--app", "--columns=application,version")
		if err != nil {
			log.WithError(err).Error("Failed to list installed flatpak apps")
			return true, "Flatpak installed apps check failed"
		}
		installedApps := f.parseFlatpak(string(installedOutput))
		updatableApps := f.parseFlatpak(string(updatesOutput))
		log.WithField("updates", updatesOutput).WithField("installed", installedOutput).Debug("Flatpak updates")

		for app, version := range installedApps {
			if installed, ok := updatableApps[app]; ok && version != installed {
				updates = append(updates, "Flatpak")
				break
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
		if out, _ := shared.RunCommand("dnf", "updateinfo", "list", "--security", "--quiet"); !lo.IsEmpty(out) {
			outStr := string(out)
			if strings.Contains(outStr, "security") && strings.Count(outStr, "\n") > 0 {
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
			if err == nil && !lo.IsEmpty(output) && !strings.Contains(string(output), "All snaps up to date") {
				log.WithField("output", string(output)).Info("Snap updates found")
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
