package requestpane

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/components/bodyeditor"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func keyPress(k tea.Key) tea.KeyPressMsg {
	return tea.KeyPressMsg(k)
}

func TestNew_Defaults(t *testing.T) {
	m := New()
	if m.tabs.ActiveLabel() != headersTabLabel {
		t.Fatalf("first tab = %q, want %q", m.tabs.ActiveLabel(), headersTabLabel)
	}
	if len(m.Headers()) != 0 {
		t.Fatalf("Headers() = %v, want empty", m.Headers())
	}
	if m.width != 0 {
		t.Fatalf("width = %d, want 0 before SetSize", m.width)
	}
}

func TestSetSize_SetFocused_Propagate(t *testing.T) {
	m := New()
	m.SetSize(72, 0)
	if m.width != 72 {
		t.Fatalf("width = %d, want 72", m.width)
	}
	m.SetFocused(true)
	if !m.focused {
		t.Fatal("focused not set")
	}
}

func TestSetHeaders_RoundTrip(t *testing.T) {
	m := New()
	m.SetHeaders(map[string]string{"X-Test": "1"})
	if m.Headers()["X-Test"] != "1" {
		t.Fatalf("Headers() = %v", m.Headers())
	}
}

func TestSetBody_RoundTrip(t *testing.T) {
	m := New()
	m.SetBody("{\"ok\":true}")
	if got := m.Body(); got != "{\"ok\":true}" {
		t.Fatalf("Body() = %q", got)
	}
}

func TestSetHeaders_SelectsBodySyntax(t *testing.T) {
	m := New()
	m.SetHeaders(map[string]string{"Content-Type": "application/json"})
	if got := m.body.Syntax(); got != bodyeditor.SyntaxJSON {
		t.Fatalf("body syntax = %q, want %q", got, bodyeditor.SyntaxJSON)
	}
}

func TestSetURL(t *testing.T) {
	m := New()
	m.SetURL("https://example.com")
	if m.url != "https://example.com" {
		t.Fatalf("url = %q", m.url)
	}
}

func TestUpdate_Unfocused_NotHandled(t *testing.T) {
	m := New()
	m.SetFocused(false)
	_, _, handled := m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	if handled {
		t.Fatal("unfocused pane should not handle keys")
	}
}

func TestUpdate_TabBracketNavigation(t *testing.T) {
	m := New()
	m.SetFocused(true)

	m, _, handled := m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	if !handled {
		t.Fatal("] should be handled")
	}
	if m.tabs.ActiveLabel() != paramsTabLabel {
		t.Fatalf("after ], active = %q, want %q", m.tabs.ActiveLabel(), paramsTabLabel)
	}

	m, _, handled = m.Update(keyPress(tea.Key{Code: '[', Text: "["}))
	if !handled {
		t.Fatal("[ should be handled")
	}
	if m.tabs.ActiveLabel() != headersTabLabel {
		t.Fatalf("after [, active = %q, want %q", m.tabs.ActiveLabel(), headersTabLabel)
	}
}

func TestUpdate_NonHeadersTab_UnhandledContentKey(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))

	_, _, handled := m.Update(keyPress(tea.Key{Code: 'k', Text: "k"}))
	if handled {
		t.Fatal("content keys on non-headers tab should not be handled yet")
	}
}

func TestUpdate_NonHeadersTab_BracketsStillWork(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))

	m, _, handled := m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	if !handled {
		t.Fatal("] should still advance tabs from params")
	}
	if m.tabs.ActiveLabel() != bodyTabLabel {
		t.Fatalf("active = %q, want %q", m.tabs.ActiveLabel(), bodyTabLabel)
	}
}

func TestUpdate_BodyTabEditorAllowsBrackets(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: 'i', Text: "i"}))
	m, _, handled := m.Update(keyPress(tea.Key{Code: '[', Text: "["}))
	if !handled {
		t.Fatal("[ should be handled by the editor")
	}
	m, _, handled = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	if !handled {
		t.Fatal("] should be handled by the editor")
	}
	if got := m.Body(); got != "[]" {
		t.Fatalf("Body() = %q, want []", got)
	}
	m, _, handled = m.Update(keyPress(tea.Key{Code: tea.KeyEsc}))
	if !handled {
		t.Fatal("esc should exit editor focus")
	}
}

func TestUpdate_BodyTabEditorConsumesTab(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: 'i', Text: "i"}))

	before := m.tabs.ActiveLabel()
	m2, _, handled := m.Update(keyPress(tea.Key{Code: tea.KeyTab}))
	if !handled {
		t.Fatal("tab should be handled by the editor")
	}
	if got := m2.tabs.ActiveLabel(); got != before {
		t.Fatalf("active tab = %q, want %q", got, before)
	}
}

func TestUpdate_BodyTabSelectorTabBubblesForMainPaneNav(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))

	m2, _, handled := m.Update(keyPress(tea.Key{Code: tea.KeyTab}))
	if handled {
		t.Fatal("tab should not be consumed in body selector mode (root moves focus between panes)")
	}
	if got := m2.tabs.ActiveLabel(); got != bodyTabLabel {
		t.Fatalf("active tab = %q, want %q", got, bodyTabLabel)
	}
}

func TestUpdate_BodyTabSelectorShiftTabBubblesForMainPaneNav(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))

	m2, _, handled := m.Update(keyPress(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	if handled {
		t.Fatal("shift+tab should not be consumed in body selector mode")
	}
	if got := m2.tabs.ActiveLabel(); got != bodyTabLabel {
		t.Fatalf("active tab = %q, want %q", got, bodyTabLabel)
	}
}

func TestUpdate_BodyTabSelectorBracketsSwitchInnerTabs(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))

	m2, _, handled := m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	if !handled {
		t.Fatal("] should advance inner tabs from body selector")
	}
	if got := m2.tabs.ActiveLabel(); got != authTabLabel {
		t.Fatalf("active tab = %q, want %q", got, authTabLabel)
	}
	m3, _, handled := m2.Update(keyPress(tea.Key{Code: '[', Text: "["}))
	if !handled {
		t.Fatal("[ should go to previous inner tab from body selector")
	}
	if got := m3.tabs.ActiveLabel(); got != bodyTabLabel {
		t.Fatalf("active tab = %q, want %q", got, bodyTabLabel)
	}
}

func TestUpdate_BodyTabCtrlVHandledInEditor(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: 'i', Text: "i"}))

	m, _, handled := m.Update(keyPress(tea.Key{Code: 'v', Mod: tea.ModCtrl}))
	if !handled {
		t.Fatal("ctrl+v should be handled in body editor mode")
	}
}

func TestUpdate_BodyTabPasteMsgInsertsText(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: ']', Text: "]"}))
	m, _, _ = m.Update(keyPress(tea.Key{Code: 'i', Text: "i"}))

	next, _, handled := m.Update(tea.PasteMsg{Content: "{\"ok\":true}"})
	if !handled {
		t.Fatal("PasteMsg should be handled in body editor mode")
	}
	if got := next.Body(); got != "{\"ok\":true}" {
		t.Fatalf("Body() = %q, want pasted content", got)
	}
}

func TestUpdate_HeadersInsert_EnterAndExit(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)

	m, _, handled := m.Update(keyPress(tea.Key{Text: "i", Code: 'i'}))
	if !handled || !m.headers.IsInserting() {
		t.Fatal("i should enter headers insert mode")
	}

	m, _, handled = m.Update(keyPress(tea.Key{Code: tea.KeyEsc}))
	if !handled || m.headers.IsInserting() {
		t.Fatal("esc should cancel insert mode")
	}
}

func TestUpdate_Headers_AddRowViaInsert(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(60, 0)

	m, _, _ = m.Update(keyPress(tea.Key{Text: "i", Code: 'i'}))
	for _, r := range "X-Custom" {
		m, _, _ = m.Update(keyPress(tea.Key{Text: string(r), Code: r}))
	}
	m, _, _ = m.Update(keyPress(tea.Key{Code: tea.KeyTab}))
	for _, r := range "v" {
		m, _, _ = m.Update(keyPress(tea.Key{Text: string(r), Code: r}))
	}
	m, _, handled := m.Update(keyPress(tea.Key{Code: tea.KeyEnter}))
	if !handled {
		t.Fatal("enter should be handled in insert mode")
	}
	if m.Headers()["X-Custom"] != "v" {
		t.Fatalf("Headers() = %v", m.Headers())
	}
}

func TestUpdate_Headers_MoveCursorKeysHandled(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetHeaders(map[string]string{"a": "1", "b": "2"})

	_, _, handled := m.Update(keyPress(tea.Key{Code: 'j', Text: "j"}))
	if !handled {
		t.Fatal("j should be handled when headers tab has rows")
	}
	_, _, handled = m.Update(keyPress(tea.Key{Code: 'k', Text: "k"}))
	if !handled {
		t.Fatal("k should be handled")
	}
	_, _, handled = m.Update(keyPress(tea.Key{Code: tea.KeyUp}))
	if !handled {
		t.Fatal("up should be handled on headers tab with rows")
	}
}

func TestUpdate_Headers_CapitalA_StartsInsert(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.SetSize(50, 0)
	m, _, handled := m.Update(keyPress(tea.Key{Text: "A", Code: 'a'}))
	if !handled || !m.headers.IsInserting() {
		t.Fatal("A should start insert like i/I")
	}
}

func TestTabsView_IncludesActiveTab(t *testing.T) {
	tuitest.UseStableTheme(t)
	m := New()
	m.SetFocused(true)
	m.SetSize(50, 0)
	out := tuitest.StripANSI(m.TabsView().Content)
	if !strings.Contains(strings.ToLower(out), "headers") {
		t.Fatalf("TabsView() = %q", out)
	}
}

func TestActiveTabSpan_OnHeaders(t *testing.T) {
	m := New()
	start, w := m.ActiveTabSpan()
	if start < 1 || w < 1 {
		t.Fatalf("ActiveTabSpan() = (%d, %d), want positive span", start, w)
	}
}

func TestView_BodyTabRendersBodyEditor(t *testing.T) {
	m := New()
	m.SetSize(40, 0)
	m.tabs.SetActive(2) // body
	out := tuitest.StripANSI(m.View().Content)
	if !strings.Contains(out, "binary") {
		t.Fatalf("View() on body tab = %q, want body type line", out)
	}
	if !strings.Contains(out, "request body") {
		t.Fatalf("View() on body tab = %q, want body editor placeholder", out)
	}
}

func TestView_HeadersTabRendersTable(t *testing.T) {
	tuitest.UseStableTheme(t)
	m := New()
	m.SetSize(40, 0)
	out := tuitest.StripANSI(m.View().Content)
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Value") {
		t.Fatalf("View() = %q, want headers column titles", out)
	}
	if ansi.StringWidth(out) < 10 {
		t.Fatalf("View() unexpectedly short")
	}
}
