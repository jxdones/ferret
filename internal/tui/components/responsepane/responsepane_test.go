package responsepane

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/exec"
)

func TestCharacterization_BodyPagingAndClamping(t *testing.T) {
	m := New()
	m.SetSize(80, 10)
	m.SetFocused(true)

	bodyLines := make([]string, 30)
	for i := range bodyLines {
		bodyLines[i] = "line"
	}
	m.SetResponse([]byte(strings.Join(bodyLines, "\n")), map[string][]string{"Content-Type": {"text/plain"}}, exec.Trace{}, false, 0)

	var handled bool
	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'd', Mod: tea.ModCtrl}))
	if !handled {
		t.Fatal("ctrl+d should be handled in body tab")
	}
	if m.offset != 5 {
		t.Fatalf("offset = %d, want 5 after ctrl+d", m.offset)
	}

	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'G', Text: "G"}))
	if !handled {
		t.Fatal("G should be handled in body tab")
	}
	if m.offset != 21 {
		t.Fatalf("offset = %d, want 21 at bottom clamp", m.offset)
	}

	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
	if !handled {
		t.Fatal("g should be handled in body tab")
	}
	if m.offset != 0 {
		t.Fatalf("offset = %d, want 0 after g", m.offset)
	}
}

func TestCharacterization_HeadersPagingAndClamping(t *testing.T) {
	m := New()
	m.SetSize(80, 10)
	m.SetFocused(true)

	headersMap := make(map[string][]string, 24)
	for i := 0; i < 24; i++ {
		headersMap[fmt.Sprintf("X-Header-%02d", i)] = []string{strings.Repeat("v", i+1)}
	}
	m.SetResponse([]byte("ok"), headersMap, exec.Trace{}, false, 0)

	// Move from body -> headers tab.
	m, _, _ = m.Update(tea.KeyPressMsg(tea.Key{Code: ']', Text: "]"}))

	var handled bool
	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'd', Mod: tea.ModCtrl}))
	if !handled {
		t.Fatal("ctrl+d should be handled in headers tab")
	}
	if m.headersOffset == 0 {
		t.Fatal("headersOffset should advance after ctrl+d")
	}

	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'G', Text: "G"}))
	if !handled {
		t.Fatal("G should be handled in headers tab")
	}
	maxOffset := max(0, len(m.headersContentLines())-m.height)
	if m.headersOffset != maxOffset {
		t.Fatalf("headersOffset = %d, want %d bottom clamp", m.headersOffset, maxOffset)
	}

	m, _, handled = m.Update(tea.KeyPressMsg(tea.Key{Code: 'g', Text: "g"}))
	if !handled {
		t.Fatal("g should be handled in headers tab")
	}
	if m.headersOffset != 0 {
		t.Fatalf("headersOffset = %d, want 0 after g", m.headersOffset)
	}
}

func TestKeys_ShortHelp(t *testing.T) {
	tests := []struct {
		name     string
		wantKey  string
		wantDesc string
	}{
		{name: "tab_navigation", wantKey: "]/[", wantDesc: "next/prev tab"},
		{name: "scroll", wantKey: "j/k", wantDesc: "scroll"},
		{name: "half_page", wantKey: "ctrl+d", wantDesc: "half page"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBindingExists(t, Keys.ShortHelp(), tt.wantKey, tt.wantDesc)
		})
	}
}

func TestKeys_FullHelp_ContainsAllShortHelp(t *testing.T) {
	short := Keys.ShortHelp()
	var full []key.Binding
	for _, g := range Keys.FullHelp() {
		full = append(full, g...)
	}
	for _, b := range short {
		h := b.Help()
		assertBindingExists(t, full, h.Key, h.Desc)
	}
}

func TestKeys_FullHelp_HasJumpBindings(t *testing.T) {
	tests := []struct {
		name     string
		wantKey  string
		wantDesc string
	}{
		{name: "jump_top", wantKey: "g", wantDesc: "top"},
		{name: "jump_bottom", wantKey: "G", wantDesc: "bottom"},
	}
	var full []key.Binding
	for _, g := range Keys.FullHelp() {
		full = append(full, g...)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertBindingExists(t, full, tt.wantKey, tt.wantDesc)
		})
	}
}

func assertBindingExists(t *testing.T, bindings []key.Binding, wantKey, wantDesc string) {
	t.Helper()
	for _, b := range bindings {
		h := b.Help()
		if h.Key == wantKey {
			if h.Desc != wantDesc {
				t.Fatalf("binding %q desc = %q, want %q", wantKey, h.Desc, wantDesc)
			}
			return
		}
	}
	t.Fatalf("binding %q not found in bindings", wantKey)
}
