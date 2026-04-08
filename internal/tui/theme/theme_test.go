package theme

import (
	"image/color"
	"testing"
)

func TestMethodColor(t *testing.T) {
	prev := Current
	t.Cleanup(func() { Current = prev })

	Current = Theme{
		MethodGET:    color.RGBA{0x01, 0x00, 0x00, 0xff},
		MethodPOST:   color.RGBA{0x02, 0x00, 0x00, 0xff},
		MethodPUT:    color.RGBA{0x03, 0x00, 0x00, 0xff},
		MethodDELETE: color.RGBA{0x04, 0x00, 0x00, 0xff},
		MethodPATCH:  color.RGBA{0x05, 0x00, 0x00, 0xff},
	}

	tests := []struct {
		name   string
		method string
		want   color.Color
	}{
		{
			name:   "default_is_get",
			method: "",
			want:   Current.MethodGET,
		},
		{
			name:   "trims_spaces_and_uppercases",
			method: "  post  ",
			want:   Current.MethodPOST,
		},
		{
			name:   "put",
			method: "PUT",
			want:   Current.MethodPUT,
		},
		{
			name:   "delete_mixed_case",
			method: "dElEtE",
			want:   Current.MethodDELETE,
		},
		{
			name:   "patch",
			method: "patch",
			want:   Current.MethodPATCH,
		},
		{
			name:   "unknown_falls_back_to_get",
			method: "HEAD",
			want:   Current.MethodGET,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Current.MethodColor(tt.method)
			if got != tt.want {
				t.Fatalf("MethodColor(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func themeColors(th Theme) []struct {
	name string
	c    color.Color
} {
	return []struct {
		name string
		c    color.Color
	}{
		{name: "DividerBorder", c: th.DividerBorder},
		{name: "TextMuted", c: th.TextMuted},
		{name: "TextDim", c: th.TextDim},
		{name: "TextPrimary", c: th.TextPrimary},
		{name: "TextAccent", c: th.TextAccent},
		{name: "SyntaxKeyword", c: th.SyntaxKeyword},
		{name: "SyntaxString", c: th.SyntaxString},
		{name: "SyntaxNumber", c: th.SyntaxNumber},
		{name: "SyntaxComment", c: th.SyntaxComment},
		{name: "SyntaxOperator", c: th.SyntaxOperator},
		{name: "MethodGET", c: th.MethodGET},
		{name: "MethodPOST", c: th.MethodPOST},
		{name: "MethodPUT", c: th.MethodPUT},
		{name: "MethodDELETE", c: th.MethodDELETE},
		{name: "MethodPATCH", c: th.MethodPATCH},
		{name: "TabsActiveText", c: th.TabsActiveText},
		{name: "TabsInactiveText", c: th.TabsInactiveText},
		{name: "StatusInfo", c: th.StatusInfo},
		{name: "StatusSuccess", c: th.StatusSuccess},
		{name: "StatusWarning", c: th.StatusWarning},
		{name: "StatusError", c: th.StatusError},
		{name: "StatusCodeOK", c: th.StatusCodeOK},
		{name: "StatusCodeError", c: th.StatusCodeError},
		{name: "OverlayBorder", c: th.OverlayBorder},
		{name: "OverlayFooter", c: th.OverlayFooter},
		{name: "RequestPaneLabel", c: th.RequestPaneLabel},
		{name: "ResponsePaneLabel", c: th.ResponsePaneLabel},
		{name: "TitleBarWorkspace", c: th.TitleBarWorkspace},
		{name: "TitleBarCollection", c: th.TitleBarCollection},
		{name: "TitleBarSeparator", c: th.TitleBarSeparator},
		{name: "TitleBarEntry", c: th.TitleBarEntry},
		{name: "RequestCancel", c: th.RequestCancel},
	}
}

func TestDefaultTheme_HasNonNilColors(t *testing.T) {
	th := DefaultTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("DefaultTheme().%s is nil", c.name)
			}
		})
	}
}

func TestPrincessTheme_HasNonNilColors(t *testing.T) {
	th := PrincessTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("PrincessTheme().%s is nil", c.name)
			}
		})
	}
}

func TestDraculaTheme_HasNonNilColors(t *testing.T) {
	th := DraculaTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("DraculaTheme().%s is nil", c.name)
			}
		})
	}
}

func TestCatppuccinTheme_HasNonNilColors(t *testing.T) {
	th := CatppuccinTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("CatppuccinTheme().%s is nil", c.name)
			}
		})
	}
}

func TestGruvboxTheme_HasNonNilColors(t *testing.T) {
	th := GruvboxTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("GruvboxTheme().%s is nil", c.name)
			}
		})
	}
}

func TestSolarizedTheme_HasNonNilColors(t *testing.T) {
	th := SolarizedTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("SolarizedTheme().%s is nil", c.name)
			}
		})
	}
}

func TestEverforestTheme_HasNonNilColors(t *testing.T) {
	th := EverforestTheme()
	for _, c := range themeColors(th) {
		t.Run(c.name, func(t *testing.T) {
			if c.c == nil {
				t.Fatalf("EverforestTheme().%s is nil", c.name)
			}
		})
	}
}

func TestThemeByName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTheme Theme
	}{
		{
			name:      "princess_lowercase",
			input:     "princess",
			wantTheme: PrincessTheme(),
		},
		{
			name:      "princess_uppercase",
			input:     "PRINCESS",
			wantTheme: PrincessTheme(),
		},
		{
			name:      "princess_mixed_case",
			input:     "Princess",
			wantTheme: PrincessTheme(),
		},
		{
			name:      "empty_falls_back_to_default",
			input:     "",
			wantTheme: DefaultTheme(),
		},
		{
			name:      "unknown_falls_back_to_default",
			input:     "nonexistent",
			wantTheme: DefaultTheme(),
		},
		{
			name:      "dracula",
			input:     "dracula",
			wantTheme: DraculaTheme(),
		},
		{
			name:      "catppuccin",
			input:     "catppuccin",
			wantTheme: CatppuccinTheme(),
		},
		{
			name:      "gruvbox",
			input:     "gruvbox",
			wantTheme: GruvboxTheme(),
		},
		{
			name:      "solarized",
			input:     "solarized",
			wantTheme: SolarizedTheme(),
		},
		{
			name:      "everforest",
			input:     "everforest",
			wantTheme: EverforestTheme(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThemeByName(tt.input)
			if got.TextAccent != tt.wantTheme.TextAccent {
				t.Fatalf("ThemeByName(%q).TextAccent = %v, want %v", tt.input, got.TextAccent, tt.wantTheme.TextAccent)
			}
		})
	}
}
