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

	"github.com/ParetoSecurity/pareto-core/claims"
	shared "github.com/ParetoSecurity/pareto-core/shared"
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

	for _, claim := range all {
		for _, check := range claim.Checks {
			if check.IsRunnable() {
				if check.Passed() {
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
	log.Debug(spew.Sdump(report))
	res := ""
	errRes := ""
	method := http.MethodPatch
	if initial {
		method = http.MethodPut
		report = shared.CurrentReportingDevice()
	} else {
		report = NowReport(claims.All)
	}
	err := requests.URL(reportURL).
		Pathf("/api/v1/team/%s/device", shared.Config.TeamID).
		Method(method).
		Transport(shared.HTTPTransport()).
		Header("X-Device-Auth", "Bearer "+shared.Config.AuthToken).
		BodyJSON(&report).
		ToString(&res).
		AddValidator(
			requests.ValidatorHandler(
				requests.DefaultValidator,
				requests.ToString(&errRes),
			)).
		Fetch(context.Background())
	if err != nil {
		log.WithField("response", errRes).
			WithError(err).
			Warnf("Failed to report to team: %s", shared.Config.TeamID)
		return err
	}
	log.WithField("response", res).Debug("API Response")
	return nil
}
