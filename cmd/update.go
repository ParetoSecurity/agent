//go:build windows

package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Pareto Security Agent",
	Long:  `Update the Pareto Security Agent to the latest version.`,
	Run: func(cmd *cobra.Command, args []string) {
		updateCommand()
	},
}

// UpdateConfig holds the configuration for the update command
type UpdateConfig struct {
	UpdateApp func() error
}

// DefaultUpdateConfig returns the default configuration
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		UpdateApp: shared.UpdateApp,
	}
}

func updateCommand() {
	config := DefaultUpdateConfig()
	runUpdateCommand(config)
}

func runUpdateCommand(config *UpdateConfig) {
	config.UpdateApp()
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
