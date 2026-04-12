package auth

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func keyChar(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: string(r), Code: r})
}

func keySpecial(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func keyShift(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Mod: tea.ModShift})
}

func TestNew_Defaults(t *testing.T) {
	m := New()
	if m.authType != authTypeNone {
		t.Fatalf("authType = %d, want authTypeNone", m.authType)
	}
	if m.inserting {
		t.Fatal("inserting should be false")
	}
	if m.focused {
		t.Fatal("focused should be false")
	}
	if m.focusedID != 0 {
		t.Fatalf("focusedID = %d, want 0", m.focusedID)
	}
}

func TestSetAuth(t *testing.T) {
	collAuth := &collection.Auth{Type: "bearer", Token: "coll-token"}

	tests := []struct {
		name         string
		reqAuth      *collection.Auth
		collAuth     *collection.Auth
		wantType     authTypeID
		wantInherit  inheritState
		wantToken    string
		wantUsername string
		wantPassword string
		wantKey      string
		wantValue    string
		wantKeyIn    apiKeyIn
	}{
		{
			name:        "nil_req_inherits",
			reqAuth:     nil,
			collAuth:    collAuth,
			wantType:    authTypeNone,
			wantInherit: inheritStateInherited,
		},
		{
			name:        "empty_type_inherits",
			reqAuth:     &collection.Auth{},
			wantType:    authTypeNone,
			wantInherit: inheritStateInherited,
		},
		{
			name:        "inherit_type_inherits",
			reqAuth:     &collection.Auth{Type: "inherit"},
			wantType:    authTypeNone,
			wantInherit: inheritStateInherited,
		},
		{
			name:        "explicit_none_overrides",
			reqAuth:     &collection.Auth{Type: "none"},
			wantType:    authTypeNone,
			wantInherit: inheritStateNone,
		},
		{
			name:        "bearer_populates_token",
			reqAuth:     &collection.Auth{Type: "bearer", Token: "tok123"},
			wantType:    authTypeBearer,
			wantInherit: inheritStateOverriding,
			wantToken:   "tok123",
		},
		{
			name:         "basic_populates_credentials",
			reqAuth:      &collection.Auth{Type: "basic", Username: "alice", Password: "s3cr3t"},
			wantType:     authTypeBasic,
			wantInherit:  inheritStateOverriding,
			wantUsername: "alice",
			wantPassword: "s3cr3t",
		},
		{
			name:        "apikey_in_header",
			reqAuth:     &collection.Auth{Type: "apikey", Key: "X-API-Key", Value: "abc", In: "header"},
			wantType:    authTypeApiKey,
			wantInherit: inheritStateOverriding,
			wantKey:     "X-API-Key",
			wantValue:   "abc",
			wantKeyIn:   apiKeyInHeader,
		},
		{
			name:        "apikey_in_query",
			reqAuth:     &collection.Auth{Type: "apikey", Key: "api_key", Value: "xyz", In: "query"},
			wantType:    authTypeApiKey,
			wantInherit: inheritStateOverriding,
			wantKey:     "api_key",
			wantValue:   "xyz",
			wantKeyIn:   apiKeyInQuery,
		},
		{
			name:        "apikey_empty_in_defaults_to_header",
			reqAuth:     &collection.Auth{Type: "apikey", Key: "k", Value: "v"},
			wantType:    authTypeApiKey,
			wantInherit: inheritStateOverriding,
			wantKey:     "k",
			wantValue:   "v",
			wantKeyIn:   apiKeyInHeader,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetAuth(tt.reqAuth, tt.collAuth)

			if m.authType != tt.wantType {
				t.Errorf("authType = %d, want %d", m.authType, tt.wantType)
			}
			if m.inherit != tt.wantInherit {
				t.Errorf("inherit = %d, want %d", m.inherit, tt.wantInherit)
			}
			if tt.collAuth != nil && m.collectionAuth != tt.collAuth {
				t.Errorf("collectionAuth not stored")
			}
			if tt.wantToken != "" && m.token.Value() != tt.wantToken {
				t.Errorf("token = %q, want %q", m.token.Value(), tt.wantToken)
			}
			if tt.wantUsername != "" && m.username.Value() != tt.wantUsername {
				t.Errorf("username = %q, want %q", m.username.Value(), tt.wantUsername)
			}
			if tt.wantPassword != "" && m.password.Value() != tt.wantPassword {
				t.Errorf("password = %q, want %q", m.password.Value(), tt.wantPassword)
			}
			if tt.wantKey != "" && m.key.Value() != tt.wantKey {
				t.Errorf("key = %q, want %q", m.key.Value(), tt.wantKey)
			}
			if tt.wantValue != "" && m.value.Value() != tt.wantValue {
				t.Errorf("value = %q, want %q", m.value.Value(), tt.wantValue)
			}
			if m.keyIn != tt.wantKeyIn {
				t.Errorf("keyIn = %d, want %d", m.keyIn, tt.wantKeyIn)
			}
		})
	}
}

func TestActiveInputCount(t *testing.T) {
	tests := []struct {
		name     string
		authType authTypeID
		want     int
	}{
		{name: "none", authType: authTypeNone, want: 0},
		{name: "bearer", authType: authTypeBearer, want: 1},
		{name: "basic", authType: authTypeBasic, want: 2},
		{name: "apikey", authType: authTypeApiKey, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.authType = tt.authType
			if got := m.activeInputCount(); got != tt.want {
				t.Fatalf("activeInputCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAuthTypeNext(t *testing.T) {
	tests := []struct {
		name    string
		current authTypeID
		want    authTypeID
	}{
		{name: "none_to_bearer", current: authTypeNone, want: authTypeBearer},
		{name: "bearer_to_basic", current: authTypeBearer, want: authTypeBasic},
		{name: "basic_to_apikey", current: authTypeBasic, want: authTypeApiKey},
		{name: "apikey_wraps_to_none", current: authTypeApiKey, want: authTypeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.current.next(); got != tt.want {
				t.Fatalf("next() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAuthTypePrevious(t *testing.T) {
	tests := []struct {
		name    string
		current authTypeID
		want    authTypeID
	}{
		{name: "none_wraps_to_apikey", current: authTypeNone, want: authTypeApiKey},
		{name: "bearer_to_none", current: authTypeBearer, want: authTypeNone},
		{name: "basic_to_bearer", current: authTypeBasic, want: authTypeBearer},
		{name: "apikey_to_basic", current: authTypeApiKey, want: authTypeBasic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.current.previous(); got != tt.want {
				t.Fatalf("previous() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestApiKeyInToggle(t *testing.T) {
	tests := []struct {
		name    string
		current apiKeyIn
		want    apiKeyIn
	}{
		{name: "header_to_query", current: apiKeyInHeader, want: apiKeyInQuery},
		{name: "query_to_header", current: apiKeyInQuery, want: apiKeyInHeader},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.current.toggle(); got != tt.want {
				t.Fatalf("toggle() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestUpdate_Unfocused_NotHandled(t *testing.T) {
	m := New()
	m.SetFocused(false)
	_, _, handled := m.Update(keyChar('l'))
	if handled {
		t.Fatal("unfocused model should not handle keys")
	}
}

func TestUpdate_TypeNavigation(t *testing.T) {
	tests := []struct {
		name     string
		start    authTypeID
		key      rune
		wantType authTypeID
	}{
		{name: "l_advances_none_to_bearer", start: authTypeNone, key: 'l', wantType: authTypeBearer},
		{name: "right_advances_none_to_bearer", start: authTypeNone, key: 0, wantType: authTypeBearer}, // handled below
		{name: "l_wraps_apikey_to_none", start: authTypeApiKey, key: 'l', wantType: authTypeNone},
		{name: "h_retreats_bearer_to_none", start: authTypeBearer, key: 'h', wantType: authTypeNone},
		{name: "h_wraps_none_to_apikey", start: authTypeNone, key: 'h', wantType: authTypeApiKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = tt.start

			var msg tea.KeyPressMsg
			if tt.key == 0 {
				msg = keySpecial(tea.KeyRight)
			} else {
				msg = keyChar(tt.key)
			}

			m, _, handled := m.Update(msg)
			if !handled {
				t.Fatal("key should be handled")
			}
			if m.authType != tt.wantType {
				t.Fatalf("authType = %d, want %d", m.authType, tt.wantType)
			}
		})
	}
}

func TestUpdate_TypeNavigation_Right_Left(t *testing.T) {
	tests := []struct {
		name     string
		start    authTypeID
		sym      rune
		wantType authTypeID
	}{
		{name: "right_advances_type", start: authTypeNone, sym: tea.KeyRight, wantType: authTypeBearer},
		{name: "left_retreats_type", start: authTypeBearer, sym: tea.KeyLeft, wantType: authTypeNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = tt.start
			m, _, handled := m.Update(keySpecial(tt.sym))
			if !handled {
				t.Fatal("arrow key should be handled")
			}
			if m.authType != tt.wantType {
				t.Fatalf("authType = %d, want %d", m.authType, tt.wantType)
			}
		})
	}
}

func TestUpdate_ApiKey_HLTogglesInWhenFocusedID2(t *testing.T) {
	tests := []struct {
		name       string
		key        rune
		startKeyIn apiKeyIn
		wantKeyIn  apiKeyIn
	}{
		{name: "l_toggles_to_query", key: 'l', startKeyIn: apiKeyInHeader, wantKeyIn: apiKeyInQuery},
		{name: "h_toggles_to_header", key: 'h', startKeyIn: apiKeyInQuery, wantKeyIn: apiKeyInHeader},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = authTypeApiKey
			m.focusedID = 2
			m.keyIn = tt.startKeyIn
			m, _, handled := m.Update(keyChar(tt.key))
			if !handled {
				t.Fatal("key should be handled")
			}
			if m.keyIn != tt.wantKeyIn {
				t.Fatalf("keyIn = %d, want %d", m.keyIn, tt.wantKeyIn)
			}
			if m.authType != authTypeApiKey {
				t.Fatal("authType should remain apikey when toggling In")
			}
		})
	}
}

func TestUpdate_RowNavigation(t *testing.T) {
	tests := []struct {
		name     string
		authType authTypeID
		start    int
		key      rune
		wantID   int
	}{
		{name: "j_advances_focus_bearer", authType: authTypeBearer, start: 0, key: 'j', wantID: 0}, // clamps at max (0)
		{name: "j_advances_focus_basic", authType: authTypeBasic, start: 0, key: 'j', wantID: 1},
		{name: "k_retreats_focus_basic", authType: authTypeBasic, start: 1, key: 'k', wantID: 0},
		{name: "k_clamps_at_zero", authType: authTypeBasic, start: 0, key: 'k', wantID: 0},
		{name: "down_advances_focus", authType: authTypeBasic, start: 0, key: 0, wantID: 1}, // handled via sym
		{name: "j_advances_apikey", authType: authTypeApiKey, start: 1, key: 'j', wantID: 2},
		{name: "j_clamps_apikey_at_max", authType: authTypeApiKey, start: 2, key: 'j', wantID: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = tt.authType
			m.focusedID = tt.start

			var msg tea.KeyPressMsg
			if tt.key == 0 {
				msg = keySpecial(tea.KeyDown)
			} else {
				msg = keyChar(tt.key)
			}

			m, _, handled := m.Update(msg)
			if !handled {
				t.Fatal("navigation key should be handled")
			}
			if m.focusedID != tt.wantID {
				t.Fatalf("focusedID = %d, want %d", m.focusedID, tt.wantID)
			}
		})
	}
}

func TestUpdate_InsertMode_EnterAndEsc(t *testing.T) {
	tests := []struct {
		name          string
		authType      authTypeID
		enterKey      tea.KeyPressMsg
		wantInserting bool
	}{
		{name: "i_starts_insert_bearer", authType: authTypeBearer, enterKey: keyChar('i'), wantInserting: true},
		{name: "enter_starts_insert_bearer", authType: authTypeBearer, enterKey: keySpecial(tea.KeyEnter), wantInserting: true},
		{name: "i_noop_on_none", authType: authTypeNone, enterKey: keyChar('i'), wantInserting: false},
		{name: "enter_noop_on_none", authType: authTypeNone, enterKey: keySpecial(tea.KeyEnter), wantInserting: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = tt.authType
			m, _, _ = m.Update(tt.enterKey)
			if m.Inserting() != tt.wantInserting {
				t.Fatalf("Inserting() = %v, want %v", m.Inserting(), tt.wantInserting)
			}
		})
	}
}

func TestUpdate_InsertMode_EscExits(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.authType = authTypeBearer
	m, _, _ = m.Update(keyChar('i'))
	if !m.Inserting() {
		t.Fatal("i should enter insert mode")
	}
	m, _, handled := m.Update(keySpecial(tea.KeyEsc))
	if !handled {
		t.Fatal("esc should be handled in insert mode")
	}
	if m.Inserting() {
		t.Fatal("esc should exit insert mode")
	}
}

func TestUpdate_InsertMode_TabCyclesFocusedID(t *testing.T) {
	tests := []struct {
		name     string
		authType authTypeID
		start    int
		key      tea.KeyPressMsg
		wantID   int
	}{
		{
			name:     "tab_advances_basic",
			authType: authTypeBasic,
			start:    0,
			key:      keySpecial(tea.KeyTab),
			wantID:   1,
		},
		{
			name:     "tab_clamps_at_max_basic",
			authType: authTypeBasic,
			start:    1,
			key:      keySpecial(tea.KeyTab),
			wantID:   1,
		},
		{
			name:     "shift_tab_retreats_basic",
			authType: authTypeBasic,
			start:    1,
			key:      keyShift(tea.KeyTab),
			wantID:   0,
		},
		{
			name:     "shift_tab_clamps_at_zero",
			authType: authTypeBasic,
			start:    0,
			key:      keyShift(tea.KeyTab),
			wantID:   0,
		},
		{
			name:     "tab_advances_apikey",
			authType: authTypeApiKey,
			start:    1,
			key:      keySpecial(tea.KeyTab),
			wantID:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetFocused(true)
			m.authType = tt.authType
			m.inserting = true
			m.focusedID = tt.start
			m, _, handled := m.Update(tt.key)
			if !handled {
				t.Fatal("tab key should be handled in insert mode")
			}
			if m.focusedID != tt.wantID {
				t.Fatalf("focusedID = %d, want %d", m.focusedID, tt.wantID)
			}
		})
	}
}

func TestUpdate_TypeNavigation_ResetsToFirstRow(t *testing.T) {
	m := New()
	m.SetFocused(true)
	m.authType = authTypeBasic
	m.focusedID = 1

	m, _, _ = m.Update(keyChar('l'))
	if m.focusedID != 0 {
		t.Fatalf("focusedID = %d, want 0 after type switch", m.focusedID)
	}
}

func TestInserting_Getter(t *testing.T) {
	m := New()
	if m.Inserting() {
		t.Fatal("Inserting() should be false on new model")
	}
	m.inserting = true
	if !m.Inserting() {
		t.Fatal("Inserting() should reflect inserting field")
	}
}

func TestView(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name     string
		setup    func(*Model)
		wantSubs []string
		wantNot  []string
	}{
		{
			name:  "type_selector_shows_all_labels",
			setup: func(m *Model) {},
			wantSubs: []string{
				noneTabLabel,
				bearerTabLabel,
				basicTabLabel,
				apiKeyTabLabel,
			},
		},
		{
			name: "none_shows_no_auth_message",
			setup: func(m *Model) {
				m.authType = authTypeNone
			},
			wantSubs: []string{"no auth"},
		},
		{
			name: "bearer_shows_token_label",
			setup: func(m *Model) {
				m.authType = authTypeBearer
			},
			wantSubs: []string{"Token"},
		},
		{
			name: "basic_shows_username_and_password_labels",
			setup: func(m *Model) {
				m.authType = authTypeBasic
			},
			wantSubs: []string{"Username", "Password"},
		},
		{
			name: "apikey_shows_key_value_and_in_labels",
			setup: func(m *Model) {
				m.authType = authTypeApiKey
			},
			wantSubs: []string{"Key", "Value", "In"},
		},
		{
			name: "apikey_shows_header_by_default",
			setup: func(m *Model) {
				m.authType = authTypeApiKey
				m.keyIn = apiKeyInHeader
			},
			wantSubs: []string{"header"},
			wantNot:  []string{"query"},
		},
		{
			name: "apikey_shows_query_when_set",
			setup: func(m *Model) {
				m.authType = authTypeApiKey
				m.keyIn = apiKeyInQuery
			},
			wantSubs: []string{"query"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			tt.setup(&m)
			out := tuitest.StripANSI(m.View().Content)
			for _, sub := range tt.wantSubs {
				if !strings.Contains(out, sub) {
					t.Errorf("View() missing %q in:\n%s", sub, out)
				}
			}
			for _, sub := range tt.wantNot {
				if strings.Contains(out, sub) {
					t.Errorf("View() should not contain %q", sub)
				}
			}
		})
	}
}

func TestSetSize_PropagatesWidth(t *testing.T) {
	m := New()
	m.SetSize(80)
	if m.width != 80 {
		t.Fatalf("width = %d, want 80", m.width)
	}
}

func TestSetFocused_UpdatesFocusState(t *testing.T) {
	m := New()
	m.SetFocused(true)
	if !m.focused {
		t.Fatal("focused should be true")
	}
	m.SetFocused(false)
	if m.focused {
		t.Fatal("focused should be false")
	}
}
