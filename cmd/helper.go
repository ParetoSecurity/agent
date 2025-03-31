package cmd

import (
	"net"
	"os"

	"github.com/ParetoSecurity/agent/runner"
	shared "github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// runHelperServer listens on a socket (passed via file descriptor 0) and handles incoming connections.
// It's designed to be run in a systemd context where systemd provides the socket.
// The server accepts a single connection, handles it using runner.HandleConnection, and then exits.
// It logs the socket path and version information upon startup and logs any errors encountered during socket creation or connection acceptance.
func runHelperServer() {
	// Get the socket from file descriptor 0
	file := os.NewFile(0, "socket")
	listener, err := net.FileListener(file)
	if err != nil {
		log.WithError(err).Fatal("Failed to create listener, not running in systemd context")
	}
	defer listener.Close()
	log.WithField("socket", runner.SocketPath).WithField("version", shared.Version).Info("Listening on socket")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithError(err).Warn("Failed to accept connection")
			continue
		}

		runner.HandleConnection(conn)
		break
	}
}

var helperCmd = &cobra.Command{
	Use:   "helper [--socket]",
	Short: "A root helper",
	Long:  `A root helper that listens on a Unix domain socket and responds to authenticated requests.`,
	Run: func(cmd *cobra.Command, args []string) {

		socketFlag, _ := cmd.Flags().GetString("socket")
		if lo.IsNotEmpty(socketFlag) {
			runner.SocketPath = socketFlag
		}

		runHelperServer()
	},
}

func init() {
	rootCmd.AddCommand(helperCmd)
	helperCmd.Flags().Bool("socket", false, "socket path")
}
