package tabs

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/tui/common"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	tabSlotWidth = 10
	minTabsWidth = 8
)

// Model represents a set of tabs for request and response panes.
type Model struct {
	labels  []string
	active  int
	width   int
	focused bool

	// activeForeground styles the active tab when focused. If nil, TextAccent is used.
	activeForeground color.Color
}

// New initializes a tabs model with the given labels. The first tab is active
// by default.
func New(labels []string) Model {
	return Model{
		labels: labels,
		width:  common.ClampMin(len(labels)*tabSlotWidth, minTabsWidth),
	}
}

// SetSize sets the width of the tabs model.
func (m *Model) SetSize(width int) {
	m.width = common.ClampMin(width, 1)
}

// SetFocused sets the focused state of the tabs model.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
}

// SetActiveForeground sets the focused active tab text color. Pass nil to use
// the default accent color.
func (m *Model) SetActiveForeground(c color.Color) {
	m.activeForeground = c
}

// SetActive sets the active tab by index. If the index is out of bounds, the
// active tab is left unchanged.
func (m *Model) SetActive(index int) {
	if index >= 0 && index < len(m.labels) {
		m.active = index
	}
}

// Active returns the index of the active tab.
func (m Model) Active() int {
	return m.active
}

// ActiveLabel returns the label of the active tab.
func (m Model) ActiveLabel() string {
	if m.active < 0 || m.active >= len(m.labels) {
		return ""
	}
	return m.labels[m.active]
}

// Next cycles to the next tab.
func (m *Model) Next() {
	if len(m.labels) == 0 {
		return
	}
	m.active++
	if m.active >= len(m.labels) {
		m.active = 0
	}
}

// Previous cycles to the previous tab.
func (m *Model) Previous() {
	if len(m.labels) == 0 {
		return
	}
	m.active--
	if m.active < 0 {
		m.active = len(m.labels) - 1
	}
}

// ActiveSpan returns the start column and width of the active tab label.
func (m Model) ActiveSpan() (startColumn, activeWidth int) {
	if len(m.labels) == 0 || m.active < 0 || m.active >= len(m.labels) {
		return 0, 0
	}

	// mirrors View(): leading space + "  " between labels.
	col := 1
	for i, label := range m.labels {
		if i == m.active {
			return col, lipgloss.Width(label)
		}
		col += lipgloss.Width(label)
		if i < len(m.labels)-1 {
			col += 2
		}
	}
	return 0, 0
}

// View renders the tabs view as a single line padded to the full width.
func (m Model) View() tea.View {
	var activeStyle, inactiveStyle lipgloss.Style
	if m.focused {
		activeFG := theme.Current.TextAccent
		if m.activeForeground != nil {
			activeFG = m.activeForeground
		}
		activeStyle = lipgloss.NewStyle().Foreground(activeFG).Bold(true)
		inactiveStyle = lipgloss.NewStyle().Foreground(theme.Current.TabsInactiveText)
	} else {
		activeStyle = lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
		inactiveStyle = lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	}

	var sb strings.Builder
	sb.WriteString(" ")
	for i, label := range m.labels {
		if i > 0 {
			sb.WriteString("  ")
		}
		if i == m.active {
			sb.WriteString(activeStyle.Render(label))
		} else {
			sb.WriteString(inactiveStyle.Render(label))
		}
	}

	builtString := sb.String()
	tabsLine := builtString + strings.Repeat(" ", max(0, m.width-lipgloss.Width(builtString)))
	return tea.NewView(tabsLine)
}
