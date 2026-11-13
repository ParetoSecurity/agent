package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/fatih/color"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
)

// wrapStatusRoot formats the check status with color-coded indicators.
func wrapStatusRoot(status *CheckStatus, chk check.Check, err error) string {
	if err != nil {
		return fmt.Sprintf("%s %s", color.RedString("[ERROR]"), err.Error())
	}

	msg := status.Details
	if lo.IsEmpty(msg) {
		msg = chk.Status()
	}
	if status.Passed {
		return fmt.Sprintf("%s %s", color.GreenString("[OK]"), msg)
	}
	return fmt.Sprintf("%s %s", color.RedString("[FAIL]"), msg)
}

// wrapStatus formats the check status with color-coded indicators.
func wrapStatus(chk check.Check, err error) string {
	if err != nil {
		return fmt.Sprintf("%s %s", color.RedString("[ERROR]"), err.Error())
	}

	if chk.Passed() {
		return fmt.Sprintf("%s %s", color.GreenString("[OK]"), chk.Status())
	}
	return fmt.Sprintf("%s %s", color.RedString("[FAIL]"), chk.Status())
}

// inFlight holds the checks that have started but not finished yet, keyed by
// UUID, so a timed out run can say which checks it was still waiting for.
var inFlight sync.Map

func markStarted(claim claims.Claim, chk check.Check) {
	inFlight.Store(chk.UUID(), fmt.Sprintf("%s: %s", claim.Title, chk.Name()))
}

func markFinished(chk check.Check) {
	inFlight.Delete(chk.UUID())
}

// InFlightChecks returns the names of the checks that are still running,
// sorted, as "Claim: Check".
func InFlightChecks() []string {
	names := []string{}
	inFlight.Range(func(_, value any) bool {
		names = append(names, value.(string))
		return true
	})
	sort.Strings(names)
	return names
}

// Check runs a series of checks concurrently for a list of claims.
//
// It iterates over each claim provided in claimsTorun and, for each claim,
// over its associated checks. Each check is executed in its own goroutine.
func Check(ctx context.Context, claimsTorun []claims.Claim, skipUUIDs []string, onlyUUID string) {

	var checkLogger = log.New(os.Stdout)
	var wg sync.WaitGroup
	checkLogger.Info("Starting checks...")

	for _, claim := range claimsTorun {
		for _, chk := range claim.Checks {
			// Skip checks that are skipped
			if lo.Contains(skipUUIDs, chk.UUID()) {
				checkLogger.Warn(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), fmt.Sprintf("%s Skipped by the command rule", color.YellowString("[SKIP]"))))
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
						checkLogger.Debug(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), fmt.Sprintf("%s Skipped by the command rule", color.YellowString("[SKIP]"))))
						return
					}

					// Skip checks that are not runnable or are disabled
					if !chk.IsRunnable() || shared.IsCheckDisabled(chk.UUID()) {
						reason := chk.Status()
						if shared.IsCheckDisabled(chk.UUID()) {
							reason = "Disabled by the config file"
						}
						checkLogger.Warn(fmt.Sprintf("%s: %s > %s %s", claim.Title, chk.Name(), color.YellowString("[DISABLED]"), reason))
						return
					}
					hasError := false
					markStarted(claim, chk)
					defer markFinished(chk)
					if chk.RequiresRoot() {
						log.Debug("Running check via root helper")
						// Run as root
						status, err := RunCheckViaRoot(chk.UUID())
						if err != nil {
							hasError = true
							checkLogger.Info(fmt.Sprintf("[root] %s: %s > %s", claim.Title, chk.Name(), wrapStatusRoot(status, chk, err)))
						} else {
							if status.Passed {
								checkLogger.Info(fmt.Sprintf("[root] %s: %s > %s", claim.Title, chk.Name(), wrapStatusRoot(status, chk, err)))
							} else {
								checkLogger.Warn(fmt.Sprintf("[root] %s: %s > %s", claim.Title, chk.Name(), wrapStatusRoot(status, chk, err)))
							}

						}
						shared.UpdateLastState(shared.LastState{
							UUID:     chk.UUID(),
							Name:     chk.Name(),
							Passed:   status.Passed,
							HasError: hasError,
							Details:  status.Details,
						})
					} else {
						if err := chk.Run(); err != nil {
							hasError = true
							checkLogger.Info(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), wrapStatus(chk, err)))
						} else {
							if chk.Passed() {
								checkLogger.Info(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), wrapStatus(chk, err)))
							} else {
								checkLogger.Warn(fmt.Sprintf("%s: %s > %s", claim.Title, chk.Name(), wrapStatus(chk, err)))
							}
						}
						shared.UpdateLastState(shared.LastState{
							UUID:     chk.UUID(),
							Name:     chk.Name(),
							Passed:   chk.Passed(),
							HasError: hasError,
							Details:  chk.Status(),
						})
					}

				}
			}(claim, chk)
		}
	}
	wg.Wait()
	if err := shared.CommitLastState(); err != nil {
		log.WithError(err).Warn("failed to commit last state")
	}

	checkLogger.Info("Checks completed.")
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
