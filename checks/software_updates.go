package checks

import (
	"os/exec"
	"strings"

	"github.com/ParetoSecurity/pareto-core/shared"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
)

type SoftwareUpdates struct {
	passed  bool
	details string
}

// Name returns the name of the check
func (f *SoftwareUpdates) Name() string {
	return "Apps are up to date"
}

func (f *SoftwareUpdates) checkUpdates() (bool, string) {
	updates := []string{}

	// Check flatpak
	if _, err := lookPath("flatpak"); err == nil {
		output, err := shared.RunCommand("flatpak", "remote-ls", "--updates")
		log.WithField("output", string(output)).Debug("Flatpak updates")
		if err == nil && len(output) > 0 {
			updates = append(updates, "Flatpak")
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
		output, err := shared.RunCommand("snap", "refresh", "--list")
		log.WithField("output", string(output)).Debug("Snap updates")
		if err == nil && len(output) > 0 && !strings.Contains(string(output), "All snaps up to date.") {
			updates = append(updates, "Snap")
		}
	}

	if len(updates) == 0 {
		return true, "All packages are up to date"
	}
	updates = lo.Uniq(updates)
	return false, "Updates available for: " + strings.Join(updates, ", ")
}

// Run executes the check
func (f *SoftwareUpdates) Run() error {
	var ok bool
	ok, f.details = f.checkUpdates()
	f.passed = ok
	return nil
}

// Passed returns the status of the check
func (f *SoftwareUpdates) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *SoftwareUpdates) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *SoftwareUpdates) UUID() string {
	return "7436553a-ae52-479b-937b-2ae14d15a520"
}

// PassedMessage returns the message to return if the check passed
func (f *SoftwareUpdates) PassedMessage() string {
	return "All apps are up to date"
}

// FailedMessage returns the message to return if the check failed
func (f *SoftwareUpdates) FailedMessage() string {
	return "Some apps are out of date"
}

// RequiresRoot returns whether the check requires root access
func (f *SoftwareUpdates) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *SoftwareUpdates) Status() string {
	return f.details
}
