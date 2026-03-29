package workspacepicker

import (
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jxdones/ferret/internal/tui/theme"
)

const maxVisible = 8

type item struct {
	path string
	name string
}

// Model is a fuzzy finder for workspace collection directories.
type Model struct {
	input    textinput.Model
	all      []item
	filtered []item
	cursor   int
	width    int
}

// New returns an initialized picker with no items.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "search collections…"
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

// Load replaces the list with absolute collection directory paths.
func (m *Model) Load(paths []string) {
	m.all = make([]item, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m.all = append(m.all, item{path: p, name: filepath.Base(p)})
	}
	m.refilter()
	m.cursor = 0
}

// Selected returns the highlighted directory path, if any.
func (m Model) Selected() (string, bool) {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return "", false
	}
	return m.filtered[m.cursor].path, true
}

// SetSize sets the inner content width.
func (m *Model) SetSize(width int) {
	m.width = width
	m.input.SetWidth(width - 2)
}

// Reset clears the search input and cursor. Call when opening the modal.
func (m *Model) Reset() {
	m.input.SetValue("")
	m.cursor = 0
	m.refilter()
	m.input.Focus()
}

// SetActive moves the cursor to a path equal to activePath (if present in the filtered list).
func (m *Model) SetActive(activePath string) {
	ap := strings.TrimSpace(activePath)
	if ap == "" {
		return
	}
	for i, it := range m.filtered {
		if it.path == ap {
			m.cursor = i
			return
		}
	}
}

// MoveCursor moves the list cursor by delta.
func (m *Model) MoveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.filtered)-1))
}

// Update forwards keys to the search field and re-filters on change.
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

// View renders search input and the filtered list.
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
		lineStyle := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
		if i == m.cursor {
			prefix = lipgloss.NewStyle().Foreground(theme.Current.TextAccent).Render("▶ ")
			lineStyle = lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
		}
		rows = append(rows, prefix+lineStyle.Render(it.name))
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

func (m *Model) refilter() {
	q := strings.ToLower(m.input.Value())
	if q == "" {
		m.filtered = m.all
		return
	}
	m.filtered = nil
	for _, it := range m.all {
		if strings.Contains(strings.ToLower(it.name), q) ||
			strings.Contains(strings.ToLower(it.path), q) {
			m.filtered = append(m.filtered, it)
		}
	}
}
