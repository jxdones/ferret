package responsepane

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	chromaQuick "github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/x/ansi"

	"github.com/jxdones/ferret/internal/exec"
	"github.com/jxdones/ferret/internal/tui/components/headers"
	"github.com/jxdones/ferret/internal/tui/components/tabs"
	"github.com/jxdones/ferret/internal/tui/theme"
)

const (
	bodyTabLabel    = "body"
	headersTabLabel = "headers"
	cookiesTabLabel = "cookies"
	traceTabLabel   = "trace"
)

type responseTabID int

const (
	responseTabBody responseTabID = iota
	responseTabHeaders
	responseTabCookies
	responseTabTrace
)

func responseTabLabels() []string {
	return []string{
		responseTabBody.label(),
		responseTabHeaders.label(),
		responseTabCookies.label(),
		responseTabTrace.label(),
	}
}

func (t responseTabID) label() string {
	switch t {
	case responseTabBody:
		return bodyTabLabel
	case responseTabHeaders:
		return headersTabLabel
	case responseTabCookies:
		return cookiesTabLabel
	case responseTabTrace:
		return traceTabLabel
	default:
		return bodyTabLabel
	}
}

func responseTabFromLabel(label string) responseTabID {
	switch label {
	case headersTabLabel:
		return responseTabHeaders
	case cookiesTabLabel:
		return responseTabCookies
	case traceTabLabel:
		return responseTabTrace
	default:
		return responseTabBody
	}
}

// Model is the response viewer pane.
type Model struct {
	tabs    tabs.Model
	body    []byte
	headers map[string][]string
	trace   exec.Trace
	lines   []string
	offset  int // body tab vertical scroll

	headersOffset int // headers tab vertical scroll
	width         int
	height        int
	focused       bool
}

// New returns an initialized response pane with no data.
func New() Model {
	t := tabs.New(responseTabLabels())
	t.SetActiveForeground(theme.Current.ResponsePaneLabel)
	return Model{
		tabs: t,
	}
}

// SetSize propagates available width and height. Height is used for scroll
// clamping and for deciding how many body lines to render.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.tabs.SetSize(width)
}

// SetFocused propagates focus to the tab strip.
func (m *Model) SetFocused(focused bool) {
	m.focused = focused
	m.tabs.SetFocused(focused)
}

// SetResponse stores response data and resets the scroll position.
// Call this whenever a new response arrives.
func (m *Model) SetResponse(body []byte, headers map[string][]string, trace exec.Trace) {
	m.body = body
	m.headers = headers
	m.trace = trace
	m.offset = 0
	m.headersOffset = 0
	m.rebuildBodyCache()
}

// Reset clears all response data and scroll state.
// Call this when loading a new request from the collection.
func (m *Model) Reset() {
	m.body = nil
	m.headers = nil
	m.trace = exec.Trace{}
	m.lines = nil
	m.offset = 0
	m.headersOffset = 0
}

// TabsView returns the rendered tab strip line.
func (m Model) TabsView() tea.View {
	return m.tabs.View()
}

// ActiveTabSpan returns the start column and width of the active tab label.
func (m Model) ActiveTabSpan() (int, int) {
	return m.tabs.ActiveSpan()
}

// Update handles key events for the response pane.
// Returns (updated model, cmd, handled).
func (m Model) Update(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if !m.focused {
		return m, nil, false
	}

	// Tab strip navigation.
	switch msg.String() {
	case "]":
		m.tabs.Next()
		return m, nil, true
	case "[":
		m.tabs.Previous()
		return m, nil, true
	}

	switch m.activeTab() {
	case responseTabBody:
		switch msg.String() {
		case "j", "down":
			m.scrollBody(1)
			return m, nil, true
		case "k", "up":
			m.scrollBody(-1)
			return m, nil, true
		case "ctrl+d":
			m.scrollBody(m.height / 2)
			return m, nil, true
		case "g":
			m.offset = 0
			return m, nil, true
		case "G":
			m.scrollBody(99999)
			return m, nil, true
		}
	case responseTabHeaders:
		if len(m.headers) == 0 {
			return m, nil, false
		}
		switch msg.String() {
		case "j", "down":
			m.scrollHeaders(1)
			return m, nil, true
		case "k", "up":
			m.scrollHeaders(-1)
			return m, nil, true
		case "ctrl+d":
			m.scrollHeaders(m.height / 2)
			return m, nil, true
		case "g":
			m.headersOffset = 0
			return m, nil, true
		case "G":
			m.scrollHeaders(99999)
			return m, nil, true
		}
	}

	return m, nil, false
}

// View renders the active tab's content as a multi-line string.
func (m Model) View() tea.View {
	switch m.activeTab() {
	case responseTabBody:
		return tea.NewView(m.bodyView())
	case responseTabHeaders:
		return tea.NewView(m.headersView())
	case responseTabCookies:
		return tea.NewView(m.cookiesView())
	case responseTabTrace:
		return tea.NewView(m.traceView())
	}
	return tea.NewView("")
}

func (m Model) activeTab() responseTabID {
	return responseTabFromLabel(m.tabs.ActiveLabel())
}

// bodyView renders the body tab content as a multi-line string.
func (m Model) bodyView() string {
	if len(m.lines) == 0 {
		return lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
			Render("  send a request to see the response")
	}

	all := m.lines
	total := len(all)

	// Info header: line count on the left, scroll range on the right.
	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	left := muted.Render(fmt.Sprintf("  %d lines", total))
	right := ""

	viewLines := m.height - 1
	if total > viewLines {
		end := min(m.offset+viewLines, total)
		right = muted.Render(fmt.Sprintf("%d–%d  ", m.offset+1, end))
	}
	leftWidth := ansi.StringWidth(left)
	rightWidth := ansi.StringWidth(right)
	infoLine := left + strings.Repeat(" ", max(0, m.width-leftWidth-rightWidth)) + right

	// Line number gutter: right-aligned within the digit count of total lines.
	numDigits := len(fmt.Sprintf("%d", total))
	gutterWidth := numDigits + 2 // space + digits + space
	contentWidth := max(1, m.width-gutterWidth)
	numStyle := lipgloss.NewStyle().Foreground(theme.Current.TextDim)

	start := min(m.offset, total)
	end := min(start+viewLines, total)

	lines := make([]string, end-start)
	for i, line := range all[start:end] {
		lineNum := start + i + 1
		gutter := numStyle.Render(fmt.Sprintf(" %*d ", numDigits, lineNum))
		content := ansi.Truncate(line, contentWidth, "")
		cw := ansi.StringWidth(content)
		if cw < contentWidth {
			content += strings.Repeat(" ", contentWidth-cw)
		}
		lines[i] = gutter + content
	}

	return strings.Join(append([]string{infoLine}, lines...), "\n")
}

// headersView renders the headers tab content as a multi-line string.
func (m Model) headersView() string {
	if len(m.headers) == 0 {
		return lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
			Render("  no response headers yet")
	}

	all := m.headersContentLines()
	total := len(all)
	viewLines := max(1, m.height)

	start := min(m.headersOffset, total)
	end := min(start+viewLines, total)
	chunk := all[start:end]
	return strings.Join(chunk, "\n")
}

// sortedHeaderRows flattens and lexicographically sorts response headers into rows.
func (m Model) sortedHeaderRows() []headers.Row {
	if len(m.headers) == 0 {
		return nil
	}
	var pairs []string
	for key, values := range m.headers {
		for _, v := range values {
			pairs = append(pairs, fmt.Sprintf("%s\x00%s", key, v))
		}
	}
	sort.Strings(pairs)
	rows := make([]headers.Row, 0, len(pairs))
	for _, pair := range pairs {
		key, value, _ := strings.Cut(pair, "\x00")
		rows = append(rows, headers.Row{Name: key, Value: value})
	}
	return rows
}

// headersContentLines builds render-ready lines for the headers tab body.
func (m Model) headersContentLines() []string {
	rows := m.sortedHeaderRows()
	if len(rows) == 0 {
		return nil
	}
	return strings.Split(headers.ReadOnlyView(m.width, rows).Content, "\n")
}

// cookiesView renders the cookies tab content as a multi-line string.
func (m Model) cookiesView() string {
	return lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
		Render("  no cookies received from the server")
}

// traceView renders the trace tab content as a multi-line string.
func (m Model) traceView() string {
	if len(m.trace.Events) == 0 {
		return lipgloss.NewStyle().Foreground(theme.Current.TextMuted).
			Render("  no trace data for this request")
	}

	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	primary := lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	accent := lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
	fastBar := lipgloss.NewStyle().Foreground(theme.Current.MethodGET)
	mediumBar := lipgloss.NewStyle().Foreground(theme.Current.MethodPATCH)
	slowBar := lipgloss.NewStyle().Foreground(theme.Current.MethodDELETE)

	totalMs := m.trace.Events[len(m.trace.Events)-1].Elapsed.Milliseconds()
	if totalMs < 1 {
		totalMs = 1
	}
	stepMs := make([]int64, len(m.trace.Events))
	var prevElapsed time.Duration
	var maxStepMs int64 = 1
	for i, event := range m.trace.Events {
		delta := max(0, event.Elapsed-prevElapsed)
		ms := delta.Milliseconds()
		if ms < 1 && delta > 0 {
			ms = 1
		}
		stepMs[i] = ms
		if ms > maxStepMs {
			maxStepMs = ms
		}
		prevElapsed = event.Elapsed
	}
	lines := []string{
		muted.Render("  total: ") + primary.Render(fmt.Sprintf("%dms", totalMs)) + muted.Render("  redirects: ") + primary.Render(fmt.Sprintf("%d", len(m.trace.Redirects))),
		"",
		fastBar.Render("  timeline"),
	}

	nameCol := 24
	switch {
	case m.width >= 80:
		nameCol = 32
	case m.width <= 55:
		nameCol = 16
	}
	barMax := max(6, min(20, m.width-nameCol-16))
	for i, event := range m.trace.Events {
		eventMs := stepMs[i]
		ratio := float64(eventMs) / float64(maxStepMs)
		name := ansi.Truncate(event.Name, nameCol, "")
		dotCount := max(2, nameCol-ansi.StringWidth(name))
		barLen := int(ratio * float64(barMax))
		if eventMs > 0 && barLen == 0 {
			barLen = 1
		}
		bar := strings.Repeat("█", barLen)
		timeText := fmt.Sprintf("%5dms", eventMs)
		barStyle := fastBar
		switch {
		case eventMs >= 600:
			barStyle = slowBar
		case eventMs >= 200:
			barStyle = mediumBar
		}
		lines = append(lines,
			muted.Render("  ")+primary.Render(name)+muted.Render(" "+strings.Repeat(".", dotCount)+" "+timeText+" |")+barStyle.Render(bar),
		)
	}

	if m.trace.Proto != "" || m.trace.RemoteAddr != "" {
		lines = append(lines, "", muted.Render("  connection"))
		if m.trace.Proto != "" {
			lines = append(lines, muted.Render("  proto: ")+accent.Render(m.trace.Proto))
		}
		if m.trace.RemoteAddr != "" {
			lines = append(lines, muted.Render("  remote: ")+accent.Render(m.trace.RemoteAddr))
		}
	}

	if len(m.trace.Redirects) > 0 {
		lines = append(lines, "", muted.Render("  redirects"))
		for _, u := range m.trace.Redirects {
			lines = append(lines, muted.Render("  -> ")+primary.Render(u))
		}
	}

	return strings.Join(lines, "\n")
}

// rebuildBodyCache prepares the body lines for display.
func (m *Model) rebuildBodyCache() {
	if len(m.body) == 0 {
		m.lines = nil
		return
	}

	body := prettyBody(m.body)
	highlighted := syntaxHighlight(body, m.detectLexer())

	lines := strings.Split(highlighted, "\n")
	// Chroma always appends a trailing newline; strip the phantom empty element.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	m.lines = lines
}

// detectLexer picks a chroma lexer from the Content-Type header, falling back
// to content sniffing, then plaintext.
func (m Model) detectLexer() string {
	ct := m.headers["Content-Type"]
	if len(ct) > 0 {
		v := strings.ToLower(ct[0])
		switch {
		case strings.Contains(v, "json"):
			return "json"
		case strings.Contains(v, "xml"):
			return "xml"
		case strings.Contains(v, "html"):
			return "html"
		}
	}
	s := strings.TrimSpace(string(m.body))
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		return "json"
	}
	if strings.HasPrefix(s, "<") {
		return "xml"
	}
	return "plaintext"
}

// syntaxHighlight applies chroma terminal256/evergarden highlighting.
// Falls back to plain text if the lexer or formatter is unavailable.
func syntaxHighlight(body []byte, lexer string) string {
	var buf bytes.Buffer
	if err := chromaQuick.Highlight(&buf, string(body), lexer, "terminal256", "evergarden"); err != nil {
		return string(body)
	}
	return buf.String()
}

// prettyBody returns raw formatted for display: JSON is indented, other content unchanged.
func prettyBody(raw []byte) []byte {
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		if pretty, err := json.MarshalIndent(v, "", "  "); err == nil {
			return pretty
		}
	}
	return raw
}

// scrollBody adjusts the body tab scroll offset by the given delta.
func (m *Model) scrollBody(delta int) {
	if len(m.lines) == 0 {
		return
	}
	total := len(m.lines)
	viewLines := m.height - 1 // one line reserved for the info header
	maxOffset := max(0, total-viewLines)
	m.offset = max(0, min(m.offset+delta, maxOffset))
}

// scrollHeaders adjusts the headers tab scroll offset by the given delta.
func (m *Model) scrollHeaders(delta int) {
	all := m.headersContentLines()
	total := len(all)
	if total == 0 {
		return
	}
	viewLines := max(1, m.height)
	maxOffset := max(0, total-viewLines)
	m.headersOffset = max(0, min(m.headersOffset+delta, maxOffset))
}
