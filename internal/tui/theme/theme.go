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

// ThemeByName returns the theme matching name, falling back to DefaultTheme.
func ThemeByName(name string) Theme {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "princess":
		return PrincessTheme()
	case "dracula":
		return DraculaTheme()
	case "catppuccin":
		return CatppuccinTheme()
	case "gruvbox":
		return GruvboxTheme()
	case "solarized":
		return SolarizedTheme()
	case "everforest":
		return EverforestTheme()
	default:
		return DefaultTheme()
	}
}

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

// DraculaTheme is inspired by the Dracula color scheme.
func DraculaTheme() Theme {
	return Theme{
		DividerBorder: cc("#5f5f87", "60", "8"),

		TextMuted:      cc("#afafaf", "145", "8"),
		TextDim:        cc("#44475a", "59", "8"),
		TextPrimary:    cc("#87afd7", "110", "7"),
		TextAccent:     cc("#87d7ff", "117", "14"),
		SyntaxKeyword:  cc("#ff79c6", "212", "13"),
		SyntaxString:   cc("#f1fa8c", "228", "11"),
		SyntaxNumber:   cc("#bd93f9", "183", "5"),
		SyntaxComment:  cc("#6272a4", "60", "8"),
		SyntaxOperator: cc("#ff79c6", "212", "13"),

		MethodGET:    cc("#87d7ff", "117", "14"),
		MethodPOST:   cc("#ff79c6", "212", "13"),
		MethodPUT:    cc("#f1fa8c", "228", "11"),
		MethodDELETE: cc("#ff5f5f", "203", "1"),
		MethodPATCH:  cc("#ffaf5f", "215", "3"),

		TabsActiveText:   cc("#ffffff", "231", "15"),
		TabsInactiveText: cc("#afafaf", "145", "7"),

		StatusInfo:    cc("#87d7ff", "117", "14"),
		StatusSuccess: cc("#50fa7b", "84", "2"),
		StatusWarning: cc("#ffaf5f", "215", "3"),
		StatusError:   cc("#ff5f5f", "203", "1"),

		StatusCodeOK:    cc("#50fa7b", "84", "2"),
		StatusCodeError: cc("#ff5f5f", "203", "1"),

		OverlayBorder: cc("#875fff", "99", "5"),
		OverlayFooter: cc("#afafaf", "145", "7"),

		RequestPaneLabel:  cc("#875fff", "99", "5"),
		ResponsePaneLabel: cc("#87d7ff", "117", "14"),

		TitleBarWorkspace:  cc("#875fff", "99", "5"),
		TitleBarCollection: cc("#ff79c6", "212", "13"),
		TitleBarSeparator:  cc("#5f5f87", "60", "8"),
		TitleBarEntry:      cc("#ffaf5f", "215", "3"),

		RequestCancel: cc("#ff5f5f", "203", "1"),
	}
}

// CatppuccinTheme is the Catppuccin Mocha color scheme.
func CatppuccinTheme() Theme {
	return Theme{
		DividerBorder: cc("#6c7086", "60", "8"),

		TextMuted:      cc("#a6adc8", "146", "8"),
		TextDim:        cc("#45475a", "59", "8"),
		TextPrimary:    cc("#cdd6f4", "189", "7"),
		TextAccent:     cc("#cba6f7", "183", "5"),
		SyntaxKeyword:  cc("#cba6f7", "183", "5"),
		SyntaxString:   cc("#a6e3a1", "150", "2"),
		SyntaxNumber:   cc("#fab387", "216", "3"),
		SyntaxComment:  cc("#6c7086", "60", "8"),
		SyntaxOperator: cc("#89dceb", "116", "6"),

		MethodGET:    cc("#89dceb", "116", "14"),
		MethodPOST:   cc("#f38ba8", "211", "13"),
		MethodPUT:    cc("#f9e2af", "223", "11"),
		MethodDELETE: cc("#f38ba8", "211", "1"),
		MethodPATCH:  cc("#fab387", "216", "3"),

		TabsActiveText:   cc("#cdd6f4", "231", "15"),
		TabsInactiveText: cc("#a6adc8", "146", "7"),

		StatusInfo:    cc("#89dceb", "116", "14"),
		StatusSuccess: cc("#a6e3a1", "150", "2"),
		StatusWarning: cc("#f9e2af", "223", "3"),
		StatusError:   cc("#f38ba8", "211", "1"),

		StatusCodeOK:    cc("#a6e3a1", "150", "2"),
		StatusCodeError: cc("#f38ba8", "211", "1"),

		OverlayBorder: cc("#cba6f7", "183", "5"),
		OverlayFooter: cc("#a6adc8", "146", "7"),

		RequestPaneLabel:  cc("#cba6f7", "183", "5"),
		ResponsePaneLabel: cc("#89dceb", "116", "6"),

		TitleBarWorkspace:  cc("#cba6f7", "183", "5"),
		TitleBarCollection: cc("#89dceb", "116", "6"),
		TitleBarSeparator:  cc("#6c7086", "60", "8"),
		TitleBarEntry:      cc("#fab387", "216", "3"),

		RequestCancel: cc("#f38ba8", "211", "1"),
	}
}

// GruvboxTheme is the Gruvbox Dark color scheme.
func GruvboxTheme() Theme {
	return Theme{
		DividerBorder: cc("#7c6f64", "241", "8"),

		TextMuted:      cc("#928374", "102", "8"),
		TextDim:        cc("#504945", "59", "8"),
		TextPrimary:    cc("#ebdbb2", "223", "7"),
		TextAccent:     cc("#83a598", "109", "6"),
		SyntaxKeyword:  cc("#fb4934", "203", "1"),
		SyntaxString:   cc("#b8bb26", "142", "2"),
		SyntaxNumber:   cc("#d3869b", "175", "5"),
		SyntaxComment:  cc("#928374", "102", "8"),
		SyntaxOperator: cc("#83a598", "109", "6"),

		MethodGET:    cc("#83a598", "109", "6"),
		MethodPOST:   cc("#d3869b", "175", "13"),
		MethodPUT:    cc("#fabd2f", "214", "11"),
		MethodDELETE: cc("#fb4934", "203", "1"),
		MethodPATCH:  cc("#fe8019", "173", "3"),

		TabsActiveText:   cc("#fbf1c7", "229", "15"),
		TabsInactiveText: cc("#928374", "102", "7"),

		StatusInfo:    cc("#83a598", "109", "14"),
		StatusSuccess: cc("#b8bb26", "142", "2"),
		StatusWarning: cc("#fabd2f", "214", "3"),
		StatusError:   cc("#fb4934", "203", "1"),

		StatusCodeOK:    cc("#b8bb26", "142", "2"),
		StatusCodeError: cc("#fb4934", "203", "1"),

		OverlayBorder: cc("#83a598", "109", "6"),
		OverlayFooter: cc("#928374", "102", "7"),

		RequestPaneLabel:  cc("#83a598", "109", "6"),
		ResponsePaneLabel: cc("#8ec07c", "108", "2"),

		TitleBarWorkspace:  cc("#83a598", "109", "6"),
		TitleBarCollection: cc("#8ec07c", "108", "2"),
		TitleBarSeparator:  cc("#7c6f64", "241", "8"),
		TitleBarEntry:      cc("#fabd2f", "214", "3"),

		RequestCancel: cc("#fb4934", "203", "1"),
	}
}

// SolarizedTheme is the Solarized Dark color scheme.
func SolarizedTheme() Theme {
	return Theme{
		DividerBorder: cc("#585858", "240", "8"),

		TextMuted:      cc("#808080", "244", "8"),
		TextDim:        cc("#3a3a3a", "237", "8"),
		TextPrimary:    cc("#afafaf", "145", "7"),
		TextAccent:     cc("#00afaf", "37", "6"),
		SyntaxKeyword:  cc("#268bd2", "32", "4"),
		SyntaxString:   cc("#2aa198", "36", "6"),
		SyntaxNumber:   cc("#d33682", "125", "5"),
		SyntaxComment:  cc("#657b83", "66", "8"),
		SyntaxOperator: cc("#859900", "100", "2"),

		MethodGET:    cc("#00afaf", "37", "6"),
		MethodPOST:   cc("#d33682", "125", "13"),
		MethodPUT:    cc("#d75f00", "166", "11"),
		MethodDELETE: cc("#d70000", "160", "1"),
		MethodPATCH:  cc("#859900", "100", "2"),

		TabsActiveText:   cc("#ffffd7", "230", "15"),
		TabsInactiveText: cc("#808080", "244", "7"),

		StatusInfo:    cc("#00afaf", "37", "14"),
		StatusSuccess: cc("#859900", "100", "2"),
		StatusWarning: cc("#d75f00", "166", "3"),
		StatusError:   cc("#d70000", "160", "1"),

		StatusCodeOK:    cc("#859900", "100", "2"),
		StatusCodeError: cc("#d70000", "160", "1"),

		OverlayBorder: cc("#00afaf", "37", "6"),
		OverlayFooter: cc("#808080", "244", "7"),

		RequestPaneLabel:  cc("#00afaf", "37", "6"),
		ResponsePaneLabel: cc("#268bd2", "32", "4"),

		TitleBarWorkspace:  cc("#00afaf", "37", "6"),
		TitleBarCollection: cc("#268bd2", "32", "4"),
		TitleBarSeparator:  cc("#585858", "240", "8"),
		TitleBarEntry:      cc("#d75f00", "166", "3"),

		RequestCancel: cc("#d70000", "160", "1"),
	}
}

// EverforestTheme is the Everforest Dark Medium color scheme.
func EverforestTheme() Theme {
	return Theme{
		DividerBorder: cc("#4f585e", "239", "8"),

		TextMuted:      cc("#7a8478", "243", "8"),
		TextDim:        cc("#374145", "237", "8"),
		TextPrimary:    cc("#d3c6aa", "223", "7"),
		TextAccent:     cc("#a7c080", "142", "2"),
		SyntaxKeyword:  cc("#e67e80", "167", "1"),
		SyntaxString:   cc("#a7c080", "142", "2"),
		SyntaxNumber:   cc("#dbbc7f", "214", "3"),
		SyntaxComment:  cc("#7a8478", "243", "8"),
		SyntaxOperator: cc("#7fbbb3", "109", "6"),

		MethodGET:    cc("#a7c080", "142", "2"),
		MethodPOST:   cc("#e69875", "208", "13"),
		MethodPUT:    cc("#dbbc7f", "214", "11"),
		MethodDELETE: cc("#e67e80", "167", "1"),
		MethodPATCH:  cc("#7fbbb3", "109", "6"),

		TabsActiveText:   cc("#d3c6aa", "223", "15"),
		TabsInactiveText: cc("#7a8478", "243", "7"),

		StatusInfo:    cc("#7fbbb3", "109", "14"),
		StatusSuccess: cc("#a7c080", "142", "2"),
		StatusWarning: cc("#dbbc7f", "214", "3"),
		StatusError:   cc("#e67e80", "167", "1"),

		StatusCodeOK:    cc("#a7c080", "142", "2"),
		StatusCodeError: cc("#e67e80", "167", "1"),

		OverlayBorder: cc("#a7c080", "142", "2"),
		OverlayFooter: cc("#e69875", "208", "3"),

		RequestPaneLabel:  cc("#a7c080", "142", "2"),
		ResponsePaneLabel: cc("#7fbbb3", "109", "6"),

		TitleBarWorkspace:  cc("#a7c080", "142", "2"),
		TitleBarCollection: cc("#7fbbb3", "109", "6"),
		TitleBarSeparator:  cc("#4f585e", "239", "8"),
		TitleBarEntry:      cc("#dbbc7f", "214", "3"),

		RequestCancel: cc("#e67e80", "167", "1"),
	}
}

// PrincessTheme is inspired by the M365Princess oh-my-posh theme.
// Warm palette of plum, blush, salmon, and sky blue on a dark purple base.
func PrincessTheme() Theme {
	return Theme{
		DividerBorder: cc("#3d2040", "236", "8"),

		TextMuted:      cc("#c4a0c8", "182", "13"),
		TextDim:        cc("#5a3060", "54", "5"),
		TextPrimary:    cc("#f8e8ff", "255", "15"),
		TextAccent:     cc("#DA627D", "168", "13"),
		SyntaxKeyword:  cc("#9A348E", "127", "5"),
		SyntaxString:   cc("#DA627D", "168", "13"),
		SyntaxNumber:   cc("#FCA17D", "215", "3"),
		SyntaxComment:  cc("#7a5a7e", "96", "5"),
		SyntaxOperator: cc("#c4a0c8", "182", "7"),

		MethodGET:    cc("#86BBD8", "110", "14"),
		MethodPOST:   cc("#DA627D", "168", "13"),
		MethodPUT:    cc("#FCA17D", "215", "3"),
		MethodDELETE: cc("#9A348E", "127", "5"),
		MethodPATCH:  cc("#047E84", "30", "6"),

		TabsActiveText:   cc("#ffffff", "231", "15"),
		TabsInactiveText: cc("#c4a0c8", "182", "13"),

		StatusInfo:    cc("#86BBD8", "110", "14"),
		StatusSuccess: cc("#047E84", "30", "6"),
		StatusWarning: cc("#FCA17D", "215", "3"),
		StatusError:   cc("#CC3802", "166", "1"),

		StatusCodeOK:    cc("#047E84", "30", "6"),
		StatusCodeError: cc("#CC3802", "166", "1"),

		OverlayBorder: cc("#DA627D", "168", "13"),
		OverlayFooter: cc("#c4a0c8", "182", "13"),

		RequestPaneLabel:  cc("#9A348E", "127", "5"),
		ResponsePaneLabel: cc("#86BBD8", "110", "14"),

		TitleBarWorkspace:  cc("#9A348E", "127", "5"),
		TitleBarCollection: cc("#DA627D", "168", "13"),
		TitleBarSeparator:  cc("#7a5a7e", "96", "8"),
		TitleBarEntry:      cc("#FCA17D", "215", "3"),

		RequestCancel: cc("#CC3802", "166", "1"),
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
