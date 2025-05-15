package shared

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommitLastState(t *testing.T) {

	tempFile, err := os.CreateTemp("", "test-state-*.toml")
	assert.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	StatePath = tempFile.Name()

	// Reset state for test
	lastState = LastState{
		Checks:       make(map[string]CheckState),
		RunningState: time.Time{},
	}

	// Add a test check
	testCheck := CheckState{
		Name:    "TestCheck",
		UUID:    "test-uuid",
		State:   true,
		Details: "Test details",
	}

	lastState.Checks[testCheck.UUID] = testCheck

	// Test
	err = CommitLastState()
	assert.NoError(t, err)

	// Verify file exists and can be loaded
	_, err = os.Stat(StatePath)
	assert.NoError(t, err)

	// Reset state and load from file
	oldLastState := lastState
	lastState = LastState{
		Checks:       make(map[string]CheckState),
		RunningState: time.Time{},
	}

	loadStates()

	// Verify content matches
	assert.Equal(t, oldLastState.Checks, lastState.Checks)

	// Test error case
	StatePath = "/nonexistent/directory/file.toml"
	err = CommitLastState()
	assert.Error(t, err)
}

func TestAllChecksPassed(t *testing.T) {

	tests := []struct {
		name     string
		setupFn  func()
		expected bool
	}{
		{
			name: "all checks pass",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", State: true},
						"test2": {UUID: "test2", State: true},
						"test3": {UUID: "test3", State: true},
					},
				}
			},
			expected: true,
		},
		{
			name: "one check fails",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", State: true},
						"test2": {UUID: "test2", State: false},
						"test3": {UUID: "test3", State: true},
					},
				}
			},
			expected: false,
		},
		{
			name: "all checks fail",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", State: false},
						"test2": {UUID: "test2", State: false},
					},
				}
			},
			expected: false,
		},
		{
			name: "no checks",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{},
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFn()
			result := AllChecksPassed()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFailedChecks(t *testing.T) {

	tests := []struct {
		name     string
		setupFn  func()
		expected []CheckState
	}{
		{
			name: "no failed checks",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", Name: "Check1", State: true},
						"test2": {UUID: "test2", Name: "Check2", State: true},
					},
				}
			},
			expected: []CheckState{},
		},
		{
			name: "some failed checks",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", Name: "Check1", State: true},
						"test2": {UUID: "test2", Name: "Check2", State: false, Details: "Failed"},
						"test3": {UUID: "test3", Name: "Check3", State: false, Details: "Also failed"},
					},
				}
			},
			expected: []CheckState{
				{UUID: "test2", Name: "Check2", State: false, Details: "Failed"},
				{UUID: "test3", Name: "Check3", State: false, Details: "Also failed"},
			},
		},
		{
			name: "all failed checks",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{
						"test1": {UUID: "test1", Name: "Check1", State: false, Details: "Failed 1"},
						"test2": {UUID: "test2", Name: "Check2", State: false, Details: "Failed 2"},
					},
				}
			},
			expected: []CheckState{
				{UUID: "test1", Name: "Check1", State: false, Details: "Failed 1"},
				{UUID: "test2", Name: "Check2", State: false, Details: "Failed 2"},
			},
		},
		{
			name: "no checks",
			setupFn: func() {
				lastState = LastState{
					Checks: map[string]CheckState{},
				}
			},
			expected: []CheckState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFn()
			result := GetFailedChecks()

			// Since map iteration order is not guaranteed, we need to check that all expected items are in the result
			// rather than checking the exact order
			assert.Len(t, result, len(tt.expected))

			if len(tt.expected) > 0 {
				// Create a map of expected CheckStates by UUID for easier comparison
				expectedMap := make(map[string]CheckState)
				for _, check := range tt.expected {
					expectedMap[check.UUID] = check
				}

				// Verify each result is in the expected map
				for _, check := range result {
					expectedCheck, exists := expectedMap[check.UUID]
					assert.True(t, exists, "Unexpected check in results: %s", check.UUID)
					assert.Equal(t, expectedCheck, check)
				}
			}
		})
	}
}

func TestUpdateLastState(t *testing.T) {

	// Reset state for test
	lastState = LastState{
		Checks:       make(map[string]CheckState),
		RunningState: time.Time{},
	}

	// Test cases
	tests := []struct {
		name     string
		newState CheckState
		checkFn  func(t *testing.T)
	}{
		{
			name: "add new check state",
			newState: CheckState{
				Name:    "TestCheck1",
				UUID:    "test-uuid-1",
				State:   true,
				Details: "Test details 1",
			},
			checkFn: func(t *testing.T) {
				state, exists, _ := GetLastState("test-uuid-1")
				assert.True(t, exists)
				assert.Equal(t, "TestCheck1", state.Name)
				assert.True(t, state.State)
				assert.Equal(t, "Test details 1", state.Details)
			},
		},
		{
			name: "update existing check state",
			newState: CheckState{
				Name:    "TestCheck1-Updated",
				UUID:    "test-uuid-1",
				State:   false,
				Details: "Updated details",
			},
			checkFn: func(t *testing.T) {
				state, exists, _ := GetLastState("test-uuid-1")
				assert.True(t, exists)
				assert.Equal(t, "TestCheck1-Updated", state.Name)
				assert.False(t, state.State)
				assert.Equal(t, "Updated details", state.Details)
			},
		},
		{
			name: "add another check state",
			newState: CheckState{
				Name:    "TestCheck2",
				UUID:    "test-uuid-2",
				State:   false,
				Details: "Test details 2",
			},
			checkFn: func(t *testing.T) {
				// Verify first check still exists
				state1, exists1, _ := GetLastState("test-uuid-1")
				assert.True(t, exists1)
				assert.Equal(t, "TestCheck1-Updated", state1.Name)

				// Verify second check was added
				state2, exists2, _ := GetLastState("test-uuid-2")
				assert.True(t, exists2)
				assert.Equal(t, "TestCheck2", state2.Name)
				assert.False(t, state2.State)
				assert.Equal(t, "Test details 2", state2.Details)

				// Verify we have exactly two checks
				assert.Equal(t, 2, len(lastState.Checks))
			},
		},
	}

	// Before any updates, lastModTime should be zero or earlier than now
	initialTime := lastModTime

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Record time before update
			beforeUpdate := time.Now()

			// Perform update
			UpdateLastState(tt.newState)

			// Run specific checks for this test case
			tt.checkFn(t)

			// Verify lastModTime was updated
			assert.True(t, lastModTime.After(initialTime))
			assert.True(t, !lastModTime.Before(beforeUpdate))
		})
	}
}

func TestAreChecksRunning(t *testing.T) {

	tests := []struct {
		name     string
		setupFn  func()
		expected bool
	}{
		{
			name: "checks are running",
			setupFn: func() {
				lastState = LastState{
					Checks:       make(map[string]CheckState),
					RunningState: time.Now(),
				}
			},
			expected: true,
		},
		{
			name: "checks are not running",
			setupFn: func() {
				lastState = LastState{
					Checks:       make(map[string]CheckState),
					RunningState: time.Time{},
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test case
			tt.setupFn()

			// Test
			result := AreChecksRunning()

			// Verify
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStartRunningChecks(t *testing.T) {

	// Create temp file for state
	tempFile, err := os.CreateTemp("", "test-running-*.toml")
	assert.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())
	StatePath = tempFile.Name()

	// Reset state for test
	lastState = LastState{
		Checks:       make(map[string]CheckState),
		RunningState: time.Time{}, // Not running
	}

	// Verify checks are not running initially
	assert.False(t, AreChecksRunning())

}
