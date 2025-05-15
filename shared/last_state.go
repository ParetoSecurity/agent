package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/caarlos0/log"
	"github.com/olekukonko/tablewriter"
	"github.com/pelletier/go-toml"
)

type CheckState struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	State   bool   `json:"state"`
	Details string `json:"details"`
}

type LastState struct {
	Checks       map[string]CheckState `json:"states"`
	RunningState time.Time             `json:"running_state"` // Zero time means not running, otherwise start time
}

var (
	mutex       sync.RWMutex
	lastState   LastState
	lastModTime time.Time
	StatePath   string
)

func init() {
	lastState = LastState{
		Checks:       make(map[string]CheckState),
		RunningState: time.Time{},
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Warn("failed to get user home directory, using current directory instead")
		homeDir = "."
	}
	StatePath = filepath.Join(homeDir, ".paretosecurity.state")
}

// Commit writes the current state map to the TOML file.
func CommitLastState() error {
	mutex.Lock()
	defer mutex.Unlock()

	file, err := os.Create(StatePath)
	if err != nil {
		return err
	}
	defer file.Close()
	lastModTime = time.Now()
	encoder := toml.NewEncoder(file)
	return encoder.Encode(lastState)
}

// AllChecksPassed returns true if all checks have passed.
func AllChecksPassed() bool {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, state := range lastState.Checks {
		if !state.State {
			return false
		}
	}
	return true
}

// GetFailedChecks returns a slice of failed checks.
func GetFailedChecks() []CheckState {
	mutex.RLock()
	defer mutex.RUnlock()

	var failedChecks []CheckState
	for _, state := range lastState.Checks {
		if !state.State {
			failedChecks = append(failedChecks, state)
		}
	}
	return failedChecks
}

// PrintStates loads and prints all stored states with their UUIDs, state values, and details.
func PrintStates() {
	loadStates()

	fmt.Printf("Loaded %d states from %s\n", len(lastState.Checks), StatePath)
	fmt.Printf("Last modified time: %s\n\n", lastModTime.Format(time.RFC3339))

	data := [][]string{}
	for uuid, state := range lastState.Checks {
		stateStr := "Pass"
		if !state.State {
			stateStr = "Fail"
		}
		data = append(data, []string{uuid, state.Name, stateStr, state.Details})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"UUID", "Name", "State", "Details"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(data)
	table.Render()
}

// UpdateState updates the LastState struct in the in-memory map and commits to the TOML file.
func UpdateLastState(newState CheckState) {
	mutex.Lock()
	defer mutex.Unlock()
	lastModTime = time.Now()
	lastState.Checks[newState.UUID] = newState
}

// GetState retrieves the LastState struct by UUID.
func GetLastState(uuid string) (CheckState, bool, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	loadStates()

	state, exists := lastState.Checks[uuid]
	return state, exists, nil
}

func GetLastStates() map[string]CheckState {
	mutex.RLock()
	defer mutex.RUnlock()
	loadStates()

	return lastState.Checks
}

// GetModifiedTime returns the last modified time of the state file.
func GetModifiedTime() time.Time {
	mutex.RLock()
	defer mutex.RUnlock()
	loadStates()

	return lastModTime
}

// loadStates loads the states from the TOML file if it has been modified since the last load.
func loadStates() {
	fileInfo, err := os.Stat(StatePath)
	if err != nil {
		return
	}

	if fileInfo.ModTime().After(lastModTime) {
		file, err := os.Open(StatePath)
		if err != nil {
			log.WithError(err).Warnf("failed to open state file: %s", StatePath)
			return
		}
		defer file.Close()

		decoder := toml.NewDecoder(file)
		if err := decoder.Decode(&lastState); err != nil {
			log.WithError(err).Warnf("failed to decode state file: %s", StatePath)
			return
		}
		lastModTime = fileInfo.ModTime()
	}
}

// SetModifiedTime sets the last modified time of the state file.
func SetModifiedTime(t time.Time) {
	mutex.Lock()
	defer mutex.Unlock()

	lastModTime = t
}

// AreChecksRunning checks if any checks are currently running.
func AreChecksRunning() bool {
	loadStates()
	return !lastState.RunningState.IsZero()
}

// StartRunningChecks marks all checks as running.
func StartRunningChecks() {
	lastState.RunningState = time.Now()
	CommitLastState()
}

// StopRunningChecks marks all checks as not running.
func StopRunningChecks() {
	lastState.RunningState = time.Time{}
	CommitLastState()
}
