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
	"github.com/golang-jwt/jwt/v5"

	"github.com/ParetoSecurity/agent/check"
	"github.com/ParetoSecurity/agent/claims"
	shared "github.com/ParetoSecurity/agent/shared"
)

const reportURL = "https://cloud.paretosecurity.com"

type DeviceEnrollmentRequest struct {
	InviteID string `json:"invite_id"`
}

type DeviceEnrollmentResponse struct {
	Auth string `json:"auth"`
}

type Report struct {
	PassedCount       int                         `json:"passedCount"`
	FailedCount       int                         `json:"failedCount"`
	DisabledCount     int                         `json:"disabledCount"`
	Device            shared.ReportingDevice      `json:"device"`
	Version           string                      `json:"version"`
	LastCheck         string                      `json:"lastCheck"`
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

func EnrollDevice(inviteID string) (string, string, error) {
	request := DeviceEnrollmentRequest{
		InviteID: inviteID,
	}

	var response DeviceEnrollmentResponse
	errRes := ""

	// Create a context with a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.WithField("inviteID", inviteID).Debug("Enrolling device")

	err := requests.URL(reportURL).
		Path("/api/v1/team/enroll").
		Method(http.MethodPost).
		Header("User-Agent", shared.UserAgent()).
		BodyJSON(&request).
		ToJSON(&response).
		AddValidator(
			requests.ValidatorHandler(
				requests.DefaultValidator,
				requests.ToString(&errRes),
			)).
		Fetch(ctx)

	if err != nil {
		log.WithField("response", errRes).
			WithError(err).
			Warn("Failed to enroll device")
		return "", "", err
	}

	log.WithField("response", response).Debug("Device enrollment successful")

	// Extract team ID from the JWT token
	teamID, err := extractTeamIDFromToken(response.Auth)
	if err != nil {
		log.WithError(err).Warn("Failed to extract team ID from auth token")
		return "", "", fmt.Errorf("failed to extract team ID from auth token: %w", err)
	}

	return response.Auth, teamID, nil
}

func extractTeamIDFromToken(token string) (string, error) {
	// Parse JWT token without verification to extract claims
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse JWT token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to extract claims from JWT token")
	}

	// Extract team_id from claims
	teamID, ok := claims["team_id"].(string)
	if !ok {
		return "", fmt.Errorf("team_id not found in JWT claims")
	}

	return teamID, nil
}
