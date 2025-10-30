// Package views provides view rendering functions for the presentation layer.
package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/curly/internal/domain"
	"github.com/williajm/curly/internal/presentation/models"
	"github.com/williajm/curly/internal/presentation/styles"
)

// RenderRequestView renders the request builder form.
func RenderRequestView(m models.RequestModel) string {
	var sections []string

	// Title.
	title := styles.TitleStyle.Render("Request Builder")
	sections = append(sections, title)
	sections = append(sections, "")

	// Method selector.
	methodSection := renderMethodSelector(m)
	sections = append(sections, methodSection)
	sections = append(sections, "")

	// URL input.
	urlSection := renderURLInput(m)
	sections = append(sections, urlSection)
	sections = append(sections, "")

	// Name input.
	nameSection := renderNameInput(m)
	sections = append(sections, nameSection)
	sections = append(sections, "")

	// Headers section (simplified for MVP).
	headersSection := renderHeadersSection(m)
	sections = append(sections, headersSection)
	sections = append(sections, "")

	// Query params section (simplified for MVP).
	querySection := renderQueryParamsSection(m)
	sections = append(sections, querySection)
	sections = append(sections, "")

	// Body section.
	bodySection := renderBodySection(m)
	sections = append(sections, bodySection)
	sections = append(sections, "")

	// Auth section.
	authSection := renderAuthSection(m)
	sections = append(sections, authSection)
	sections = append(sections, "")

	// Send button.
	sendSection := renderSendButton(m)
	sections = append(sections, sendSection)

	// Error message if any.
	if m.IsLoading() {
		sections = append(sections, "")
		sections = append(sections, styles.TextStyle.Render("⠋ Sending request..."))
	}
	// TODO: Show error if present (access via reflection or add getter).

	// Help text.
	sections = append(sections, "")
	helpText := renderRequestHelp()
	sections = append(sections, helpText)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderMethodSelector renders the HTTP method selector.
func renderMethodSelector(_ models.RequestModel) string {
	label := styles.RenderLabel("Method")

	// Get current method index via reflection or add getter.
	// For now, we'll show all methods with a simple selector UI.
	methods := domain.SupportedMethods
	var methodButtons []string

	// In a real implementation, we'd track which method is selected.
	// For now, display all methods as buttons.
	for i, method := range methods {
		var style lipgloss.Style
		// Assume first method (GET) is default for display.
		if i == 0 {
			style = styles.ButtonStyle
		} else {
			style = styles.ButtonDisabledStyle
		}
		methodButtons = append(methodButtons, style.Render(method))
	}

	methodsRow := lipgloss.JoinHorizontal(lipgloss.Left, methodButtons...)

	return label + "\n" + methodsRow
}

// renderURLInput renders the URL input field.
func renderURLInput(_ models.RequestModel) string {
	label := styles.RenderLabel("URL")
	// The model's urlInput will be rendered here.
	// We'll need to access it via a public method.
	inputView := "[URL Input Field]" // Placeholder

	return label + "\n" + inputView
}

// renderNameInput renders the name input field.
func renderNameInput(_ models.RequestModel) string {
	label := styles.RenderLabel("Name (optional)")
	inputView := "[Name Input Field]" // Placeholder

	return label + "\n" + inputView
}

// renderHeadersSection renders the headers editor.
func renderHeadersSection(_ models.RequestModel) string {
	label := styles.RenderLabel("Headers")
	help := styles.HelpStyle.Render("Add custom HTTP headers (key: value, one per line)")

	// For MVP, we'll show a simple text area.
	content := styles.DimmedStyle.Render("Headers editor coming soon...")

	return label + "\n" + help + "\n" + content
}

// renderQueryParamsSection renders the query parameters editor.
func renderQueryParamsSection(_ models.RequestModel) string {
	label := styles.RenderLabel("Query Parameters")
	help := styles.HelpStyle.Render("Add query parameters (key=value, one per line)")

	// For MVP, we'll show a simple text area.
	content := styles.DimmedStyle.Render("Query params editor coming soon...")

	return label + "\n" + help + "\n" + content
}

// renderBodySection renders the request body editor.
func renderBodySection(_ models.RequestModel) string {
	label := styles.RenderLabel("Body")
	help := styles.HelpStyle.Render("Request body (JSON, text, etc.)")

	// The model's bodyTextArea will be rendered here.
	bodyView := "[Body Text Area]" // Placeholder

	return label + "\n" + help + "\n" + bodyView
}

// renderAuthSection renders the authentication selector.
func renderAuthSection(_ models.RequestModel) string {
	label := styles.RenderLabel("Authentication")

	authTypes := []string{"None", "Basic", "Bearer", "API Key"}
	var authButtons []string

	// Display auth type options.
	for i, authType := range authTypes {
		var style lipgloss.Style
		// Assume None is default.
		if i == 0 {
			style = styles.ButtonStyle
		} else {
			style = styles.ButtonDisabledStyle
		}
		authButtons = append(authButtons, style.Render(authType))
	}

	authRow := lipgloss.JoinHorizontal(lipgloss.Left, authButtons...)

	return label + "\n" + authRow
}

// renderSendButton renders the send request button.
func renderSendButton(m models.RequestModel) string {
	buttonText := "Send Request"
	if m.IsLoading() {
		return styles.ButtonDisabledStyle.Render(buttonText)
	}
	return styles.ButtonStyle.Render(buttonText)
}

// renderRequestHelp renders help text for the request view.
func renderRequestHelp() string {
	shortcuts := []string{
		styles.RenderShortcut("Tab", "next field"),
		styles.RenderShortcut("Shift+Tab", "previous field"),
		styles.RenderShortcut("Ctrl+Enter", "send request"),
		styles.RenderShortcut("Ctrl+S", "save request"),
		styles.RenderShortcut("?", "help"),
	}

	return strings.Join(shortcuts, " • ")
}

// Note: This is a simplified version for the initial implementation.
// The actual view will need to access model fields properly.
// We'll need to add getter methods to the model or expose fields publicly.
