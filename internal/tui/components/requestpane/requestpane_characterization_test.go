package requestpane

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestCharacterization_ParamsTabReadOnlyGuards(t *testing.T) {
	tuitest.UseStableTheme(t)

	m := New()
	m.SetFocused(true)
	m.SetSize(60, 10)
	m.SetURL("https://example.com/search?q=ferret")

	var handled bool
	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: ']', Text: "]"}))
	if !handled {
		t.Fatal("] should be handled to switch to params tab")
	}
	if m.tabs.ActiveLabel() != paramsTabLabel {
		t.Fatalf("active tab = %q, want %q", m.tabs.ActiveLabel(), paramsTabLabel)
	}

	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'i', Text: "i"}))
	if handled {
		t.Fatal("i should be unhandled on params tab (read-only)")
	}
	if m.headers.IsInserting() {
		t.Fatal("headers insert mode should remain false on params tab")
	}
}

func TestCharacterization_QueryTableTransitions(t *testing.T) {
	tuitest.UseStableTheme(t)

	m := New()
	m.SetFocused(true)
	m.SetSize(72, 10)
	m.SetURL("https://api.example.com/items?limit=10&cursor=abc")

	m, _, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ']', Text: "]"}))
	paramsView := tuitest.StripANSI(m.View().Content)
	if !strings.Contains(paramsView, "limit") || !strings.Contains(paramsView, "cursor") {
		t.Fatalf("params view = %q, want parsed query keys", paramsView)
	}

	m, _, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: '[', Text: "["}))
	headersView := tuitest.StripANSI(m.View().Content)
	if !strings.Contains(headersView, "Name") || !strings.Contains(headersView, "Value") {
		t.Fatalf("headers view = %q, want header table columns", headersView)
	}
}
