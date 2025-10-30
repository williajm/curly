package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/curly/internal/app"
	"github.com/williajm/curly/internal/infrastructure/repository"
)

// HistoryModel represents the history browser.
type HistoryModel struct {
	// Services.
	historyService *app.HistoryService

	// History entries.
	entries       []*repository.HistoryEntry
	selectedIndex int
	loading       bool
	errorMsg      string

	// UI dimensions.
	width  int
	height int
}

// Custom messages.
type historyLoadedMsg struct {
	entries []*repository.HistoryEntry
	err     error
}

type historyDeletedMsg struct {
	err error
}

// NewHistoryModel creates a new history browser model.
func NewHistoryModel(historyService *app.HistoryService) HistoryModel {
	return HistoryModel{
		historyService: historyService,
		entries:        []*repository.HistoryEntry{},
		selectedIndex:  0,
		loading:        false,
	}
}

// Init initializes the model and loads history.
func (m HistoryModel) Init() tea.Cmd {
	return m.loadHistory()
}

// Update handles messages and updates the model.
func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case KeyCtrlC:
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}

		case "down", "j":
			if m.selectedIndex < len(m.entries)-1 {
				m.selectedIndex++
			}

		case "enter":
			// Load selected history entry into request builder.
			// This will be handled by the main model.
			return m, nil

		case "delete", "d":
			// Delete selected history entry.
			if len(m.entries) > 0 {
				return m, m.deleteEntry(m.entries[m.selectedIndex].ID)
			}

		case "r":
			// Refresh history.
			return m, m.loadHistory()

		case "home", "g":
			m.selectedIndex = 0

		case "end", "G":
			if len(m.entries) > 0 {
				m.selectedIndex = len(m.entries) - 1
			}
		}

	case historyLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.entries = msg.entries
			m.errorMsg = ""
			// Ensure selected index is valid.
			if m.selectedIndex >= len(m.entries) {
				m.selectedIndex = max(0, len(m.entries)-1)
			}
		}

	case historyDeletedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			// Reload history after deletion.
			return m, m.loadHistory()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the history browser.
func (m HistoryModel) View() string {
	var sections []string

	sections = append(sections, "══ History ══")
	sections = append(sections, "")

	if m.loading {
		sections = append(sections, "Loading history...")
		return strings.Join(sections, "\n")
	}

	if m.errorMsg != "" {
		sections = append(sections, "Error: "+m.errorMsg)
		sections = append(sections, "")
	}

	if len(m.entries) == 0 {
		sections = append(sections, "No history entries yet.")
		sections = append(sections, "")
		sections = append(sections, "r: refresh • q: quit")
		return strings.Join(sections, "\n")
	}

	// Header.
	header := fmt.Sprintf("%-20s %-8s %-40s %-8s", "Time", "Method", "URL", "Status")
	sections = append(sections, header)
	sections = append(sections, strings.Repeat("─", 80))

	// Entries.
	for i, entry := range m.entries {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = "> "
		}

		// Truncate URL if too long.
		url := entry.RequestID // We don't have URL in history entry, use ID for now
		if len(url) > 40 {
			url = url[:37] + "..."
		}

		status := fmt.Sprintf("%d", entry.StatusCode)
		if entry.StatusCode == 0 {
			status = "Error"
		}

		// Format timestamp safely - parse RFC3339 and format to date+time only.
		timestamp := entry.ExecutedAt
		if t, err := time.Parse(time.RFC3339, entry.ExecutedAt); err == nil {
			timestamp = t.Format("2006-01-02 15:04:05")
		} else if len(entry.ExecutedAt) > 19 {
			// Fallback to truncation if parse fails but string is long enough.
			timestamp = entry.ExecutedAt[:19]
		}

		line := fmt.Sprintf("%s%-20s %-8s %-40s %-8s",
			cursor,
			timestamp,
			"GET", // We don't have method in history, default to GET
			url,
			status,
		)
		sections = append(sections, line)
	}

	sections = append(sections, "")
	sections = append(sections, "↑↓: navigate • Enter: load • d: delete • r: refresh • q: quit")

	return strings.Join(sections, "\n")
}

// loadHistory creates a command to load history from the service.
func (m *HistoryModel) loadHistory() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx := context.Background()
		entries, err := m.historyService.GetHistory(ctx, 100) // Load last 100 entries
		return historyLoadedMsg{entries: entries, err: err}
	}
}

// deleteEntry creates a command to delete a history entry.
func (m *HistoryModel) deleteEntry(id string) tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx := context.Background()
		err := m.historyService.DeleteHistory(ctx, id)
		return historyDeletedMsg{err: err}
	}
}

// GetSelectedEntry returns the currently selected history entry.
func (m *HistoryModel) GetSelectedEntry() *repository.HistoryEntry {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.entries) {
		return m.entries[m.selectedIndex]
	}
	return nil
}

// GetEntries returns all history entries.
func (m *HistoryModel) GetEntries() []*repository.HistoryEntry {
	return m.entries
}

// IsLoading returns whether history is currently being loaded.
func (m *HistoryModel) IsLoading() bool {
	return m.loading
}
