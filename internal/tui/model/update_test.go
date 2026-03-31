package model

import (
	"errors"
	"path/filepath"
	"strings"
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

func TestOnRequestFinished(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Model)
		msg   func(Model) RequestFinishedMsg
		check func(*testing.T, Model)
	}{
		{
			name:  "response_lands_on_correct_tab_leaves_other_tab_empty",
			setup: func(m *Model) { m.newTab(); m.activeTab = 0 },
			msg: func(m Model) RequestFinishedMsg {
				return RequestFinishedMsg{TabID: m.tabs[0].id, Body: []byte(`{"ok":true}`)}
			},
			check: func(t *testing.T, next Model) {
				if strings.Contains(next.tabs[0].responsePane.View().Content, "send a request") {
					t.Fatal("tab 0 should have a response, not the empty placeholder")
				}
				if !strings.Contains(next.tabs[1].responsePane.View().Content, "send a request") {
					t.Fatal("tab 1 should still show the empty placeholder")
				}
			},
		},
		{
			name: "does_not_steal_focus_when_finished_tab_is_not_active",
			setup: func(m *Model) {
				m.newTab()
				m.activeTab = 1
				m.focus = focusRequestPane
			},
			msg: func(m Model) RequestFinishedMsg {
				return RequestFinishedMsg{TabID: m.tabs[0].id, Body: []byte("hi")}
			},
			check: func(t *testing.T, next Model) {
				if next.focus != focusRequestPane {
					t.Fatalf("focus = %v, want focusRequestPane", next.focus)
				}
			},
		},
		{
			name:  "moves_focus_to_response_pane_when_finished_tab_is_active",
			setup: func(m *Model) { m.focus = focusRequestPane },
			msg: func(m Model) RequestFinishedMsg {
				return RequestFinishedMsg{TabID: m.tabs[0].id, Body: []byte("hi")}
			},
			check: func(t *testing.T, next Model) {
				if next.focus != focusResponsePane {
					t.Fatalf("focus = %v, want focusResponsePane", next.focus)
				}
			},
		},
		{
			name:  "ignores_unknown_tab_id",
			setup: func(m *Model) {},
			msg: func(m Model) RequestFinishedMsg {
				return RequestFinishedMsg{TabID: 999, Body: []byte("stale")}
			},
			check: func(t *testing.T, next Model) {
				if next.focus != focusRequestPane {
					t.Fatalf("focus = %v, want focusRequestPane (unchanged)", next.focus)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			tt.setup(&m)
			next, _ := m.onRequestFinished(tt.msg(m))
			tt.check(t, next)
		})
	}
}

func TestOnRequestStarted(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Model)
		wantSpinner bool
	}{
		{
			name:        "starts_spinner_for_active_tab",
			setup:       func(m *Model) {},
			wantSpinner: true,
		},
		{
			name:        "no_spinner_for_background_tab",
			setup:       func(m *Model) { m.newTab() }, // activeTab = 1, tab 0 is background
			wantSpinner: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			tt.setup(&m)
			_, cmd := m.onRequestStarted(RequestStartedMsg{TabID: m.tabs[0].id})
			if tt.wantSpinner && cmd == nil {
				t.Fatal("expected spinner cmd")
			}
			if !tt.wantSpinner && cmd != nil {
				t.Fatal("did not expect spinner cmd")
			}
		})
	}
}

func TestOnRequestFailed(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Model)
		isLoading   bool
		wantLoading bool
		wantCmd     bool
	}{
		{
			name:        "shows_error_and_clears_loading_on_active_tab",
			setup:       func(m *Model) {},
			isLoading:   true,
			wantLoading: false,
			wantCmd:     true,
		},
		{
			name:        "clears_loading_silently_on_background_tab",
			setup:       func(m *Model) { m.newTab() }, // activeTab = 1, tab 0 is background
			isLoading:   true,
			wantLoading: false,
			wantCmd:     false,
		},
		{
			name:        "no_op_when_tab_is_not_loading",
			setup:       func(m *Model) {},
			isLoading:   false,
			wantLoading: false,
			wantCmd:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			tt.setup(&m)
			m.tabs[0].isLoading = tt.isLoading

			next, cmd := m.onRequestFailed(RequestFailedMsg{TabID: m.tabs[0].id, Error: errors.New("err")})

			if next.tabs[0].isLoading != tt.wantLoading {
				t.Fatalf("isLoading = %v, want %v", next.tabs[0].isLoading, tt.wantLoading)
			}
			if tt.wantCmd && cmd == nil {
				t.Fatal("expected a cmd")
			}
			if !tt.wantCmd && cmd != nil {
				t.Fatal("did not expect a cmd")
			}
		})
	}
}

func TestCloseTab(t *testing.T) {
	tests := []struct {
		name          string
		isLoading     bool
		wantCancelled bool
	}{
		{
			name:          "cancels_in_flight_request",
			isLoading:     true,
			wantCancelled: true,
		},
		{
			name:          "no_cancel_call_when_tab_is_not_loading",
			isLoading:     false,
			wantCancelled: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.newTab() // need 2 tabs so closing is allowed

			cancelled := false
			m.tabs[0].isLoading = tt.isLoading
			if tt.isLoading {
				m.tabs[0].cancel = func() { cancelled = true }
			}

			m.switchTab(0)
			m.closeTab()

			if cancelled != tt.wantCancelled {
				t.Fatalf("cancelled = %v, want %v", cancelled, tt.wantCancelled)
			}
		})
	}
}

func TestSwitchTab_StatusbarSync(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Model)
		wantCmd bool
	}{
		{
			name: "returns_spinner_cmd_when_switching_to_loading_tab",
			setup: func(m *Model) {
				m.newTab()
				m.tabs[0].isLoading = true
			},
			wantCmd: true,
		},
		{
			name: "no_spinner_when_switching_to_tab_with_completed_response",
			setup: func(m *Model) {
				m.newTab()
				resp := statusbar.Response{StatusCode: 200}
				m.tabs[0].lastResponse = &resp
			},
			wantCmd: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			tt.setup(&m)
			cmd := m.switchTab(0)
			if tt.wantCmd && cmd == nil {
				t.Fatal("expected a cmd")
			}
			if !tt.wantCmd && cmd != nil {
				t.Fatalf("did not expect a cmd, got %v", cmd)
			}
		})
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
			id:             1,
			title:          "new request",
			collectionRoot: collectionRoot,
			urlbar:         u,
			requestPane:    rp,
			responsePane:   responsepane.New(),
		}},
		activeTab:       0,
		nextTabID:       2,
		collection:      collection.New(),
		workspacePicker: workspacepicker.New(),
		statusbar:       statusbar.New(),
		methods:         methodpicker.New(),
		workspaceRoot:   workspace,
		collectionDirs:  []string{collectionRoot},
		focus:           focusRequestPane,
		lastPane:        requestPane,
	}

	m.syncChildStateWithLayout()
	return m
}
