package requestpane

import (
	"net/url"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/components/bodyeditor"
	"github.com/jxdones/ferret/internal/tui/components/headers"
	"github.com/jxdones/ferret/internal/tui/components/tabs"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	headersTabLabel = "headers"
	paramsTabLabel  = "params"
	bodyTabLabel    = "body"
	authTabLabel    = "auth"
)

type requestTabID int

const (
	requestTabHeaders requestTabID = iota
	requestTabParams
	requestTabBody
	requestTabAuth
)

type bodyFocusID int

const (
	bodyFocusType bodyFocusID = iota
	bodyFocusEditor
)

type queryParam struct {
	key   string
	value string
}

type bodyTypeID int

const (
	bodyTypeNone bodyTypeID = iota
	bodyTypeRaw
)

// Model represents a request pane.
type Model struct {
	tabs    tabs.Model
	headers headers.Model
	body    bodyeditor.Model

	url       string
	width     int
	height    int
	focused   bool
	bodyFocus bodyFocusID
	bodyType  bodyTypeID
}

// New returns an initialized request pane model.
func New() Model {
	t := tabs.New(requestTabLabels())
	t.SetActiveForeground(theme.Current.RequestPaneLabel)
	return Model{
		tabs:      t,
		headers:   headers.New(),
		body:      bodyeditor.New(),
		bodyFocus: bodyFocusType,
	}
}

// SetSize sets the dimensions of the request pane model and its child components.
func (m *Model) SetSize(width int, height int) {
	m.width = width
	m.height = height
	m.tabs.SetSize(width)
	m.headers.SetSize(width)
	m.body.SetSize(width, max(1, height-2))
}

// SetFocused sets the focused state of the request pane model and its child
// components.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	m.tabs.SetFocused(focused)
	m.syncEditorFocus()
}

// SetHeaders sets the headers of the request pane model.
func (m *Model) SetHeaders(h map[string]string) {
	m.headers.SetHeaders(h)
	m.syncBodySyntax()
	m.syncEditorFocus()
}

// ResetBodyFocus returns the body tab to selector focus.
func (m *Model) ResetBodyFocus() {
	m.bodyFocus = bodyFocusType
	m.syncEditorFocus()
}

// SetBody sets the request body for the body tab.
func (m *Model) SetBody(body string) {
	m.body.SetValue(body)
	if body != "" {
		m.bodyType = bodyTypeRaw
	} else {
		m.bodyType = bodyTypeNone
	}
	m.syncBodySyntax()
	m.syncEditorFocus()
}

// Headers returns the current request headers as a name→value map.
func (m Model) Headers() map[string]string {
	return m.headers.Headers()
}

// Body returns the current request body text.
func (m Model) Body() string {
	if m.bodyType == bodyTypeNone {
		return ""
	}
	return m.body.Value()
}

// SetURL stores the current request URL used by the params tab.
func (m *Model) SetURL(rawURL string) {
	m.url = rawURL
}

// TabsView returns the rendered tab strip line.
func (m Model) TabsView() tea.View {
	return m.tabs.View()
}

// ActiveTabSpan returns the start column and width of the active tab label,
// used by the root to render the cross-pane tab underline.
func (m Model) ActiveTabSpan() (int, int) {
	return m.tabs.ActiveSpan()
}

// View renders the active tab's content area as a multi-line string.
// The root splits this on newlines and positions it in the layout grid.
func (m Model) View() tea.View {
	switch m.activeTab() {
	case requestTabHeaders:
		return m.headers.View()
	case requestTabParams:
		return tea.NewView(m.paramsView())
	case requestTabBody:
		typeLine := m.renderBodyTypeLine()
		if m.bodyType == bodyTypeNone {
			return tea.NewView(typeLine)
		}
		return tea.NewView(strings.Join([]string{typeLine, "", m.body.View()}, "\n"))
	}
	return tea.NewView("")
}

// Update handles tab navigation, paste, and delegates interactive edits.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, bool) {
	if !m.focused {
		return m, nil, false
	}

	switch msg := msg.(type) {
	case tea.PasteMsg:
		if m.activeTab() == requestTabBody && m.bodyFocus == bodyFocusEditor {
			var cmd tea.Cmd
			m.body, cmd = m.body.Update(msg)
			m.syncBodySyntax()
			return m, cmd, true
		}
		return m, nil, false
	case tea.KeyPressMsg:
		return m.updateKeyPress(msg)
	default:
		return m, nil, false
	}
}

// updateKeyPress handles key presses for the request pane.
func (m Model) updateKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch m.activeTab() {
	case requestTabHeaders:
		switch msg.String() {
		case "]":
			m.tabs.Next()
			m.syncEditorFocus()
			return m, nil, true
		case "[":
			m.tabs.Previous()
			m.syncEditorFocus()
			return m, nil, true
		}
	case requestTabBody:
		// Body tab has two modes. Only the outer (selector) mode uses [ ] for inner tab
		// navigation; in the editor, [ ] are passed through for JSON and similar payloads.
		if m.bodyFocus == bodyFocusEditor {
			return m.updateBodyKeyPress(msg)
		}
		switch msg.String() {
		case "]":
			m.tabs.Next()
			m.syncEditorFocus()
			return m, nil, true
		case "[":
			m.tabs.Previous()
			m.syncEditorFocus()
			return m, nil, true
		case "esc":
			return m, nil, false
		case "enter", "i":
			if m.bodyType == bodyTypeRaw {
				m.bodyFocus = bodyFocusEditor
				m.syncEditorFocus()
				return m, nil, true
			}
			return m, nil, false
		case "h", "left":
			m.bodyType = m.bodyType.cycle()
			m.syncEditorFocus()
			return m, nil, true
		case "l", "right":
			m.bodyType = m.bodyType.cycle()
			m.syncEditorFocus()
			return m, nil, true
		}
		// Tab / shift+tab: same as params/auth — let the root move focus between URL bar and panes.
		return m, nil, false
	default:
		switch msg.String() {
		case "]":
			m.tabs.Next()
			m.syncEditorFocus()
			return m, nil, true
		case "[":
			m.tabs.Previous()
			m.syncEditorFocus()
			return m, nil, true
		}
		return m, nil, false
	}

	// When in insert mode, forward the key to the headers model.
	if m.headers.IsInserting() {
		var cmd tea.Cmd
		m.headers, cmd = m.headers.Update(msg)
		m.syncBodySyntax()
		return m, cmd, true
	}

	// Normal mode, navigation and insert mode shortcuts are handled here.
	switch msg.String() {
	case "up", "k":
		m.headers.MoveCursor(-1)
		return m, nil, true
	case "down", "j":
		m.headers.MoveCursor(1)
		return m, nil, true
	case "d":
		m.headers.DeleteCursorRow()
		m.syncBodySyntax()
		return m, nil, true
	case "i", "I", "A":
		var cmd tea.Cmd
		m.headers, cmd = m.headers.Update(msg)
		m.syncBodySyntax()
		return m, cmd, true
	}
	return m, nil, false
}

// updateBodyKeyPress handles key presses for the body tab.
func (m Model) updateBodyKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "esc":
		m.bodyFocus = bodyFocusType
		m.syncEditorFocus()
		return m, nil, true
	case "tab":
		indent := strings.Repeat(" ", 4)
		m.body.InsertString(indent)
		return m, nil, true
	case "ctrl+l":
		m.body.SetValue("")
		m.syncBodySyntax()
		return m, nil, true
	default:
		var cmd tea.Cmd
		m.body, cmd = m.body.Update(msg)
		m.syncBodySyntax()
		return m, cmd, true
	}
}

// BodyFocused reports whether the body editor is focused.
func (m Model) BodyFocused() bool {
	return m.focused && m.activeTab() == requestTabBody && m.bodyFocus == bodyFocusEditor
}

// activeTab returns the active tab based on the tab label.
func (m Model) activeTab() requestTabID {
	return requestTabFromLabel(m.tabs.ActiveLabel())
}

// syncEditorFocus updates the focus state of the editor based on the active tab.
func (m *Model) syncEditorFocus() {
	m.headers.SetFocused(m.focused && m.activeTab() == requestTabHeaders)
	bodyActive := m.focused && m.activeTab() == requestTabBody
	m.body.SetFocused(bodyActive && m.bodyFocus == bodyFocusEditor)
}

// syncBodySyntax updates the body syntax based on the headers.
func (m *Model) syncBodySyntax() {
	m.body.SetSyntax(bodySyntaxFor(m.headers.Headers(), m.body.Value()))
}

// renderBodyTypeLine renders the body type line.
func (m Model) renderBodyTypeLine() string {
	types := []bodyTypeID{
		bodyTypeNone,
		bodyTypeRaw,
	}

	selected := lipgloss.NewStyle().Foreground(theme.Current.TitleBarEntry).Bold(true)
	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	dim := lipgloss.NewStyle().Foreground(theme.Current.TextDim)

	var parts []string
	for _, t := range types {
		if t == m.bodyType {
			parts = append(parts, selected.Render(t.label()))
		} else {
			parts = append(parts, muted.Render(t.label()))
		}
	}
	inner := strings.Join(parts, dim.Render(" │ "))
	return muted.Render(" [ ") + inner + muted.Render(" ] ")
}

// paramsView renders query parameters derived from the current URL.
func (m Model) paramsView() string {
	params := parseQueryParams(m.url)
	if len(params) == 0 {
		return lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
			Render("  no query params in URL")
	}

	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	dim := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	primary := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)

	keyW := max(8, m.width*2/5)
	valW := max(8, m.width*2/5)
	descW := max(0, m.width-keyW-valW)
	divider := dim.Render(strings.Repeat("─", m.width))

	lines := []string{
		muted.Render(padRight("  Key", keyW)) +
			muted.Render(padRight("Value", valW)) +
			muted.Render(padRight("Description", descW)),
		divider,
	}

	for _, p := range params {
		lines = append(lines,
			muted.Render(padRight("  "+ansi.Truncate(p.key, max(1, keyW-2), "…"), keyW))+
				primary.Render(padRight(ansi.Truncate(p.value, valW, "…"), valW))+
				dim.Render(padRight("", descW)),
		)
	}
	return strings.Join(lines, "\n")
}

// parseQueryParams parses and decodes query params from a raw URL string.
func parseQueryParams(rawURL string) []queryParam {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	rawQuery := strings.TrimPrefix(u.RawQuery, "?")
	if rawQuery == "" {
		return nil
	}

	parts := strings.Split(rawQuery, "&")
	out := make([]queryParam, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		k, v, hasEq := strings.Cut(part, "=")
		if !hasEq {
			v = ""
		}
		key, err := url.QueryUnescape(strings.ReplaceAll(k, "+", " "))
		if err != nil {
			key = k
		}
		value, err := url.QueryUnescape(strings.ReplaceAll(v, "+", " "))
		if err != nil {
			value = v
		}
		out = append(out, queryParam{key: key, value: value})
	}
	return out
}

func bodySyntaxFor(headers map[string]string, body string) bodyeditor.Syntax {
	for name, value := range headers {
		if !strings.EqualFold(name, "Content-Type") {
			continue
		}
		switch {
		case strings.Contains(strings.ToLower(value), "json"):
			return bodyeditor.SyntaxJSON
		case strings.Contains(strings.ToLower(value), "yaml"):
			return bodyeditor.SyntaxYAML
		case strings.Contains(strings.ToLower(value), "graphql"):
			return bodyeditor.SyntaxGraphQL
		}
	}

	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return bodyeditor.SyntaxText
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return bodyeditor.SyntaxJSON
	}
	if strings.HasPrefix(trimmed, "---") || strings.Contains(trimmed, ": ") {
		return bodyeditor.SyntaxYAML
	}
	first := strings.Fields(trimmed)
	if len(first) > 0 {
		switch first[0] {
		case "query", "mutation", "subscription", "fragment":
			return bodyeditor.SyntaxGraphQL
		}
	}
	return bodyeditor.SyntaxText
}

// KeyMap defines the key bindings for the request pane.
type KeyMap struct{}

// ShortHelp returns the primary bindings shown in the collapsed shortcuts bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("]", "["), key.WithHelp("]/[", "next/prev tab")),
		key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "navigate")),
		key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "insert / edit")),
	}
}

// FullHelp returns all bindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			key.NewBinding(key.WithKeys("]", "["), key.WithHelp("]/[", "next/prev tab")),
		},
		{
			key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "navigate")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete row")),
			key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "insert / edit")),
		},
		{
			key.NewBinding(key.WithKeys("enter", "i"), key.WithHelp("enter/i", "edit body")),
			key.NewBinding(key.WithKeys("h", "l"), key.WithHelp("h/l", "cycle body type")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "exit editor")),
			key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear body")),
		},
	}
}

// Keys is the default KeyMap for the request pane.
var Keys = KeyMap{}

// padRight truncates or pads text to exactly n visible columns.
func padRight(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := ansi.Truncate(s, n, "")
	if w := ansi.StringWidth(out); w < n {
		out += strings.Repeat(" ", n-w)
	}
	return out
}

// requestTabLabels returns the labels for the request tabs.
func requestTabLabels() []string {
	return []string{
		requestTabHeaders.label(),
		requestTabParams.label(),
		requestTabBody.label(),
		requestTabAuth.label(),
	}
}

// label returns the label for the request tab.
func (t requestTabID) label() string {
	switch t {
	case requestTabHeaders:
		return headersTabLabel
	case requestTabParams:
		return paramsTabLabel
	case requestTabBody:
		return bodyTabLabel
	case requestTabAuth:
		return authTabLabel
	default:
		return headersTabLabel
	}
}

// requestTabFromLabel returns the request tab ID from the label.
func requestTabFromLabel(label string) requestTabID {
	switch label {
	case paramsTabLabel:
		return requestTabParams
	case bodyTabLabel:
		return requestTabBody
	case authTabLabel:
		return requestTabAuth
	default:
		return requestTabHeaders
	}
}

// label returns the label for the body type.
func (t bodyTypeID) label() string {
	switch t {
	case bodyTypeRaw:
		return "raw"
	default:
		return "none"
	}
}

// cycle cycles the body type.
// The cycle is: none -> raw -> none.
func (t bodyTypeID) cycle() bodyTypeID {
	if t == bodyTypeNone {
		return bodyTypeRaw
	}
	return bodyTypeNone
}
