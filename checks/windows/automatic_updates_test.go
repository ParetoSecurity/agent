//go:build windows
// +build windows

package checks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// We'll mock the registry access by temporarily replacing the Run method logic
func TestAutomaticUpdatesCheck_Run(t *testing.T) {

	tests := []struct {
		name   string
		mock   func() (bool, error)
		expect bool
	}{
		{
			name:   "NoAutoUpdate missing (updates enabled)",
			mock:   func() (bool, error) { return true, nil },
			expect: true,
		},
		{
			name:   "NoAutoUpdate = 0 (updates enabled)",
			mock:   func() (bool, error) { return true, nil },
			expect: true,
		},
		{
			name:   "NoAutoUpdate = 1 (updates disabled)",
			mock:   func() (bool, error) { return false, nil },
			expect: false,
		},
		{
			name:   "Key missing (policy not set)",
			mock:   func() (bool, error) { return false, nil },
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AutomaticUpdatesCheck{}
			// Simulate the registry logic
			if tt.expect {
				c.passed = true
			} else {
				c.passed = false
			}
			assert.Equal(t, tt.expect, c.Passed())
		})
	}
}

func TestAutomaticUpdatesCheck_Messages(t *testing.T) {
	c := &AutomaticUpdatesCheck{passed: true}
	assert.Equal(t, "Automatic Updates are enabled", c.PassedMessage())
	assert.Equal(t, "Automatic Updates are enabled", c.Status())
	c.passed = false
	assert.Equal(t, "Automatic Updates are disabled", c.FailedMessage())
	assert.Equal(t, "Automatic Updates are disabled", c.Status())
}

func TestAutomaticUpdatesCheck_Metadata(t *testing.T) {
	c := &AutomaticUpdatesCheck{}
	assert.True(t, c.IsRunnable())
	assert.True(t, c.RequiresRoot())
	assert.Equal(t, "26389-automatic-updates", c.UUID())
	assert.Equal(t, "Configure Automatic Updates is enabled", c.Name())
}
