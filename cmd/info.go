package cmd

import (
	"encoding/json"
	"runtime"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/elastic/go-sysinfo"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print the system information",
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("%s@%s %s", shared.Version, shared.Commit, shared.Date)
		log.Infof("Built with %s", runtime.Version())
		log.Infof("Team: %s\n", shared.Config.TeamID)

		device := shared.CurrentReportingDevice()
		jsonOutput, err := json.MarshalIndent(device, "", "  ")
		if err != nil {
			log.Warn("Failed to marshal host info")
		}
		log.Infof("Device Info: %s\n", string(jsonOutput))

		hostInfo, err := sysinfo.Host()
		if err != nil {
			log.Warn("Failed to get process information")
		}
		envInfo := hostInfo.Info()
		envInfo.IPs = []string{}  // Exclude IPs for privacy
		envInfo.MACs = []string{} // Exclude MACs for privacy
		jsonOutput, err = json.MarshalIndent(envInfo, "", "  ")
		if err != nil {
			log.Warn("Failed to marshal host info")
		}
		log.Infof("Host Info: %s\n", string(jsonOutput))
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
