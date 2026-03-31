package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Theme defines a set of colors for the ferret UI.
type Theme struct {
	DividerBorder color.Color

	TextMuted      color.Color
	TextDim        color.Color
	TextPrimary    color.Color
	TextAccent     color.Color
	SyntaxKeyword  color.Color
	SyntaxString   color.Color
	SyntaxNumber   color.Color
	SyntaxComment  color.Color
	SyntaxOperator color.Color

	MethodGET    color.Color
	MethodPOST   color.Color
	MethodPUT    color.Color
	MethodDELETE color.Color
	MethodPATCH  color.Color

	TabsActiveText   color.Color
	TabsInactiveText color.Color

	// Status bar (left) text colors by kind.
	StatusInfo    color.Color
	StatusSuccess color.Color
	StatusWarning color.Color
	StatusError   color.Color

	// HTTP status code colors.
	StatusCodeOK    color.Color
	StatusCodeError color.Color

	OverlayBorder color.Color
	OverlayFooter color.Color

	RequestPaneLabel  color.Color
	ResponsePaneLabel color.Color

	// TitleBarWorkspace is the config workspace name (left segment).
	TitleBarWorkspace color.Color

	// TitleBarCollection is the active collection folder name in the title bar.
	TitleBarCollection color.Color

	// TitleBarSeparator is the " / " between workspace, collection, and entry.
	TitleBarSeparator color.Color

	// TitleBarEntry highlights the loaded collection entry name.
	TitleBarEntry color.Color

	// RequestCancel is the color for the "^x to cancel" text in the status bar.
	RequestCancel color.Color
}

// Current is the process-wide active theme. Components read it for styling.
var Current = DefaultTheme()

// MethodColor returns the themed color for an HTTP method.
func MethodColor(method string) color.Color {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "POST":
		return Current.MethodPOST
	case "PUT":
		return Current.MethodPUT
	case "DELETE":
		return Current.MethodDELETE
	case "PATCH":
		return Current.MethodPATCH
	default:
		return Current.MethodGET
	}
}

// cc creates a CompleteColor with true color, ANSI256, and ANSI (16-color)
// fallbacks.
func cc(hex, ansi256, ansi string) color.Color {
	return compat.CompleteColor{
		TrueColor: lipgloss.Color(hex),
		ANSI256:   lipgloss.Color(ansi256),
		ANSI:      lipgloss.Color(ansi),
	}
}

// DefaultTheme is ferret's default color scheme. The accent teal (#5fd7af) is
// the primary brand color, with HTTP method colors chosen for quick scanning.
func DefaultTheme() Theme {
	return Theme{
		DividerBorder: cc("#333333", "236", "8"),

		TextMuted:      cc("#8a9099", "246", "7"),
		TextDim:        cc("#4f4f4f", "239", "8"),
		TextPrimary:    cc("#d7d7d7", "188", "7"),
		TextAccent:     cc("#5fd7af", "79", "6"),
		SyntaxKeyword:  cc("#5fd7af", "79", "6"),
		SyntaxString:   cc("#87ceeb", "117", "14"),
		SyntaxNumber:   cc("#ffb86c", "215", "3"),
		SyntaxComment:  cc("#6b7a8f", "246", "8"),
		SyntaxOperator: cc("#9aa2ab", "247", "7"),

		MethodGET:    cc("#5fd7af", "79", "6"),
		MethodPOST:   cc("#ff87d7", "213", "13"),
		MethodPUT:    cc("#f1fa8c", "228", "11"),
		MethodDELETE: cc("#ff5f87", "204", "1"),
		MethodPATCH:  cc("#ffb86c", "215", "3"),

		TabsActiveText:   cc("#ffffff", "231", "15"),
		TabsInactiveText: cc("#9aa2ab", "247", "7"),

		StatusInfo:    cc("#ffffff", "231", "15"),
		StatusSuccess: cc("#5fd7af", "79", "6"),
		StatusWarning: cc("#ffb86c", "215", "3"),
		StatusError:   cc("#ff5f87", "204", "1"),

		StatusCodeOK:    cc("#5fd7af", "79", "6"),
		StatusCodeError: cc("#ff5f87", "204", "1"),

		OverlayBorder: cc("#5fd7af", "79", "6"),
		OverlayFooter: cc("#8a9099", "246", "7"),

		RequestPaneLabel:  cc("#5fd7af", "79", "6"),
		ResponsePaneLabel: cc("#ff87d7", "215", "3"),

		TitleBarWorkspace:  cc("#5fd7af", "79", "6"),
		TitleBarCollection: cc("#87ceeb", "117", "14"),
		TitleBarSeparator:  cc("#6b7a8f", "246", "8"),
		TitleBarEntry:      cc("#ff9500", "208", "3"),

		RequestCancel: cc("#ff5f87", "204", "1"),
	}
}
