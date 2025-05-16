package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml"
)

func TestCommitLastState_Success(t *testing.T) {
	// Create a temporary directory for our test file.
	tmpDir, err := os.MkdirTemp("", "commitlaststate_success")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override statePath to a file in the temporary directory.
	testFile := filepath.Join(tmpDir, "test.state")
	StatePath = testFile

	// Prepare a test state.
	testState := LastState{
		UUID:    "test-uuid",
		State:   true,
		Details: "all good",
	}

	// Clear and set the states map for a clean test.
	mutex.Lock()
	states = make(map[string]LastState)
	states[testState.UUID] = testState
	mutex.Unlock()

	// Commit the state to the file.
	if err := CommitLastState(); err != nil {
		t.Fatalf("CommitLastState failed: %v", err)
	}

	// Open the file and decode its contents.
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	var decoded map[string]LastState
	decoder := toml.NewDecoder(file)
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatalf("failed to decode TOML file: %v", err)
	}

	// Validate that the decoded state matches the test state.
	got, exists := decoded[testState.UUID]
	if !exists {
		t.Fatalf("expected state with UUID %s not found", testState.UUID)
	}
	if got != testState {
		t.Fatalf("expected state %+v, got %+v", testState, got)
	}
}

func TestCommitLastState_Error(t *testing.T) {
	// Simulate an error by setting statePath to a directory path.
	tmpDir, err := os.MkdirTemp("", "commitlaststate_error")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	StatePath = tmpDir // os.Create on a directory should fail

	// Clear the states map.
	mutex.Lock()
	states = make(map[string]LastState)
	mutex.Unlock()

	// Attempt to commit; it should return an error.
	if err := CommitLastState(); err == nil {
		t.Fatalf("expected error when committing to a directory, got none")
	}
}

func TestAllChecksPassed(t *testing.T) {
	tests := []struct {
		name     string
		testData map[string]LastState
		want     bool
	}{
		{
			name: "all checks pass",
			testData: map[string]LastState{
				"uuid1": {UUID: "uuid1", State: true, Details: "passed"},
				"uuid2": {UUID: "uuid2", State: true, Details: "passed"},
				"uuid3": {UUID: "uuid3", State: true, Details: "passed"},
			},
			want: true,
		},
		{
			name: "one check fails",
			testData: map[string]LastState{
				"uuid1": {UUID: "uuid1", State: true, Details: "passed"},
				"uuid2": {UUID: "uuid2", State: false, Details: "failed"},
				"uuid3": {UUID: "uuid3", State: true, Details: "passed"},
			},
			want: false,
		},
		{
			name: "all checks fail",
			testData: map[string]LastState{
				"uuid1": {UUID: "uuid1", State: false, Details: "failed"},
				"uuid2": {UUID: "uuid2", State: false, Details: "failed"},
			},
			want: false,
		},
		{
			name:     "no checks",
			testData: map[string]LastState{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test state
			mutex.Lock()
			states = make(map[string]LastState)
			for k, v := range tt.testData {
				states[k] = v
			}
			mutex.Unlock()

			// Run the function
			got := AllChecksPassed()

			// Check result
			if got != tt.want {
				t.Errorf("AllChecksPassed() = %v, want %v", got, tt.want)
			}
		})
	}
}
