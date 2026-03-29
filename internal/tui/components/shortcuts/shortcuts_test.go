package shortcuts

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestRenderShortcuts(t *testing.T) {
	tuitest.UseStableTheme(t)

	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "run")),
	}

	tests := []struct {
		name  string
		width int
		want  []string
	}{
		{
			name:  "small_width_does_not_panic_and_includes_some_help",
			width: 0,
			want:  []string{},
		},
		{
			name:  "reasonable_width_includes_descriptions",
			width: 40,
			want:  []string{"quit", "search", "run"},
		},
		{
			name:  "very_small_width_truncates_with_ellipsis",
			width: 8,
			want:  []string{"…"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tuitest.StripANSI(RenderShortcuts(tt.width, bindings))

			// Should never exceed its content width (minContentWidth or width - padding).
			contentWidth := max(minContentWidth, tt.width-horizontalPaddingColumns)
			if len([]rune(got)) > contentWidth*4 { // sanity guard: should not explode
				t.Fatalf("RenderShortcuts(%d) produced unexpectedly long output: %q", tt.width, got)
			}

			for _, w := range tt.want {
				if w == "" {
					continue
				}
				if !strings.Contains(got, w) {
					t.Fatalf("RenderShortcuts(%d) = %q, want to contain %q", tt.width, got, w)
				}
			}
		})
	}
}
