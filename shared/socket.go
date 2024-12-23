package shared

import (
	"encoding/json"
	"net"
	"os/exec"

	"github.com/caarlos0/log"
	"go.uber.org/ratelimit"
)

var SocketPath = "/var/run/pareto-linux.sock"
var rateLimitCall = ratelimit.New(1)

func IsSocketServicePresent() bool {
	cmd := exec.Command("systemctl", "is-enabled", "--quiet", "pareto-linux.socket")
	err := cmd.Run()
	return err == nil
}

func RunCheckViaHelper(uuid string) (bool, error) {

	rateLimitCall.Take()
	log.WithField("uuid", uuid).Debug("Running check via helper")

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.WithError(err).Warn("Failed to connect to helper")
		return false, err
	}
	defer conn.Close()

	// Send UUID
	input := map[string]string{"uuid": uuid}
	encoder := json.NewEncoder(conn)
	log.WithField("input", input).Debug("Sending input to helper")
	if err := encoder.Encode(input); err != nil {
		log.WithError(err).Warn("Failed to encode JSON")
		return false, err
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var status map[string]bool
	if err := decoder.Decode(&status); err != nil {
		log.WithError(err).Warn("Failed to decode JSON")
		return false, err
	}
	log.WithField("status", status).Debug("Received status from helper")
	return status[uuid], nil
}
