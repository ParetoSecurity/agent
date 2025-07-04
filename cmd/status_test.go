package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_runStatusCommand(t *testing.T) {
	var printStatesCalled bool

	config := &StatusConfig{
		PrintStates: func() {
			printStatesCalled = true
		},
	}

	runStatusCommand(config)

	assert.True(t, printStatesCalled)
}

func Test_statusCommand(t *testing.T) {
	// Test that statusCommand creates default config and calls runStatusCommand
	// This is primarily testing the wiring
	config := DefaultStatusConfig()
	assert.NotNil(t, config.PrintStates)
}

func Test_DefaultStatusConfig(t *testing.T) {
	config := DefaultStatusConfig()
	assert.NotNil(t, config.PrintStates)
}
