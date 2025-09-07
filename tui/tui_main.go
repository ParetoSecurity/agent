//go:build linux

package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/log"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run() {
	// Redirect all logs to discard to prevent TUI scrambling
	log.Log = log.New(io.Discard)

	program := tea.NewProgram(initialModel(),
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
