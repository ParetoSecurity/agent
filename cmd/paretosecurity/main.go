// Package main provides the entry point for the application.
package main

import (
	"github.com/ParetoSecurity/agent/cmd"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
)

func main() {
	if err := shared.LoadConfig(); err != nil {
		if !shared.IsRoot() {
			log.WithError(err).Warn("failed to load config")
		}
	}
	cmd.Execute()
}
