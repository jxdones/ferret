package statusbar

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/common"
	"github.com/jxdones/ferret/internal/tui/theme"
)

// Kind is the status severity level (info, success, warning, error).
type Kind int

const (
	Info Kind = iota
	Success
	Warning
	Error
)

const defaultStatusMessage = " Ready"

// ExpiredMsg is emitted when a flash status TTL elapses.
type ExpiredMsg struct {
	Seq int
}

// Response holds the metadata from a completed HTTP response.
type Response struct {
	StatusCode int
	StatusText string
	Duration   time.Duration
	Size       int64
	Format     string
}

// Model is the status bar component rendered between the content area and
// the shortcuts bar. It shows the current state and, after a request completes,
// response metadata (status, duration, size, format).
type Model struct {
	text string
	kind Kind
	seq  int

	response *Response
	spinner  spinner.Model
	spinning bool
	width    int
}

// New initializes a status bar in the idle "Ready" state.
func New() Model {
	return Model{
		text:    defaultStatusMessage,
		kind:    Info,
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot)),
	}
}

// SetWidth sets the width of the status bar.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// SetStatus sets a sticky status message and kind.
func (m *Model) SetStatus(text string, kind Kind) {
	m.text = strings.TrimSpace(text)
	if m.text == "" {
		m.text = defaultStatusMessage
	}
	m.kind = kind
	m.seq++
}

// SetStatusWithTTL sets the status and schedules expiration when ttl > 0
// seq is incremented per status update so each timer can be matched to the status
// generation that created it. Older timers are safely ignored on arrival.
func (m *Model) SetStatusWithTTL(text string, kind Kind, ttl time.Duration) tea.Cmd {
	m.text = strings.TrimSpace(text)
	if m.text == "" {
		m.text = defaultStatusMessage
	}
	m.kind = kind
	m.seq++

	seq := m.seq
	if ttl <= 0 {
		return nil
	}
	return tea.Tick(ttl, func(time.Time) tea.Msg {
		return ExpiredMsg{Seq: seq}
	})
}

// HandleExpired clears a flashed status message if the timer belongs to the
// current sequence.
func (m *Model) HandleExpired(msg ExpiredMsg) {
	if msg.Seq != m.seq {
		return
	}
	m.text = defaultStatusMessage
	m.kind = Info
}

// SetSending transitions to the sending state and starts the spinner.
// Clears any previous response metadata so the right side is blank while in
// flight.
func (m *Model) SetSending() tea.Cmd {
	m.spinning = true
	m.response = nil
	m.SetStatus("Making request", Info)
	return m.spinner.Tick
}

// SetResponse stores response metadata and returns the left side to Ready.
// The metadata persists on the right side until the next request starts.
func (m *Model) SetResponse(resp Response) {
	m.spinning = false
	m.response = &resp
	m.text = defaultStatusMessage
	m.kind = Info
}

// SetError flashes an error message on the left for a few seconds with no
// response metadata.
// Network-level failures where there is no HTTP response to display are handled
// here.
func (m *Model) SetError(text string) tea.Cmd {
	m.spinning = false
	m.response = nil
	return m.SetStatusWithTTL(text, Error, 5*time.Second)
}

// SetIdle returns to the idle "Ready" state with no response metadata.
func (m *Model) SetIdle() {
	m.spinning = false
	m.response = nil
	m.SetStatus(defaultStatusMessage, Info)
}

// Update handles spinner tick messages. Call from the root model's Update
// function.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.spinning {
		return nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return cmd
}

// View renders the status bar as a bordered single line padded to m.width
func (m *Model) View() tea.View {
	innerWidth := max(1, m.width-2) // -2 for padding(0,1)

	left := m.renderLeft()
	right := m.renderRight()

	leftWidth := ansi.StringWidth(left)
	rightWidth := ansi.StringWidth(right)
	gap := max(0, innerWidth-leftWidth-rightWidth)

	content := left + strings.Repeat(" ", gap) + right

	return tea.NewView(lipgloss.NewStyle().
		Width(m.width).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Current.DividerBorder).
		Padding(0, 1).
		Render(content))

}

// renderLeft renders the left side of the status bar with the current status
// text and kind.
func (m Model) renderLeft() string {
	style := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	switch m.kind {
	case Info:
		style = lipgloss.NewStyle().Foreground(theme.Current.StatusInfo).Bold(true)
	case Success:
		style = lipgloss.NewStyle().Foreground(theme.Current.StatusSuccess).Bold(true)
	case Warning:
		style = lipgloss.NewStyle().Foreground(theme.Current.StatusWarning).Bold(true)
	case Error:
		style = lipgloss.NewStyle().Foreground(theme.Current.StatusError).Bold(true)
	}

	text := m.text
	if m.spinning {
		text = " " + m.spinner.View() + " " + text
	}
	return style.Render(text)
}

// renderRight renders the right side of the status bar with the response
// metadata.
func (m Model) renderRight() string {
	if m.spinning {
		cancel := lipgloss.NewStyle().Foreground(theme.Current.RequestCancel)
		return cancel.Render("^x to cancel")
	}
	if m.response == nil {
		return ""
	}

	statusText := strings.TrimSpace(m.response.StatusText)
	codePrefix := fmt.Sprintf("%d ", m.response.StatusCode)
	statusText = strings.TrimPrefix(statusText, codePrefix)
	if statusText == "" {
		statusText = http.StatusText(m.response.StatusCode)
	}

	statusCodeColor := theme.Current.StatusCodeOK
	if m.response.StatusCode >= 400 {
		statusCodeColor = theme.Current.StatusCodeError
	}
	status := lipgloss.NewStyle().Foreground(statusCodeColor).Bold(true).
		Render(fmt.Sprintf("%d %s", m.response.StatusCode, statusText))

	muted := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)
	meta := []string{
		formatDuration(m.response.Duration),
		common.FormatSize(m.response.Size),
	}
	if m.response.Format != "" {
		meta = append(meta, m.response.Format)
	}

	return status + " " + muted.Render(strings.Join(meta, " "))
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
