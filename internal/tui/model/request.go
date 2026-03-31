package model

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	collectiondata "github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
	"github.com/jxdones/ferret/internal/exec"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/keys"
)

const (
	formatJSON = "JSON"
	formatXML  = "XML"
	formatHTML = "HTML"
)

type RequestStartedMsg struct{}

type RequestFinishedMsg struct {
	TabID          int
	Response       statusbar.Response
	Body           []byte
	Headers        map[string][]string
	Trace          exec.Trace
	ResponseTooBig bool
	ResponseSize   int64
}

type RequestFailedMsg struct {
	Error error
}

// handleRequestKey handles the key press for the request.
func (m Model) handleRequestKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Default.Send):
		if strings.TrimSpace(m.tab().urlbar.URL()) == "" {
			return m, m.statusbar.SetError("url is required")
		}
		return m, sendRequestCmd(m.buildRequest(), m.env, m.tab().id)
	case key.Matches(msg, keys.Default.NewRequest):
		cmd := m.startNewRequest()
		return m, cmd
	case key.Matches(msg, keys.Default.MethodCycle):
		m.tab().urlbar.SetMethod(nextMethod(m.tab().urlbar.Method()))
		m.tab().refreshTitle()
	case key.Matches(msg, keys.Default.MethodPicker):
		m.methods.SetActive(m.tab().urlbar.Method())
		m.activeModal = modalMethod
		m.lastPane = m.activePane()
		m.focus = focusModalMethod
		m.syncChildState()
	}
	return m, nil
}

// handleMethodModalKey handles the key press for the method modal.
func (m Model) handleMethodModalKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.methods.MoveCursor(-1)
	case "down", "j":
		m.methods.MoveCursor(1)
	case "enter":
		if sel := m.methods.Selected(); sel != "" {
			m.tab().urlbar.SetMethod(strings.ToUpper(sel))
			m.tab().refreshTitle()
		}
		m.activeModal = modalNone
		m.focus = m.focusedPaneTarget()
		m.syncChildState()
	}
	return m, nil
}

// onRequestStarted handles the request started event.
func (m Model) onRequestStarted() (Model, tea.Cmd) {
	m.activeModal = modalNone
	m.focus = m.focusedPaneTarget()
	m.syncChildState()
	return m, m.statusbar.SetSending()
}

// onRequestFinished handles the request finished event.
func (m Model) onRequestFinished(msg RequestFinishedMsg) (Model, tea.Cmd) {
	m.statusbar.SetResponse(msg.Response)
	for i := range m.tabs {
		if m.tabs[i].id == msg.TabID {
			m.tabs[i].responsePane.SetResponse(
				msg.Body, msg.Headers, msg.Trace, msg.ResponseTooBig, msg.ResponseSize,
			)
			if i == m.activeTab {
				m.focus = focusResponsePane
				m.lastPane = responsePane
				m.syncChildState()
			}
			break
		}
	}
	return m, nil
}

// onRequestFailed handles the request failed event.
func (m Model) onRequestFailed(msg RequestFailedMsg) (Model, tea.Cmd) {
	return m, m.statusbar.SetError(msg.Error.Error())
}

// startNewRequest starts a new request.
func (m *Model) startNewRequest() tea.Cmd {
	m.titlebar.SetEntry("")
	m.tab().urlbar.SetMethod("GET")
	m.tab().urlbar.SetURL("")
	m.tab().requestPane.SetURL("")
	m.tab().requestPane.SetHeaders(nil)
	m.tab().requestPane.SetBody("")
	m.tab().requestPane.ResetBodyFocus()
	m.tab().responsePane.Reset()
	m.tab().requestName = ""
	m.tab().title = "new request"
	m.statusbar.SetIdle()
	m.lastPane = requestPane
	m.focus = focusURLBar
	m.syncChildState()
	return m.statusbar.SetStatusWithTTL("new request", statusbar.Info, 2*time.Second)
}

// sendRequestCmd sends a command to send a request and returns the appropriate message based on the result.
func sendRequestCmd(req collectiondata.Request, e *env.Env, tabIndex int) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return RequestStartedMsg{} },
		func() tea.Msg {
			result, err := exec.Execute(context.Background(), req, e)
			if err != nil {
				return RequestFailedMsg{Error: err}
			}
			return RequestFinishedMsg{
				TabID: tabIndex,
				Response: statusbar.Response{
					StatusCode: result.Status,
					StatusText: result.StatusText,
					Duration:   result.Duration,
					Size:       int64(result.Size),
					Format:     detectFormatFromHeaders(result.Headers),
				},
				Body:           result.Body,
				Headers:        result.Headers,
				Trace:          result.Trace,
				ResponseTooBig: result.ResponseTooBig,
				ResponseSize:   result.ResponseSize,
			}
		},
	)
}

// buildRequest builds a request from the current URL bar and request pane.
func (m Model) buildRequest() collectiondata.Request {
	return collectiondata.Request{
		Method:  m.tab().urlbar.Method(),
		URL:     m.tab().urlbar.URL(),
		Headers: m.tab().requestPane.Headers(),
		Body:    m.tab().requestPane.Body(),
	}
}

// nextMethod returns the next method in the order of GET, POST, PUT, PATCH, DELETE.
func nextMethod(method string) string {
	m := strings.ToUpper(strings.TrimSpace(method))
	order := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for i, v := range order {
		if v == m {
			next := i + 1
			if next >= len(order) {
				next = 0
			}
			return order[next]
		}
	}
	return "GET"
}

// titleFromMethodAndURL builds a tab title from the method and URL.
func titleFromMethodAndURL(method, rawURL string) string {
	if rawURL == "" {
		return "new request"
	}
	return method + " " + clampTabTitle(rawURL)
}

// refreshTitle rebuilds the tab title from the current method and either the
// loaded request name (if present) or the typed URL.
func (t *requestTab) refreshTitle() {
	method := t.urlbar.Method()
	if t.requestName != "" {
		t.title = method + " " + t.requestName
	} else {
		t.title = titleFromMethodAndURL(method, t.urlbar.URL())
	}
}

// detectFormatFromHeaders detects the format of the response from the Content-Type header.
func detectFormatFromHeaders(headers map[string][]string) string {
	ct := headers["Content-Type"]
	if len(ct) == 0 {
		return ""
	}
	v := strings.ToLower(ct[0])
	switch {
	case strings.Contains(v, "json"):
		return formatJSON
	case strings.Contains(v, "xml"):
		return formatXML
	case strings.Contains(v, "html"):
		return formatHTML
	}
	return ""
}

// clampTabTitle truncates a title to 10 characters and adds an ellipsis if it is longer.
func clampTabTitle(title string) string {
	if len(title) > 10 {
		return title[:10] + "…"
	}
	return title
}
