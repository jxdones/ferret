package bodyeditor

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func keyPress(k tea.Key) tea.KeyPressMsg {
	return tea.KeyPressMsg(k)
}

func TestNew_Defaults(t *testing.T) {
	m := New()
	if m.Syntax() != SyntaxText {
		t.Fatalf("Syntax() = %q, want %q", m.Syntax(), SyntaxText)
	}
}

func TestSetValueAndSyntax(t *testing.T) {
	m := New()
	m.SetSyntax(SyntaxJSON)
	m.SetValue("{\"ok\":true}")
	if got := m.Value(); got != "{\"ok\":true}" {
		t.Fatalf("Value() = %q", got)
	}
	if got := m.Syntax(); got != SyntaxJSON {
		t.Fatalf("Syntax() = %q, want %q", got, SyntaxJSON)
	}
}

func TestUpdate_HandlesTyping(t *testing.T) {
	m := New()
	m.SetFocused(true)

	m, _ = m.Update(keyPress(tea.Key{Code: 'a', Text: "a"}))
	m, _ = m.Update(keyPress(tea.Key{Code: 'b', Text: "b"}))

	if got := m.Value(); got != "ab" {
		t.Fatalf("Value() = %q, want ab", got)
	}
}

func TestView_EmptyShowsPlaceholder(t *testing.T) {
	tuitest.UseStableTheme(t)
	m := New()
	out := tuitest.StripANSI(m.View())
	if !strings.Contains(out, "request body") {
		t.Fatalf("View() = %q, want placeholder", out)
	}
}
