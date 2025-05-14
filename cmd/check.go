package cmd

import (
	"context"
	"time"

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

func checkCommand(skipUUIDs []string, onlyUUID string) {
	if shared.IsRoot() {
		log.Warn("Please run this command as a normal user, as it won't report all checks correctly.")
	}

	// Set running state for all checks
	for _, claim := range claims.All {
		for _, chk := range claim.Checks {
			shared.SetRunningState(chk.UUID(), true)
		}
	}
	if err := shared.CommitLastState(); err != nil {
		log.WithError(err).Warn("failed to commit running state")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	done := make(chan struct{})
	go func() {
		runner.Check(ctx, claims.All, skipUUIDs, onlyUUID)
		close(done)
	}()

	select {
	case <-done:

		if shared.IsLinked() {
			err := team.ReportToTeam(false)
			if err != nil {
				log.WithError(err).Warn("failed to report to team")
			}
		}

		// if checks failed, exit with a non-zero status code
		if !shared.AllChecksPassed() {
			// Log the failed checks
			if failedChecks := shared.GetFailedChecks(); len(failedChecks) > 0 && verbose {
				for _, check := range failedChecks {
					log.Errorf("Failed check: %s (UUID: %s)", check.Name, check.UUID)
				}
			}
			log.Fatal("You can use `paretosecurity check --verbose` to get a detailed report.")
		}
		// Clear running state for all checks
		shared.StopAllRunningChecks()
		if err := shared.CommitLastState(); err != nil {
			log.WithError(err).Warn("failed to commit running state")
		}

	case <-ctx.Done():
		// Clear running state for all checks if timed out
		shared.StopAllRunningChecks()
		if err := shared.CommitLastState(); err != nil {
			log.WithError(err).Warn("failed to commit running state")
		}
		log.Fatal("Check run timed out")

	}
}
