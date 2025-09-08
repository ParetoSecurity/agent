package tui

import (
	"fmt"
	"os"

	"github.com/caarlos0/log"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run() {
	// Create initial model with log buffer
	m := initialModel()

	// Set up log writer to capture logs in the buffer
	logWriter := newLogWriter(&m.logBuffer, 1000)
	m.logWriter = logWriter

	// Replace the global logger with one that writes to our buffer
	log.Log = log.New(logWriter)
	log.SetLevel(log.DebugLevel)

	// Test that logging is working
	log.Info("Starting Pareto Security TUI")
	log.Debug("Debug logging enabled")

	program := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stderr), // Use stderr to avoid conflicts
	)
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
