package urlbar

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/theme"
)

// Model renders and edits the HTTP method badge + URL.
type Model struct {
	method  string
	input   textinput.Model
	width   int
	focused bool
}

// New creates a initialized urlbar model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "type your URL here..."
	ti.Prompt = ""
	ti.CharLimit = 2048 // safe limit for most URLs
	ti.SetValue("")
	ti.Blur()

	styles := ti.Styles()
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(theme.Current.TextPrimary)
	styles.Focused.Text = styles.Focused.Text.Foreground(theme.Current.TextPrimary)
	styles.Blurred.Text = styles.Blurred.Text.Foreground(theme.Current.TextMuted)
	ti.SetStyles(styles)

	return Model{
		method: "GET",
		input:  ti,
	}
}

// SetSize sets the width of the urlbar.
func (m *Model) SetSize(width int) {
	m.width = width
	m.applyInputWidth()
}

// SetMethod sets the HTTP method.
func (m *Model) SetMethod(method string) {
	m.method = method
	m.applyInputWidth()
}

// Method returns the HTTP method.
func (m Model) Method() string {
	return m.method
}

// SetURL sets the URL.
func (m *Model) SetURL(url string) {
	m.input.SetValue(url)
}

// URL returns the current URL string.
func (m Model) URL() string {
	return m.input.Value()
}

// SetFocused sets the focus state of the urlbar.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	if focused {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
	m.applyInputWidth()
}

// Update handles text input editing while focused.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View returns the urlbar view.
func (m Model) View() tea.View {
	method := lipgloss.NewStyle().Bold(true).Foreground(theme.MethodColor(m.method)).Render(m.method)

	prefix := " " + method + "  "
	available := max(0, m.width-ansi.StringWidth(prefix))

	var url string
	if m.focused {
		// textinput.View includes cursor and styling; keep it within the URL field.
		url = fit(m.input.View(), available)
	} else {
		url = lipgloss.NewStyle().Foreground(theme.Current.TextPrimary).
			Render(ansi.Truncate(m.input.Value(), available, "…"))
		url = fit(url, available)
	}

	line := prefix + url
	return tea.NewView(fit(line, m.width))
}

// applyInputWidth sets the input width from available space and method width.
func (m *Model) applyInputWidth() {
	method := lipgloss.NewStyle().Bold(true).Render(m.method)
	prefix := " " + method + "  "
	available := max(0, m.width-ansi.StringWidth(prefix))
	m.input.SetWidth(available)
}

// KeyMap defines the key bindings for the URL bar.
type KeyMap struct{}

// ShortHelp returns the primary bindings shown in the collapsed shortcuts bar.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear URL")),
	}
}

// FullHelp returns all bindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear URL")),
	}}
}

// Keys is the default KeyMap for the URL bar.
var Keys = KeyMap{}

// fit ensures a string fits within a given width, truncating if necessary.
func fit(s string, width int) string {
	output := ansi.Truncate(s, width, "")
	if w := ansi.StringWidth(output); w < width {
		output += strings.Repeat(" ", width-w)
	}
	return output
}
