package cmd

import (
	"context"

	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/runner"
	shared "github.com/ParetoSecurity/agent/shared"
	team "github.com/ParetoSecurity/agent/team"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [--skip <uuid>] [--only <uuid>]",
	Short: "Run checks on your system",
	Run: func(cc *cobra.Command, args []string) {
		skipUUIDs, _ := cc.Flags().GetStringArray("skip")
		onlyUUID, _ := cc.Flags().GetString("only")
		checkCommand(skipUUIDs, onlyUUID)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringArray("skip", []string{}, "skip checks by UUID")
	checkCmd.Flags().String("only", "", "only run checks by UUID")
}

// CheckConfig holds the configuration for the check command
type CheckConfig struct {
	IsRoot          func() bool
	IsLinked        func() bool
	AllChecksPassed func() bool
	GetFailedChecks func() []shared.LastState
	ReportToTeam    func(bool) error
	RunnerCheck     func(context.Context, []claims.Claim, []string, string)
	LogFatal        func(string)
	LogWarn         func(string)
	LogErrorf       func(string, ...interface{})
	LogWithError    func(error) *log.Entry
}

// DefaultCheckConfig returns the default configuration
func DefaultCheckConfig() *CheckConfig {
	return &CheckConfig{
		IsRoot:          shared.IsRoot,
		IsLinked:        shared.IsLinked,
		AllChecksPassed: shared.AllChecksPassed,
		GetFailedChecks: shared.GetFailedChecks,
		ReportToTeam:    team.ReportToTeam,
		RunnerCheck:     runner.Check,
		LogFatal:        log.Fatal,
		LogWarn:         log.Warn,
		LogErrorf:       log.Errorf,
		LogWithError:    log.WithError,
	}
}

func checkCommand(skipUUIDs []string, onlyUUID string) {
	config := DefaultCheckConfig()
	runCheckCommand(config, skipUUIDs, onlyUUID)
}

func runCheckCommand(config *CheckConfig, skipUUIDs []string, onlyUUID string) {
	if config.IsRoot() {
		config.LogWarn("Please run this command as a normal user, as it won't report all checks correctly.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), shared.CheckTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		config.RunnerCheck(ctx, claims.All, skipUUIDs, onlyUUID)
		close(done)
	}()

	select {
	case <-done:
		if config.IsLinked() {
			err := config.ReportToTeam(false)
			if err != nil {
				config.LogWithError(err).Warn("failed to report to team")
			}
		}

		// if checks failed, exit with a non-zero status code
		if !config.AllChecksPassed() {
			// Log the failed checks
			if failedChecks := config.GetFailedChecks(); len(failedChecks) > 0 && verbose {
				for _, check := range failedChecks {
					config.LogErrorf("Failed check: %s (UUID: %s)", check.Name, check.UUID)
				}
			}
			config.LogFatal("You can use `paretosecurity check --verbose` to get a detailed report.")
		}

	case <-ctx.Done():
		config.LogFatal("Check run timed out")

	}
}
