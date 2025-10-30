package views

import (
	"strings"
)

// RenderHelp renders the help screen with keyboard shortcuts.
func RenderHelp(_, _ int) string {
	var sections []string

	sections = append(sections, "")
	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "                    CURLY - HELP & SHORTCUTS")
	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "")

	// Global shortcuts.
	sections = append(sections, "GLOBAL SHORTCUTS:")
	sections = append(sections, "")
	sections = append(sections, "  q, Ctrl+C     Quit application")
	sections = append(sections, "  ?             Toggle help screen")
	sections = append(sections, "  Tab           Switch to next tab")
	sections = append(sections, "  Shift+Tab     Switch to previous tab")
	sections = append(sections, "  1             Jump to Request tab")
	sections = append(sections, "  2             Jump to Response tab")
	sections = append(sections, "  3             Jump to History tab")
	sections = append(sections, "")

	// Request tab shortcuts.
	sections = append(sections, "REQUEST TAB:")
	sections = append(sections, "")
	sections = append(sections, "  Tab           Move to next field")
	sections = append(sections, "  Shift+Tab     Move to previous field")
	sections = append(sections, "  Ctrl+Enter    Send request")
	sections = append(sections, "  Ctrl+R        Send request (alternative)")
	sections = append(sections, "  Ctrl+S        Save request (coming soon)")
	sections = append(sections, "  ←/→ or h/l    Change method selection")
	sections = append(sections, "  ←/→ or h/l    Change auth type")
	sections = append(sections, "")

	// Response tab shortcuts.
	sections = append(sections, "RESPONSE TAB:")
	sections = append(sections, "")
	sections = append(sections, "  h             Toggle between headers and body view")
	sections = append(sections, "  ↑/↓           Scroll response content")
	sections = append(sections, "  PgUp/PgDn     Page up/down")
	sections = append(sections, "")

	// History tab shortcuts.
	sections = append(sections, "HISTORY TAB:")
	sections = append(sections, "")
	sections = append(sections, "  ↑/↓ or k/j    Navigate history entries")
	sections = append(sections, "  Enter         Load selected entry (coming soon)")
	sections = append(sections, "  d, Delete     Delete selected entry")
	sections = append(sections, "  r             Refresh history")
	sections = append(sections, "  g, Home       Jump to first entry")
	sections = append(sections, "  G, End        Jump to last entry")
	sections = append(sections, "")

	sections = append(sections, "════════════════════════════════════════════════════════════")
	sections = append(sections, "")
	sections = append(sections, "                   Press ESC or ? to close")
	sections = append(sections, "")

	return strings.Join(sections, "\n")
}
