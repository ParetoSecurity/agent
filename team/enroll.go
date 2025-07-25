package team

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
)

// EnrollmentDevice is like ReportingDevice but without the Auth field
type EnrollmentDevice struct {
	MachineUUID      string `json:"machineUUID"`
	MachineName      string `json:"machineName"`
	OSVersion        string `json:"macOSVersion,omitempty"`
	LinuxOSVersion   string `json:"linuxOSVersion,omitempty"`
	WindowsOSVersion string `json:"windowsOSVersion,omitempty"`
	ModelName        string `json:"modelName"`
	ModelSerial      string `json:"modelSerial"`
}

type DeviceEnrollmentRequest struct {
	InviteID string           `json:"invite_id"`
	Device   EnrollmentDevice `json:"device"`
}

type DeviceEnrollmentResponse struct {
	Auth string `json:"auth"`
}

// EnrollDevice enrolls a device using an invite ID
func EnrollDevice(inviteID string, host string) error {
	if inviteID == "" {
		return errors.New("invite ID is required")
	}

	// Use provided host or default
	enrollURL := defaultReportURL
	if host != "" {
		enrollURL = host
	}

	// Get current device info
	reportingDevice := shared.CurrentReportingDevice()

	// Create enrollment device without auth field
	enrollDevice := EnrollmentDevice{
		MachineUUID: reportingDevice.MachineUUID,
		MachineName: reportingDevice.MachineName,
		ModelName:   reportingDevice.ModelName,
		ModelSerial: reportingDevice.ModelSerial,
	}

	// Set the appropriate OS version field based on platform
	switch runtime.GOOS {
	case "darwin":
		enrollDevice.OSVersion = reportingDevice.OSVersion
	case "linux":
		enrollDevice.LinuxOSVersion = reportingDevice.OSVersion
	case "windows":
		enrollDevice.WindowsOSVersion = reportingDevice.OSVersion
	}

	req := DeviceEnrollmentRequest{
		InviteID: inviteID,
		Device:   enrollDevice,
	}

	var resp DeviceEnrollmentResponse
	var errResp string

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	enrollPath := "/api/v1/enroll"
	fullURL := enrollURL + enrollPath

	log.WithField("inviteID", inviteID).
		WithField("enrollURL", enrollURL).
		WithField("fullURL", fullURL).
		WithField("device", enrollDevice).
		Debug("Enrolling device")

	err := requests.URL(enrollURL).
		Path(enrollPath).
		Method("POST").
		Header("User-Agent", shared.UserAgent()).
		BodyJSON(&req).
		ToJSON(&resp).
		AddValidator(
			requests.ValidatorHandler(
				requests.DefaultValidator,
				requests.ToString(&errResp),
			)).
		Fetch(ctx)

	if err != nil {
		log.WithError(err).
			WithField("response", errResp).
			Error("Failed to enroll device")
		return fmt.Errorf("enrollment failed: %w", err)
	}

	log.WithField("auth", resp.Auth).Debug("Device enrolled successfully")

	// Extract team ID from auth token (JWT)
	teamID, err := extractTeamIDFromAuth(resp.Auth)
	if err != nil {
		log.WithError(err).Error("Failed to extract team ID from auth token")
		return err
	}

	// Update config
	shared.Config.TeamID = teamID
	shared.Config.AuthToken = resp.Auth
	shared.Config.TeamAPI = enrollURL

	return nil
}

// extractTeamIDFromAuth extracts the team ID from the JWT auth token
func extractTeamIDFromAuth(auth string) (string, error) {
	// The auth token is a JWT that contains the team ID
	// For now, we'll parse it without verification since we just received it from the server
	// In production, you might want to verify the signature

	// Simple JWT parsing without verification
	// JWT format: header.payload.signature
	parts := strings.Split(auth, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid auth token format")
	}

	// Decode the payload (base64url)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode auth token: %w", err)
	}

	// Parse JSON payload to extract team ID
	var claims struct {
		TeamID string `json:"team_id"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse auth token claims: %w", err)
	}

	if claims.TeamID == "" {
		return "", errors.New("team ID not found in auth token")
	}

	return claims.TeamID, nil
}
