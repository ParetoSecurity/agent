//go:build linux || darwin
// +build linux darwin

package cmd

import (
	"fmt"
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestCheckStatusNotifierSupport(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      error
		expectedResult bool
	}{
		{
			name:           "KDE StatusNotifierWatcher available",
			mockOutput:     `string "org.kde.StatusNotifierWatcher"`,
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "Freedesktop StatusNotifierWatcher available",
			mockOutput:     `string "org.freedesktop.StatusNotifierWatcher"`,
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "Both StatusNotifierWatcher implementations available",
			mockOutput:     `string "org.kde.StatusNotifierWatcher" string "org.freedesktop.StatusNotifierWatcher"`,
			mockError:      nil,
			expectedResult: true,
		},
		{
			name:           "No StatusNotifierWatcher available",
			mockOutput:     `string "org.freedesktop.DBus" string "org.freedesktop.PowerManagement"`,
			mockError:      nil,
			expectedResult: false,
		},
		{
			name:           "D-Bus command fails",
			mockOutput:     "",
			mockError:      fmt.Errorf("dbus-send not found"),
			expectedResult: true, // Should assume support if can't check
		},
		{
			name:           "Empty output",
			mockOutput:     "",
			mockError:      nil,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up RunCommand mock
			shared.RunCommandMocks = []shared.RunCommandMock{
				{
					Command: "dbus-send",
					Args:    []string{"--session", "--dest=org.freedesktop.DBus", "--type=method_call", "--print-reply", "/org/freedesktop/DBus", "org.freedesktop.DBus.ListNames"},
					Out:     tt.mockOutput,
					Err:     tt.mockError,
				},
			}
			defer func() { shared.RunCommandMocks = nil }()

			result := checkStatusNotifierSupport()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestTrayiconCommandRegistration(t *testing.T) {
	// Test that the trayicon command is properly registered
	assert.NotNil(t, trayiconCmd)
	assert.Equal(t, "trayicon", trayiconCmd.Use)
	assert.Equal(t, "Display the status of the checks in the system tray", trayiconCmd.Short)
	assert.NotNil(t, trayiconCmd.Run)
}

func TestDBusCommandConstruction(t *testing.T) {
	// Test that we construct the correct dbus-send command
	shared.RunCommandMocks = []shared.RunCommandMock{
		{
			Command: "dbus-send",
			Args:    []string{"--session", "--dest=org.freedesktop.DBus", "--type=method_call", "--print-reply", "/org/freedesktop/DBus", "org.freedesktop.DBus.ListNames"},
			Out:     `string "org.kde.StatusNotifierWatcher"`,
			Err:     nil,
		},
	}
	defer func() { shared.RunCommandMocks = nil }()

	result := checkStatusNotifierSupport()

	// If we get here without error, the command was constructed correctly
	assert.True(t, result)
}

func TestStatusNotifierWatcherPatternMatching(t *testing.T) {
	// Test various output formats that might be returned by dbus-send
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "KDE format with quotes",
			output:   `string "org.kde.StatusNotifierWatcher"`,
			expected: true,
		},
		{
			name:     "Freedesktop format with quotes",
			output:   `string "org.freedesktop.StatusNotifierWatcher"`,
			expected: true,
		},
		{
			name:     "KDE format without quotes",
			output:   `org.kde.StatusNotifierWatcher`,
			expected: true,
		},
		{
			name:     "Freedesktop format without quotes",
			output:   `org.freedesktop.StatusNotifierWatcher`,
			expected: true,
		},
		{
			name:     "Multiple services including StatusNotifierWatcher",
			output:   `string "org.freedesktop.DBus" string "org.kde.StatusNotifierWatcher" string "org.freedesktop.PowerManagement"`,
			expected: true,
		},
		{
			name:     "Similar but not exact match",
			output:   `string "org.kde.StatusNotifier" string "org.freedesktop.StatusNotifierItem"`,
			expected: false,
		},
		{
			name:     "Empty output",
			output:   "",
			expected: false,
		},
		{
			name:     "Only other services",
			output:   `string "org.freedesktop.DBus" string "org.freedesktop.PowerManagement"`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shared.RunCommandMocks = []shared.RunCommandMock{
				{
					Command: "dbus-send",
					Args:    []string{"--session", "--dest=org.freedesktop.DBus", "--type=method_call", "--print-reply", "/org/freedesktop/DBus", "org.freedesktop.DBus.ListNames"},
					Out:     tt.output,
					Err:     nil,
				},
			}
			defer func() { shared.RunCommandMocks = nil }()

			result := checkStatusNotifierSupport()
			assert.Equal(t, tt.expected, result)
		})
	}
}
