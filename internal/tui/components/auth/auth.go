package auth

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	noneTabLabel   = "none"
	bearerTabLabel = "bearer"
	basicTabLabel  = "basic"
	apiKeyTabLabel = "api key"
)

type authTypeID int

const (
	authTypeNone authTypeID = iota
	authTypeBearer
	authTypeBasic
	authTypeApiKey
)

type inheritState int

const (
	inheritStateInherited inheritState = iota
	inheritStateOverriding
	inheritStateNone // no collection-level auth either
)

type apiKeyIn int

const (
	apiKeyInHeader apiKeyIn = iota
	apiKeyInQuery
)

// Model represents the auth component
type Model struct {
	authType  authTypeID
	inherit   inheritState
	keyIn     apiKeyIn
	focusedID int

	token    textinput.Model
	username textinput.Model
	password textinput.Model
	key      textinput.Model
	value    textinput.Model

	width          int
	collectionAuth *collection.Auth

	focused   bool
	inserting bool

	theme theme.Theme
}

// New creates a new auth model.
func New() Model {
	t := theme.Current
	m := Model{
		theme: t,
	}
	m.token = newInput("Token", t)
	m.username = newInput("Username", t)

	m.password = newInput("Password", t)
	m.password.EchoMode = textinput.EchoPassword
	m.password.EchoCharacter = '*'

	m.key = newInput("Key", t)
	m.value = newInput("Value", t)
	return m
}

// View renders the auth type picker and the type-specific fields.
func (m Model) View() tea.View {
	typeLine := m.renderTypeLine()
	fields := m.renderFields()
	return tea.NewView(strings.Join([]string{typeLine, "", fields}, "\n"))
}

// Update handles auth navigation, paste and interactive edits.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, bool) {
	if !m.focused {
		return m, nil, false
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.updateKeyPress(msg)
	default:
		return m, nil, false
	}
}

// updateKeyPress handles key presses for the auth model.
func (m Model) updateKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if m.inserting {
		switch msg.String() {
		case "esc":
			m.inserting = false
			m.applyInputFocus()
			return m, nil, true
		case "tab":
			m.focusedID = min(m.focusedID+1, m.activeInputCount()-1)
			m.applyInputFocus()
			return m, nil, true
		case "shift+tab":
			m.focusedID = max(m.focusedID-1, 0)
			m.applyInputFocus()
			return m, nil, true
		default:
			return m.activeInputUpdate(msg)
		}
	}

	switch msg.String() {
	case "h", "left":
		if m.authType == authTypeApiKey && m.focusedID == 2 {
			m.keyIn = m.keyIn.toggle()
			return m, nil, true
		}
		m.authType = m.authType.previous()
		m.focusedID = 0
		m.applyInputFocus()
		return m, nil, true
	case "l", "right":
		if m.authType == authTypeApiKey && m.focusedID == 2 {
			m.keyIn = m.keyIn.toggle()
			return m, nil, true
		}
		m.authType = m.authType.next()
		m.focusedID = 0
		m.applyInputFocus()
		return m, nil, true
	case "j", "down":
		m.focusedID = min(m.focusedID+1, m.activeInputCount()-1)
		m.applyInputFocus()
		return m, nil, true
	case "k", "up":
		m.focusedID = max(m.focusedID-1, 0)
		m.applyInputFocus()
		return m, nil, true
	case "i", "enter":
		if m.activeInputCount() == 0 {
			return m, nil, false
		}
		m.inserting = true
		m.applyInputFocus()
		return m, nil, true
	}

	return m, nil, false
}

// activeInputUpdate calls update for the active input.
func (m Model) activeInputUpdate(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch m.authType {
	case authTypeBearer:
		if m.focusedID == 0 {
			var cmd tea.Cmd
			m.token, cmd = m.token.Update(msg)
			return m, cmd, true
		}
	case authTypeBasic:
		switch m.focusedID {
		case 0:
			var cmd tea.Cmd
			m.username, cmd = m.username.Update(msg)
			return m, cmd, true
		case 1:
			var cmd tea.Cmd
			m.password, cmd = m.password.Update(msg)
			return m, cmd, true
		}
	case authTypeApiKey:
		switch m.focusedID {
		case 0:
			var cmd tea.Cmd
			m.key, cmd = m.key.Update(msg)
			return m, cmd, true
		case 1:
			var cmd tea.Cmd
			m.value, cmd = m.value.Update(msg)
			return m, cmd, true
		}
	}
	return m, nil, false
}

// activeInputCount returns the number of inputs from each auth type.
func (m Model) activeInputCount() int {
	switch m.authType {
	case authTypeBearer:
		return 1
	case authTypeBasic:
		return 2
	case authTypeApiKey:
		return 3
	default:
		return 0
	}
}

// SetTheme updates the theme used for rendering.
func (m *Model) SetTheme(t theme.Theme) {
	m.theme = t
	applyInputTheme(&m.token, t)
	applyInputTheme(&m.username, t)
	applyInputTheme(&m.password, t)
	applyInputTheme(&m.key, t)
	applyInputTheme(&m.value, t)
}

// SetAuth populates the auth fields from the request and collection auth.
// reqAuth is the request-level auth; collectionAuth is the collection default.
func (m *Model) SetAuth(reqAuth *collection.Auth, collectionAuth *collection.Auth) {
	m.collectionAuth = collectionAuth
	if reqAuth == nil || reqAuth.Type == "" || reqAuth.Type == "inherit" {
		m.authType = authTypeNone
		m.inherit = inheritStateInherited
		m.applyInputFocus()
		return
	}
	m.inherit = inheritStateOverriding
	switch reqAuth.Type {
	case "bearer":
		m.authType = authTypeBearer
		m.token.SetValue(reqAuth.Token)
	case "basic":
		m.authType = authTypeBasic
		m.username.SetValue(reqAuth.Username)
		m.password.SetValue(reqAuth.Password)
	case "apikey":
		m.authType = authTypeApiKey
		m.key.SetValue(reqAuth.Key)
		m.value.SetValue(reqAuth.Value)
		if reqAuth.In == "query" {
			m.keyIn = apiKeyInQuery
		} else {
			m.keyIn = apiKeyInHeader
		}
	case "none":
		m.authType = authTypeNone
		m.inherit = inheritStateNone
	}
	m.applyInputFocus()
}

// SetSize sets the width of the auth model inputs.
func (m *Model) SetSize(width int) {
	m.width = width
	m.token.SetWidth(width - 2)
	m.username.SetWidth(width - 2)
	m.password.SetWidth(width - 2)
	m.key.SetWidth(width - 2)
	m.value.SetWidth(width - 2)
}

// SetFocused sets the focused state of the auth model
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	m.applyInputFocus()
}

// Inserting returns true if the auth model is in inserting mode.
func (m Model) Inserting() bool {
	return m.inserting
}

// newInput creates a textinput model with the given placeholder.
func newInput(placeholder string, t theme.Theme) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	applyInputTheme(&ti, t)
	return ti
}

// applyInputTheme sets the themed prompt colors on a textinput.
func applyInputTheme(ti *textinput.Model, t theme.Theme) {
	styles := ti.Styles()
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(t.TextAccent)
	styles.Blurred.Prompt = styles.Blurred.Prompt.Foreground(t.TextDim)
	ti.SetStyles(styles)
}

// applyInputFocus sets the focus state of the inputs on the auth model.
func (m *Model) applyInputFocus() {
	m.token.Blur()
	m.username.Blur()
	m.password.Blur()
	m.key.Blur()
	m.value.Blur()

	if !m.inserting || !m.focused {
		return
	}

	switch m.authType {
	case authTypeBearer:
		if m.focusedID == 0 {
			m.token.Focus()
		}
	case authTypeBasic:
		switch m.focusedID {
		case 0:
			m.username.Focus()
		case 1:
			m.password.Focus()
		}
	case authTypeApiKey:
		switch m.focusedID {
		case 0:
			m.key.Focus()
		case 1:
			m.value.Focus()
		}
	}
}

// renderTypeLine renders the auth type selector line.
func (m Model) renderTypeLine() string {
	types := []authTypeID{
		authTypeNone,
		authTypeBearer,
		authTypeBasic,
		authTypeApiKey,
	}

	selected := lipgloss.NewStyle().Foreground(m.theme.TitleBarEntry).Bold(true)
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	dim := lipgloss.NewStyle().Foreground(m.theme.TextDim)

	var parts []string
	for _, t := range types {
		if t == m.authType {
			parts = append(parts, selected.Render(t.label()))
		} else {
			parts = append(parts, muted.Render(t.label()))
		}
	}
	inner := strings.Join(parts, dim.Render(" │ "))
	return muted.Render(" [ ") + inner + muted.Render(" ] ")
}

// label returns the display label for the auth type.
func (t authTypeID) label() string {
	switch t {
	case authTypeBearer:
		return bearerTabLabel
	case authTypeBasic:
		return basicTabLabel
	case authTypeApiKey:
		return apiKeyTabLabel
	case authTypeNone:
		return noneTabLabel
	default:
		return noneTabLabel
	}
}

// next returns the next authTypeID.
func (t authTypeID) next() authTypeID {
	if t == authTypeApiKey {
		return authTypeNone
	}
	return t + 1
}

// previous returns the previous authTypeID.
func (t authTypeID) previous() authTypeID {
	if t == authTypeNone {
		return authTypeApiKey
	}
	return t - 1
}

// toggle returns either api key in query or headers.
func (k apiKeyIn) toggle() apiKeyIn {
	if k == apiKeyInHeader {
		return apiKeyInQuery
	}
	return apiKeyInHeader
}

// renderFields renders the type-specific input fields for the active auth type.
func (m Model) renderFields() string {
	muted := lipgloss.NewStyle().Foreground(m.theme.TextMuted)
	primary := lipgloss.NewStyle().Foreground(m.theme.TextPrimary)
	dim := lipgloss.NewStyle().Foreground(m.theme.TextDim)

	labelStyle := func(id int) lipgloss.Style {
		if id == m.focusedID && !m.inserting {
			return primary.Bold(true)
		}
		return muted
	}

	switch m.authType {
	case authTypeNone:
		return muted.Render("  no auth")
	case authTypeBearer:
		return labelStyle(0).Render("  Token    ") + m.token.View()
	case authTypeBasic:
		return strings.Join([]string{
			labelStyle(0).Render("  Username ") + m.username.View(),
			labelStyle(1).Render("  Password ") + m.password.View(),
		}, "\n")
	case authTypeApiKey:
		inLabel := "header"
		if m.keyIn == apiKeyInQuery {
			inLabel = "query"
		}
		inLine := labelStyle(2).Render("  In       ") + dim.Render("[ ") + primary.Render(inLabel) + dim.Render(" ]")
		return strings.Join([]string{
			labelStyle(0).Render("  Key      ") + m.key.View(),
			labelStyle(1).Render("  Value    ") + m.value.View(),
			inLine,
		}, "\n")
	}
	return ""
}
