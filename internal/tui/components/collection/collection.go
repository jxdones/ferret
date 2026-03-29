package collection

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	collectiondata "github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const maxVisible = 8

type item struct {
	method string
	path   string
	title  string
	entry  collectiondata.Entry
}

// Model is the collection fuzzy finder component.
type Model struct {
	input    textinput.Model
	all      []item
	filtered []item
	cursor   int
	width    int
}

// New returns an initialized collection model with no items.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "search requests…"
	ti.CharLimit = 128
	styles := ti.Styles()
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(theme.Current.TextAccent)
	ti.SetStyles(styles)
	ti.Focus()

	m := Model{
		input: ti,
		width: 50,
	}
	m.refilter()
	return m
}

// Load replaces the item list with entries from a loaded collection.
func (m *Model) Load(entries []collectiondata.Entry) {
	m.all = make([]item, 0, len(entries))
	for _, e := range entries {
		title := strings.TrimSpace(e.Request.Name)
		if title == "" {
			title = e.Path
		}
		m.all = append(m.all, item{
			method: e.Request.Method,
			path:   e.Path,
			title:  title,
			entry:  e,
		})
	}
	m.refilter()
	m.cursor = 0
}

// Selected returns the currently highlighted entry, if any.
func (m *Model) Selected() (collectiondata.Entry, bool) {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return collectiondata.Entry{}, false
	}
	return m.filtered[m.cursor].entry, true
}

// SetSize sets the inner content width of the component.
func (m *Model) SetSize(width int) {
	m.width = width
	m.input.SetWidth(width - 2) // leave room for cursor prefix
}

// Reset clears the search input and resets the cursor. Call when opening the modal.
func (m *Model) Reset() {
	m.input.SetValue("")
	m.cursor = 0
	m.refilter()
	m.input.Focus()
}

// MoveCursor moves the list cursor by delta, clamped to the filtered list bounds.
func (m *Model) MoveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.filtered)-1))
}

// Update forwards key events to the text input and re-filters on value change.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	prev := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != prev {
		m.refilter()
		m.cursor = 0
	}
	return m, cmd
}

// View renders the input and filtered result list.
func (m Model) View() tea.View {
	divider := lipgloss.NewStyle().
		Foreground(theme.Current.DividerBorder).
		Render(strings.Repeat("─", m.width))

	var rows []string
	visible := m.filtered
	if len(visible) > maxVisible {
		visible = visible[:maxVisible]
	}

	for i, it := range visible {
		prefix := "  "
		titleStyle := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(theme.Current.TextAccent).Render("▶ ")
			titleStyle = lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
		}
		method := methodStyle(it.method).Render(fmt.Sprintf("%-7s", it.method))
		title := titleStyle.Render(it.title)
		rows = append(rows, prefix+method+title)
	}

	if len(rows) == 0 {
		rows = []string{
			lipgloss.NewStyle().Foreground(theme.Current.TextMuted).Render("  no results"),
		}
	}

	return tea.NewView(strings.Join(append(
		[]string{m.input.View(), divider},
		rows...,
	), "\n"))
}

// refilter applies the current query to all items.
func (m *Model) refilter() {
	q := strings.ToLower(m.input.Value())
	if q == "" {
		m.filtered = m.all
		return
	}
	m.filtered = nil
	for _, it := range m.all {
		if strings.Contains(strings.ToLower(it.method), q) ||
			strings.Contains(strings.ToLower(it.path), q) ||
			strings.Contains(strings.ToLower(it.title), q) {
			m.filtered = append(m.filtered, it)
		}
	}
}

// methodStyle returns a method-colored style for request method labels.
func methodStyle(method string) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.MethodColor(method))
}
