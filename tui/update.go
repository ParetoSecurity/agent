package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update implements the tea.Model interface and handles incoming messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case "down", "j":
			if m.selectedIdx < len(m.displayItems)-1 {
				m.selectedIdx++
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

		case "right", "l", "enter":
			// Expand current claim if on header
			if m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if item.IsHeader {
					m.claims[item.ClaimIndex].Expanded = !m.claims[item.ClaimIndex].Expanded
					m.rebuildDisplayItems()
				}
			}

		case "r":
			if !m.running {
				m.running = true
				return m, m.runAllChecks()
			}

		case "space", " ":
			if !m.running && m.selectedIdx < len(m.displayItems) {
				item := m.displayItems[m.selectedIdx]
				if !item.IsHeader {
					m.running = true
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
			m.updateClaimCounts(msg.claimIdx)
			m.rebuildDisplayItems()
		}

	case batchRunMsg:
		m.updateAllResults(msg.results)
		m.running = false
		m.lastUpdate = time.Now()
		m.rebuildDisplayItems()

	case runCompleteMsg:
		m.running = false
		m.lastUpdate = time.Now()
		m.rebuildDisplayItems()
	}

	return m, nil
}
