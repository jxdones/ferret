package methodpicker

import (
	"strings"
	"testing"

	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestModel_BaseBehavior(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(*Model)
		assertFn func(*testing.T, Model)
	}{
		{
			name: "new_model_has_expected_defaults",
			assertFn: func(t *testing.T, m Model) {
				if got := m.Selected(); got != "GET" {
					t.Fatalf("Selected() = %q, want %q", got, "GET")
				}
				if m.width != 30 {
					t.Fatalf("width = %d, want 30", m.width)
				}
				if len(m.methods) != len(defaultMethods) {
					t.Fatalf("methods len = %d, want %d", len(m.methods), len(defaultMethods))
				}
			},
		},
		{
			name: "set_size_clamps_to_minimum",
			prepare: func(m *Model) {
				m.SetSize(3)
			},
			assertFn: func(t *testing.T, m Model) {
				if m.width != 10 {
					t.Fatalf("width = %d, want 10", m.width)
				}
			},
		},
		{
			name: "set_active_matches_trimmed_case_insensitive_value",
			prepare: func(m *Model) {
				m.SetActive("  patch ")
			},
			assertFn: func(t *testing.T, m Model) {
				if got := m.Selected(); got != "PATCH" {
					t.Fatalf("Selected() = %q, want %q", got, "PATCH")
				}
			},
		},
		{
			name: "set_active_unknown_keeps_current_selection",
			prepare: func(m *Model) {
				m.SetActive("PUT")
				m.SetActive("NOT_A_METHOD")
			},
			assertFn: func(t *testing.T, m Model) {
				if got := m.Selected(); got != "PUT" {
					t.Fatalf("Selected() = %q, want %q", got, "PUT")
				}
			},
		},
		{
			name: "move_cursor_clamps_to_bounds",
			prepare: func(m *Model) {
				m.MoveCursor(100)
				m.MoveCursor(-100)
			},
			assertFn: func(t *testing.T, m Model) {
				if got := m.Selected(); got != "GET" {
					t.Fatalf("Selected() = %q, want %q", got, "GET")
				}
			},
		},
		{
			name: "selected_returns_empty_when_cursor_out_of_range",
			prepare: func(m *Model) {
				m.cursor = len(m.methods)
			},
			assertFn: func(t *testing.T, m Model) {
				if got := m.Selected(); got != "" {
					t.Fatalf("Selected() = %q, want empty string", got)
				}
			},
		},
		{
			name: "move_cursor_noop_when_methods_empty",
			prepare: func(m *Model) {
				m.methods = nil
				m.cursor = 2
				m.MoveCursor(1)
			},
			assertFn: func(t *testing.T, m Model) {
				if m.cursor != 2 {
					t.Fatalf("cursor = %d, want 2", m.cursor)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			if tt.prepare != nil {
				tt.prepare(&m)
			}
			tt.assertFn(t, m)
		})
	}
}

func TestModelView(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name    string
		model   Model
		want    []string
		wantNot []string
	}{
		{
			name:  "renders_all_methods_and_selected_indicator",
			model: New(),
			want:  []string{"▶GET", " POST", " PUT", " DELETE", " PATCH"},
		},
		{
			name: "renders_no_methods_placeholder",
			model: Model{
				methods: nil,
			},
			want: []string{"(no methods)"},
		},
		{
			name: "selected_indicator_moves_with_active_method",
			model: func() Model {
				m := New()
				m.SetActive("DELETE")
				return m
			}(),
			want:    []string{"▶DELETE", " GET"},
			wantNot: []string{"▶GET"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tuitest.StripANSI(tt.model.View().Content)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Fatalf("View() = %q, want to contain %q", got, w)
				}
			}
			for _, w := range tt.wantNot {
				if strings.Contains(got, w) {
					t.Fatalf("View() = %q, want not to contain %q", got, w)
				}
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name       string
		value      int
		low, high  int
		wantResult int
	}{
		{
			name:       "below_low",
			value:      0,
			low:        1,
			high:       5,
			wantResult: 1,
		},
		{
			name:       "within_range",
			value:      3,
			low:        1,
			high:       5,
			wantResult: 3,
		},
		{
			name:       "above_high",
			value:      6,
			low:        1,
			high:       5,
			wantResult: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp(tt.value, tt.low, tt.high); got != tt.wantResult {
				t.Fatalf("clamp(%d, %d, %d) = %d, want %d", tt.value, tt.low, tt.high, got, tt.wantResult)
			}
		})
	}
}
