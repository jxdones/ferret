package model

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/exec"
	"github.com/jxdones/ferret/internal/tui/components/collection"
	"github.com/jxdones/ferret/internal/tui/components/methodpicker"
	"github.com/jxdones/ferret/internal/tui/components/requestpane"
	"github.com/jxdones/ferret/internal/tui/components/responsepane"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/components/titlebar"
	"github.com/jxdones/ferret/internal/tui/components/urlbar"
	"github.com/jxdones/ferret/internal/tui/components/workspacepicker"
)

func TestUpdate_RequestPaneHandlesEscBeforeGlobal(t *testing.T) {
	m := newTestModel(t)
	m.focus = focusRequestPane
	m.lastPane = requestPane
	m.syncChildStateWithLayout()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "i", Code: 'i'}))
	m2, ok := updated.(Model)
	if !ok {
		t.Fatalf("unexpected model type: %T", updated)
	}

	updated, _ = m2.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	m3, ok := updated.(Model)
	if !ok {
		t.Fatalf("unexpected model type: %T", updated)
	}

	if m3.focus != focusRequestPane {
		t.Fatalf("focus = %v, want focusRequestPane", m3.focus)
	}
}

func TestUpdate_ModeTransitions_GlobalURLModal(t *testing.T) {
	m := newTestModel(t)
	m.focus = focusGlobal
	m.lastPane = requestPane
	m.syncChildStateWithLayout()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	m2 := updated.(Model)
	if m2.focus != focusURLBar {
		t.Fatalf("after tab from global, focus = %v, want focusURLBar", m2.focus)
	}

	updated, _ = m2.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	m3 := updated.(Model)
	if m3.focus != focusResponsePane {
		t.Fatalf("after shift+tab from URL, focus = %v, want focusResponsePane", m3.focus)
	}
	if m3.lastPane != responsePane {
		t.Fatalf("after shift+tab from URL, lastPane = %v, want responsePane", m3.lastPane)
	}

	updated, _ = m3.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	m4 := updated.(Model)
	if m4.focus != focusGlobal {
		t.Fatalf("after esc from response pane, focus = %v, want focusGlobal", m4.focus)
	}

	updated, _ = m4.Update(tea.KeyPressMsg(tea.Key{Text: "/", Code: '/'}))
	m5 := updated.(Model)
	if m5.focus != focusModalCollection || m5.activeModal != modalCollection {
		t.Fatalf("after /, focus/modal = (%v, %v), want (focusModalCollection, modalCollection)", m5.focus, m5.activeModal)
	}

	updated, _ = m5.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	m6 := updated.(Model)
	if m6.focus != focusResponsePane || m6.activeModal != modalNone {
		t.Fatalf("after esc close modal, focus/modal = (%v, %v), want (focusResponsePane, modalNone)", m6.focus, m6.activeModal)
	}
}

func TestUpdate_NewRequestShortcutClearsLoadedState(t *testing.T) {
	m := newTestModel(t)
	m.tab().urlbar.SetMethod("POST")
	m.tab().urlbar.SetURL("https://example.com/users")
	m.tab().requestPane.SetURL("https://example.com/users")
	m.tab().requestPane.SetHeaders(map[string]string{"Authorization": "Bearer token"})
	m.tab().requestPane.SetBody("{\"id\":1}")
	m.titlebar.SetEntry("existing request")
	m.tab().responsePane.SetResponse([]byte(`{"ok":true}`), map[string][]string{"Content-Type": {"application/json"}}, exec.Trace{
		Events: []exec.TraceEvent{{Name: "request started", Elapsed: 0}},
	}, false, 0)
	m.focus = focusRequestPane
	m.lastPane = requestPane
	m.syncChildStateWithLayout()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Code: 'n', Text: "n"}))
	m2 := updated.(Model)

	if got := m2.tab().urlbar.Method(); got != "GET" {
		t.Fatalf("method = %q, want GET", got)
	}
	if got := m2.tab().urlbar.URL(); got != "" {
		t.Fatalf("url = %q, want empty", got)
	}
	if got := m2.tab().requestPane.Headers(); len(got) != 0 {
		t.Fatalf("headers = %#v, want empty", got)
	}
	if got := m2.tab().requestPane.Body(); got != "" {
		t.Fatalf("body = %q, want empty", got)
	}
	if m2.focus != focusURLBar {
		t.Fatalf("focus = %v, want focusURLBar", m2.focus)
	}
	if view := m2.tab().responsePane.View().Content; view == "" {
		t.Fatalf("response view should still render after reset")
	}
}

func TestUpdate_URLBarTypingDoesNotTriggerMethodShortcuts(t *testing.T) {
	m := newTestModel(t)
	m.focus = focusURLBar
	m.syncChildStateWithLayout()

	updated, _ := m.Update(tea.KeyPressMsg(tea.Key{Text: "m", Code: 'm'}))
	m2 := updated.(Model)
	if got := m2.tab().urlbar.URL(); got != "m" {
		t.Fatalf("url after typing m = %q, want %q", got, "m")
	}
	if got := m2.tab().urlbar.Method(); got != "GET" {
		t.Fatalf("method after typing m = %q, want GET", got)
	}

	updated, _ = m2.Update(tea.KeyPressMsg(tea.Key{Text: "M", Code: 'M', Mod: tea.ModShift}))
	m3 := updated.(Model)
	if got := m3.tab().urlbar.URL(); got != "mM" {
		t.Fatalf("url after typing M = %q, want %q", got, "mM")
	}
	if got := m3.tab().urlbar.Method(); got != "GET" {
		t.Fatalf("method after typing M = %q, want GET", got)
	}
}

func TestUpdate_PasteMsgRoutesToBodyEditor(t *testing.T) {
	m := newTestModel(t)
	m.focus = focusRequestPane
	m.tab().requestPane.SetFocused(true)
	m.tab().requestPane, _, _ = m.tab().requestPane.Update(tea.KeyPressMsg(tea.Key{Code: ']', Text: "]"}))
	m.tab().requestPane, _, _ = m.tab().requestPane.Update(tea.KeyPressMsg(tea.Key{Code: ']', Text: "]"}))
	m.tab().requestPane, _, _ = m.tab().requestPane.Update(tea.KeyPressMsg(tea.Key{Code: 'l', Text: "l"}))
	m.tab().requestPane, _, _ = m.tab().requestPane.Update(tea.KeyPressMsg(tea.Key{Code: 'i', Text: "i"}))
	m.syncChildStateWithLayout()

	updated, _ := m.Update(tea.PasteMsg{Content: "{\"name\":\"ferret\"}"})
	m2 := updated.(Model)
	if got := m2.tab().requestPane.Body(); got != "{\"name\":\"ferret\"}" {
		t.Fatalf("Body() = %q, want pasted content", got)
	}
}

func TestBuildRequest_IncludesBody(t *testing.T) {
	m := newTestModel(t)
	m.tab().urlbar.SetMethod("POST")
	m.tab().urlbar.SetURL("https://example.com/users")
	m.tab().requestPane.SetHeaders(map[string]string{"Content-Type": "application/json"})
	m.tab().requestPane.SetBody("{\"name\":\"ferret\"}")

	req := m.buildRequest()
	if got := req.Body; got != "{\"name\":\"ferret\"}" {
		t.Fatalf("Body = %q, want JSON body", got)
	}
}

func newTestModel(t *testing.T) Model {
	t.Helper()

	workspace := t.TempDir()
	collectionRoot := filepath.Join(workspace, "collection")

	u := urlbar.New()
	u.SetMethod("GET")
	rp := requestpane.New()
	m := Model{
		width:    120,
		height:   40,
		titlebar: titlebar.New(),
		tabs: []requestTab{{
			title:        "new request",
			urlbar:       u,
			requestPane:  rp,
			responsePane: responsepane.New(),
		}},
		activeTab:       0,
		collection:      collection.New(),
		workspacePicker: workspacepicker.New(),
		statusbar:       statusbar.New(),
		methods:         methodpicker.New(),
		workspaceRoot:   workspace,
		collectionDirs:  []string{collectionRoot},
		collectionIndex: 0,
		collectionRoot:  collectionRoot,
		focus:           focusRequestPane,
		lastPane:        requestPane,
	}

	m.syncChildStateWithLayout()
	return m
}
