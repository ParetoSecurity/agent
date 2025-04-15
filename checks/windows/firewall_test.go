//go:build windows
// +build windows

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirewall_Name(t *testing.T) {
	f := &Firewall{}
	assert.Equal(t, "Windows Firewal is enabled", f.Name())
}

func TestFirewall_UUID(t *testing.T) {
	f := &Firewall{}
	assert.Equal(t, "b7e2e1c2-8e2a-4c1a-9e2b-ffb2e1c2a1b2", f.UUID())
}

func TestFirewall_PassedMessage(t *testing.T) {
	f := &Firewall{}
	assert.Equal(t, "Firewall is on", f.PassedMessage())
}

func TestFirewall_FailedMessage(t *testing.T) {
	f := &Firewall{}
	assert.Equal(t, "Firewall is off", f.FailedMessage())
}

func TestFirewall_IsRunnable(t *testing.T) {
	f := &Firewall{}
	assert.True(t, f.IsRunnable())
}

func TestFirewall_RequiresRoot(t *testing.T) {
	f := &Firewall{}
	assert.True(t, f.RequiresRoot())
}

func TestFirewall_Status(t *testing.T) {
	f := &Firewall{passed: true}
	assert.Equal(t, "Firewall is on", f.Status())
	f.passed = false
	assert.Equal(t, "Firewall is off", f.Status())
}

func TestFirewall_Run(t *testing.T) {
	orig := checkFirewallProfile
	defer func() { checkFirewallProfile = orig }()

	checkFirewallProfile = func(profile string) bool {
		if profile == "PublicProfile" {
			return true
		}
		if profile == "PrivateProfile" {
			return true
		}
		return false
	}

	f := &Firewall{}
	assert.NoError(t, f.Run())
	assert.True(t, f.Passed())

	checkFirewallProfile = func(profile string) bool { return false }
	f = &Firewall{}
	assert.NoError(t, f.Run())
	assert.False(t, f.Passed())
}
