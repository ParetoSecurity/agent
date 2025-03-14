package shared

import (
	"testing"

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
			RunCommandMocks = map[string]string{
				"systemctl is-enabled pareto-socket": tt.mockOutput,
			}

			// Run test
			result := IsSocketServicePresent()

			// Assert
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
