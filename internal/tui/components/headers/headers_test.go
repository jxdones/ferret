package headers

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func keyPressRune(r rune) tea.Msg {
	return tea.KeyPressMsg(tea.Key{Text: string(r), Code: r})
}

func keyEnter() tea.Msg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
}

func keyEsc() tea.Msg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})
}

func keyTab() tea.Msg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
}

func keyShiftTab() tea.Msg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
}

func TestNew_Defaults(t *testing.T) {
	m := New()
	if m.width != 40 {
		t.Fatalf("width = %d, want 40", m.width)
	}
	if len(m.Headers()) != 0 {
		t.Fatalf("Headers() = %v, want empty", m.Headers())
	}
	if m.IsInserting() {
		t.Fatal("IsInserting() = true, want false")
	}
}

func TestSetSize_MinWidth(t *testing.T) {
	m := New()
	m.SetSize(3)
	if m.width != minHeaderRowWidth {
		t.Fatalf("width = %d, want %d", m.width, minHeaderRowWidth)
	}
}

func TestHeaders_SkipsEmptyName(t *testing.T) {
	m := New()
	m.items = []header{
		{name: "", value: "orphan"},
		{name: "X", value: "y"},
	}
	got := m.Headers()
	if len(got) != 1 || got["X"] != "y" {
		t.Fatalf("Headers() = %v, want map[X:y]", got)
	}
}

func TestSetHeaders_ResetsState(t *testing.T) {
	m := New()
	m.inserting = true
	m.cursor = 3
	m.SetHeaders(map[string]string{"A": "1"})
	if m.IsInserting() {
		t.Fatal("SetHeaders should clear inserting")
	}
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	if m.Headers()["A"] != "1" {
		t.Fatalf("Headers() missing A")
	}
}

func TestDeleteCursorRow(t *testing.T) {
	m := New()
	m.items = []header{
		{name: "a", value: "1"},
		{name: "b", value: "2"},
	}
	m.cursor = 0
	m.DeleteCursorRow()
	if len(m.items) != 1 || m.items[0].name != "b" || m.cursor != 0 {
		t.Fatalf("after delete first: items=%v cursor=%d", m.items, m.cursor)
	}
	m.DeleteCursorRow()
	if len(m.items) != 0 {
		t.Fatalf("want empty items, got %v", m.items)
	}
	m.DeleteCursorRow() // noop
}

func TestMoveCursor(t *testing.T) {
	tests := []struct {
		name    string
		items   []header
		start   int
		delta   int
		wantIdx int
	}{
		{
			name:    "empty_items_noop",
			items:   nil,
			start:   0,
			delta:   5,
			wantIdx: 0,
		},
		{
			name:    "clamps_high",
			items:   []header{{"a", ""}, {"b", ""}, {"c", ""}},
			start:   1,
			delta:   99,
			wantIdx: 2,
		},
		{
			name:    "clamps_low",
			items:   []header{{"a", ""}, {"b", ""}},
			start:   1,
			delta:   -99,
			wantIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.items = tt.items
			m.cursor = tt.start
			m.MoveCursor(tt.delta)
			if m.cursor != tt.wantIdx {
				t.Fatalf("cursor = %d, want %d", m.cursor, tt.wantIdx)
			}
		})
	}
}

func TestUpdate_InsertToggle(t *testing.T) {
	m := New()
	m.SetFocused(true)

	m, _ = m.Update(keyPressRune('i'))
	if !m.IsInserting() {
		t.Fatal("i should enter insert mode")
	}

	m, _ = m.Update(keyPressRune('I'))
	if !m.IsInserting() {
		t.Fatal("I should keep insert mode (already inserting)")
	}

	m, _ = m.Update(keyEsc())
	if m.IsInserting() {
		t.Fatal("esc should exit insert mode")
	}

	m, _ = m.Update(tea.KeyPressMsg(tea.Key{Text: "A", Code: 'a'}))
	if !m.IsInserting() {
		t.Fatal("A should enter insert mode")
	}

	m, _ = m.Update(keyEsc())
	if m.IsInserting() {
		t.Fatal("esc should exit insert mode")
	}
}

func TestUpdate_EnterAddsHeader(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60)

	m, _ = m.Update(keyPressRune('i'))
	m.nameInput.SetValue("  X-Name  ")
	m.valueInput.SetValue("  val  ")
	m, _ = m.Update(keyEnter())

	got := m.Headers()
	if got["X-Name"] != "val" {
		t.Fatalf("Headers() = %v, want X-Name -> val", got)
	}
	if !m.IsInserting() {
		t.Fatal("want insert mode after enter so another row can be added")
	}
	// Inputs cleared after successful add
	if m.nameInput.Value() != "" || m.valueInput.Value() != "" {
		t.Fatalf("inputs should clear after add (%q, %q)", m.nameInput.Value(), m.valueInput.Value())
	}
}

func TestUpdate_EnterSkipsEmptyName(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m, _ = m.Update(keyPressRune('i'))
	m.nameInput.SetValue("   ")
	m, _ = m.Update(keyEnter())
	if len(m.Headers()) != 0 {
		t.Fatalf("Headers() = %v, want empty", m.Headers())
	}
}

func TestUpdate_SwitchNameValueColumn(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m, _ = m.Update(keyPressRune('i'))
	m.nameInput.SetValue("N")

	m, _ = m.Update(keyTab())
	if m.activeCol != 1 {
		t.Fatalf("tab activeCol = %d, want 1", m.activeCol)
	}
	m.valueInput.SetValue("V")

	m, _ = m.Update(keyShiftTab())
	if m.activeCol != 0 {
		t.Fatalf("shift+tab activeCol = %d, want 0", m.activeCol)
	}

	m, _ = m.Update(keyTab())
	m, _ = m.Update(keyEnter())
	if m.Headers()["N"] != "V" {
		t.Fatalf("Headers() = %v", m.Headers())
	}
}

func TestUpdate_IgnoresInsertKeysWhenNotInserting(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m, _ = m.Update(keyEnter())
	if len(m.Headers()) != 0 {
		t.Fatal("enter without insert should not add")
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{
			name: "pads_spaces",
			s:    "ab",
			n:    5,
			want: "ab   ",
		},
		{
			name: "truncates",
			s:    "abcdef",
			n:    3,
			want: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.s, tt.n)
			if got != tt.want {
				t.Fatalf("padRight(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
			if w := ansi.StringWidth(got); w != tt.n {
				t.Fatalf("visible width = %d, want %d", w, tt.n)
			}
		})
	}
}

func TestReadOnlyView_Layout(t *testing.T) {
	tuitest.UseStableTheme(t)
	out := tuitest.StripANSI(ReadOnlyView(50, []Row{
		{Name: "Content-Type", Value: "application/json"},
	}).Content)
	for _, sub := range []string{"Name", "Value", "Content-Type", "application/json"} {
		if !strings.Contains(out, sub) {
			t.Fatalf("ReadOnlyView missing %q in:\n%s", sub, out)
		}
	}
}

func TestModel_View(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name     string
		setup    func(*Model)
		wantSubs []string
		wantNot  []string
	}{
		{
			name: "shows_column_headers_and_hint_when_not_inserting",
			setup: func(m *Model) {
				m.SetSize(50)
			},
			wantSubs: []string{
				"Name",
				"Value",
				"add",
				"d delete",
			},
		},
		{
			name: "shows_computed_defaults_when_not_defined",
			setup: func(m *Model) {
				m.SetSize(50)
			},
			wantSubs: []string{
				"computed at send time",
				"Accept",
				"User-Agent",
			},
		},
		{
			name: "hides_computed_header_when_defined_case_insensitive",
			setup: func(m *Model) {
				m.SetSize(50)
				m.SetHeaders(map[string]string{"accept": "application/json"})
			},
			wantSubs: []string{
				"User-Agent",
			},
			wantNot: []string{
				"*/*", // default Accept value row should not appear when Accept set
			},
		},
		{
			name: "inserting_row_shows_input_line",
			setup: func(m *Model) {
				m.SetFocused(true)
				m.SetSize(50)
				m.inserting = true
				m.applyInputFocus()
			},
			wantSubs: []string{
				"Name",
				"Value", // column headers
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			tt.setup(&m)
			out := tuitest.StripANSI(m.View().Content)
			for _, sub := range tt.wantSubs {
				if !strings.Contains(out, sub) {
					t.Fatalf("View() missing %q in:\n%s", sub, out)
				}
			}
			for _, sub := range tt.wantNot {
				if strings.Contains(out, sub) {
					t.Fatalf("View() should not contain %q", sub)
				}
			}
		})
	}
}
