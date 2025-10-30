// Package models contains Bubble Tea models for the presentation layer.
package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williajm/curly/internal/app"
	"github.com/williajm/curly/internal/domain"
)

// Field indices for focus management.
const (
	fieldMethod = iota
	fieldURL
	fieldName
	fieldHeaders
	fieldQueryParams
	fieldBody
	fieldAuthType
	fieldSend
	fieldCount // Total number of fields
)

// UI indicator constants.
const (
	// focusedIndicator is displayed next to focused input fields.
	focusedIndicator = " (*)"
)

// RequestModel represents the request builder form.
type RequestModel struct {
	// Services.
	requestService *app.RequestService
	authService    *app.AuthService

	// Current request being built.
	request *domain.Request

	// Form inputs.
	urlInput     textinput.Model
	nameInput    textinput.Model
	bodyTextArea textarea.Model

	// State.
	methodIndex  int // Index into supported methods
	focusedField int
	loading      bool
	errorMsg     string

	// UI dimensions.
	width  int
	height int

	// Simple key-value editors for headers and query params.
	// For MVP, we'll use simple string editing.
	headersText     string
	queryParamsText string
	authTypeIndex   int // Index into auth types
}

// Custom messages for async operations.
type requestSentMsg struct {
	response *domain.Response
	err      error
}

// NewRequestModel creates a new request builder model.
func NewRequestModel(requestService *app.RequestService, authService *app.AuthService) RequestModel {
	// Initialize text inputs.
	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.example.com/endpoint"
	urlInput.Width = 60
	urlInput.Focus()

	nameInput := textinput.New()
	nameInput.Placeholder = "My Request"
	nameInput.Width = 60

	// Initialize text area for body.
	bodyTextArea := textarea.New()
	bodyTextArea.Placeholder = "Request body (JSON, etc.)"
	bodyTextArea.SetWidth(60)
	bodyTextArea.SetHeight(8)
	// Disable ctrl+enter in textarea so we can handle it globally.
	bodyTextArea.KeyMap.InsertNewline.SetEnabled(false)

	return RequestModel{
		requestService:  requestService,
		authService:     authService,
		request:         domain.NewRequest(),
		urlInput:        urlInput,
		nameInput:       nameInput,
		bodyTextArea:    bodyTextArea,
		methodIndex:     0, // GET by default
		focusedField:    fieldURL,
		headersText:     "",
		queryParamsText: "",
		authTypeIndex:   0, // NoAuth by default
	}
}

// Init initializes the model.
func (m RequestModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model.
func (m RequestModel) Update(msg tea.Msg) (RequestModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys first (before field-specific handling).
		switch msg.String() {
		case KeyCtrlC:
			return m, tea.Quit

		case "ctrl+enter", "ctrl+r":
			// Send request - handle before field-specific keys.
			if !m.loading {
				return m, m.sendRequest()
			}
			return m, nil

		case "tab":
			// Move focus to next field.
			m.focusedField = (m.focusedField + 1) % fieldCount
			m.updateFocus()
			return m, nil

		case "shift+tab":
			// Move focus to previous field.
			m.focusedField = (m.focusedField - 1 + fieldCount) % fieldCount
			m.updateFocus()
			return m, nil
		}

		// Handle field-specific keys.
		switch m.focusedField {
		case fieldMethod:
			switch msg.String() {
			case "left", "h":
				if m.methodIndex > 0 {
					m.methodIndex--
				}
			case "right", "l":
				if m.methodIndex < len(domain.SupportedMethods)-1 {
					m.methodIndex++
				}
			}

		case fieldURL:
			var cmd tea.Cmd
			m.urlInput, cmd = m.urlInput.Update(msg)
			cmds = append(cmds, cmd)

		case fieldName:
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			cmds = append(cmds, cmd)

		case fieldBody:
			// Don't pass ctrl+enter/ctrl+r to textarea (handled globally above).
			if msg.String() != "ctrl+enter" && msg.String() != "ctrl+r" {
				var cmd tea.Cmd
				m.bodyTextArea, cmd = m.bodyTextArea.Update(msg)
				cmds = append(cmds, cmd)
			}

		case fieldAuthType:
			// Simple auth type selector (NoAuth, Basic, Bearer, APIKey).
			switch msg.String() {
			case "left", "h":
				if m.authTypeIndex > 0 {
					m.authTypeIndex--
				}
			case "right", "l":
				if m.authTypeIndex < 3 { // 4 auth types: NoAuth, Basic, Bearer, APIKey
					m.authTypeIndex++
				}
			}

		case fieldSend:
			if msg.String() == "enter" || msg.String() == " " {
				if !m.loading {
					return m, m.sendRequest()
				}
			}
		}

	case requestSentMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.errorMsg = ""
			// Response will be handled by the parent model.
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust input widths.
		m.urlInput.Width = min(60, msg.Width-20)
		m.nameInput.Width = min(60, msg.Width-20)
		m.bodyTextArea.SetWidth(min(60, msg.Width-20))
	}

	return m, tea.Batch(cmds...)
}

// View renders the request builder form.
func (m RequestModel) View() string {
	return m.RenderView()
}

// RenderView renders the full request view with all components.
func (m RequestModel) RenderView() string {
	var sections []string

	// Import the view package would cause a cycle, so we'll implement here.
	sections = append(sections, m.renderTitle())
	sections = append(sections, "")
	sections = append(sections, m.renderMethod())
	sections = append(sections, "")
	sections = append(sections, m.renderURL())
	sections = append(sections, "")
	sections = append(sections, m.renderName())
	sections = append(sections, "")
	sections = append(sections, m.renderBody())
	sections = append(sections, "")
	sections = append(sections, m.renderAuth())
	sections = append(sections, "")
	sections = append(sections, m.renderSendButton())

	if m.loading {
		sections = append(sections, "")
		sections = append(sections, "⠋ Sending request...")
	}

	if m.errorMsg != "" {
		sections = append(sections, "")
		sections = append(sections, "Error: "+m.errorMsg)
	}

	sections = append(sections, "")
	sections = append(sections, m.renderHelp())

	return strings.Join(sections, "\n")
}

func (m RequestModel) renderTitle() string {
	return "══ Request Builder ══"
}

func (m RequestModel) renderMethod() string {
	label := "Method: "
	methods := domain.SupportedMethods
	var parts []string
	for i, method := range methods {
		if i == m.methodIndex {
			parts = append(parts, "["+method+"]")
		} else {
			parts = append(parts, method)
		}
	}
	return label + strings.Join(parts, " ")
}

func (m RequestModel) renderURL() string {
	label := "URL:"
	focused := ""
	if m.focusedField == fieldURL {
		focused = focusedIndicator
	}
	return label + focused + "\n" + m.urlInput.View()
}

func (m RequestModel) renderName() string {
	label := "Name (optional):"
	focused := ""
	if m.focusedField == fieldName {
		focused = focusedIndicator
	}
	return label + focused + "\n" + m.nameInput.View()
}

func (m RequestModel) renderBody() string {
	label := "Body:"
	focused := ""
	if m.focusedField == fieldBody {
		focused = focusedIndicator
	}
	return label + focused + "\n" + m.bodyTextArea.View()
}

func (m RequestModel) renderAuth() string {
	authTypes := []string{"None", "Basic", "Bearer", "API Key"}
	label := "Auth: "
	var parts []string
	for i, authType := range authTypes {
		if i == m.authTypeIndex {
			parts = append(parts, "["+authType+"]")
		} else {
			parts = append(parts, authType)
		}
	}
	return label + strings.Join(parts, " ")
}

func (m RequestModel) renderSendButton() string {
	if m.loading {
		return "[Sending...]"
	}
	focused := ""
	if m.focusedField == fieldSend {
		focused = focusedIndicator
	}
	return "[Send Request]" + focused
}

func (m RequestModel) renderHelp() string {
	return "Tab: next • Shift+Tab: prev • Ctrl+Enter: send • ?: help • q: quit"
}

// updateFocus updates which input field has focus.
func (m *RequestModel) updateFocus() {
	// Blur all inputs.
	m.urlInput.Blur()
	m.nameInput.Blur()
	m.bodyTextArea.Blur()

	// Focus the active field.
	switch m.focusedField {
	case fieldURL:
		m.urlInput.Focus()
	case fieldName:
		m.nameInput.Focus()
	case fieldBody:
		m.bodyTextArea.Focus()
	}
}

// sendRequest creates a command to send the HTTP request.
func (m *RequestModel) sendRequest() tea.Cmd {
	// Build request from form inputs.
	req := m.buildRequest()

	// Validate request.
	if err := req.Validate(); err != nil {
		return func() tea.Msg {
			return requestSentMsg{err: err}
		}
	}

	m.loading = true

	return func() tea.Msg {
		ctx := context.Background()
		resp, err := m.requestService.ExecuteAndSave(ctx, req)
		return requestSentMsg{response: resp, err: err}
	}
}

// buildRequest constructs a domain.Request from the form inputs.
func (m *RequestModel) buildRequest() *domain.Request {
	req := m.request

	// Set method.
	req.Method = domain.SupportedMethods[m.methodIndex]

	// Set URL.
	req.URL = m.urlInput.Value()

	// Set name.
	req.Name = m.nameInput.Value()
	if req.Name == "" {
		req.Name = fmt.Sprintf("%s %s", req.Method, req.URL)
	}

	// Set body.
	req.Body = m.bodyTextArea.Value()

	// Parse headers from text (simple format: "Key: Value" per line).
	// For MVP, we'll skip complex parsing.

	// Parse query params from text (simple format: "key=value" per line).
	// For MVP, we'll skip complex parsing.

	// Set auth based on authTypeIndex.
	switch m.authTypeIndex {
	case 0:
		req.AuthConfig = domain.NewNoAuth()
	case 1:
		// Basic auth - for MVP, hardcoded or skipped.
		req.AuthConfig = domain.NewNoAuth()
	case 2:
		// Bearer token - for MVP, hardcoded or skipped.
		req.AuthConfig = domain.NewNoAuth()
	case 3:
		// API Key - for MVP, hardcoded or skipped.
		req.AuthConfig = domain.NewNoAuth()
	}

	return req
}

// GetRequest returns the current request being built.
func (m *RequestModel) GetRequest() *domain.Request {
	return m.buildRequest()
}

// SetError sets an error message to display.
func (m *RequestModel) SetError(err string) {
	m.errorMsg = err
}

// IsLoading returns whether a request is currently being sent.
func (m *RequestModel) IsLoading() bool {
	return m.loading
}
