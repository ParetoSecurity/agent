package cmd

import (
	"context"
	"os"
	"time"

	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/runner"
	shared "github.com/ParetoSecurity/agent/shared"
	team "github.com/ParetoSecurity/agent/team"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [--skip <uuid>]",
	Short: "Run checks on your system",
	Run: func(cc *cobra.Command, args []string) {
		skipUUIDs, _ := cc.Flags().GetStringArray("skip")

		checkCommand(skipUUIDs)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringArray("skip", []string{}, "skip checks by UUID")
}

func checkCommand(skipUUIDs []string) {
	if shared.IsRoot() {
		log.Warn("Please run this command as a normal user, as it won't report all checks correctly.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	done := make(chan struct{})
	go func() {
		runner.Check(ctx, claims.All, skipUUIDs)
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
			os.Exit(1)
		}

	case <-ctx.Done():
		log.Warn("Check run timed out")
		os.Exit(1)
	}
}
