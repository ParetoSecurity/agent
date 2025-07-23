package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

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
	if strings.Contains(teamURL, "https://") {
		return errors.New("team URL should not contain the protocol")
	}
	if shared.IsLinked() {
		log.Warn("Already linked to a team")
		log.Warn("Unlink first with `paretosecurity unlink`")
		log.Infof("Team ID: %s", shared.Config.TeamID)
		return errors.New("already linked to a team")
	}

	if lo.IsNotEmpty(teamURL) {
		inviteID, err := getInviteIDFromURL(teamURL)
		if err != nil {
			log.WithError(err).Warn("failed to get invite ID from URL")
			return err
		}

		authToken, teamID, err := team.EnrollDevice(inviteID)
		if err != nil {
			log.WithError(err).Warn("failed to enroll device")
			return err
		}

		shared.Config.TeamID = teamID
		shared.Config.AuthToken = authToken

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

		log.Infof("Device successfully linked to team: %s", teamID)
	}
	return nil
}

func getInviteIDFromURL(teamURL string) (string, error) {
	parsedURL, err := url.Parse(teamURL)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	inviteID := queryParams.Get("invite_id")
	if inviteID == "" {
		return "", fmt.Errorf("invite_id not found in URL")
	}

	return inviteID, nil
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
