package shared

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
	"github.com/samber/lo"
)

type ParetoRelease struct {
	Version     string    `json:"tag_name,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Draft       bool      `json:"draft,omitempty"`
	Prerelease  bool      `json:"prerelease,omitempty"`
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
	res := []ParetoRelease{}
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
			Header("User-Agent", shared.UserAgent()).
			ToJSON(&res).
			Fetch(context.Background())
		if err != nil {
			log.WithError(err).
				Warnf("Failed to check for updates")
			return err
		}

		latestVersion, latest := f.checkVersion(res)
		f.passed = latest
		f.details = fmt.Sprintf("Current version: %s, Latest version: %s", shared.Version, latestVersion)
		return nil
	}

	err := requests.URL("https://api.github.com/repos/ParetoSecurity/agent/releases").
		Header("Accept", "application/vnd.github+json").
		Header("X-GitHub-Api-Version", "2022-11-28").
		Header("User-Agent", shared.UserAgent()).
		ToJSON(&res).
		Fetch(context.Background())
	if err != nil {
		log.WithError(err).
			Warnf("Failed to check for updates")
		return err
	}

	latestVersion, latest := f.checkVersion(res)
	f.passed = latest
	f.details = fmt.Sprintf("Current version: %s, Latest version: %s", shared.Version, latestVersion)
	return nil

}

func (f *ParetoUpdated) checkVersion(res []ParetoRelease) (string, bool) {

	// Sort releases by published date (newest first)
	slices.SortFunc(res, func(a, b ParetoRelease) int {
		return strings.Compare(b.PublishedAt.Format(time.RFC3339), a.PublishedAt.Format(time.RFC3339))
	})

	// Find the latest stable release
	latestRelease, found := lo.Find(res, func(release ParetoRelease) bool {
		return !release.Draft && !release.Prerelease
	})

	if !found {
		return "Could not compare versions", false
	}

	// Only fail if latest release is older than 10 days and current version does not match
	tenDaysAgo := time.Now().AddDate(0, 0, -10)
	if latestRelease.PublishedAt.Before(tenDaysAgo) {
		currentVersion := shared.Version
		if strings.Contains(currentVersion, "-") {
			// Strip any pre-release suffix for comparison
			currentVersion = strings.Split(currentVersion, "-")[0]
		}
		if currentVersion != latestRelease.Version {
			return latestRelease.Version, false
		}
	}

	// Within 10 days grace period or version matches
	return latestRelease.Version, true
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
