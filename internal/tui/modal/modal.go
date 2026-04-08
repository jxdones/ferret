package modal

import (
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/tui/theme"
)

// boxStyle matches the outer frame used by [Render] (border + padding). Keep
// this in sync with Render so [InnerWidth] subtracts the correct horizontal
// frame.
func boxStyle(outerWidth int, t theme.Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(outerWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.OverlayBorder).
		Padding(0, 1)
}

// Render produces a styled modal box with a title, body content, and an
// optional footer line. Width is the total outer width including borders and
// padding.
func Render(title, content, footer string, width int, t theme.Theme) string {
	titleLine := lipgloss.NewStyle().
		Foreground(t.TextAccent).
		Bold(true).
		Render(title)

	body := titleLine + "\n\n" + content
	if footer != "" {
		footerLine := lipgloss.NewStyle().
			Foreground(t.OverlayFooter).
			Render(footer)
		body += "\n\n" + footerLine
	}

	return boxStyle(width, t).Render(body)
}

// InnerWidth returns usable content width inside a modal of the given outer
// width.
func InnerWidth(outerWidth int) int {
	frame := boxStyle(outerWidth, theme.DefaultTheme()).GetHorizontalFrameSize()
	inner := outerWidth - frame
	if inner < 1 {
		return 1
	}
	return inner
}
