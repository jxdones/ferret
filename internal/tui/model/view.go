package model

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jxdones/ferret/internal/tui/common"
	"github.com/jxdones/ferret/internal/tui/components/requestpane"
	"github.com/jxdones/ferret/internal/tui/components/responsepane"
	"github.com/jxdones/ferret/internal/tui/components/shortcuts"
	"github.com/jxdones/ferret/internal/tui/components/urlbar"
	"github.com/jxdones/ferret/internal/tui/keys"
	"github.com/jxdones/ferret/internal/tui/modal"
	"github.com/jxdones/ferret/internal/tui/theme"
)

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

// render renders the full screen.
func (m Model) render() string {
	if m.width == 0 {
		return ""
	}
	base := common.NormalizeCanvas(m.renderBase(), m.width, m.height)
	out := m.renderModal(base)
	return common.NormalizeCanvas(out, m.width, m.height)
}

// renderBase renders the full screen without any modal overlay.
func (m Model) renderBase() string {
	// Fixed top rows: titlebar, divider, requestTabs, urlbar, dividerSplit, paneLabels, tabsRow, tabsDivider = 8
	// Bottom: statusbar (2) + dynamic options height (see fixedFrameRows in layout.go)
	sections := []string{
		m.titlebar.View().Content,
		m.renderDivider(),
		m.renderRequestTabs(),
		m.tab().urlbar.View().Content,
		m.renderDividerSplit(),
		m.renderPaneLabels(),
		m.renderTabsRow(),
		m.renderTabsDivider(),
	}
	sections = append(sections, m.renderContentLines()...)
	sections = append(sections,
		m.statusbar.View().Content,
		m.renderOptions(),
	)
	return strings.Join(sections, "\n")
}

// renderModal composites a modal overlay onto the base canvas.
func (m Model) renderModal(base string) string {
	switch m.activeModal {
	case modalCollection:
		footer := shortcuts.RenderShortcuts(modal.InnerWidth(m.modalOuterWidth()), []key.Binding{
			key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
		})
		overlay := modal.Render("requests", m.collection.View().Content, footer, m.modalOuterWidth())
		return overlayAtCenter(base, overlay, m.width, m.height)
	case modalWorkspace:
		footer := shortcuts.RenderShortcuts(modal.InnerWidth(m.modalOuterWidth()), []key.Binding{
			key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
		})
		overlay := modal.Render("collections", m.workspacePicker.View().Content, footer, m.modalOuterWidth())
		return overlayAtCenter(base, overlay, m.width, m.height)
	case modalMethod:
		footer := shortcuts.RenderShortcuts(modal.InnerWidth(m.modalOuterWidth()), []key.Binding{
			key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "navigate")),
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
		})
		overlay := modal.Render("method", m.methods.View().Content, footer, m.modalOuterWidth())
		return overlayAtCenter(base, overlay, m.width, m.height)
	default:
		return base
	}
}

// renderOptions renders the bottom shortcuts/help bar.
func (m Model) renderOptions() string {
	var content string
	if m.helpExpanded {
		h := help.New()
		h.ShowAll = true
		h.SetWidth(m.width)
		h.Styles.FullKey = lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
		h.Styles.FullDesc = lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
		h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(theme.Current.DividerBorder)
		h.Styles.Ellipsis = lipgloss.NewStyle().Foreground(theme.Current.DividerBorder)
		content = h.FullHelpView(m.fullHelpBindings())
	} else {
		content = shortcuts.RenderShortcuts(m.width, m.statusBindings())
	}

	return lipgloss.NewStyle().
		Width(m.width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Current.DividerBorder).
		Padding(0, 1).
		Render(content)
}

// optionsHeight returns the number of rows renderOptions() will occupy.
func (m Model) optionsHeight() int {
	if !m.helpExpanded {
		return 2
	}
	h := help.New()
	h.SetWidth(m.width)
	return 1 + lipgloss.Height(h.FullHelpView(m.fullHelpBindings()))
}

// statusBindings returns the bindings for the collapsed shortcuts bar.
func (m Model) statusBindings() []key.Binding {
	return append(m.paneKeyMap().ShortHelp(), m.shortGlobalBindings()...)
}

// fullHelpBindings returns the bindings for the expanded help view.
func (m Model) fullHelpBindings() [][]key.Binding {
	return append(m.paneKeyMap().FullHelp(), m.fullGlobalBindings())
}

// paneKeyMap returns the help.KeyMap for the currently focused pane.
func (m Model) paneKeyMap() help.KeyMap {
	switch m.focus {
	case focusURLBar:
		return urlbar.Keys
	case focusRequestPane:
		return requestpane.Keys
	case focusResponsePane:
		return responsepane.Keys
	default:
		return emptyKeyMap{}
	}
}

// shortGlobalBindings returns the 2–3 global bindings always shown in the bar.
func (m Model) shortGlobalBindings() []key.Binding {
	return []key.Binding{
		keys.Default.SendRequest,
		keys.Default.Help,
	}
}

// fullGlobalBindings returns all global bindings for the expanded help view.
func (m Model) fullGlobalBindings() []key.Binding {
	return []key.Binding{
		keys.Default.SendRequest,
		keys.Default.NewRequest,
		keys.Default.URLFocus,
		keys.Default.MethodCycle,
		keys.Default.MethodPicker,
		keys.Default.EnvCycle,
		keys.Default.Collection,
		keys.Default.CollectionCycle,
		keys.Default.WorkspacePick,
		keys.Default.NextTab,
		keys.Default.PrevTab,
		keys.Default.NewTab,
		keys.Default.CloseTab,
		keys.Default.Help,
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
		key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev pane")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear focus")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

// emptyKeyMap is returned when no pane is focused (global focus or modal states).
type emptyKeyMap struct{}

func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }

// renderDivider renders the full-width horizontal divider below the URL bar.
func (m Model) renderDivider() string {
	return lipgloss.NewStyle().Foreground(theme.Current.TextDim).Render(strings.Repeat("─", m.width))
}

// renderDividerSplit renders the split divider between left and right panes.
func (m Model) renderDividerSplit() string {
	mid := m.width / 2
	center := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	return center.Render(strings.Repeat("─", mid)) +
		center.Render("┬") +
		center.Render(strings.Repeat("─", m.width-mid-1))
}

// renderTabsDivider renders pane dividers with active-tab highlight segments.
func (m Model) renderTabsDivider() string {
	mid := m.width / 2
	dim := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	reqTabLine := lipgloss.NewStyle().Foreground(theme.Current.RequestPaneLabel)
	resTabLine := lipgloss.NewStyle().Foreground(theme.Current.ResponsePaneLabel)

	leftStart, leftWidth := m.tab().requestPane.ActiveTabSpan()
	rightStart, rightWidth := m.tab().responsePane.ActiveTabSpan()

	leftActive := dim
	rightActive := dim
	if m.focus == focusRequestPane {
		leftActive = reqTabLine
	}
	if m.focus == focusResponsePane {
		rightActive = resTabLine
	}

	left := paneDivider(mid, leftStart, leftWidth, dim, leftActive)
	right := paneDivider(m.width-mid-1, rightStart, rightWidth, dim, rightActive)
	return left + dim.Render("┼") + right
}

// renderRequestTabs renders the top-level request tab strip.
func (m Model) renderRequestTabs() string {
	var sb strings.Builder
	sb.WriteString(" ")
	for i, t := range m.tabs {
		if i > 0 {
			sb.WriteString("  ")
		}
		title := t.title
		if title == "" {
			title = "new request"
		}
		if i == m.activeTab {
			method := t.urlbar.Method()
			activeStyle := lipgloss.NewStyle().
				Background(theme.MethodColor(method)).
				Foreground(lipgloss.Color("#111111")).
				Bold(true).
				Padding(0, 1)
			sb.WriteString(activeStyle.Render(title))
		} else {
			method := t.urlbar.Method()
			nameStyle := lipgloss.NewStyle().Foreground(theme.Current.TabsInactiveText)
			sb.WriteString(" ")
			if method != "" && strings.HasPrefix(title, method+" ") {
				methodStyle := lipgloss.NewStyle().Foreground(theme.MethodColor(method))
				sb.WriteString(methodStyle.Render(method))
				sb.WriteString(nameStyle.Render(" " + title[len(method)+1:]))
			} else {
				sb.WriteString(nameStyle.Render(title))
			}
			sb.WriteString(" ")
		}
	}
	built := sb.String()
	return built + strings.Repeat(" ", max(0, m.width-lipgloss.Width(built)))
}

// renderTabsRow renders both pane tab strips separated by a vertical divider.
func (m Model) renderTabsRow() string {
	divider := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	return m.tab().requestPane.TabsView().Content + divider.Render("│") + m.tab().responsePane.TabsView().Content
}

// renderPaneLabels renders request/response labels with focus highlighting.
func (m Model) renderPaneLabels() string {
	mid := m.width / 2
	rightWidth := m.width - mid - 1

	leftLabelStyle := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	rightLabelStyle := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	if m.focus == focusRequestPane {
		leftLabelStyle = lipgloss.NewStyle().Foreground(theme.Current.RequestPaneLabel).Bold(true)
	}
	if m.focus == focusResponsePane {
		rightLabelStyle = lipgloss.NewStyle().Foreground(theme.Current.ResponsePaneLabel).Bold(true)
	}

	divider := lipgloss.NewStyle().Foreground(theme.Current.TextDim)
	left := fitToWidth([]string{" " + leftLabelStyle.Render("request")}, mid)[0]
	right := fitToWidth([]string{" " + rightLabelStyle.Render("response")}, rightWidth)[0]
	return left + divider.Render("│") + right
}

// renderContentLines builds the split content area by delegating to each pane.
func (m Model) renderContentLines() []string {
	contentHeight := m.contentHeight()
	mid := m.width / 2
	rightWidth := m.width - mid - 1
	divider := lipgloss.NewStyle().Foreground(theme.Current.TextDim)

	leftLines := splitAndFit(m.tab().requestPane.View().Content, mid)
	rightLines := splitAndFit(m.tab().responsePane.View().Content, rightWidth)

	lines := make([]string, contentHeight)
	for i := range lines {
		left := strings.Repeat(" ", mid)
		if i < len(leftLines) {
			left = leftLines[i]
		}
		right := strings.Repeat(" ", rightWidth)
		if i < len(rightLines) {
			right = rightLines[i]
		}
		lines[i] = left + divider.Render("│") + right
	}
	return lines
}

// splitAndFit splits content on newlines and pads/truncates each line to width.
func splitAndFit(content string, width int) []string {
	if content == "" {
		return nil
	}
	return fitToWidth(strings.Split(content, "\n"), width)
}

// fitToWidth pads or truncates each line to exactly width visible columns.
func fitToWidth(lines []string, width int) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		w := ansi.StringWidth(line)
		switch {
		case w < width:
			result[i] = line + strings.Repeat(" ", width-w)
		case w > width:
			result[i] = ansi.Truncate(line, width, "")
		default:
			result[i] = line
		}
	}
	return result
}

// overlayAtCenter composites overlay centered over a dimmed base canvas.
func overlayAtCenter(base, overlay string, width, height int) string {
	dimStyle := lipgloss.NewStyle().
		Foreground(theme.Current.TextMuted).
		Faint(true)

	canvas := lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base)
	baseLines := strings.Split(canvas, "\n")
	for i := range baseLines {
		baseLines[i] = fitStyled(ansi.Strip(baseLines[i]), width)
	}

	overlayLines := strings.Split(overlay, "\n")
	overlayWidth := 1
	for _, ln := range overlayLines {
		if w := ansi.StringWidth(ln); w > overlayWidth {
			overlayWidth = w
		}
	}

	x := (width - overlayWidth) / 2
	y := (height - len(overlayLines)) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	overlayRows := make(map[int]struct{}, len(overlayLines))
	for i, line := range overlayLines {
		row := y + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		dst := []rune(baseLines[row])
		prefixEnd := min(x, len(dst))
		suffixStart := min(x+overlayWidth, len(dst))
		prefix := string(dst[:prefixEnd])
		suffix := string(dst[suffixStart:])
		baseLines[row] = dimStyle.Render(prefix) + fitStyled(line, overlayWidth) + dimStyle.Render(suffix)
		overlayRows[row] = struct{}{}
	}

	for i := range baseLines {
		if _, ok := overlayRows[i]; !ok {
			baseLines[i] = dimStyle.Render(baseLines[i])
		}
		baseLines[i] = fitStyled(baseLines[i], width)
	}

	return strings.Join(baseLines, "\n")
}

// fitStyled truncates or pads s to exactly width visible columns, preserving ANSI styles.
func fitStyled(s string, width int) string {
	out := ansi.Truncate(s, width, "")
	if w := ansi.StringWidth(out); w < width {
		out += strings.Repeat(" ", width-w)
	}
	return out
}

// paneDivider renders a divider line with an optional highlighted active span.
func paneDivider(width, activeStart, activeWidth int, dim, active lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	if activeWidth <= 0 || activeStart >= width {
		return dim.Render(strings.Repeat("─", width))
	}
	if activeStart < 0 {
		activeStart = 0
	}
	end := min(width, activeStart+activeWidth)
	if end <= activeStart {
		return dim.Render(strings.Repeat("─", width))
	}

	var b strings.Builder
	if activeStart > 0 {
		b.WriteString(dim.Render(strings.Repeat("─", activeStart)))
	}
	b.WriteString(active.Render(strings.Repeat("─", end-activeStart)))
	if end < width {
		b.WriteString(dim.Render(strings.Repeat("─", width-end)))
	}
	return b.String()
}
