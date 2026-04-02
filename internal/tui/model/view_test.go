package model

import (
	"testing"

	"charm.land/bubbles/v2/key"
)

func TestStatusBindings(t *testing.T) {
	tests := []struct {
		name        string
		focus       focusedTarget
		wantKey     string
		wantPresent bool
	}{
		{
			name:        "url_bar_includes_enter",
			focus:       focusURLBar,
			wantKey:     "enter",
			wantPresent: true,
		},
		{
			name:        "url_bar_includes_clear",
			focus:       focusURLBar,
			wantKey:     "ctrl+l",
			wantPresent: true,
		},
		{
			name:        "request_pane_includes_tab_navigation",
			focus:       focusRequestPane,
			wantKey:     "]/[",
			wantPresent: true,
		},
		{
			name:        "response_pane_includes_scroll",
			focus:       focusResponsePane,
			wantKey:     "j/k",
			wantPresent: true,
		},
		{
			name:        "any_pane_always_includes_global_send",
			focus:       focusRequestPane,
			wantKey:     "^r",
			wantPresent: true,
		},
		{
			name:        "any_pane_always_includes_help",
			focus:       focusResponsePane,
			wantKey:     "?",
			wantPresent: true,
		},
		{
			name:        "global_focus_has_no_pane_bindings",
			focus:       focusGlobal,
			wantKey:     "]/[",
			wantPresent: false,
		},
		{
			name:        "global_focus_still_has_send",
			focus:       focusGlobal,
			wantKey:     "^r",
			wantPresent: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.focus = tt.focus
			bindings := m.statusBindings()
			found := containsKey(bindings, tt.wantKey)
			if found != tt.wantPresent {
				t.Fatalf("binding %q present=%v, want %v", tt.wantKey, found, tt.wantPresent)
			}
		})
	}
}

func TestFullHelpBindings(t *testing.T) {
	tests := []struct {
		name    string
		focus   focusedTarget
		wantKey string
	}{
		{name: "request_pane_has_clear_body", focus: focusRequestPane, wantKey: "ctrl+l"},
		{name: "request_pane_has_cycle_body_type", focus: focusRequestPane, wantKey: "h/l"},
		{name: "response_pane_has_jump_to_bottom", focus: focusResponsePane, wantKey: "G"},
		{name: "any_focus_has_global_tab_key", focus: focusRequestPane, wantKey: "tab"},
		{name: "any_focus_has_global_shift_tab_key", focus: focusResponsePane, wantKey: "shift+tab"},
		{name: "any_focus_has_esc", focus: focusURLBar, wantKey: "esc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.focus = tt.focus
			var all []key.Binding
			for _, g := range m.fullHelpBindings() {
				all = append(all, g...)
			}
			if !containsKey(all, tt.wantKey) {
				t.Fatalf("expected binding %q in fullHelpBindings, not found", tt.wantKey)
			}
		})
	}
}

func TestPaneKeyMap_ReturnsEmptyForModalAndGlobal(t *testing.T) {
	tests := []struct {
		name  string
		focus focusedTarget
	}{
		{name: "global_focus", focus: focusGlobal},
		{name: "modal_collection_focus", focus: focusModalCollection},
		{name: "modal_workspace_focus", focus: focusModalWorkspace},
		{name: "modal_method_focus", focus: focusModalMethod},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel(t)
			m.focus = tt.focus
			km := m.paneKeyMap()
			if len(km.ShortHelp()) != 0 {
				t.Fatalf("ShortHelp() for %v = %v, want empty", tt.focus, km.ShortHelp())
			}
			if len(km.FullHelp()) != 0 {
				t.Fatalf("FullHelp() for %v = %v, want empty", tt.focus, km.FullHelp())
			}
		})
	}
}

func containsKey(bindings []key.Binding, wantKey string) bool {
	for _, b := range bindings {
		if b.Help().Key == wantKey {
			return true
		}
	}
	return false
}
