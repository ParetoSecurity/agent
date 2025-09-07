//go:build linux

package tui

import (
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// isTerminal checks if stdin, stdout, and stderr are connected to a terminal
func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// Run starts the TUI application
func Run() {
	// Check if we're in a proper terminal
	if !isTerminal() {
		fmt.Println("Error: TUI requires a proper terminal environment")
		fmt.Println("Please run this command in an interactive terminal session")
		os.Exit(1)
	}

	// Redirect all logs to discard to prevent TUI scrambling
	log.SetOutput(io.Discard)

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
