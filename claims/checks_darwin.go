package claims

import (
	"github.com/ParetoSecurity/pareto-core/check"
	shared "github.com/ParetoSecurity/pareto-core/checks/shared"
	checks "github.com/ParetoSecurity/pareto-core/checks/darwin"
)

var All = []Claim{
	{"Access Security", []check.Check{
		check.Register(&shared.SSHKeys{}),
		check.Register(&shared.SSHKeysAlgo{}),
		check.Register(&checks.PasswordManagerCheck{}),
	}},
	{"Software Updates", []check.Check{
		check.Register(&shared.ParetoUpdated{}),
	}},
	{"Firewall & Sharing", []check.Check{
		check.Register(&shared.RemoteLogin{}),
	}},
	{"System Integrity", []check.Check{}},
}
