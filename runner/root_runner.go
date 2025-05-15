package runner

import (
	"encoding/json"
	"net"

	"github.com/ParetoSecurity/agent/claims"
	"github.com/caarlos0/log"
)

type CheckStatus struct {
	UUID    string `json:"uuid"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

// handleConnection handles an incoming network connection.
// It reads input from the connection, processes the input to run checks,
// and sends back the status of the checks as a JSON response.
//
// The input is expected to be a JSON object containing a "uuid" key.
// The function will look for checks that are runnable, require root,
// and match the provided UUID. It will run those checks and collect their status.
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	log.Info("Connection received")

	// Read input from connection
	decoder := json.NewDecoder(conn)
	var input map[string]string
	if err := decoder.Decode(&input); err != nil {
		log.Debugf("Failed to decode input: %v\n", err)
		return
	}
	uuid, ok := input["uuid"]
	if !ok {
		log.Debugf("UUID not found in input")
		return
	}
	log.Debugf("Received UUID: %s", uuid)

	status := &CheckStatus{
		UUID:    uuid,
		Passed:  false,
		Details: "Check not found",
	}
	for _, claim := range claims.All {
		for _, chk := range claim.Checks {
			if chk.RequiresRoot() && uuid == chk.UUID() {
				log.Infof("Running check %s\n", chk.UUID())
				if chk.Run() != nil {
					log.Warnf("Failed to run check %s\n", chk.UUID())
					continue
				}
				log.Infof("Check %s completed\n", chk.UUID())
				status.Passed = chk.Passed()
				status.Details = chk.Status()
				log.Infof("Check %s status: %v\n", chk.UUID(), status.Passed)
			}
		}
	}

	// Handle the request
	response, err := json.Marshal(status)
	if err != nil {
		log.Debugf("Failed to marshal response: %v\n", err)
		return
	}
	if _, err = conn.Write(response); err != nil {
		log.Debugf("Failed to write to connection: %v\n", err)
	}
}

// IsRootHelperRunning checks if the root helper service is running when needed.
func IsRootHelperRunning(claimsTorun []claims.Claim) bool {
	for _, claim := range claimsTorun {
		for _, chk := range claim.Checks {
			if chk.RequiresRoot() {
				conn, err := net.Dial("unix", SocketPath)
				if err != nil {
					log.WithError(err).Warn("Failed to connect to root helper")
					return false
				}
				defer conn.Close()
				return true
			}
		}
	}
	log.Debug("No checks require root")
	return true
}

// RunCheckViaRoot connects to a Unix socket, sends a UUID, and receives a boolean status.
// It is used to execute a check with root privileges via a helper process.
// The function establishes a connection to the socket specified by SocketPath,
// sends the UUID as a JSON-encoded string, and then decodes the JSON response
// to determine the status of the check. It returns the boolean status associated
// with the UUID and any error encountered during the process.
func RunCheckViaRoot(uuid string) (*CheckStatus, error) {

	rateLimitCall.Take()
	log.WithField("uuid", uuid).Debug("Running check via root helper")

	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.WithError(err).Warn("Failed to connect to root helper")
		return &CheckStatus{}, err
	}
	defer conn.Close()

	// Send UUID
	input := map[string]string{"uuid": uuid}
	encoder := json.NewEncoder(conn)
	log.WithField("input", input).Debug("Sending input to helper")
	if err := encoder.Encode(input); err != nil {
		log.WithError(err).Warn("Failed to encode JSON")
		return &CheckStatus{}, err
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var status = &CheckStatus{}
	if err := decoder.Decode(status); err != nil {
		log.WithError(err).Warn("Failed to decode JSON")
		return &CheckStatus{}, err
	}
	log.WithField("status", status).Debug("Received status from helper")
	return status, nil
}
