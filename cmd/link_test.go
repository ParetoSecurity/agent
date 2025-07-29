package cmd

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestParseEnrollmentURL(t *testing.T) {
	t.Run("valid URL with invite_id", func(t *testing.T) {
		inviteID, host, err := parseEnrollmentURL("paretosecurity://linkDevice?invite_id=test-invite-123")
		assert.NoError(t, err)
		assert.Equal(t, "test-invite-123", inviteID)
		assert.Equal(t, "", host)
	})

	t.Run("valid URL with invite_id and host", func(t *testing.T) {
		inviteID, host, err := parseEnrollmentURL("paretosecurity://linkDevice?invite_id=test-invite-123&host=https://api.example.com")
		assert.NoError(t, err)
		assert.Equal(t, "test-invite-123", inviteID)
		assert.Equal(t, "https://api.example.com", host)
	})

	t.Run("missing invite_id", func(t *testing.T) {
		_, _, err := parseEnrollmentURL("paretosecurity://linkDevice?host=https://api.example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invite_id not found")
	})

	t.Run("legacy token parameter not supported", func(t *testing.T) {
		_, _, err := parseEnrollmentURL("http://example.com?token=legacy-token-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invite_id not found")
	})
}

func TestRunLinkCommand_Success(t *testing.T) {
	defer gock.Off()

	// Setup temporary config file
	tempDir := t.TempDir()
	originalConfigPath := shared.ConfigPath
	shared.ConfigPath = tempDir + "/config.toml"
	defer func() {
		shared.ConfigPath = originalConfigPath
	}()

	// Mock the enrollment endpoint
	gock.New("https://cloud.paretosecurity.com").
		Post("/api/v1/team/enroll").
		Reply(200).
		JSON(map[string]string{
			"auth": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoiMjQyOWM0OWUtMzdiYi00MWJiLTkwNzctNmJiNjIwMmUyNTViIiwic3ViIjoianRAZXhhbXBsZS5jb20iLCJpYXQiOjE3MzY0MTc0MTB9.test",
		})

	// Mock the device report endpoint
	gock.New("https://cloud.paretosecurity.com").
		Patch("/api/v1/team/2429c49e-37bb-41bb-9077-6bb6202e255b/device").
		Reply(200).
		JSON(map[string]string{"status": "ok"})

	// Reset config
	shared.Config.TeamID = ""
	shared.Config.AuthToken = ""
	shared.Config.TeamAPI = ""

	defer func() {
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		shared.Config.TeamAPI = ""
	}()

	// Construct the URL with an invite_id
	url := "paretosecurity://linkDevice?invite_id=test-invite-123"

	// Call the function under test
	err := runLinkCommand(url)
	assert.NoError(t, err)

	// Assert that shared.Config was updated
	assert.Equal(t, "2429c49e-37bb-41bb-9077-6bb6202e255b", shared.Config.TeamID)
	assert.NotEmpty(t, shared.Config.AuthToken)
	assert.Equal(t, "https://cloud.paretosecurity.com", shared.Config.TeamAPI)
}

func TestRunLinkCommand_RelinkSuccess(t *testing.T) {
	defer gock.Off()

	// Setup temporary config file
	tempDir := t.TempDir()
	originalConfigPath := shared.ConfigPath
	shared.ConfigPath = tempDir + "/config.toml"
	defer func() {
		shared.ConfigPath = originalConfigPath
	}()

	// Mock the enrollment endpoint for the new team
	gock.New("https://cloud.paretosecurity.com").
		Post("/api/v1/team/enroll").
		Reply(200).
		JSON(map[string]string{
			"auth": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoibmV3LXRlYW0taWQiLCJzdWIiOiJqc0BleGFtcGxlLmNvbSIsImlhdCI6MTczNjQxNzQxMH0.test",
		})

	// Mock the device report endpoint for the new team
	gock.New("https://cloud.paretosecurity.com").
		Patch("/api/v1/team/new-team-id/device").
		Reply(200).
		JSON(map[string]string{"status": "ok"})

	// Set up initial linked state (simulating already linked device)
	shared.Config.TeamID = "old-team-id"
	shared.Config.AuthToken = "old-auth-token"
	shared.Config.TeamAPI = "https://cloud.paretosecurity.com"

	defer func() {
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		shared.Config.TeamAPI = ""
	}()

	// Construct the URL with a new invite_id
	url := "paretosecurity://linkDevice?invite_id=new-invite-123"

	// Call the function under test - should succeed without error
	err := runLinkCommand(url)
	assert.NoError(t, err)

	// Assert that shared.Config was updated to the new team
	assert.Equal(t, "new-team-id", shared.Config.TeamID)
	assert.NotEmpty(t, shared.Config.AuthToken)
	assert.NotEqual(t, "old-auth-token", shared.Config.AuthToken)
	assert.Equal(t, "https://cloud.paretosecurity.com", shared.Config.TeamAPI)
}
