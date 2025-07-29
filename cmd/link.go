package cmd

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ParetoSecurity/agent/notify"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/team"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <url>",
	Short: "Link this device to a team",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if shared.IsRoot() {
			return fmt.Errorf("please run this command as a normal user")
		}
		err := runLinkCommand(args[0])
		if err != nil {
			log.WithError(err).Error("Failed to link team")
			notify.Toast("Failed to add device to the team!")
			return err
		}
		notify.Toast("Device successfully linked to the team!")
		return nil
	},
}

func runLinkCommand(teamURL string) error {
	if lo.IsEmpty(teamURL) {
		log.Warn("Please provide a team URL")
		return errors.New("no team URL provided")
	}
	if shared.IsLinked() {
		log.Info("Device already linked to a team, unlinking first")
		log.Infof("Previous Team ID: %s", shared.Config.TeamID)
		// Automatically unlink the device
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		shared.Config.TeamAPI = ""
		shared.Config.TeamAPI = ""
		log.Info("Device unlinked, proceeding with new team linking")
	}

	// Parse the URL to extract invite_id and host
	inviteID, host, err := parseEnrollmentURL(teamURL)
	if err != nil {
		log.WithError(err).Warn("failed to parse enrollment URL")
		return err
	}

	// Enroll the device
	err = team.EnrollDevice(inviteID, host)
	if err != nil {
		log.WithError(err).Warn("failed to enroll device")
		return err
	}

	// Save config
	err = shared.SaveConfig()
	if err != nil {
		log.Errorf("Error saving config: %v", err)
		return err
	}

	// Report to team
	if shared.IsLinked() {
		err := team.ReportToTeam(false)
		if err != nil {
			log.WithError(err).Warn("failed to report to team")
		}
	}

	log.Infof("Device successfully linked to team: %s", shared.Config.TeamID)
	return nil
}

func parseEnrollmentURL(enrollURL string) (inviteID, host string, err error) {
	// Expected format: paretosecurity://linkDevice?invite_id=<ID>&host=<optional>
	// or just a URL with query parameters

	parsedURL, err := url.Parse(enrollURL)
	if err != nil {
		return "", "", err
	}

	queryParams := parsedURL.Query()
	inviteID = queryParams.Get("invite_id")
	if inviteID == "" {
		return "", "", fmt.Errorf("invite_id not found in URL")
	}

	// Extract optional host parameter
	host = queryParams.Get("host")

	return inviteID, host, nil
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
