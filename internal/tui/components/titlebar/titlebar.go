package titlebar

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/tui/theme"
)

// Model represents the titlebar model.
type Model struct {
	workspace  string
	collection string
	env        string
	entry      string
	width      int
}

// New creates a initialized titlebar model.
func New() Model {
	return Model{}
}

// SetSize sets the width of the titlebar.
func (m *Model) SetSize(width int) {
	m.width = width
}

// SetWorkspace sets the workspace name on the left.
func (m *Model) SetWorkspace(workspace string) {
	m.workspace = strings.TrimSpace(workspace)
}

// SetCollection sets the collection name.
func (m *Model) SetCollection(collection string) {
	m.collection = collection
}

// SetEnv sets the file-backed environment name shown on the right. An empty
// string shows "shell only" (no environments/<name>.yaml layer).
func (m *Model) SetEnv(env string) {
	m.env = env
}

// SetEntry sets the name for the currently loaded collection entry.
func (m *Model) SetEntry(name string) {
	m.entry = strings.TrimSpace(name)
}

// View returns the titlebar view.
func (m *Model) View() tea.View {
	workspaceStyle := lipgloss.NewStyle().Foreground(theme.Current.TitleBarWorkspace)
	collectionStyle := lipgloss.NewStyle().Foreground(theme.Current.TitleBarCollection)
	entryStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.TitleBarEntry)
	sepStyle := lipgloss.NewStyle().Foreground(theme.Current.TitleBarSeparator)
	fileEnv := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.TextAccent)
	shellEnv := lipgloss.NewStyle().Foreground(theme.Current.MethodPATCH)

	sep := sepStyle.Render(" / ")

	var segments []string
	if m.workspace != "" {
		segments = append(segments, workspaceStyle.Render(m.workspace))
	}
	if m.collection != "" {
		segments = append(segments, collectionStyle.Render(m.collection))
	}
	if m.entry != "" {
		segments = append(segments, entryStyle.Render(m.entry))
	}

	left := " "
	if len(segments) > 0 {
		left += strings.Join(segments, sep)
	}

	right := ""
	if strings.TrimSpace(m.env) == "" {
		right = shellEnv.Render("shell only ")
	} else {
		right = fileEnv.Render(m.env + " ")
	}

	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return tea.NewView(left + strings.Repeat(" ", gap) + right)
}
