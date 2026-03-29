package tabs

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestNew_DefaultWidthAndActive(t *testing.T) {
	tests := []struct {
		name       string
		labels     []string
		wantWidth  int
		wantActive int
	}{
		{
			name:       "empty_labels_uses_minimum_width",
			labels:     nil,
			wantWidth:  minTabsWidth,
			wantActive: 0,
		},
		{
			name:       "one_label_slot_width",
			labels:     []string{"Request"},
			wantWidth:  tabSlotWidth,
			wantActive: 0,
		},
		{
			name:       "two_labels_double_slot",
			labels:     []string{"Request", "Response"},
			wantWidth:  2 * tabSlotWidth,
			wantActive: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.labels)
			if m.width != tt.wantWidth {
				t.Fatalf("width = %d, want %d", m.width, tt.wantWidth)
			}
			if m.Active() != tt.wantActive {
				t.Fatalf("Active() = %d, want %d", m.Active(), tt.wantActive)
			}
		})
	}
}

func TestSetSize_ClampMinOne(t *testing.T) {
	m := New([]string{"A"})
	m.SetSize(0)
	if m.width != 1 {
		t.Fatalf("SetSize(0): width = %d, want 1", m.width)
	}
}

func TestModel_NavigationAndSelection(t *testing.T) {
	tests := []struct {
		name    string
		labels  []string
		prepare func(*Model)
		wantIdx int
		wantLbl string
	}{
		{
			name:    "set_active_in_bounds",
			labels:  []string{"a", "b", "c"},
			prepare: func(m *Model) { m.SetActive(1) },
			wantIdx: 1,
			wantLbl: "b",
		},
		{
			name:    "set_active_negative_unchanged",
			labels:  []string{"a", "b"},
			prepare: func(m *Model) { m.SetActive(0); m.SetActive(-1) },
			wantIdx: 0,
			wantLbl: "a",
		},
		{
			name:    "set_active_out_of_range_unchanged",
			labels:  []string{"a", "b"},
			prepare: func(m *Model) { m.SetActive(0); m.SetActive(9) },
			wantIdx: 0,
			wantLbl: "a",
		},
		{
			name:    "next_wraps_to_zero",
			labels:  []string{"a", "b"},
			prepare: func(m *Model) { m.SetActive(1); m.Next() },
			wantIdx: 0,
			wantLbl: "a",
		},
		{
			name:    "previous_wraps_to_last",
			labels:  []string{"a", "b"},
			prepare: func(m *Model) { m.Previous() },
			wantIdx: 1,
			wantLbl: "b",
		},
		{
			name:    "next_noop_when_no_labels",
			labels:  nil,
			prepare: func(m *Model) { m.Next() },
			wantIdx: 0,
			wantLbl: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.labels)
			if tt.prepare != nil {
				tt.prepare(&m)
			}
			if got := m.Active(); got != tt.wantIdx {
				t.Fatalf("Active() = %d, want %d", got, tt.wantIdx)
			}
			if got := m.ActiveLabel(); got != tt.wantLbl {
				t.Fatalf("ActiveLabel() = %q, want %q", got, tt.wantLbl)
			}
		})
	}
}

func TestActiveLabel_OutOfRangeCursor(t *testing.T) {
	m := New([]string{"x"})
	m.active = 1
	if got := m.ActiveLabel(); got != "" {
		t.Fatalf("ActiveLabel() = %q, want empty", got)
	}
}

func TestActiveSpan(t *testing.T) {
	tests := []struct {
		name      string
		model     Model
		wantStart int
		wantWidth int
	}{
		{
			name:      "empty_returns_zero",
			model:     New(nil),
			wantStart: 0,
			wantWidth: 0,
		},
		{
			name: "first_tab_one_based_start",
			model: func() Model {
				m := New([]string{"Foo", "Bar"})
				m.SetActive(0)
				return m
			}(),
			wantStart: 1,
			wantWidth: len("Foo"),
		},
		{
			name: "second_tab_includes_gap",
			model: func() Model {
				m := New([]string{"Foo", "Bar"})
				m.SetActive(1)
				return m
			}(),
			wantStart: 1 + len("Foo") + 2,
			wantWidth: len("Bar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotW := tt.model.ActiveSpan()
			if gotStart != tt.wantStart || gotW != tt.wantWidth {
				t.Fatalf("ActiveSpan() = (%d, %d), want (%d, %d)",
					gotStart, gotW, tt.wantStart, tt.wantWidth)
			}
		})
	}
}

func TestModel_View(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name      string
		setup     func(*Model)
		wantSubs  []string
		wantWidth int
	}{
		{
			name: "focused_contains_labels_and_full_width",
			setup: func(m *Model) {
				m.labels = []string{"Req", "Res"}
				m.SetSize(40)
				m.SetFocused(true)
				m.SetActive(0)
			},
			wantSubs:  []string{"Req", "Res"},
			wantWidth: 40,
		},
		{
			name: "blurred_contains_labels",
			setup: func(m *Model) {
				m.labels = []string{"Req", "Res"}
				m.SetSize(30)
				m.SetFocused(false)
			},
			wantSubs:  []string{"Req", "Res"},
			wantWidth: 30,
		},
		{
			name: "empty_labels_pads_to_width",
			setup: func(m *Model) {
				m.labels = nil
				m.SetSize(12)
			},
			wantSubs:  nil,
			wantWidth: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New([]string{"placeholder"})
			tt.setup(&m)
			out := tuitest.StripANSI(m.View().Content)
			if w := ansi.StringWidth(out); w != tt.wantWidth {
				t.Fatalf("View() width = %d, want %d (out=%q)", w, tt.wantWidth, out)
			}
			for _, sub := range tt.wantSubs {
				if !strings.Contains(out, sub) {
					t.Fatalf("View() = %q, want to contain %q", out, sub)
				}
			}
		})
	}
}
