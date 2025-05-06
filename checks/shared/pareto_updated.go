package shared

import (
	"context"
	"fmt"
	"runtime"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
)

type ParetoReleases []struct {
	TagName string `json:"tag_name,omitempty"`
}

type ParetoUpdated struct {
	passed  bool
	details string
}

// Name returns the name of the check
func (f *ParetoUpdated) Name() string {
	return "Pareto Security is up to date"
}

// Run executes the check
func (f *ParetoUpdated) Run() error {
	f.passed = false
	res := ParetoReleases{}
	device := shared.CurrentReportingDevice()
	platform := "linux"
	if runtime.GOOS == "darwin" {
		platform = "macos"
	}
	if runtime.GOOS == "windows" {
		platform = "windows"
	}

	if shared.IsLinked() {
		err := requests.URL("https://paretosecurity.com/api/updates").
			Param("uuid", device.MachineUUID).
			Param("version", shared.Version).
			Param("os_version", device.OSVersion).
			Param("platform", platform).
			Param("app", "auditor").
			Param("distribution", func() string {
				if shared.IsLinked() {
					return "app-live-team"
				}
				return "app-live-opensource"
			}()).
			Header("Accept", "application/vnd.github+json").
			Header("X-GitHub-Api-Version", "2022-11-28").
			Header("User-Agent", "ParetoSecurity-agent"). // needs to be org/repo name to allow requests
			ToJSON(&res).
			Fetch(context.Background())
		if err != nil {
			log.WithError(err).
				Warnf("Failed to check for updates")
			return err
		}

		if len(res) == 0 {
			f.details = "No releases found"
		}

		if res[0].TagName == shared.Version {
			f.passed = true
		}
		f.details = fmt.Sprintf("Current version: %s, Latest version: %s", shared.Version, res[0].TagName)
		return nil
	}

	err := requests.URL("https://api.github.com/repos/ParetoSecurity/agent/releases").
		Header("Accept", "application/vnd.github+json").
		Header("X-GitHub-Api-Version", "2022-11-28").
		Header("User-Agent", "ParetoSecurity-agent"). // needs to be org/repo name to allow requests
		ToJSON(&res).
		Fetch(context.Background())
	if err != nil {
		log.WithError(err).
			Warnf("Failed to check for updates")
		return err
	}

	if len(res) == 0 {
		f.details = "No releases found"
		return nil
	}

	if res[0].TagName == shared.Version {
		f.passed = true
	}
	f.details = fmt.Sprintf("Current version: %s, Latest version: %s", shared.Version, res[0].TagName)
	return nil

}

// Passed returns the status of the check
func (f *ParetoUpdated) Passed() bool {
	return f.passed
}

// CanRun returns whether the check can run
func (f *ParetoUpdated) IsRunnable() bool {
	return true
}

// UUID returns the UUID of the check
func (f *ParetoUpdated) UUID() string {
	return "44e4754a-0b42-4964-9cc2-b88b2023cb1e"
}

// PassedMessage returns the message to return if the check passed
func (f *ParetoUpdated) PassedMessage() string {
	return "Pareto Security is up to date"
}

// FailedMessage returns the message to return if the check failed
func (f *ParetoUpdated) FailedMessage() string {
	return "Pareto Security is outdated " + f.details
}

// RequiresRoot returns whether the check requires root access
func (f *ParetoUpdated) RequiresRoot() bool {
	return false
}

// Status returns the status of the check
func (f *ParetoUpdated) Status() string {
	if f.passed {
		return f.PassedMessage()
	}
	return f.FailedMessage()
}
