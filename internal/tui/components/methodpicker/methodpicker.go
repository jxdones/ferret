package methodpicker

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/tui/theme"
)

var defaultMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

// Model is a selectable list of HTTP methods.
type Model struct {
	methods []string
	cursor  int
	width   int
}

// New initializes a method picker model with the default methods.
func New() Model {
	return Model{
		methods: defaultMethods,
		cursor:  0,
		width:   30,
	}
}

// SetSize sets the picker width, clamping to a readable minimum.
func (m *Model) SetSize(width int) {
	m.width = max(10, width)
}

// SetActive sets the active method. If the method is not found, the cursor is
// left unchanged.
func (m *Model) SetActive(method string) {
	for i, mtd := range m.methods {
		if strings.EqualFold(mtd, strings.TrimSpace(method)) {
			m.cursor = i
			break
		}
	}
}

// MoveCursor adjusts the cursor position by the given delta.
func (m *Model) MoveCursor(delta int) {
	if len(m.methods) == 0 {
		return
	}
	m.cursor = clamp(m.cursor+delta, 0, len(m.methods)-1)
}

// Selected returns the currently selected method.
func (m Model) Selected() string {
	if m.cursor < 0 || m.cursor >= len(m.methods) {
		return ""
	}
	return m.methods[m.cursor]
}

// View returns the method picker view.
func (m Model) View() tea.View {
	var rows []string
	for i, method := range m.methods {
		prefix := " "
		style := lipgloss.NewStyle().
			Foreground(theme.MethodColor(method)).
			Bold(true)
		if i == m.cursor {
			prefix = lipgloss.NewStyle().
				Foreground(theme.Current.TextAccent).
				Render("▶")
			style = style.Foreground(theme.Current.TextAccent)
		}
		// format the method with a fixed width of 7 characters
		rows = append(rows, prefix+style.Render(fmt.Sprintf("%-7s", method)))
	}

	if len(rows) == 0 {
		return tea.NewView(lipgloss.NewStyle().
			Foreground(theme.Current.TextMuted).
			Render(" (no methods)"))
	}
	return tea.NewView(strings.Join(rows, "\n"))
}

// clamp ensures a value is within a given range.
func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
