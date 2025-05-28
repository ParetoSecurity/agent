package team

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
	"github.com/davecgh/go-spew/spew"

	"github.com/ParetoSecurity/agent/claims"
	shared "github.com/ParetoSecurity/agent/shared"
)

const reportURL = "https://dash.paretosecurity.com"

type Report struct {
	PassedCount       int                    `json:"passedCount"`
	FailedCount       int                    `json:"failedCount"`
	DisabledCount     int                    `json:"disabledCount"`
	Device            shared.ReportingDevice `json:"device"`
	Version           string                 `json:"version"`
	LastCheck         string                 `json:"lastCheck"`
	SignificantChange string                 `json:"significantChange"`
	State             map[string]string      `json:"state"`
}

// NowReport compiles and returns a Report that summarizes the results of all runnable checks.
func NowReport(all []claims.Claim) Report {
	passed := 0
	failed := 0
	disabled := 0
	disabledSeed, _ := shared.SystemUUID()
	failedSeed, _ := shared.SystemUUID()
	checkStates := make(map[string]string)
	lastCheckStates := shared.GetLastStates()

	for _, claim := range all {
		for _, check := range claim.Checks {
			lastState, found := lastCheckStates[check.UUID()]
			if check.IsRunnable() && found {
				if lastState.State {
					passed++
					checkStates[check.UUID()] = "pass"
				} else {
					failed++
					failedSeed += check.UUID()
					checkStates[check.UUID()] = "fail"
				}
			} else {
				disabled++
				disabledSeed += check.UUID()
				checkStates[check.UUID()] = "off"
			}
		}
	}

	significantChange := sha256.Sum256([]byte(disabledSeed + "." + failedSeed))
	return Report{
		PassedCount:       passed,
		FailedCount:       failed,
		DisabledCount:     disabled,
		Device:            shared.CurrentReportingDevice(),
		Version:           shared.Version,
		LastCheck:         time.Now().Format(time.RFC3339),
		SignificantChange: hex.EncodeToString(significantChange[:]),
		State:             checkStates,
	}
}

// ReportAndSave generates a report and saves it to the configuration file.
func ReportToTeam(initial bool) error {
	var report interface{}

	res := ""
	errRes := ""
	method := http.MethodPatch
	if initial {
		method = http.MethodPut
		report = shared.CurrentReportingDevice()
	} else {
		report = NowReport(claims.All)
	}

	// Create a context with a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.WithField("report", spew.Sdump(report)).
		WithField("method", method).
		WithField("teamID", shared.Config.TeamID).
		Debug("Reporting to team")
	err := requests.URL(reportURL).
		Pathf("/api/v1/team/%s/device", shared.Config.TeamID).
		Method(method).
		Header("X-Device-Auth", "Bearer "+shared.Config.AuthToken).
		Header("User-Agent", shared.UserAgent()).
		BodyJSON(&report).
		ToString(&res).
		AddValidator(
			requests.ValidatorHandler(
				requests.DefaultValidator,
				requests.ToString(&errRes),
			)).
		Fetch(ctx)
	if err != nil {
		log.WithField("response", errRes).
			WithError(err).
			Warnf("Failed to report to team: %s", shared.Config.TeamID)
		return err
	}
	log.WithField("response", res).Debug("API Response")
	return nil
}
