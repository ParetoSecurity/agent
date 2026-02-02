package claims

import (
	"github.com/ParetoSecurity/agent/check"
	shared "github.com/ParetoSecurity/agent/checks/shared"
	checks "github.com/ParetoSecurity/agent/checks/windows"
)

var All = []Claim{
	{"Access Security", []check.Check{
		&checks.PasswordManagerCheck{},
		&checks.ScreensaverTimeout{},
		&checks.ScreensaverPassword{},
	}},
	{"Application Updates", []check.Check{
		&shared.ParetoUpdated{},
		&checks.AutomaticUpdatesCheck{},
	}},
	{"Firewall & Sharing", []check.Check{
		&shared.RemoteLogin{},
		&checks.WindowsFirewall{},
	}},
	{"System Integrity", []check.Check{
		&checks.WindowsDefender{},
		&checks.DiskEncryption{},
	}},
}
