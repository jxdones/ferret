package responsepane

import (
	"fmt"
	"strings"
	"testing"

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
	m.SetResponse([]byte(strings.Join(bodyLines, "\n")), map[string][]string{"Content-Type": {"text/plain"}}, exec.Trace{})

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
	m.SetResponse([]byte("ok"), headersMap, exec.Trace{})

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
