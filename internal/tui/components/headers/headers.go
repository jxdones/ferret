package headers

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	minHeaderRowWidth = 10

	nameColumnWeight  = 2
	valueColumnWeight = 3
	totalColumnWeight = nameColumnWeight + valueColumnWeight

	// textinput.SetWidth is the editable width; reserve space for prompt / framing in View.
	nameInputReservedCols  = 2
	valueInputReservedCols = 1
)

type header struct {
	name  string
	value string
}

type computedHeader struct {
	name  string
	value string
}

// defaultComputedHeaders are the headers that are added to all requests.
var defaultComputedHeaders = []computedHeader{
	{
		name:  "Accept",
		value: "*/*",
	},
	{
		name:  "User-Agent",
		value: "ferret",
	},
}

// Model is a key-value headers editor. Existing rows are navigable;
type Model struct {
	items     []header
	cursor    int
	activeCol int

	nameInput  textinput.Model
	valueInput textinput.Model

	width int

	focused   bool
	inserting bool
}

// New creates a new headers model.
func New() Model {
	m := Model{
		items:  nil,
		width:  40,
		cursor: 0,
	}
	m.nameInput = newInput("Name")
	m.valueInput = newInput("Value")
	m.applyInputFocus()
	return m
}

// newInput creates a new textinput model with the given placeholder.
func newInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256

	styles := ti.Styles()
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(theme.Current.TextAccent)
	styles.Blurred.Prompt = styles.Blurred.Prompt.Foreground(theme.Current.TextDim)
	ti.SetStyles(styles)

	return ti
}

// SetSize sets the width of the headers model.
func (m *Model) SetSize(width int) {
	m.width = max(minHeaderRowWidth, width)
	nameColumnWidth, valueColumnWidth := m.columnWidths()
	m.nameInput.SetWidth(nameColumnWidth - nameInputReservedCols)
	m.valueInput.SetWidth(valueColumnWidth - valueInputReservedCols)
}

// SetFocused sets the focused state of the headers model.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	m.applyInputFocus()
}

// IsInserting returns true if the headers model is in inserting mode.
func (m Model) IsInserting() bool {
	return m.inserting
}

// MoveCursor adjusts the cursor position by the given delta.
func (m *Model) MoveCursor(delta int) {
	if len(m.items) == 0 {
		return
	}
	m.cursor = max(0, min(m.cursor+delta, len(m.items)-1))
	m.applyInputFocus()
}

// DeleteCursorRow removes the header at the cursor. Default/computed headers are
// not listed in the cursor list — only user-added rows can be deleted.
func (m *Model) DeleteCursorRow() {
	if len(m.items) == 0 {
		return
	}
	m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)
	if m.cursor >= len(m.items) && m.cursor > 0 {
		m.cursor--
	}
	m.applyInputFocus()
}

// Headers returns the current headers as a name→value map.
func (m Model) Headers() map[string]string {
	out := make(map[string]string, len(m.items))
	for _, header := range m.items {
		if header.name != "" {
			out[header.name] = header.value
		}
	}
	return out
}

// SetHeaders sets the headers of the headers model.
func (m *Model) SetHeaders(headers map[string]string) {
	m.items = make([]header, 0, len(headers))
	for name, value := range headers {
		m.items = append(m.items, header{name: name, value: value})
	}
	m.cursor = 0
	m.inserting = false
}

// Row is a name/value pair for ReadOnlyView.
type Row struct {
	Name  string
	Value string
}

// ReadOnlyView renders a Name/Value table without editing chrome, matching the
// layout of Model.View (column titles, dividers, aligned rows).
func ReadOnlyView(width int, rows []Row) tea.View {
	width = max(minHeaderRowWidth, width)
	nameW, valW := columnWidthsFor(width)
	accent := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	dim := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	bodyName := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	bodyValue := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	divider := dim.Render(strings.Repeat("─", width))

	var lines []string
	nameHeader := accent.Render(padRight("  Name", nameW))
	valHeader := accent.Render(padRight("Value", valW))
	lines = append(lines, nameHeader+valHeader)
	lines = append(lines, divider)

	for _, h := range rows {
		name := bodyName.Render(padRight(ansi.Truncate("  "+h.Name, nameW, "…"), nameW))
		val := bodyValue.Render(padRight(ansi.Truncate(h.Value, valW, "…"), valW))
		lines = append(lines, name+val)
		lines = append(lines, divider)
	}

	return tea.NewView(strings.Join(lines, "\n"))
}

// View renders the headers table as a multi-line string padded to m.width.
func (m Model) View() tea.View {
	nameColumnWidth, valueColumnWidth := m.columnWidths()
	accent := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	dim := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	bodyNameOff := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	bodyValueOff := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	bodyNameSel := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	bodyValueSel := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	divider := dim.Render(strings.Repeat("─", m.width))

	var lines []string

	// Title row: same name/value column widths as data rows; muted + dim only.
	nameHdr := accent.Render(padRight("  Name", nameColumnWidth))
	valHdr := accent.Render(padRight("Value", valueColumnWidth))
	lines = append(lines, nameHdr+valHdr)
	lines = append(lines, divider)

	hasBody := false

	computed := m.missingComputedHeaders()
	if len(computed) > 0 {
		hasBody = true
		lines = append(lines, muted.Render(padRight("  computed at send time", m.width)))
		lines = append(lines, divider)
		for _, h := range computed {
			name := muted.Render(padRight(ansi.Truncate("  "+h.name, nameColumnWidth, "…"), nameColumnWidth))
			val := muted.Render(padRight(ansi.Truncate(h.value, valueColumnWidth, "…"), valueColumnWidth))
			lines = append(lines, name+val)
			lines = append(lines, divider)
		}
	}

	for i, h := range m.items {
		hasBody = true
		nameStyle, valStyle := bodyNameOff, bodyValueOff
		if i == m.cursor {
			nameStyle, valStyle = bodyNameSel, bodyValueSel
			if m.focused {
				nameStyle = nameStyle.Bold(true)
				valStyle = valStyle.Bold(true)
			}
		}
		name := nameStyle.Render(padRight(ansi.Truncate("  "+h.name, nameColumnWidth, "…"), nameColumnWidth))
		val := valStyle.Render(padRight(ansi.Truncate(h.value, valueColumnWidth, "…"), valueColumnWidth))
		lines = append(lines, name+val)
		lines = append(lines, divider)
	}

	if m.inserting {
		nameView := m.nameInput.View()
		valueView := m.valueInput.View()
		lines = append(lines, " "+nameView+" "+valueView)
	} else {
		hint := lipgloss.NewStyle().
			Foreground(theme.Current.TextMuted).
			Render("  i/I/A add · d delete row · tab/shift+tab fields")
		if !hasBody {
			lines = append(lines, divider)
		}
		lines = append(lines, padRight(hint, m.width))
	}

	return tea.NewView(strings.Join(lines, "\n"))
}

// Update handles key events from the parent model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		next, cmd, handled := m.tryHandleKey(key)
		if handled {
			return next, cmd
		}
	}

	var cmd tea.Cmd
	if m.activeCol == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
	} else {
		m.valueInput, cmd = m.valueInput.Update(msg)
	}
	return m, cmd
}

// tryHandleKey runs built-in shortcuts (insert mode, column focus, row commit).
// If handled is true, the event must not be forwarded to textinput. If handled is
// false, the parent should pass the same message to the active field editor.
func (m Model) tryHandleKey(key tea.KeyMsg) (Model, tea.Cmd, bool) {
	ks := key.String()

	if !m.inserting {
		if ks == "i" || ks == "I" || ks == "A" {
			m.inserting = true
			m.activeCol = 0
			m.applyInputFocus()
			return m, nil, true
		}
		return m, nil, false
	}

	if ks == "esc" {
		m.inserting = false
		m.nameInput.SetValue("")
		m.valueInput.SetValue("")
		m.activeCol = 0
		m.applyInputFocus()
		return m, nil, true
	}

	return m.consumeInsertRowKey(key)
}

// consumeInsertRowKey handles navigation and insert-mode shortcuts. If handled is
// true, the event must not be forwarded to textinput. If handled is false, the
// parent should pass the same message to the active field editor.
func (m Model) consumeInsertRowKey(key tea.KeyMsg) (Model, tea.Cmd, bool) {
	switch key.String() {
	case "tab":
		m.activeCol = 1
		m.applyInputFocus()
		return m, nil, true
	case "shift+tab":
		m.activeCol = 0
		m.applyInputFocus()
		return m, nil, true
	case "enter":
		return m.commitInsertRow()
	default:
		return m, nil, false
	}
}

// commitInsertRow commits the current insert row and resets the insert mode.
func (m Model) commitInsertRow() (Model, tea.Cmd, bool) {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		return m, nil, true
	}
	m.items = append(m.items, header{
		name:  name,
		value: strings.TrimSpace(m.valueInput.Value()),
	})
	m.nameInput.SetValue("")
	m.valueInput.SetValue("")
	m.activeCol = 0
	m.applyInputFocus()
	return m, nil, true
}

// columnWidths returns the width of the name and value columns.
func (m Model) columnWidths() (int, int) {
	return columnWidthsFor(m.width)
}

// columnWidthsFor splits total width into name and value columns.
func columnWidthsFor(width int) (nameColumnWidth, valueColumnWidth int) {
	nameColumnWidth = width * nameColumnWeight / totalColumnWeight
	return nameColumnWidth, width - nameColumnWidth
}

// applyInputFocus sets the focus state of the name and value inputs based
// on the current active column and the focused state.
func (m *Model) applyInputFocus() {
	if m.inserting && m.focused {
		if m.activeCol == 0 {
			m.nameInput.Focus()
			m.valueInput.Blur()
		} else {
			m.nameInput.Blur()
			m.valueInput.Focus()
		}
	} else {
		m.nameInput.Blur()
		m.valueInput.Blur()
	}
}

// padRight pads or truncates s to exactly n visible characters.
func padRight(s string, n int) string {
	out := ansi.Truncate(s, n, "")
	if w := ansi.StringWidth(out); w < n {
		out += strings.Repeat(" ", n-w)
	}
	return out
}

// missingComputedHeaders returns default computed headers not overridden by user rows.
func (m Model) missingComputedHeaders() []computedHeader {
	if len(defaultComputedHeaders) == 0 {
		return nil
	}

	defined := make(map[string]struct{}, len(m.items))
	for _, h := range m.items {
		k := strings.ToLower(strings.TrimSpace(h.name))
		if k != "" {
			defined[k] = struct{}{}
		}
	}

	out := make([]computedHeader, 0, len(defaultComputedHeaders))
	for _, h := range defaultComputedHeaders {
		if _, ok := defined[strings.ToLower(h.name)]; ok {
			continue
		}
		out = append(out, h)
	}
	return out
}
