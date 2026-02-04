package team

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
	"github.com/davecgh/go-spew/spew"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	shared "github.com/ParetoSecurity/agent/shared"
)

const defaultReportURL = "https://cloud.paretosecurity.com"

type Report struct {
	PassedCount       int                         `json:"passedCount"`
	FailedCount       int                         `json:"failedCount"`
	DisabledCount     int                         `json:"disabledCount"`
	Device            shared.ReportingDevice      `json:"device"`
	Version           string                      `json:"version"`
	SignificantChange string                      `json:"significantChange"`
	State             map[string]check.CheckState `json:"state"`
}

// NowReport compiles and returns a Report that summarizes the results of all runnable checks.
func NowReport(all []claims.Claim) Report {

	device := shared.CurrentReportingDevice()
	passed := 0
	failed := 0
	disabled := 0
	disabledSeed := device.MachineUUID
	failedSeed := device.MachineUUID
	checkStates := make(map[string]check.CheckState)
	lastCheckStates := shared.GetLastStates()

	for _, claim := range all {
		for _, checkS := range claim.Checks {
			lastState, found := lastCheckStates[checkS.UUID()]
			if checkS.IsRunnable() && found {
				if lastState.HasError {
					failed++
					failedSeed += checkS.UUID()
					checkStates[checkS.UUID()] = check.CheckStateError
				} else {
					if lastState.Passed {
						passed++
						checkStates[checkS.UUID()] = check.CheckStatePassed
					} else {
						failed++
						failedSeed += checkS.UUID()
						checkStates[checkS.UUID()] = check.CheckStateFailed
					}
				}
			} else {
				disabled++
				disabledSeed += checkS.UUID()
				checkStates[checkS.UUID()] = check.CheckStateDisabled
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

	// Use TeamAPI from config if available, otherwise use default
	reportURL := defaultReportURL
	if shared.Config.TeamAPI != "" {
		reportURL = shared.Config.TeamAPI
	}

	log.WithField("report", spew.Sdump(report)).
		WithField("method", method).
		WithField("teamID", shared.Config.TeamID).
		WithField("reportURL", reportURL).
		Debug("Reporting to team")
	requestURL := fmt.Sprintf("%s/api/v1/team/%s/device", reportURL, shared.Config.TeamID)
	log.WithField("url", requestURL).
		WithField("method", method).
		Debug("Making API request")

	err := requests.URL(reportURL).
		Pathf("/api/v1/team/%s/device", shared.Config.TeamID).
		Method(method).
		Header("X-Device-Auth", shared.Config.AuthToken).
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

	shared.Config.LastTeamReportSuccess = time.Now().UnixMilli()
	if err := shared.SaveConfig(); err != nil {
		log.WithError(err).Warn("failed to save last team report success timestamp")
	}
	return nil
}
