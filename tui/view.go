package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View implements the tea.Model interface and renders the TUI
// getStatusStyles returns the status color styles
func getStatusStyles() (passStyle, failStyle, errorStyle, disabledStyle lipgloss.Style) {
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // Green
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))             // Red
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Bright Red and bold
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))         // Dark Gray
	return
}

// styleStatus applies consistent styling to status text
func styleStatus(statusText string, isRunning bool) string {
	passStyle, failStyle, errorStyle, disabledStyle := getStatusStyles()
	runningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")) // Blue

	switch {
	case strings.Contains(statusText, "PASS"):
		return passStyle.Render(statusText)
	case strings.Contains(statusText, "FAIL"):
		return failStyle.Render(statusText)
	case strings.Contains(statusText, "ERROR"):
		return errorStyle.Render(statusText)
	case strings.Contains(statusText, "DISABLED"):
		return disabledStyle.Render(statusText)
	default:
		if isRunning {
			return runningStyle.Render("⟳ RUNNING")
		}
		return statusText
	}
}

func (m model) View() string {
	if m.viewport.width == 0 {
		return "Loading..."
	}

	// Calculate available width for content
	contentWidth := m.viewport.width

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("4")). // Blue from terminal palette (matches table border)
		Width(contentWidth).
		Align(lipgloss.Center)

	var s strings.Builder

	// Generate ASCII header
	s.WriteString(m.generateASCIIHeader(titleStyle, contentWidth))

	// Checks table with border
	var tableBuilder strings.Builder

	// Show all items or truncate if too many
	maxItems := m.viewport.height - 8 // Leave space for header and footer
	endIdx := len(m.displayItems)
	if maxItems > 0 && endIdx > maxItems {
		endIdx = maxItems
	}

	for i := 0; i < endIdx; i++ {
		item := m.displayItems[i]

		// Selection indicator
		indicator := " "
		if i == m.selectedIdx {
			indicator = ">"
		}

		// Indentation for checks under claims
		indent := ""
		if item.Indented {
			indent = "  "
		}

		// Status styling
		var statusText string
		if item.IsHeader {
			statusText = item.StatusText
		} else {
			statusText = styleStatus(item.StatusText, m.running)
		}

		var line string
		if item.IsHeader {
			// Claim header
			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")) // Cyan and bold
			if item.ClaimIndex < len(m.claims) {
				claim := m.claims[item.ClaimIndex]
				total := claim.PassCount + claim.FailCount + claim.ErrorCount + claim.DisabledCount + claim.NotRunCount
				line = fmt.Sprintf("%s %s (%d checks) %s",
					indicator,
					headerStyle.Render(item.Text),
					total,
					statusText)
			}
		} else {
			// Individual check - calculate available space for details dynamically
			checkNameWidth := 25
			statusTextWidth := 18
			fixedWidth := 1 + 2 + checkNameWidth + 1 + statusTextWidth + 1 // indicator + indent + name + spaces + status + space
			availableDetailsWidth := m.viewport.width - fixedWidth - 4     // Leave some margin

			// Ensure minimum width and truncate if needed
			if availableDetailsWidth < 20 {
				availableDetailsWidth = 20
			}

			details := item.Details
			if len(details) > availableDetailsWidth {
				details = details[:availableDetailsWidth-3] + "..."
			}

			line = fmt.Sprintf("%s %s%-25s %-18s %s",
				indicator,
				indent,
				item.Text,
				statusText,
				details)
		}

		// Highlight selected line
		if i == m.selectedIdx {
			style := lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("7")) // Black background, white text
			line = style.Render(line)
		}

		tableBuilder.WriteString(line)
		if i < endIdx-1 { // Don't add newline after last item
			tableBuilder.WriteString("\n")
		}
	}

	// Create bordered table with help text in bottom border using lipgloss table approach
	tableContent := tableBuilder.String()

	// Help text for bottom border
	var helpText string
	if m.running {
		helpText = "Running checks..."
	} else {
		helpText = "r=run all • space=run selected • b=help • q=quit"
	}

	// Use lipgloss to create a proper bordered style with custom bottom border
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(0, 1).
		Width(contentWidth - 2)

	// Apply the standard border to content
	borderedTable := borderStyle.Render(tableContent)

	// Now manually replace the bottom border line to include help text
	lines := strings.Split(borderedTable, "\n")
	if len(lines) > 0 {
		// Find the last line (bottom border) and replace it with custom help text border
		lastLineIdx := len(lines) - 1
		originalBottomBorder := lines[lastLineIdx]

		// Calculate available space for help text
		borderWidth := lipgloss.Width(originalBottomBorder)

		// Calculate the exact space needed
		helpTextLength := lipgloss.Width(helpText)
		cornersAndSpaces := 2 + 2 + 2 // ╰ + ╯ + spaces around help text
		minDashes := 2                // At least 1 dash on each side

		totalNeeded := cornersAndSpaces + helpTextLength + minDashes

		if totalNeeded <= borderWidth {
			// Create new bottom border with help text
			remainingDashes := borderWidth - cornersAndSpaces - helpTextLength
			leftDashes := remainingDashes / 2
			rightDashes := remainingDashes - leftDashes

			// Ensure at least 1 dash on each side
			if leftDashes < 1 {
				leftDashes = 1
				rightDashes = remainingDashes - leftDashes
			}
			if rightDashes < 1 {
				rightDashes = 1
				leftDashes = remainingDashes - rightDashes
			}

			// Style help text with different color (yellow and bold)
			styledHelpText := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render(helpText)

			// Create border parts with border color (blue)
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
			leftPart := borderStyle.Render("╰" + strings.Repeat("─", leftDashes) + " ")
			rightPart := borderStyle.Render(" " + strings.Repeat("─", rightDashes) + "╯")

			// Combine styled parts
			newBottomBorder := leftPart + styledHelpText + rightPart

			// Make sure the total width matches exactly by adjusting right dashes if needed
			actualWidth := lipgloss.Width(newBottomBorder)
			if actualWidth != borderWidth {
				// Adjust right dashes if there's a mismatch
				rightDashes = rightDashes + (borderWidth - actualWidth)
				if rightDashes >= 0 {
					rightPart = borderStyle.Render(" " + strings.Repeat("─", rightDashes) + "╯")
					newBottomBorder = leftPart + styledHelpText + rightPart
				}
			}

			lines[lastLineIdx] = newBottomBorder
		}

		borderedTable = strings.Join(lines, "\n")
	}

	s.WriteString(borderedTable)

	// Stats section at bottom
	passed := 0
	failed := 0
	errors := 0
	disabled := 0
	notRun := 0
	totalChecks := 0

	for _, claim := range m.claims {
		passed += claim.PassCount
		failed += claim.FailCount
		errors += claim.ErrorCount
		disabled += claim.DisabledCount
		notRun += claim.NotRunCount
		totalChecks += len(claim.Checks)
	}

	passStyle, failStyle, errorStyle, disabledStyle := getStatusStyles()
	stats := fmt.Sprintf("Total: %d | %s | %s | %s | %s",
		totalChecks,
		passStyle.Render(fmt.Sprintf("Pass: %d", passed)),
		failStyle.Render(fmt.Sprintf("Fail: %d", failed)),
		errorStyle.Render(fmt.Sprintf("Error: %d", errors)),
		disabledStyle.Render(fmt.Sprintf("Disabled: %d", disabled)),
	)

	if !m.lastUpdate.IsZero() {
		stats += fmt.Sprintf(" | Last Updated: %s", m.lastUpdate.Format("15:04:05"))
	} else if passed+failed+errors > 0 {
		_, _, _, disabledStyle := getStatusStyles()
		stats += " | " + disabledStyle.Render("(Previous Results)")
	}

	// Center the status line
	statusStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center)

	s.WriteString("\n\n")
	s.WriteString(statusStyle.Render(stats))

	return s.String()
}

// generateASCIIHeader generates the ASCII art header based on terminal width
func (m model) generateASCIIHeader(titleStyle lipgloss.Style, contentWidth int) string {
	// Use different logos based on terminal width
	var asciiHeader string

	if contentWidth >= 120 {
		// Full wide logo for wide terminals
		asciiHeader = `██████╗  █████╗ ██████╗ ███████╗████████╗ ██████╗     ███████╗███████╗ ██████╗██╗   ██╗██████╗ ██╗████████╗██╗   ██╗
██╔══██╗██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔═══██╗    ██╔════╝██╔════╝██╔════╝██║   ██║██╔══██╗██║╚══██╔══╝╚██╗ ██╔╝
██████╔╝███████║██████╔╝█████╗     ██║   ██║   ██║    ███████╗█████╗  ██║     ██║   ██║██████╔╝██║   ██║    ╚████╔╝ 
██╔═══╝ ██╔══██║██╔══██╗██╔══╝     ██║   ██║   ██║    ╚════██║██╔══╝  ██║     ██║   ██║██╔══██╗██║   ██║     ╚██╔╝  
██║     ██║  ██║██║  ██║███████╗   ██║   ╚██████╔╝    ███████║███████╗╚██████╗╚██████╔╝██║  ██║██║   ██║      ██║   
                                                                                                                    `
	} else {
		// Braille pattern logo for narrow terminals
		asciiHeader = `░█▀█░█▀█░█▀▄░█▀▀░▀█▀░█▀█░░░█▀▀░█▀▀░█▀▀░█░█░█▀▄░▀█▀░▀█▀░█░█
░█▀▀░█▀█░█▀▄░█▀▀░░█░░█░█░░░▀▀█░█▀▀░█░░░█░█░█▀▄░░█░░░█░░░█░
░▀░░░▀░▀░▀░▀░▀▀▀░░▀░░▀▀▀░░░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀░▀░▀▀▀░░▀░░░▀░`
	}

	return titleStyle.Render(asciiHeader) + "\n"
}
