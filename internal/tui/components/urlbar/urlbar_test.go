package urlbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestFit(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		want  string
	}{
		{
			name:  "pads_to_width",
			in:    "abc",
			width: 5,
			want:  "abc  ",
		},
		{
			name:  "truncates_to_width",
			in:    "abcdef",
			width: 3,
			want:  "abc",
		},
		{
			name:  "zero_width_empty",
			in:    "abcdef",
			width: 0,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fit(tt.in, tt.width)
			if got != tt.want {
				t.Fatalf("fit(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
			if w := ansi.StringWidth(got); w != tt.width {
				t.Fatalf("fit(%q, %d) width = %d, want %d", tt.in, tt.width, w, tt.width)
			}
		})
	}
}

func TestModel_BaseBehavior(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name       string
		method     string
		url        string
		width      int
		focused    bool
		wantSubs   []string
		wantWidth  int
		wantHasEll bool
	}{
		{
			name:      "renders_method_and_url_when_blurred",
			method:    "GET",
			url:       "https://example.com",
			width:     40,
			focused:   false,
			wantSubs:  []string{"GET", "https://example.com"},
			wantWidth: 40,
		},
		{
			name:       "truncates_long_url_when_blurred_with_ellipsis",
			method:     "GET",
			url:        "https://example.com/this/is/a/very/long/path/that/should/truncate",
			width:      20,
			focused:    false,
			wantSubs:   []string{"GET", "https://"},
			wantWidth:  20,
			wantHasEll: true,
		},
		{
			name:      "focused_view_does_not_use_ellipsis_truncator",
			method:    "POST",
			url:       "https://example.com/this/is/a/very/long/path/that/should/truncate",
			width:     20,
			focused:   true,
			wantSubs:  []string{"POST"},
			wantWidth: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetSize(tt.width)
			m.SetMethod(tt.method)
			m.SetURL(tt.url)
			m.SetFocused(tt.focused)

			out := tuitest.StripANSI(m.View().Content)
			if w := ansi.StringWidth(out); w != tt.wantWidth {
				t.Fatalf("View() width = %d, want %d (output=%q)", w, tt.wantWidth, out)
			}
			for _, sub := range tt.wantSubs {
				if !strings.Contains(out, sub) {
					t.Fatalf("View() = %q, want to contain %q", out, sub)
				}
			}
			if tt.wantHasEll && !strings.Contains(out, "…") {
				t.Fatalf("View() = %q, want to contain ellipsis", out)
			}
			if tt.focused && strings.Contains(out, "…") {
				// Focused path uses fit() + truncate-without-ellipsis; it may still contain unicode
				// ellipsis if user typed it, but not from truncation here.
				t.Fatalf("View() = %q, want not to contain ellipsis when focused", out)
			}
		})
	}
}
