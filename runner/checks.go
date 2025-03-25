package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fatih/color"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
)

// wrapStatus formats the check status with color-coded indicators.
func wrapStatus(chk check.Check) string {
	if chk.Passed() {
		return fmt.Sprintf("%s %s", color.GreenString("[OK]"), chk.Status())
	}
	return fmt.Sprintf("%s %s", color.RedString("[FAIL]"), chk.Status())
}

// Check runs a series of checks concurrently for a list of claims.
//
// It iterates over each claim provided in claimsTorun and, for each claim,
// over its associated checks. Each check is executed in its own goroutine.
func Check(ctx context.Context, claimsTorun []claims.Claim, skipUUIDs []string, onlyUUID string) {

	var wg sync.WaitGroup
	log.Info("Starting checks...")

	for _, claim := range claimsTorun {
		for _, chk := range claim.Checks {
			// Skip checks that are skipped
			if lo.Contains(skipUUIDs, chk.UUID()) {
				log.Warn(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), fmt.Sprintf("%s Skipped by the command rule", color.YellowString("[SKIP]"))))
				continue
			}
			wg.Add(1)
			go func(claim claims.Claim, chk check.Check) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				default:

					// Skip checks that are not in the onlyUUID list
					if onlyUUID != "" && onlyUUID != chk.UUID() {
						log.Debug(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), fmt.Sprintf("%s Skipped by the command rule", color.YellowString("[SKIP]"))))
						return
					}

					// Skip checks that are not runnable
					if !chk.IsRunnable() {
						log.Warn(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), chk.Status()))
						return
					}

					if err := chk.Run(); err != nil {
						log.WithError(err).Warnf("%s: %s > %s", claim.Title, chk.Name(), err.Error())
					}

					if chk.Passed() {
						log.Info(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), wrapStatus(chk)))
					} else {
						log.Warn(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), wrapStatus(chk)))
					}

					shared.UpdateLastState(shared.LastState{
						UUID:    chk.UUID(),
						Name:    chk.Name(),
						State:   chk.Passed(),
						Details: chk.Status(),
					})
				}
			}(claim, chk)
		}
	}
	wg.Wait()
	if err := shared.CommitLastState(); err != nil {
		log.WithError(err).Warn("failed to commit last state")
	}

	log.Info("Checks completed.")
}

// PrintSchemaJSON constructs and prints a JSON schema generated from a slice of claims.
// For each claim, the function builds a nested map where the claim's title is the key and its
// value is another map. This inner map associates each check's UUID with a slice that contains
// the check's passed message and failed message.
// The resulting schema is marshalled into an indented JSON string and printed to standard output.
// In case of an error during marshalling, the function logs a warning with the error details.
func PrintSchemaJSON(claimsTorun []claims.Claim) {
	schema := make(map[string]map[string][]string)
	for _, claim := range claimsTorun {
		checks := make(map[string][]string)
		for _, chk := range claim.Checks {
			checks[chk.UUID()] = []string{chk.PassedMessage(), chk.FailedMessage()}
		}
		schema[claim.Title] = checks
	}
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.WithError(err).Warn("cannot marshal schema")
	}
	fmt.Println(string(out))
}
