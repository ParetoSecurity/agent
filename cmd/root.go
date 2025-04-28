package cmd

import (
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:     "paretosecurity --help --version [command]",
	Short:   "Pareto Security CLI",
	Version: shared.Version,
	Long:    `Pareto Security CLI is a tool for running and reporting audits to paretosecurity.com.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "output verbose logs")
}

func Execute() {
	if rootCmd.Execute() != nil {
		log.Fatal("Failed to execute command")
	}
}
