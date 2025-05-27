package shared

import (
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/h2non/gock"
)

func TestParetoUpdated_Run(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	shared.Config.TeamID = "test-team-id"
	shared.Config.AuthToken = "test-auth-token"
	defer func() {
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
	}()

	gock.New("https://paretosecurity.com/api/updates").
		Reply(200).
		JSON([]map[string]string{{"tag_name": "1.7.91"}})
	check := &ParetoUpdated{}
	err := check.Run()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}

func TestParetoUpdated_RunPublic(t *testing.T) {
	defer gock.Off() // Flush pending mocks after test execution
	shared.Config.TeamID = ""
	shared.Config.AuthToken = ""

	gock.New("https://api.github.com").
		Reply(200).
		JSON([]map[string]string{{"tag_name": "1.7.91"}})
	check := &ParetoUpdated{}
	err := check.Run()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}

func TestParetoUpdated_Name(t *testing.T) {
	dockerAccess := &ParetoUpdated{}
	expectedName := "Pareto Security is up to date"
	if dockerAccess.Name() != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, dockerAccess.Name())
	}
}

func TestParetoUpdated_Status(t *testing.T) {
	dockerAccess := &ParetoUpdated{}
	expectedStatus := "Pareto Security is outdated "
	if dockerAccess.Status() != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, dockerAccess.Status())
	}
}

func TestParetoUpdated_UUID(t *testing.T) {
	dockerAccess := &ParetoUpdated{}
	expectedUUID := "44e4754a-0b42-4964-9cc2-b88b2023cb1e"
	if dockerAccess.UUID() != expectedUUID {
		t.Errorf("Expected UUID %s, got %s", expectedUUID, dockerAccess.UUID())
	}
}

func TestParetoUpdated_Passed(t *testing.T) {
	dockerAccess := &ParetoUpdated{passed: true}
	expectedPassed := true
	if dockerAccess.Passed() != expectedPassed {
		t.Errorf("Expected Passed %v, got %v", expectedPassed, dockerAccess.Passed())
	}
}

func TestParetoUpdated_FailedMessage(t *testing.T) {
	dockerAccess := &ParetoUpdated{}
	expectedFailedMessage := "Pareto Security is outdated "
	if dockerAccess.FailedMessage() != expectedFailedMessage {
		t.Errorf("Expected FailedMessage %s, got %s", expectedFailedMessage, dockerAccess.FailedMessage())
	}
}

func TestParetoUpdated_PassedMessage(t *testing.T) {
	dockerAccess := &ParetoUpdated{}
	expectedPassedMessage := "Pareto Security is up to date"
	if dockerAccess.PassedMessage() != expectedPassedMessage {
		t.Errorf("Expected PassedMessage %s, got %s", expectedPassedMessage, dockerAccess.PassedMessage())
	}
}

func TestParetoUpdated_checkVersion(t *testing.T) {
	tests := []struct {
		name            string
		releases        []ParetoRelease
		currentVersion  string
		expectedVersion string
		expectedPassed  bool
	}{
		{
			name: "no stable releases found",
			releases: []ParetoRelease{
				{Version: "1.0.0", PublishedAt: time.Now().AddDate(0, 0, -5), Draft: true},
				{Version: "1.1.0", PublishedAt: time.Now().AddDate(0, 0, -3), Prerelease: true},
			},
			expectedVersion: "Cound not compare versions",
			expectedPassed:  false,
		},
		{
			name: "latest release within 10 days grace period",
			releases: []ParetoRelease{
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -5), Draft: false, Prerelease: false},
				{Version: "1.1.0", PublishedAt: time.Now().AddDate(0, 0, -15), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.1.0",
			expectedVersion: "1.2.0",
			expectedPassed:  true,
		},
		{
			name: "current version matches latest older than 10 days",
			releases: []ParetoRelease{
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -15), Draft: false, Prerelease: false},
				{Version: "1.1.0", PublishedAt: time.Now().AddDate(0, 0, -20), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.2.0",
			expectedVersion: "1.2.0",
			expectedPassed:  true,
		},
		{
			name: "current version with prerelease suffix matches latest",
			releases: []ParetoRelease{
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -15), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.2.0-beta.1",
			expectedVersion: "1.2.0",
			expectedPassed:  true,
		},
		{
			name: "current version outdated and latest older than 10 days",
			releases: []ParetoRelease{
				{Version: "1.3.0", PublishedAt: time.Now().AddDate(0, 0, -15), Draft: false, Prerelease: false},
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -20), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.2.0",
			expectedVersion: "1.3.0",
			expectedPassed:  false,
		},
		{
			name: "releases sorted correctly by date",
			releases: []ParetoRelease{
				{Version: "1.1.0", PublishedAt: time.Now().AddDate(0, 0, -20), Draft: false, Prerelease: false},
				{Version: "1.3.0", PublishedAt: time.Now().AddDate(0, 0, -5), Draft: false, Prerelease: false},
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -10), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.2.0",
			expectedVersion: "1.3.0",
			expectedPassed:  true,
		},
		{
			name: "mixed draft and stable releases",
			releases: []ParetoRelease{
				{Version: "1.4.0", PublishedAt: time.Now().AddDate(0, 0, -2), Draft: true, Prerelease: false},
				{Version: "1.3.0", PublishedAt: time.Now().AddDate(0, 0, -5), Draft: false, Prerelease: false},
				{Version: "1.2.0", PublishedAt: time.Now().AddDate(0, 0, -15), Draft: false, Prerelease: false},
			},
			currentVersion:  "1.2.0",
			expectedVersion: "1.3.0",
			expectedPassed:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.Version = tt.currentVersion
			check := &ParetoUpdated{}
			version, passed := check.checkVersion(tt.releases)

			if version != tt.expectedVersion {
				t.Errorf("Expected version %s, got %s", tt.expectedVersion, version)
			}
			if passed != tt.expectedPassed {
				t.Errorf("Expected passed %v, got %v", tt.expectedPassed, passed)
			}
		})
	}
}
