package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ParetoSecurity/agent/claims"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// initialModel creates and returns the initial TUI model
func initialModel() *model {
	// Load previous state from disk
	stateMap := shared.GetLastStates()

	var claimGroups []claimGroup

	for _, claim := range claims.All {
		group := claimGroup{
			Title:    claim.Title,
			Expanded: true, // Start with all claims expanded
		}

		for _, chk := range claim.Checks {
			result := checkResult{
				Check:    chk,
				Claim:    claim.Title,
				Status:   "Not Run",
				Passed:   false,
				HasError: false,
				LastRun:  time.Time{},
				Details:  "",
			}

			// Apply previous state if available
			if lastState, exists := stateMap[chk.UUID()]; exists {
				if lastState.HasError {
					result.Status = "Error"
					group.ErrorCount++
				} else if lastState.Passed {
					result.Status = "Pass"
					group.PassCount++
				} else {
					result.Status = "Fail"
					group.FailCount++
				}
				result.Passed = lastState.Passed
				result.HasError = lastState.HasError
				result.Details = lastState.Details
				// Set a recent timestamp to indicate this is from previous run
				result.LastRun = time.Now().Add(-1 * time.Minute)
			} else {
				group.NotRunCount++
			}

			// Check if disabled
			if !chk.IsRunnable() || shared.IsCheckDisabled(chk.UUID()) {
				result.Status = "Disabled"
				group.DisabledCount++
				group.NotRunCount-- // Remove from not run count
			}

			group.Checks = append(group.Checks, result)
		}

		// Sort checks within each claim
		sort.Slice(group.Checks, func(i, j int) bool {
			return group.Checks[i].Check.Name() < group.Checks[j].Check.Name()
		})

		claimGroups = append(claimGroups, group)
	}

	// Sort claims alphabetically
	sort.Slice(claimGroups, func(i, j int) bool {
		return claimGroups[i].Title < claimGroups[j].Title
	})

	model := &model{
		claims:      claimGroups,
		selectedIdx: 0,
		logBuffer:   make([]string, 0, 1000), // Pre-allocate log buffer
	}

	model.rebuildDisplayItems()
	return model
}

// rebuildDisplayItems rebuilds the display items list from the current claims
func (m *model) rebuildDisplayItems() {
	m.displayItems = nil

	// Status colors using terminal color palette (same as in View method)
	passStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // Green
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))             // Red
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Bright Red and bold
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))         // Dark Gray

	for claimIdx, claim := range m.claims {
		// Add claim header with colored status indicators
		var statusParts []string
		if claim.PassCount > 0 {
			statusParts = append(statusParts, passStyle.Render(fmt.Sprintf("✓ %d", claim.PassCount)))
		}
		if claim.FailCount > 0 {
			statusParts = append(statusParts, failStyle.Render(fmt.Sprintf("✗ %d", claim.FailCount)))
		}
		if claim.ErrorCount > 0 {
			statusParts = append(statusParts, errorStyle.Render(fmt.Sprintf("⚠ %d", claim.ErrorCount)))
		}
		if claim.DisabledCount > 0 {
			statusParts = append(statusParts, disabledStyle.Render(fmt.Sprintf("⊘ %d", claim.DisabledCount)))
		}

		statusSummary := strings.Join(statusParts, " │ ")
		if len(statusParts) == 0 {
			statusSummary = "No checks"
		}

		expandIndicator := "▼"
		if !claim.Expanded {
			expandIndicator = "▶"
		}

		headerText := fmt.Sprintf("%s %s", expandIndicator, claim.Title)

		m.displayItems = append(m.displayItems, displayItem{
			IsHeader:   true,
			ClaimIndex: claimIdx,
			CheckIndex: -1,
			Text:       headerText,
			StatusText: statusSummary,
			Details:    "",
			Indented:   false,
		})

		// Add checks if expanded
		if claim.Expanded {
			for checkIdx, check := range claim.Checks {
				var statusText string
				switch check.Status {
				case "Pass":
					statusText = "✓ PASS"
				case "Fail":
					statusText = "✗ FAIL"
				case "Error":
					statusText = "⚠ ERROR"
				case "Disabled":
					statusText = "- DISABLED"
				default:
					statusText = ""
				}

				m.displayItems = append(m.displayItems, displayItem{
					IsHeader:   false,
					ClaimIndex: claimIdx,
					CheckIndex: checkIdx,
					Text:       check.Check.Name(),
					StatusText: statusText,
					Details:    check.Details,
					Indented:   true,
				})
			}
		}
	}
}

// updateModelState updates check results and recalculates all claim counts
func (m *model) updateModelState(results []checkResult, claimIdx int) {
	if len(results) > 0 {
		// Batch update - update all results
		resultMap := make(map[string][]checkResult)
		for _, result := range results {
			resultMap[result.Claim] = append(resultMap[result.Claim], result)
		}

		for claimIdx, claim := range m.claims {
			if claimResults, exists := resultMap[claim.Title]; exists {
				for checkIdx, check := range claim.Checks {
					for _, result := range claimResults {
						if check.Check.UUID() == result.Check.UUID() {
							m.claims[claimIdx].Checks[checkIdx] = result
							break
						}
					}
				}
			}
		}
	}

	// Recalculate counts for affected claims
	claimsToUpdate := []int{}
	if claimIdx >= 0 {
		claimsToUpdate = append(claimsToUpdate, claimIdx)
	} else {
		// Update all claims
		for i := range m.claims {
			claimsToUpdate = append(claimsToUpdate, i)
		}
	}

	for _, idx := range claimsToUpdate {
		if idx < len(m.claims) {
			claim := &m.claims[idx]
			claim.PassCount = 0
			claim.FailCount = 0
			claim.ErrorCount = 0
			claim.DisabledCount = 0
			claim.NotRunCount = 0

			for _, check := range claim.Checks {
				switch check.Status {
				case "Pass":
					claim.PassCount++
				case "Fail":
					claim.FailCount++
				case "Error":
					claim.ErrorCount++
				case "Disabled":
					claim.DisabledCount++
				default:
					claim.NotRunCount++
				}
			}
		}
	}
}

// Init implements the tea.Model interface
func (m *model) Init() tea.Cmd {
	// Test log to verify logging is working
	log.Info("TUI initialized successfully")
	log.Debug("Log capture system is active")
	return nil
}

// runAllChecks creates a command to run all checks
func (m *model) runAllChecks() tea.Cmd {
	// Collect all checks from all claims
	var allChecks []checkResult
	for _, claim := range m.claims {
		allChecks = append(allChecks, claim.Checks...)
	}
	return runAllChecks(allChecks)
}
