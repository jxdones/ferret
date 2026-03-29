// Package tuitest provides helpers shared by TUI unit tests (ANSI stripping,
// deterministic theme for styled output).
package tuitest

import (
	"image/color"
	"regexp"
	"testing"

	"github.com/jxdones/ferret/internal/tui/theme"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// StripANSI removes SGR escape sequences from s for stable assertions.
func StripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// StableTestTheme returns a fully populated theme using simple RGBA colors so
// lipgloss output is deterministic in tests.
func StableTestTheme() theme.Theme {
	return theme.Theme{
		DividerBorder: color.RGBA{0x33, 0x33, 0x33, 0xff},

		TextMuted:      color.RGBA{0x8a, 0x90, 0x99, 0xff},
		TextDim:        color.RGBA{0x4f, 0x4f, 0x4f, 0xff},
		TextPrimary:    color.RGBA{0xd7, 0xd7, 0xd7, 0xff},
		TextAccent:     color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		SyntaxKeyword:  color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		SyntaxString:   color.RGBA{0x87, 0xce, 0xeb, 0xff},
		SyntaxNumber:   color.RGBA{0xff, 0xb8, 0x6c, 0xff},
		SyntaxComment:  color.RGBA{0x6b, 0x7a, 0x8f, 0xff},
		SyntaxOperator: color.RGBA{0x9a, 0xa2, 0xab, 0xff},

		MethodGET:    color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		MethodPOST:   color.RGBA{0xff, 0x87, 0xd7, 0xff},
		MethodPUT:    color.RGBA{0xf1, 0xfa, 0x8c, 0xff},
		MethodDELETE: color.RGBA{0xff, 0x5f, 0x87, 0xff},
		MethodPATCH:  color.RGBA{0xff, 0xb8, 0x6c, 0xff},

		TabsActiveText:   color.RGBA{0xff, 0xff, 0xff, 0xff},
		TabsInactiveText: color.RGBA{0x9a, 0xa2, 0xab, 0xff},

		StatusInfo:    color.RGBA{0xff, 0xff, 0xff, 0xff},
		StatusSuccess: color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		StatusWarning: color.RGBA{0xff, 0xb8, 0x6c, 0xff},
		StatusError:   color.RGBA{0xff, 0x5f, 0x87, 0xff},

		StatusCodeOK:    color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		StatusCodeError: color.RGBA{0xff, 0x5f, 0x87, 0xff},

		OverlayBorder: color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		OverlayFooter: color.RGBA{0x8a, 0x90, 0x99, 0xff},

		RequestPaneLabel:   color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		ResponsePaneLabel:  color.RGBA{0xff, 0x87, 0xd7, 0xff},
		TitleBarWorkspace:  color.RGBA{0x5f, 0xd7, 0xaf, 0xff},
		TitleBarCollection: color.RGBA{0x87, 0xce, 0xeb, 0xff},
		TitleBarSeparator:  color.RGBA{0x6b, 0x7a, 0x8f, 0xff},
		TitleBarEntry:      color.RGBA{0xff, 0x95, 0x00, 0xff},
	}
}

// UseStableTheme sets theme.Current to StableTestTheme and restores the
// previous value when t finishes.
func UseStableTheme(t *testing.T) {
	t.Helper()
	prev := theme.Current
	t.Cleanup(func() { theme.Current = prev })
	theme.Current = StableTestTheme()
}
