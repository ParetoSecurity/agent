package team

import (
	"errors"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestExtractTeamIDFromAuth(t *testing.T) {
	t.Run("valid JWT with team_id", func(t *testing.T) {
		// JWT payload: {"team_id":"test-team-123","sub":"user@example.com"}
		auth := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoidGVzdC10ZWFtLTEyMyIsInN1YiI6InVzZXJAZXhhbXBsZS5jb20ifQ.signature"
		teamID, err := extractTeamIDFromAuth(auth)
		assert.NoError(t, err)
		assert.Equal(t, "test-team-123", teamID)
	})

	t.Run("invalid JWT format", func(t *testing.T) {
		auth := "invalid.jwt"
		_, err := extractTeamIDFromAuth(auth)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid auth token format")
	})

	t.Run("JWT without team_id", func(t *testing.T) {
		// JWT payload: {"sub":"user@example.com"}
		auth := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyQGV4YW1wbGUuY29tIn0.signature"
		_, err := extractTeamIDFromAuth(auth)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team ID not found")
	})
}

func TestEnrollDevice(t *testing.T) {
	defer gock.Off()

	// Save original config
	originalTeamAPI := shared.Config.TeamAPI
	originalTeamID := shared.Config.TeamID
	originalAuthToken := shared.Config.AuthToken
	defer func() {
		shared.Config.TeamAPI = originalTeamAPI
		shared.Config.TeamID = originalTeamID
		shared.Config.AuthToken = originalAuthToken
	}()

	t.Run("successful enrollment", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(200).
			JSON(map[string]string{
				"auth": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoidGVzdC10ZWFtLTEyMyIsInN1YiI6InVzZXJAZXhhbXBsZS5jb20ifQ.signature",
			})

		err := EnrollDevice("test-invite-123", "")
		assert.NoError(t, err)
	})

	t.Run("custom host", func(t *testing.T) {
		gock.New("https://custom.api.com").
			Post("/api/v1/team/enroll").
			Reply(200).
			JSON(map[string]string{
				"auth": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoidGVzdC10ZWFtLTEyMyIsInN1YiI6InVzZXJAZXhhbXBsZS5jb20ifQ.signature",
			})

		err := EnrollDevice("test-invite-123", "https://custom.api.com")
		assert.NoError(t, err)
	})

	t.Run("empty invite ID", func(t *testing.T) {
		err := EnrollDevice("", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invite ID is required")
	})

	t.Run("enrollment failure", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(400).
			JSON(map[string]string{
				"error": "Invalid invite ID",
			})

		err := EnrollDevice("invalid-invite", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enrollment failed")
	})

	t.Run("network timeout", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			ReplyError(errors.New("timeout"))

		err := EnrollDevice("test-invite-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enrollment failed")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(200).
			BodyString("invalid json")

		err := EnrollDevice("test-invite-123", "")
		assert.Error(t, err)
	})

	t.Run("missing auth in response", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(200).
			JSON(map[string]string{
				"auth": "",
			})

		err := EnrollDevice("test-invite-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid auth token format")
	})

	t.Run("malformed auth token", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(200).
			JSON(map[string]string{
				"auth": "malformed.token",
			})

		err := EnrollDevice("test-invite-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid auth token format")
	})

	t.Run("server error response", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/team/enroll").
			Reply(500).
			JSON(map[string]string{
				"error": "Internal server error",
			})

		err := EnrollDevice("test-invite-123", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enrollment failed")
	})
}
