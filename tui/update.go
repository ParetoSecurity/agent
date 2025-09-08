package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update implements the tea.Model interface and handles incoming messages
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.width = msg.Width
		m.viewport.height = msg.Height

	case tea.KeyMsg:
		if m.running {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.showLogs {
				// Scroll logs up
				if m.logScrollPos > 0 {
					m.logScrollPos--
				}
			} else {
				// Navigate checks list
				if m.selectedIdx > 0 {
					m.selectedIdx--
				}
			}

		case "down", "j":
			if m.showLogs {
				// Scroll logs down
				viewportHeight := m.viewport.height - 12
				if viewportHeight < 5 {
					viewportHeight = 5
				}
				maxScroll := len(m.logBuffer) - viewportHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.logScrollPos < maxScroll {
					m.logScrollPos++
				}
			} else {
				// Navigate checks list
				if m.selectedIdx < len(m.displayItems)-1 {
					m.selectedIdx++
				}
			}

		case "left", "h":
			// Collapse current claim if on header, or move to parent claim
			if m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if item.IsHeader {
					m.claims[item.ClaimIndex].Expanded = false
				} else {
					// Move to parent claim header
					for i := m.selectedIdx - 1; i >= 0; i-- {
						if m.displayItems[i].IsHeader && m.displayItems[i].ClaimIndex == item.ClaimIndex {
							m.selectedIdx = i
							break
						}
					}
				}
				m.rebuildDisplayItems()
			}

		case "right", "enter":
			// Expand current claim if on header
			if m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if item.IsHeader {
					m.claims[item.ClaimIndex].Expanded = !m.claims[item.ClaimIndex].Expanded
					m.rebuildDisplayItems()
				}
			}

		case "pgup":
			if m.showLogs {
				// Page up in logs
				viewportHeight := m.viewport.height - 12
				if viewportHeight < 5 {
					viewportHeight = 5
				}
				m.logScrollPos -= viewportHeight
				if m.logScrollPos < 0 {
					m.logScrollPos = 0
				}
			}

		case "pgdown":
			if m.showLogs {
				// Page down in logs
				viewportHeight := m.viewport.height - 12
				if viewportHeight < 5 {
					viewportHeight = 5
				}
				maxScroll := len(m.logBuffer) - viewportHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.logScrollPos += viewportHeight
				if m.logScrollPos > maxScroll {
					m.logScrollPos = maxScroll
				}
			}

		case "home":
			if m.showLogs {
				// Go to top of logs
				m.logScrollPos = 0
			}

		case "end":
			if m.showLogs {
				// Go to bottom of logs
				viewportHeight := m.viewport.height - 12
				if viewportHeight < 5 {
					viewportHeight = 5
				}
				maxScroll := len(m.logBuffer) - viewportHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.logScrollPos = maxScroll
			}

		case "l":
			// Toggle logs view
			m.showLogs = !m.showLogs
			// Reset scroll position when opening logs
			if m.showLogs {
				// Start at the bottom (latest logs)
				viewportHeight := m.viewport.height - 12
				if viewportHeight < 5 {
					viewportHeight = 5
				}
				maxScroll := len(m.logBuffer) - viewportHeight
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.logScrollPos = maxScroll
			}

		case "r":
			if !m.running {
				m.running = true
				// Clear log buffer before running checks
				if m.logWriter != nil {
					m.logWriter.Clear()
					m.logScrollPos = 0 // Reset scroll position
				}
				return m, m.runAllChecks()
			}

		case "space", " ":
			if !m.running && m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if !item.IsHeader {
					m.running = true
					// Clear log buffer before running single check
					if m.logWriter != nil {
						m.logWriter.Clear()
						m.logScrollPos = 0 // Reset scroll position
					}
					check := m.claims[item.ClaimIndex].Checks[item.CheckIndex]
					return m, runSingleCheck(item.ClaimIndex, item.CheckIndex, check)
				}
			}

		case "b":
			if !m.running && m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if !item.IsHeader {
					check := m.claims[item.ClaimIndex].Checks[item.CheckIndex]
					return m, openBrowserForCheck(check.Check)
				}
			}
		}

	case checkCompleteMsg:
		if msg.claimIdx < len(m.claims) && msg.checkIdx < len(m.claims[msg.claimIdx].Checks) {
			m.claims[msg.claimIdx].Checks[msg.checkIdx] = msg.result
			m.updateModelState(nil, msg.claimIdx)
			m.rebuildDisplayItems()
		}
		// Set running to false after single check completes
		m.running = false
		m.lastUpdate = time.Now()

	case batchRunMsg:
		m.updateModelState(msg.results, -1)
		m.running = false
		m.lastUpdate = time.Now()
		m.rebuildDisplayItems()
	}

	return m, nil
}
