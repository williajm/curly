// Package models contains Bubble Tea models for the presentation layer.
package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/curly/internal/app"
)

// Tab indices.
const (
	TabRequest = iota
	TabResponse
	TabHistory
)

// MainModel is the root model with tab navigation.
type MainModel struct {
	// Tab state.
	tabs      []string
	activeTab int

	// Sub-models.
	requestModel  RequestModel
	responseModel ResponseModel
	historyModel  HistoryModel

	// Services (injected from app initialization).
	requestService *app.RequestService
	historyService *app.HistoryService
	authService    *app.AuthService

	// UI state.
	width     int
	height    int
	showHelp  bool
	statusMsg string

	// Flags.
	quitting bool
}

// NewMainModel creates a new main model with all sub-models.
func NewMainModel(
	requestService *app.RequestService,
	historyService *app.HistoryService,
	authService *app.AuthService,
) MainModel {
	return MainModel{
		tabs:           []string{"Request", "Response", "History"},
		activeTab:      TabRequest,
		requestModel:   NewRequestModel(requestService, authService),
		responseModel:  NewResponseModel(),
		historyModel:   NewHistoryModel(historyService),
		requestService: requestService,
		historyService: historyService,
		authService:    authService,
		statusMsg:      "Press ? for help",
	}
}

// Init initializes the main model and sub-models.
func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.requestModel.Init(),
		m.responseModel.Init(),
		m.historyModel.Init(),
	)
}

// Update handles messages and updates the model.
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys.
		switch msg.String() {
		case KeyCtrlC, "q":
			if !m.showHelp {
				m.quitting = true
				return m, tea.Quit
			}

		case "?":
			// Toggle help screen.
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			// Close help screen.
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}

		case "tab":
			// Switch to next tab (but not if help is showing).
			if !m.showHelp {
				m.activeTab = (m.activeTab + 1) % len(m.tabs)
				return m, nil
			}

		case "shift+tab":
			// Switch to previous tab.
			if !m.showHelp {
				m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
				return m, nil
			}

		case "1":
			// Jump to Request tab.
			if !m.showHelp {
				m.activeTab = TabRequest
				return m, nil
			}

		case "2":
			// Jump to Response tab.
			if !m.showHelp {
				m.activeTab = TabResponse
				return m, nil
			}

		case "3":
			// Jump to History tab.
			if !m.showHelp {
				m.activeTab = TabHistory
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case requestSentMsg:
		// When a request is sent, update both request and response models.
		var cmd tea.Cmd
		m.requestModel, cmd = m.requestModel.Update(msg)
		cmds = append(cmds, cmd)

		// Also update response model with the new response.
		if msg.response != nil {
			m.responseModel.SetResponse(msg.response)
			// Switch to response tab to show the result.
			m.activeTab = TabResponse
			m.statusMsg = "Request completed successfully"
		} else if msg.err != nil {
			m.statusMsg = "Request failed"
		}

		return m, tea.Batch(cmds...)

	case historyLoadedMsg, historyDeletedMsg:
		// Pass history messages to history model.
		var cmd tea.Cmd
		m.historyModel, cmd = m.historyModel.Update(msg)
		return m, cmd
	}

	// Don't pass messages to sub-models if help is showing.
	if m.showHelp {
		return m, nil
	}

	// Delegate to active sub-model.
	var cmd tea.Cmd
	switch m.activeTab {
	case TabRequest:
		m.requestModel, cmd = m.requestModel.Update(msg)
		cmds = append(cmds, cmd)

	case TabResponse:
		m.responseModel, cmd = m.responseModel.Update(msg)
		cmds = append(cmds, cmd)

	case TabHistory:
		m.historyModel, cmd = m.historyModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the main view with tabs.
func (m MainModel) View() string {
	if m.quitting {
		return "Thanks for using curly!\n"
	}

	// Show help overlay if active.
	if m.showHelp {
		return m.renderHelp()
	}

	var sections []string

	// Render tabs.
	tabsView := m.renderTabs()
	sections = append(sections, tabsView)
	sections = append(sections, "")

	// Render active view.
	var activeView string
	switch m.activeTab {
	case TabRequest:
		activeView = m.requestModel.View()
	case TabResponse:
		activeView = m.responseModel.View()
	case TabHistory:
		activeView = m.historyModel.View()
	}
	sections = append(sections, activeView)

	// Render status bar.
	sections = append(sections, "")
	sections = append(sections, m.renderStatusBar())

	return strings.Join(sections, "\n")
}

// renderTabs renders the tab navigation.
func (m MainModel) renderTabs() string {
	var parts []string
	for i, tab := range m.tabs {
		if i == m.activeTab {
			parts = append(parts, "["+tab+"]")
		} else {
			parts = append(parts, " "+tab+" ")
		}
	}
	return strings.Join(parts, " ")
}

// renderStatusBar renders the bottom status bar.
func (m MainModel) renderStatusBar() string {
	if m.statusMsg != "" {
		return m.statusMsg
	}
	return "Press ? for help"
}

// renderHelp renders the help screen.
func (m MainModel) renderHelp() string {
	var sections []string

	sections = append(sections, "")
	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "                    CURLY - HELP & SHORTCUTS")
	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "")
	sections = append(sections, "GLOBAL: q/Ctrl+C=quit • ?=help • Tab=next tab • 1/2/3=jump to tab")
	sections = append(sections, "")
	sections = append(sections, "REQUEST: Tab=next field • Ctrl+Enter=send • ←/→=change method/auth")
	sections = append(sections, "")
	sections = append(sections, "RESPONSE: h=toggle headers/body • ↑↓=scroll")
	sections = append(sections, "")
	sections = append(sections, "HISTORY: ↑↓=navigate • Enter=load • d=delete • r=refresh")
	sections = append(sections, "")
	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "")
	sections = append(sections, "                   Press ESC or ? to close")
	sections = append(sections, "")

	return strings.Join(sections, "\n")
}

// GetActiveTab returns the currently active tab index.
func (m MainModel) GetActiveTab() int {
	return m.activeTab
}

// SetStatusMessage sets the status bar message.
func (m *MainModel) SetStatusMessage(msg string) {
	m.statusMsg = msg
}
