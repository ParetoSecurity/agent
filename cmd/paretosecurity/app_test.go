package main

import (
	"errors"
	"testing"
)

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name          string
		loadConfigErr error
		isRoot        bool
		expectLogWarn bool
		expectExecute bool
	}{
		{
			name:          "successful run with config loaded",
			loadConfigErr: nil,
			isRoot:        false,
			expectExecute: true,
		},
		{
			name:          "config load error, not root",
			loadConfigErr: errors.New("config error"),
			isRoot:        false,
			expectLogWarn: true,
			expectExecute: true,
		},
		{
			name:          "config load error, is root",
			loadConfigErr: errors.New("config error"),
			isRoot:        true,
			expectLogWarn: false,
			expectExecute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executeCalled bool

			config := &AppConfig{
				LoadConfig: func() error {
					return tt.loadConfigErr
				},
				IsRoot: func() bool {
					return tt.isRoot
				},
				Execute: func() {
					executeCalled = true
				},
			}

			app := NewApp(config)
			err := app.Run()

			if err != nil {
				t.Errorf("App.Run() returned error: %v", err)
			}

			if executeCalled != tt.expectExecute {
				t.Errorf("Execute called = %v, expected = %v", executeCalled, tt.expectExecute)
			}
		})
	}
}

func TestNewApp(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		app := NewApp(nil)
		if app.config == nil {
			t.Error("expected default config to be set")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &AppConfig{}
		app := NewApp(config)
		if app.config != config {
			t.Error("expected custom config to be set")
		}
	})
}

func TestDefaultAppConfig(t *testing.T) {
	config := DefaultAppConfig()

	if config.LoadConfig == nil {
		t.Error("LoadConfig should not be nil")
	}
	if config.IsRoot == nil {
		t.Error("IsRoot should not be nil")
	}
	if config.Execute == nil {
		t.Error("Execute should not be nil")
	}
}
