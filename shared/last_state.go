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

type LastState struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	State   bool   `json:"state"`
	Details string `json:"details"`
}

// TomlFileContent represents the structure of the TOML state file.
type TomlFileContent struct {
	RunningStateTime time.Time            `toml:"running_state_time"`
	States           map[string]LastState `toml:"states"`
}

var (
	mutex            sync.RWMutex
	states           = make(map[string]LastState)
	runningStateTime time.Time // Stores the time of the last run
	lastModTime      time.Time
	StatePath        string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.WithError(err).Warn("failed to get user home directory, using current directory instead")
		homeDir = "."
	}
	StatePath = filepath.Join(homeDir, ".paretosecurity.state")

	// Initial load from state file
	file, openErr := os.Open(StatePath)
	if openErr == nil {
		defer file.Close()
		var content TomlFileContent
		decoder := toml.NewDecoder(file)
		if decodeErr := decoder.Decode(&content); decodeErr == nil {
			if content.States != nil {
				states = content.States
			}
			runningStateTime = content.RunningStateTime

			fileInfo, statErr := file.Stat()
			if statErr == nil {
				lastModTime = fileInfo.ModTime()
			} else {
				log.WithError(statErr).Warnf("Failed to stat state file %s during init", StatePath)
			}
		} else {
			log.WithError(decodeErr).Warnf("Failed to decode state file %s during init, using defaults", StatePath)
		}
	} else if !os.IsNotExist(openErr) {
		log.WithError(openErr).Warnf("Failed to open state file %s during init, using defaults", StatePath)
	}
}

// CommitSharedState writes the current state (checks and running state) to the TOML file.
func CommitSharedState() error {
	mutex.Lock()
	defer mutex.Unlock()

	content := TomlFileContent{
		States:           states,
		RunningStateTime: runningStateTime,
	}

	file, err := os.Create(StatePath)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(file)
	encodeErr := encoder.Encode(content)

	closeErr := file.Close()

	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}

	fileInfo, statErr := os.Stat(StatePath)
	if statErr == nil {
		lastModTime = fileInfo.ModTime()
	} else {
		log.WithError(statErr).Warnf("Failed to stat state file %s after commit", StatePath)
	}
	return nil
}

// refreshStateFromDiskIfNeeded loads states from the TOML file if it has been modified
// since the last load. It must be called with mutex RLock or Lock held.
func refreshStateFromDiskIfNeeded() {
	fileInfo, err := os.Stat(StatePath)
	if err != nil {
		if os.IsNotExist(err) && !lastModTime.IsZero() {
			log.Infof("State file %s appears to have been deleted. Resetting in-memory state.", StatePath)
			states = make(map[string]LastState)
			runningStateTime = time.Time{}
			lastModTime = time.Time{}
		} else if !os.IsNotExist(err) {
			log.WithError(err).Warnf("Failed to stat state file %s. Using current in-memory state.", StatePath)
		}
		return
	}

	if fileInfo.ModTime().After(lastModTime) {
		log.Debugf("State file %s modified on disk. Reloading.", StatePath)
		rFile, openErr := os.Open(StatePath)
		if openErr != nil {
			log.WithError(openErr).Warnf("Failed to open modified state file %s. Using stale in-memory state.", StatePath)
			return
		}
		defer rFile.Close()

		var content TomlFileContent
		decoder := toml.NewDecoder(rFile)
		if decodeErr := decoder.Decode(&content); decodeErr != nil {
			log.WithError(decodeErr).Warnf("Failed to decode modified state file %s. Using stale in-memory state.", StatePath)
			return
		}

		if content.States == nil {
			states = make(map[string]LastState)
		} else {
			states = content.States
		}
		runningStateTime = content.RunningStateTime
		lastModTime = fileInfo.ModTime()
		log.Debugf("Successfully reloaded state from %s", StatePath)
	}
}

// AllChecksPassed returns true if all checks have passed.
func AllChecksPassed() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()

	for _, state := range states {
		if !state.State {
			return false
		}
	}
	return true
}

// GetFailedChecks returns a slice of failed checks.
func GetFailedChecks() []LastState {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()

	var failedChecks []LastState
	for _, state := range states {
		if !state.State {
			failedChecks = append(failedChecks, state)
		}
	}
	return failedChecks
}

// PrintStates loads and prints all stored states with their UUIDs, state values, and details.
func PrintStates() {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()

	fmt.Printf("Loaded %d states from %s\n", len(states), StatePath)
	fmt.Printf("Last modified time: %s\n\n", lastModTime.Format(time.RFC3339))

	data := [][]string{}
	for uuid, state := range states {
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

// UpdateLastState updates the LastState struct in the in-memory map.
// Caller should call CommitSharedState() to persist changes.
func UpdateLastState(newState LastState) {
	mutex.Lock()
	defer mutex.Unlock()
	refreshStateFromDiskIfNeeded()
	states[newState.UUID] = newState
}

// GetLastState retrieves the LastState struct by UUID.
func GetLastState(uuid string) (LastState, bool, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()

	state, exists := states[uuid]
	return state, exists, nil
}

func GetLastStates() map[string]LastState {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()

	copiedStates := make(map[string]LastState, len(states))
	for k, v := range states {
		copiedStates[k] = v
	}
	return copiedStates
}

// GetModifiedTime returns the last modified time of the state file.
func GetModifiedTime() time.Time {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()
	return lastModTime
}

// SetModifiedTime sets the last modified time of the state file.
func SetModifiedTime(t time.Time) {
	mutex.Lock()
	defer mutex.Unlock()
	lastModTime = t
}

// GetRunningState returns the last time the application was run.
func GetRunningState() time.Time {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()
	return runningStateTime
}

// SetRunningState sets the application's running state time in memory.
func SetRunningState(t time.Time) {
	mutex.Lock()
	defer mutex.Unlock()
	refreshStateFromDiskIfNeeded()
	runningStateTime = t
}

// IsRunning returns true if the application's running state time is non-zero.
func IsRunning() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	refreshStateFromDiskIfNeeded()
	return !runningStateTime.IsZero()
}
