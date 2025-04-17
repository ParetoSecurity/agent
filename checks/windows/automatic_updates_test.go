package checks

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ParetoSecurity/agent/shared"
)

func TestAutomaticUpdatesCheck_Run_AutoUpdatesDisabled(t *testing.T) {
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "powershell",
			Args:    []string{"-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json"},
			Out:     `{"NotificationLevel":1}`,
			Err:     nil,
		},
	}
	a := &AutomaticUpdatesCheck{}
	err := a.Run()
	if err != nil {
		t.Fatal(err)
	}
	if a.passed {
		t.Error("expected passed=false when NotificationLevel==1")
	}
	if a.status != "Automatic Updates are disabled" {
		t.Errorf("unexpected status: %s", a.status)
	}
}

func TestAutomaticUpdatesCheck_Run_QueryError(t *testing.T) {
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "powershell",
			Args:    []string{"-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json"},
			Out:     "",
			Err:     errors.New("fail"),
		},
	}
	a := &AutomaticUpdatesCheck{}
	_ = a.Run()
	if a.passed {
		t.Error("expected passed=false on query error")
	}
	if a.status != "Failed to query update settings" {
		t.Errorf("unexpected status: %s", a.status)
	}
}

func TestAutomaticUpdatesCheck_Run_ParseError(t *testing.T) {
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "powershell",
			Args:    []string{"-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json"},
			Out:     "notjson",
			Err:     nil,
		},
	}
	a := &AutomaticUpdatesCheck{}
	_ = a.Run()
	if a.passed {
		t.Error("expected passed=false on parse error")
	}
	if a.status != "Failed to parse update settings" {
		t.Errorf("unexpected status: %s", a.status)
	}
}

func TestAutomaticUpdatesCheck_Run_UpdatesPaused(t *testing.T) {
	now := time.Now().Unix()
	expiry := now + 3600
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "powershell",
			Args:    []string{"-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json"},
			Out:     `{"NotificationLevel":4}`,
			Err:     nil,
		},
		{
			Command: "powershell",
			Args:    []string{"-Command", `try { Get-ItemPropertyValue -Path "HKLM:\SOFTWARE\Microsoft\WindowsUpdate\UX\Settings" -Name "PauseUpdatesExpiryTime" } catch { 0 }`},
			Out:     fmt.Sprintf("%d", expiry),
			Err:     nil,
		},
	}
	a := &AutomaticUpdatesCheck{}
	_ = a.Run()
	if a.passed {
		t.Error("expected passed=false when updates are paused")
	}
	if a.status != "Updates are paused" {
		t.Errorf("unexpected status: %s", a.status)
	}
}

func TestAutomaticUpdatesCheck_Run_UpdatesEnabled(t *testing.T) {
	now := time.Now().Unix()
	expiry := now - 3600 // expired
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "powershell",
			Args:    []string{"-Command", "(New-Object -ComObject Microsoft.Update.AutoUpdate).Settings | ConvertTo-Json"},
			Out:     `{"NotificationLevel":4}`,
			Err:     nil,
		},
		{
			Command: "powershell",
			Args:    []string{"-Command", `try { Get-ItemPropertyValue -Path "HKLM:\SOFTWARE\Microsoft\WindowsUpdate\UX\Settings" -Name "PauseUpdatesExpiryTime" } catch { 0 }`},
			Out:     fmt.Sprintf("%d", expiry),
			Err:     nil,
		},
	}
	a := &AutomaticUpdatesCheck{}
	_ = a.Run()
	if !a.passed {
		t.Error("expected passed=true when updates enabled and not paused")
	}
	if a.status != "" && a.status != a.PassedMessage() {
		t.Errorf("unexpected status: %s", a.status)
	}
}
