package runner

import (
	"github.com/ParetoSecurity/agent/shared"
	"go.uber.org/ratelimit"
)

var SocketPath = "/run/paretosecurity.sock"
var rateLimitCall = ratelimit.New(1)

func IsSocketServicePresent() bool {
	_, err := shared.RunCommand("systemctl", "is-enabled", "--quiet", "paretosecurity.socket")
	return err == nil
}
