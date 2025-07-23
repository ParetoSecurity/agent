package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ParetoSecurity/agent/notify"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/ParetoSecurity/agent/team"
	"github.com/caarlos0/log"
	"github.com/carlmjohnson/requests"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var rsaPublicKey = `
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwGh64DK49GOq1KX+ojyg
Y9JSAZ4cfm5apavetQ42D2gTjfhDu1kivrDRwhjqj7huUWRI2ExMdMHp8CzrJI3P
zpzutEUXTEHloe0vVMZqPoP/r2f1cl4bmDkFZyHr6XTgiYPE4GgMjxUc04J2ksqU
/XbNwOVsBiuy1T2BduLYiYr1UyIx8VqEb+3tunQKlyRKF7a5LoEZatt5F/5vaMMI
4zp1yIc2PMoBdlBH4/tpJmC/PiwjBuwgp5gMIle4Hy7zwW4+rIJzF5P3Tg+Am+Lg
davB8TIZDBlqIWV7zK1kWBPj364a5cnaUP90BnOriMJBh7zPG0FNGTXTiJED2qDM
fajDrji3oAPO24mJsCCzSd8LIREK5c6iAf1X4UI/UFP+UhOBCsANrhNSXRpO2KyM
+60JYzFpMvyhdK9zMo7Tc+KM6R0YRNmBCYK/ePAGk3WU6qxN5+OmSjdTvFrqC4JQ
FyK51WJI80PKvp3B7ZB7XpH5B24wr/OhMRh5YZOcrpuBykfHaMozkDCudgaj/V+x
K79CqMF/BcSxCSBktWQmabYCM164utpmJaCSpZyDtKA4bYVv9iRCGTqFQT7jX+/h
Z37gmg/+TlIdTAeB5TG2ffHxLnRhT4AAhUgYmk+QP3a1hxP5xj2otaSTZ3DxQd6F
ZaoGJg3y8zjrxYBQDC8gF6sCAwEAAQ==
`

type InviteClaims struct {
	TeamAuth string `json:"token"`
	TeamUUID string `json:"teamID"`
	jwt.RegisteredClaims
}

type TeamInviteRequest struct {
	Email string `json:"email"`
}

type TeamTicketResponse struct {
	Token string `json:"token"`
	Link  string `json:"link"`
}

var linkCmd = &cobra.Command{
	Use:   "link <url_or_team_uuid> [email]",
	Short: "Link this device to a team",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if shared.IsRoot() {
			return fmt.Errorf("please run this command as a normal user")
		}

		var err error
		if len(args) == 1 {
			// Original URL format
			err = runLinkCommandWithURL(args[0])
		} else {
			// New team_uuid + email format
			token, err := runLinkCommandWithCredentials(args[0], args[1])
			if err != nil {
				return err
			}
			// Use the received token with the URL-based flow
			err = runLinkCommandWithURL(token)
		}

		if err != nil {
			log.WithError(err).Error("Failed to link team")
			notify.Toast("Failed to add device to the team!")
			return err
		}
		notify.Toast("Device successfully linked to the team!")
		return nil
	},
}

func runLinkCommandWithURL(teamURL string) error {
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
		token, err := getTokenFromURL(teamURL)
		if err != nil {
			log.WithError(err).Warn("failed to get token from URL")
			return err
		}

		parsedToken, err := parseJWT(token)
		if err != nil {
			log.WithError(err).Warn("failed to parse JWT")
			return err
		}

		shared.Config.TeamID = parsedToken.TeamUUID
		shared.Config.AuthToken = parsedToken.TeamAuth

		err = team.ReportToTeam(true)
		if err != nil {
			log.WithError(err).Warn("failed to link to team")
			return err
		}

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

		log.Infof("Device successfully linked to team: %s", parsedToken.TeamUUID)
	}
	return nil
}

func runLinkCommandWithCredentials(teamUUID, email string) (string, error) {
	if lo.IsEmpty(teamUUID) {
		log.Warn("Please provide a team UUID")
		return "", errors.New("no team UUID provided")
	}
	if lo.IsEmpty(email) {
		log.Warn("Please provide an email")
		return "", errors.New("no email provided")
	}
	if shared.IsLinked() {
		log.Warn("Already linked to a team")
		log.Warn("Unlink first with `paretosecurity unlink`")
		log.Infof("Team ID: %s", shared.Config.TeamID)
		return "", errors.New("already linked to a team")
	}

	token, err := linkAccount(teamUUID, email)
	if err != nil {
		log.WithError(err).Warn("failed to get invite token")
		return "", err
	}

	log.Infof("Successfully received invite token for team: %s", teamUUID)
	return token, nil
}

func linkAccount(teamUUID, email string) (string, error) {
	const inviteURL = "https://cloud.paretosecurity.com"

	inviteRequest := TeamInviteRequest{
		Email: email,
	}

	var ticketResponse TeamTicketResponse
	var errRes string

	// Create a context with a timeout for the request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.WithField("teamUUID", teamUUID).
		WithField("email", email).
		Debug("Getting invite ticket for team")

	err := requests.URL(inviteURL).
		Pathf("/team/%s/invite/public", teamUUID).
		Method(http.MethodPost).
		Header("User-Agent", shared.UserAgent()).
		BodyJSON(&inviteRequest).
		ToJSON(&ticketResponse).
		AddValidator(
			requests.ValidatorHandler(
				requests.DefaultValidator,
				requests.ToString(&errRes),
			)).
		Fetch(ctx)

	if err != nil {
		log.WithField("response", errRes).
			WithError(err).
			Warnf("Failed to get invite ticket for team: %s", teamUUID)
		return "", err
	}

	log.WithField("token", ticketResponse.Token).
		WithField("link", ticketResponse.Link).
		Debug("Received invite ticket")

	log.Infof("Successfully received invite link for team: %s", teamUUID)
	return ticketResponse.Link, nil
}

func getTokenFromURL(teamURL string) (string, error) {

	parsedURL, err := url.Parse(teamURL)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	token := queryParams.Get("token")
	if token == "" {
		return "", fmt.Errorf("token not found in URL")
	}

	return token, nil
}

func parseJWT(token string) (*InviteClaims, error) {
	jwttToken, _ := jwt.ParseWithClaims(token, &InviteClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(strings.ReplaceAll(rsaPublicKey, "\n", "")), nil
	})
	if claims, ok := jwttToken.Claims.(*InviteClaims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("failed to parse JWT")
}

func init() {
	rootCmd.AddCommand(linkCmd)
}
