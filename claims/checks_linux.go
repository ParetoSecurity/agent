package claims

import (
	"github.com/ParetoSecurity/agent/check"
	checks "github.com/ParetoSecurity/agent/checks/linux"
	shared "github.com/ParetoSecurity/agent/checks/shared"
)

var All = []Claim{
	{"Access Security", []check.Check{
		&checks.Autologin{},
		&checks.DockerAccess{},
		&checks.PasswordToUnlock{},
		&shared.SSHKeys{},
		&shared.SSHKeysAlgo{},
		&checks.PasswordManagerCheck{},
	}},
	{"Application Updates", []check.Check{
		&checks.ApplicationUpdates{},
		&shared.ParetoUpdated{},
	}},
	{"Firewall & Sharing", []check.Check{
		&checks.Firewall{},
		&checks.Printer{},
		&shared.RemoteLogin{},
		&checks.Sharing{},
	}},
	{"System Integrity", []check.Check{
		&checks.SecureBoot{},
		&checks.EncryptingFS{},
		&shared.TeamReportSentCheck{},
	}},
}
