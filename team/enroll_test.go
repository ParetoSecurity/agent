package team

import (
	"testing"

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

	t.Run("successful enrollment", func(t *testing.T) {
		gock.New("https://cloud.paretosecurity.com").
			Post("/api/v1/enroll").
			Reply(200).
			JSON(map[string]string{
				"auth": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoidGVzdC10ZWFtLTEyMyIsInN1YiI6InVzZXJAZXhhbXBsZS5jb20ifQ.signature",
			})

		err := EnrollDevice("test-invite-123", "")
		assert.NoError(t, err)
	})

	t.Run("custom host", func(t *testing.T) {
		gock.New("https://custom.api.com").
			Post("/api/v1/enroll").
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
			Post("/api/v1/enroll").
			Reply(400).
			JSON(map[string]string{
				"error": "Invalid invite ID",
			})

		err := EnrollDevice("invalid-invite", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "enrollment failed")
	})
}
