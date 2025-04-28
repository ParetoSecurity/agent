package cmd

import (
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink this device from the team",
	Run: func(cc *cobra.Command, args []string) {
		log.Info("Unlinking device ...")
		shared.Config.TeamID = ""
		shared.Config.AuthToken = ""
		if err := shared.SaveConfig(); err != nil {
			log.WithError(err).Fatal("failed to save config")
		}
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}
