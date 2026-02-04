package claims

import (
	"github.com/ParetoSecurity/agent/check"
	checks "github.com/ParetoSecurity/agent/checks/darwin"
	shared "github.com/ParetoSecurity/agent/checks/shared"
)

var All = []Claim{
	{"Access Security", []check.Check{
		&shared.SSHKeys{},
		&shared.SSHKeysAlgo{},
		&checks.PasswordManagerCheck{},
	}},
	{"Application Updates", []check.Check{
		&shared.ParetoUpdated{},
	}},
	{"Firewall & Sharing", []check.Check{
		&shared.RemoteLogin{},
	}},
	{"System Integrity", []check.Check{
		&shared.TeamReportSentCheck{},
	}},
}
