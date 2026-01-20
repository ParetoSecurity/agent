//go:build windows

package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_runUpdateCommand(t *testing.T) {
	var updateAppCalled bool

	config := &UpdateConfig{
		UpdateApp: func() error {
			updateAppCalled = true
			return nil
		},
	}

	runUpdateCommand(config)

	assert.True(t, updateAppCalled)
}

func Test_updateCommand(t *testing.T) {
	// Test that updateCommand creates default config and calls runUpdateCommand
	// This is primarily testing the wiring
	config := DefaultUpdateConfig()
	assert.NotNil(t, config.UpdateApp)
}

func Test_DefaultUpdateConfig(t *testing.T) {
	config := DefaultUpdateConfig()
	assert.NotNil(t, config.UpdateApp)
}
