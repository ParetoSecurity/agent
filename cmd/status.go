package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the status of the checks",
	Run: func(cmd *cobra.Command, args []string) {
		statusCommand()
	},
}

// StatusConfig holds the configuration for the status command
type StatusConfig struct {
	PrintStates func()
}

// DefaultStatusConfig returns the default configuration
func DefaultStatusConfig() *StatusConfig {
	return &StatusConfig{
		PrintStates: shared.PrintStates,
	}
}

func statusCommand() {
	config := DefaultStatusConfig()
	runStatusCommand(config)
}

func runStatusCommand(config *StatusConfig) {
	config.PrintStates()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
