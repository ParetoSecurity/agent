//go:build windows
// +build windows

package main

import (
	"errors"
	"testing"
	"time"
)

func TestTrayApp_Run(t *testing.T) {
	var (
		exitCalled    bool
		systrayRun    bool
		loadConfigErr error
	)

	config := &TrayAppConfig{
		LoadConfig: func() error {
			return loadConfigErr
		},
		GetModifiedTime: func() time.Time {
			return time.Now().Add(-2 * time.Hour)
		},
		RunCommand: func(command string, args ...string) (string, error) {
			return "", nil
		},
		SelfExe: func() string {
			return "test.exe"
		},
		SystrayRun: func(onReady func(), onExit func()) {
			systrayRun = true
			onExit() // Trigger exit for testing
		},
		Exit: func(code int) {
			exitCalled = true
		},
		Sleep: func(duration time.Duration) {
			// No-op for testing
		},
		Rand: func(n int) int {
			return 5
		},
	}

	app := NewTrayApp(config)
	app.Run()

	if !systrayRun {
		t.Error("Expected systray.Run to be called")
	}

	if !exitCalled {
		t.Error("Expected exit to be called")
	}
}

func TestTrayApp_RunInitialCheck(t *testing.T) {
	var commandCalled bool

	config := &TrayAppConfig{
		RunCommand: func(command string, args ...string) (string, error) {
			commandCalled = true
			if len(args) != 1 || args[0] != "check" {
				t.Errorf("Expected check command, got %v", args)
			}
			return "", nil
		},
		SelfExe: func() string {
			return "test.exe"
		},
	}

	app := NewTrayApp(config)
	app.runInitialCheck()

	if !commandCalled {
		t.Error("Expected command to be called")
	}
}

func TestTrayApp_RunInitialCheckWithError(t *testing.T) {
	config := &TrayAppConfig{
		RunCommand: func(command string, args ...string) (string, error) {
			return "", errors.New("command failed")
		},
		SelfExe: func() string {
			return "test.exe"
		},
	}

	app := NewTrayApp(config)
	// Should not panic even with error
	app.runInitialCheck()
}

func TestNewTrayApp(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		app := NewTrayApp(nil)
		if app.config == nil {
			t.Error("expected default config to be set")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &TrayAppConfig{}
		app := NewTrayApp(config)
		if app.config != config {
			t.Error("expected custom config to be set")
		}
	})
}

func TestDefaultTrayAppConfig(t *testing.T) {
	config := DefaultTrayAppConfig()

	if config.LoadConfig == nil {
		t.Error("LoadConfig should not be nil")
	}
	if config.GetModifiedTime == nil {
		t.Error("GetModifiedTime should not be nil")
	}
	if config.RunCommand == nil {
		t.Error("RunCommand should not be nil")
	}
	if config.SelfExe == nil {
		t.Error("SelfExe should not be nil")
	}
	if config.SystrayRun == nil {
		t.Error("SystrayRun should not be nil")
	}
	if config.Exit == nil {
		t.Error("Exit should not be nil")
	}
	if config.Sleep == nil {
		t.Error("Sleep should not be nil")
	}
	if config.Rand == nil {
		t.Error("Rand should not be nil")
	}
}
