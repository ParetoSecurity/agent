//go:build windows
// +build windows

package main

import (
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

//go:embed ui/dist/*
var linkAssets embed.FS

func main() {
	// Check for link command
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "link" {
		if len(os.Args) > 2 {
			handleLinkCommand(os.Args[2])
			return
		}
		log.Error("No URL provided for link command")
		os.Exit(1)
		return
	}

	// Normal tray app execution
	lockDir, _ := shared.UserHomeDir()
	if err := shared.OnlyInstance(filepath.Join(lockDir, ".paretosecurity-tray.lock")); err != nil {
		log.WithError(err).Fatal("An instance of ParetoSecurity tray application is already running.")
		return
	}

	app := NewTrayApp(nil)
	app.Run()
}

// handleLinkCommand handles the paretosecurity:// URL protocol
func handleLinkCommand(linkURL string) {
	log.WithField("url", linkURL).Info("Handling link command")

	// Parse the enrollment URL
	inviteID, host, err := parseEnrollmentURL(linkURL)
	if err != nil {
		log.WithError(err).Error("Failed to parse enrollment URL")
		// Still create the app to show an error
		createLinkApp("", "").Run()
		return
	}

	// Create and run the link app
	app := createLinkApp(inviteID, host)
	app.Run()
}

// parseEnrollmentURL parses the paretosecurity:// URL to extract invite_id and host
func parseEnrollmentURL(enrollURL string) (inviteID, host string, err error) {
	// Expected format: paretosecurity://linkDevice/?invite_id=<ID>&host=<optional>
	// Also handles: paretosecurity://linkDevice?invite_id=<ID>&host=<optional>
	parsedURL, err := url.Parse(enrollURL)
	if err != nil {
		return "", "", err
	}

	// The URL might have the query params in RawQuery or in the path after the slash
	queryParams := parsedURL.Query()

	// If no query params found, try parsing from the path (handles linkDevice/ case)
	if len(queryParams) == 0 && parsedURL.RawQuery == "" && strings.Contains(enrollURL, "?") {
		// Handle case where URL parser might not properly parse the query
		parts := strings.Split(enrollURL, "?")
		if len(parts) == 2 {
			values, _ := url.ParseQuery(parts[1])
			queryParams = values
		}
	}

	inviteID = queryParams.Get("invite_id")
	if inviteID == "" {
		return "", "", fmt.Errorf("invite_id not found in URL")
	}

	// Extract optional host parameter
	host = queryParams.Get("host")

	return inviteID, host, nil
}
