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
			notify.Blocking("Failed to add device to the team!")
			return err
		}
		notify.Blocking("Device successfully linked to the team!")
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
			log.WithError(err).Warn("failed to report to team")
			return err
		}

		err = shared.SaveConfig()
		if err != nil {
			log.Errorf("Error saving config: %v", err)
			return err
		}
		log.Infof("Device successfully linked to team: %s", parsedToken.TeamUUID)

	}
	return nil
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
