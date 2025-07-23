package cmd

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestGetInviteIDFromURL(t *testing.T) {
	validURL := "paretosecurity://linkDevice/?invite_id=4af8514a-ce63-4747-807f-5c3839d78341"
	invalidURL := "paretosecurity://linkDevice/?token=invalid"

	t.Run("valid URL with invite_id", func(t *testing.T) {
		inviteID, err := getInviteIDFromURL(validURL)
		assert.NoError(t, err)
		assert.Equal(t, "4af8514a-ce63-4747-807f-5c3839d78341", inviteID)
	})

	t.Run("invalid URL without invite_id", func(t *testing.T) {
		_, err := getInviteIDFromURL(invalidURL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invite_id not found in URL")
	})
}

func TestRunLinkCommand_Success(t *testing.T) {
	defer gock.Off()

	// Mock the enrollment endpoint
	gock.New("https://cloud.paretosecurity.com").
		Post("/api/v1/team/enroll").
		Reply(200).
		JSON(map[string]string{
			"auth": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ0ZWFtX2lkIjoiMjQyOWM0OWUtMzdiYi00MWJiLTkwNzctNmJiNjIwMmUyNTViIn0.test",
		})

	// Mock the reporting endpoint
	gock.New("https://cloud.paretosecurity.com").
		Patch("/api/v1/team/2429c49e-37bb-41bb-9077-6bb6202e255b/device").
		Reply(200).
		JSON(map[string]string{"status": "ok"})

	// Reset config
	shared.Config.TeamID = ""
	shared.Config.AuthToken = ""

	defer func() {
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		shared.SaveConfig()
	}()

	// Construct the URL with a valid invite_id
	url := "paretosecurity://linkDevice/?invite_id=4af8514a-ce63-4747-807f-5c3839d78341"

	// Call the function under test
	err := runLinkCommand(url)
	assert.NoError(t, err)

	// Assert that shared.Config was updated
	assert.Equal(t, "2429c49e-37bb-41bb-9077-6bb6202e255b", shared.Config.TeamID)
	assert.NotEmpty(t, shared.Config.AuthToken)
}
