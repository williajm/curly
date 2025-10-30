// Package presentation provides the TUI (Terminal User Interface) layer for curly.
//
// This package implements the Bubble Tea application structure and wires together.
// all the models, views, and services to create an interactive terminal UI.
package presentation

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/curly/internal/app"
	"github.com/williajm/curly/internal/presentation/models"
)

// NewApp creates and configures a new Bubble Tea application.
//
// It takes the application services as dependencies and wires them into the.
// presentation layer models. The returned tea.Program is ready to run.
//
// Example usage:.
//
//	program := presentation.NewApp(requestService, historyService, authService).
//	if err := program.Start(); err != nil {.
//		log.Fatal(err).
//	}.
func NewApp(
	requestService *app.RequestService,
	historyService *app.HistoryService,
	authService *app.AuthService,
) *tea.Program {
	// Create the main model with all services.
	model := models.NewMainModel(requestService, historyService, authService)

	// Create the Bubble Tea program with options.
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	return program
}

// RunApp is a convenience function that creates and runs the application.
//
// It creates a new Bubble Tea program and starts it immediately.
// This is the simplest way to launch the TUI from main.go.
//
// Returns an error if the program fails to start or encounters a runtime error.
func RunApp(
	requestService *app.RequestService,
	historyService *app.HistoryService,
	authService *app.AuthService,
) error {
	program := NewApp(requestService, historyService, authService)
	_, err := program.Run()
	return err
}
