package runner

import (
	"testing"

	"github.com/ParetoSecurity/agent/shared"
	"github.com/stretchr/testify/assert"
)

func TestIsSocketServicePresent(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     string
		mockError      bool
		expectedResult bool
	}{

		{
			name:           "service is not enabled",
			mockOutput:     "",
			mockError:      true,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			shared.RunCommandMocks = []shared.RunCommandMock{
				{Command: "systemctl", Args: []string{"is-enabled", "pareto-socket"}, Out: tt.mockOutput, Err: nil},
			}

			// Run test
			result := IsSocketServicePresent()

			// Assert
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
