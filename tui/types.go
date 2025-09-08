package tui

import (
	"time"

	"github.com/ParetoSecurity/agent/check"
)

// checkResult represents the result of running a security check
type checkResult struct {
	Check    check.Check
	Claim    string
	Status   string
	Passed   bool
	HasError bool
	Details  string
	LastRun  time.Time
}

// claimGroup represents a group of checks under a security claim
type claimGroup struct {
	Title         string
	Checks        []checkResult
	Expanded      bool
	PassCount     int
	FailCount     int
	ErrorCount    int
	DisabledCount int
	NotRunCount   int
}

// displayItem represents an item in the TUI display list
type displayItem struct {
	IsHeader   bool
	ClaimIndex int
	CheckIndex int
	Text       string
	StatusText string
	Details    string
	Indented   bool
}

// model represents the main TUI application state
type model struct {
	claims       []claimGroup
	displayItems []displayItem
	selectedIdx  int
	running      bool
	lastUpdate   time.Time
	showLogs     bool
	logBuffer    []string
	logWriter    *logWriter
	logScrollPos int // Current scroll position in logs
	viewport     struct {
		width  int
		height int
	}
}

// Bubble Tea messages
type checkCompleteMsg struct {
	claimIdx int
	checkIdx int
	result   checkResult
}

type batchRunMsg struct {
	results []checkResult
}
