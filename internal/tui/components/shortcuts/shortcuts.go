package shortcuts

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	horizontalPaddingColumns = 2
	minContentWidth          = 8
)

// RenderShortcuts renders a compact shortcuts bar for the given width and
// bindings.
func RenderShortcuts(width int, bindings []key.Binding) string {
	contentWidth := max(minContentWidth, width-horizontalPaddingColumns)

	h := help.New()
	h.SetWidth(contentWidth)
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	h.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(theme.Current.DividerBorder)
	h.Styles.Ellipsis = lipgloss.NewStyle().Foreground(theme.Current.DividerBorder)

	return ansi.Truncate(h.ShortHelpView(bindings), contentWidth, "…")
}
