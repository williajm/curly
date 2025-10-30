// Package components provides reusable UI components for the TUI.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/williajm/curly/internal/presentation/styles"
)

// RenderTabs renders a horizontal tab navigation component
func RenderTabs(tabs []string, activeTab int) string {
	var renderedTabs []string

	for i, tab := range tabs {
		var style lipgloss.Style
		if i == activeTab {
			style = styles.ActiveTabStyle
		} else {
			style = styles.TabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(tab))
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Add indicator line under active tab
	indicator := renderTabIndicator(tabs, activeTab)

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, indicator)
}

// renderTabIndicator creates the underline indicator for the active tab
func renderTabIndicator(tabs []string, activeTab int) string {
	var parts []string

	for i, tab := range tabs {
		// Calculate the width of each tab (including padding)
		// ActiveTabStyle and TabStyle both have padding(0, 2), so +4 characters
		width := len(tab) + 4

		if i == activeTab {
			// Active tab gets the indicator
			indicator := strings.Repeat("━", width)
			parts = append(parts, styles.TabIndicatorStyle.Render(indicator))
		} else {
			// Inactive tabs get spaces
			parts = append(parts, strings.Repeat(" ", width))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// RenderStatusBar renders a status bar with optional message and shortcuts
func RenderStatusBar(width int, message string, shortcuts []string) string {
	if width <= 0 {
		return ""
	}

	// Render shortcuts on the right
	shortcutsStr := ""
	if len(shortcuts) > 0 {
		shortcutsStr = strings.Join(shortcuts, " • ")
	}

	// Calculate available space for message
	shortcutsWidth := lipgloss.Width(shortcutsStr)
	messageWidth := width - shortcutsWidth - 4 // Account for padding and spacing

	// Truncate message if necessary
	if len(message) > messageWidth && messageWidth > 3 {
		message = message[:messageWidth-3] + "..."
	}

	// Create left (message) and right (shortcuts) parts
	leftPart := message
	rightPart := shortcutsStr

	// Calculate spacing between parts
	usedWidth := len(leftPart) + len(rightPart)
	spacing := ""
	if usedWidth < width-2 { // Account for padding
		spacing = strings.Repeat(" ", width-usedWidth-2)
	}

	content := leftPart + spacing + rightPart

	return styles.StatusBarStyle.
		Width(width).
		Render(content)
}

// RenderBox renders content in a bordered box
func RenderBox(title string, content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Render title if provided
	titleStr := ""
	if title != "" {
		titleStr = styles.SubtitleStyle.Render(title)
	}

	// Create box style with dimensions
	boxStyle := styles.BoxStyle.
		Width(width - 4).  // Account for border and padding
		Height(height - 2) // Account for border

	boxContent := content
	if titleStr != "" {
		boxContent = titleStr + "\n\n" + content
	}

	return boxStyle.Render(boxContent)
}

// RenderKeyValuePair renders a key-value pair for headers/params
func RenderKeyValuePair(key, value string, focused bool) string {
	keyStyle := styles.KeyStyle
	valueStyle := styles.ValueStyle

	if focused {
		keyStyle = keyStyle.Background(styles.ColorBackgroundActive)
		valueStyle = valueStyle.Background(styles.ColorBackgroundActive)
	}

	return keyStyle.Render(key) + ": " + valueStyle.Render(value)
}

// RenderLoadingSpinner renders a simple loading indicator
func RenderLoadingSpinner(message string) string {
	spinner := "⠋" // Simple spinner character (can be animated in Update loop)
	return styles.TextStyle.Render(spinner + " " + message)
}

// RenderEmptyState renders an empty state message
func RenderEmptyState(message string) string {
	return styles.DimmedStyle.
		Italic(true).
		Render(message)
}
