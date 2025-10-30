package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/curly/internal/domain"
)

// ResponseModel represents the response viewer
type ResponseModel struct {
	// Current response being displayed
	response *domain.Response

	// Viewport for scrollable content
	viewport viewport.Model

	// State
	showingHeaders bool // Toggle between headers and body view

	// UI dimensions
	width  int
	height int
}

// NewResponseModel creates a new response viewer model
func NewResponseModel() ResponseModel {
	vp := viewport.New(80, 20)
	vp.SetContent("No response yet. Send a request to see the response here.")

	return ResponseModel{
		viewport:       vp,
		showingHeaders: false,
	}
}

// Init initializes the model
func (m ResponseModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ResponseModel) Update(msg tea.Msg) (ResponseModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "h":
			// Toggle headers/body view
			m.showingHeaders = !m.showingHeaders
			m.updateViewportContent()
			return m, nil

		default:
			// Pass other keys to viewport for scrolling
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10 // Leave room for header and status bar
		m.updateViewportContent()

	case requestSentMsg:
		// Update response when request completes
		if msg.err == nil && msg.response != nil {
			m.response = msg.response
			m.updateViewportContent()
		}
	}

	return m, cmd
}

// View renders the response viewer
func (m ResponseModel) View() string {
	var sections []string

	sections = append(sections, "══ Response ══")
	sections = append(sections, "")

	if m.response == nil {
		sections = append(sections, "No response yet. Send a request to see the response here.")
		sections = append(sections, "")
		sections = append(sections, "h: toggle headers/body • ↑↓: scroll • q: quit")
		return strings.Join(sections, "\n")
	}

	// Status line
	statusLine := fmt.Sprintf("Status: %d %s", m.response.StatusCode, m.response.Status)
	sections = append(sections, statusLine)

	// Timing
	timingLine := fmt.Sprintf("Time: %dms", m.response.DurationMillis())
	sections = append(sections, timingLine)

	// Content length
	sizeLine := fmt.Sprintf("Size: %d bytes", m.response.ContentLength)
	sections = append(sections, sizeLine)
	sections = append(sections, "")

	// View mode indicator
	if m.showingHeaders {
		sections = append(sections, "═══ Headers ═══")
		sections = append(sections, m.renderHeaders())
	} else {
		sections = append(sections, "═══ Body ═══")
		sections = append(sections, m.viewport.View())
	}

	sections = append(sections, "")
	sections = append(sections, "h: toggle headers/body • ↑↓: scroll • q: quit")

	return strings.Join(sections, "\n")
}

func (m ResponseModel) renderHeaders() string {
	if m.response == nil || len(m.response.Headers) == 0 {
		return "No headers"
	}

	var lines []string
	for key, value := range m.response.Headers {
		lines = append(lines, fmt.Sprintf("%s: %s", key, value))
	}
	return strings.Join(lines, "\n")
}

// updateViewportContent updates the viewport with current response data
func (m *ResponseModel) updateViewportContent() {
	if m.response == nil {
		m.viewport.SetContent("No response yet. Send a request to see the response here.")
		return
	}

	// Content will be formatted in response_view.go
	// For now, use simple formatting
	content := ""
	if m.showingHeaders {
		content = "Headers view"
	} else {
		content = m.response.Body
	}

	m.viewport.SetContent(content)
}

// SetResponse sets the response to display
func (m *ResponseModel) SetResponse(response *domain.Response) {
	m.response = response
	m.showingHeaders = false
	m.updateViewportContent()
}

// GetResponse returns the current response
func (m *ResponseModel) GetResponse() *domain.Response {
	return m.response
}

// HasResponse returns whether a response is currently loaded
func (m *ResponseModel) HasResponse() bool {
	return m.response != nil
}
