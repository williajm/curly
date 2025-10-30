// Package styles provides consistent styling for the TUI using Lipgloss.
package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette.
var (
	// Primary colors.
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#06B6D4") // Cyan

	// Status colors.
	ColorSuccess = lipgloss.Color("#10B981") // Green
	ColorError   = lipgloss.Color("#EF4444") // Red
	ColorWarning = lipgloss.Color("#F59E0B") // Orange
	ColorInfo    = lipgloss.Color("#3B82F6") // Blue

	// Status code colors.
	Color2xx = lipgloss.Color("#10B981") // Green for 2xx
	Color3xx = lipgloss.Color("#F59E0B") // Yellow for 3xx
	Color4xx = lipgloss.Color("#FB923C") // Orange for 4xx
	Color5xx = lipgloss.Color("#EF4444") // Red for 5xx

	// Text colors.
	ColorText       = lipgloss.Color("#F5F5F5") // Light gray
	ColorTextDimmed = lipgloss.Color("#9CA3AF") // Dimmed gray
	ColorTextBright = lipgloss.Color("#FFFFFF") // White

	// Background colors.
	ColorBackground       = lipgloss.Color("#1F2937") // Dark gray
	ColorBackgroundLight  = lipgloss.Color("#374151") // Light gray
	ColorBackgroundActive = lipgloss.Color("#4B5563") // Active state

	// Border colors.
	ColorBorder       = lipgloss.Color("#4B5563")
	ColorBorderActive = lipgloss.Color("#7C3AED")
)

// Base text styles.
var (
	// TextStyle is the default text style.
	TextStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// TitleStyle is for main titles.
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// SubtitleStyle is for subtitles and section headers.
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// DimmedStyle is for less important text.
	DimmedStyle = lipgloss.NewStyle().
			Foreground(ColorTextDimmed)

	// BoldStyle is for emphasized text.
	BoldStyle = lipgloss.NewStyle().
			Foreground(ColorTextBright).
			Bold(true)

	// SuccessStyle is for success messages.
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	// ErrorStyle is for error messages.
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// WarningStyle is for warning messages.
	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
)

// Component styles.
var (
	// TabStyle is for inactive tabs.
	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorTextDimmed).
			Background(ColorBackground)

	// ActiveTabStyle is for the active tab.
	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorTextBright).
			Background(ColorBackgroundActive).
			Bold(true)

	// TabIndicatorStyle is for the line under the active tab.
	TabIndicatorStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// InputStyle is for text input fields.
	InputStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)

	// InputFocusedStyle is for focused input fields.
	InputFocusedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderActive)

	// ButtonStyle is for buttons.
	ButtonStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Background(ColorPrimary).
			Foreground(ColorTextBright).
			Bold(true)

	// ButtonDisabledStyle is for disabled buttons.
	ButtonDisabledStyle = lipgloss.NewStyle().
				Padding(0, 2).
				Background(ColorBackgroundLight).
				Foreground(ColorTextDimmed)

	// StatusBarStyle is for the bottom status bar.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextBright).
			Background(ColorBackgroundLight).
			Padding(0, 1)

	// BoxStyle is for general containers.
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// SelectedItemStyle is for selected items in lists.
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)
)

// Status code styles.
var (
	// Status2xxStyle is for 2xx success responses.
	Status2xxStyle = lipgloss.NewStyle().
			Foreground(Color2xx).
			Bold(true)

	// Status3xxStyle is for 3xx redirect responses.
	Status3xxStyle = lipgloss.NewStyle().
			Foreground(Color3xx).
			Bold(true)

	// Status4xxStyle is for 4xx client error responses.
	Status4xxStyle = lipgloss.NewStyle().
			Foreground(Color4xx).
			Bold(true)

	// Status5xxStyle is for 5xx server error responses.
	Status5xxStyle = lipgloss.NewStyle().
			Foreground(Color5xx).
			Bold(true)
)

// GetStatusStyle returns the appropriate style for a status code.
func GetStatusStyle(statusCode int) lipgloss.Style {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return Status2xxStyle
	case statusCode >= 300 && statusCode < 400:
		return Status3xxStyle
	case statusCode >= 400 && statusCode < 500:
		return Status4xxStyle
	case statusCode >= 500 && statusCode < 600:
		return Status5xxStyle
	default:
		return TextStyle
	}
}

// Specialized styles for specific UI elements.
var (
	// LabelStyle is for form field labels.
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	// ValueStyle is for form field values.
	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// KeyStyle is for key-value pair keys (headers, query params).
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorInfo).
			Bold(true)

	// HelpStyle is for help text.
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextDimmed).
			Italic(true)

	// ShortcutStyle is for keyboard shortcuts.
	ShortcutStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// Helper functions for common styling operations.

// RenderKeyValue renders a key-value pair with consistent styling.
func RenderKeyValue(key, value string) string {
	return KeyStyle.Render(key) + ": " + ValueStyle.Render(value)
}

// RenderLabel renders a form label.
func RenderLabel(label string) string {
	return LabelStyle.Render(label + ":")
}

// RenderStatusCode renders a status code with appropriate color.
func RenderStatusCode(statusCode int, statusText string) string {
	style := GetStatusStyle(statusCode)
	return style.Render(statusText)
}

// RenderShortcut renders a keyboard shortcut with its description.
func RenderShortcut(key, description string) string {
	return ShortcutStyle.Render(key) + " " + DimmedStyle.Render(description)
}
