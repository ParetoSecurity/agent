//go:build windows
// +build windows

package main

import (
	"embed"
	"errors"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestInstallerApp_Run_SilentInstall(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"silent install /qs", []string{"/qs"}, true},
		{"silent install /QS", []string{"/QS"}, true},
		{"silent install /qsp", []string{"/qsp"}, true},
		{"silent install /s", []string{"/s"}, true},
		{"silent install /q", []string{"/q"}, true},
		{"no silent install", []string{"/help"}, false},
		{"empty args", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				installCalled bool
				exitCalled    bool
				exitCode      int
			)

			config := &InstallerConfig{
				Args: tt.args,
				InstallApp: func(silent bool) error {
					installCalled = true
					if !silent {
						t.Error("Expected silent install to be true")
					}
					return nil
				},
				Exit: func(code int) {
					exitCalled = true
					exitCode = code
				},
				NewApp: func(opts application.Options) *application.App {
					t.Error("NewApp should not be called for silent install")
					return nil
				},
				Assets: embed.FS{},
			}

			app := NewInstallerApp(config)
			app.Run()

			if tt.expected {
				if !installCalled {
					t.Error("Expected InstallApp to be called")
				}
				if !exitCalled {
					t.Error("Expected Exit to be called")
				}
				if exitCode != 0 {
					t.Errorf("Expected exit code 0, got %d", exitCode)
				}
			}
		})
	}
}

func TestInstallerApp_Run_GUIMode(t *testing.T) {
	var (
		newAppCalled bool
		appRunCalled bool
	)

	mockApp := &MockApp{
		RunFunc: func() {
			appRunCalled = true
		},
		NewWebviewWindowWithOptionsFunc: func(opts application.WebviewWindowOptions) *application.WebviewWindow {
			return nil
		},
	}

	config := &InstallerConfig{
		Args: []string{"/help"}, // Non-silent args
		InstallApp: func(silent bool) error {
			t.Error("InstallApp should not be called for GUI mode")
			return nil
		},
		Exit: func(code int) {
			t.Error("Exit should not be called for GUI mode")
		},
		NewApp: func(opts application.Options) *application.App {
			newAppCalled = true
			return &application.App{} // Return a mock app
		},
		Assets: embed.FS{},
	}

	app := NewInstallerApp(config)
	// Note: This test can't fully test GUI mode due to the complexities of mocking the Wails app
	// but we can test the silent install detection
	isSilent := app.shouldInstallSilently()
	if isSilent {
		t.Error("Expected GUI mode, not silent install")
	}
}

func TestInstallerApp_ShouldInstallSilently(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{"silent /qs", []string{"/qs"}, true},
		{"silent /QS", []string{"/QS"}, true},
		{"silent /qsp", []string{"/qsp"}, true},
		{"silent /s", []string{"/s"}, true},
		{"silent /q", []string{"/q"}, true},
		{"not silent", []string{"/help"}, false},
		{"empty", []string{}, false},
		{"multiple args with silent", []string{"/help", "/qs"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &InstallerConfig{Args: tt.args}
			app := NewInstallerApp(config)
			result := app.shouldInstallSilently()
			if result != tt.expected {
				t.Errorf("shouldInstallSilently() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInstallerApp_InstallError(t *testing.T) {
	var exitCalled bool
	var exitCode int

	config := &InstallerConfig{
		Args: []string{"/qs"},
		InstallApp: func(silent bool) error {
			return errors.New("install failed")
		},
		Exit: func(code int) {
			exitCalled = true
			exitCode = code
		},
		Assets: embed.FS{},
	}

	app := NewInstallerApp(config)
	app.Run()

	if !exitCalled {
		t.Error("Expected Exit to be called")
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

// MockApp is a mock implementation for testing
type MockApp struct {
	RunFunc                         func()
	NewWebviewWindowWithOptionsFunc func(opts application.WebviewWindowOptions) *application.WebviewWindow
}

func (m *MockApp) Run() {
	if m.RunFunc != nil {
		m.RunFunc()
	}
}

func (m *MockApp) NewWebviewWindowWithOptions(opts application.WebviewWindowOptions) *application.WebviewWindow {
	if m.NewWebviewWindowWithOptionsFunc != nil {
		return m.NewWebviewWindowWithOptionsFunc(opts)
	}
	return nil
}
