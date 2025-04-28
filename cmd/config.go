package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure application settings",
	Long:  "Configure application settings, such as enabling or disabling specific checks.",
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration settings",
	Long:  "Reset configuration settings to their default values.",
	Run: func(cmd *cobra.Command, args []string) {
		shared.ResetConfig()
		log.WithField("config", shared.ConfigPath).Info("Configuration reset to default values.")
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable [check UUID]",
	Short: "Enable a specific check",
	Long:  "Enable a specific check by providing its id.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		check := args[0]
		err := shared.EnableCheck(check)
		if err != nil {
			log.WithError(err).Fatalf("Failed to enable check: %s", check)
		} else {
			log.WithField("check", check).Info("Check enabled successfully.")
		}
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable [check UUID]",
	Short: "Disable a specific check",
	Long:  "Disable a specific check by providing its id.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		check := args[0]
		err := shared.DisableCheck(check)
		if err != nil {
			log.WithError(err).Fatalf("Failed to disable check: %s", check)
		} else {
			log.WithField("check", check).Info("Check disabled successfully.")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(resetCmd)
	configCmd.AddCommand(enableCmd)
	configCmd.AddCommand(disableCmd)
}
