package model

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/tui/keys"
)

// activePane returns the active pane based on the current focus.
func (m Model) activePane() pane {
	switch m.focus {
	case focusResponsePane:
		return responsePane
	case focusRequestPane:
		return requestPane
	default:
		return m.lastPane
	}
}

// focusedPaneTarget returns the focused target based on the last active pane.
func (m Model) focusedPaneTarget() focusedTarget {
	if m.lastPane == responsePane {
		return focusResponsePane
	}
	return focusRequestPane
}

// focusNextMain focuses the next main pane.
func (m *Model) focusNextMain() {
	switch m.focus {
	case focusURLBar, focusGlobal:
		m.focus = focusRequestPane
		m.lastPane = requestPane
	case focusRequestPane:
		m.focus = focusResponsePane
		m.lastPane = responsePane
	case focusResponsePane:
		m.focus = focusURLBar
	}
}

// focusPrevMain focuses the previous main pane.
func (m *Model) focusPrevMain() {
	switch m.focus {
	case focusURLBar, focusGlobal:
		m.focus = focusResponsePane
		m.lastPane = responsePane
	case focusResponsePane:
		m.focus = focusRequestPane
		m.lastPane = requestPane
	case focusRequestPane:
		m.focus = focusURLBar
	}
}

// routeByFocus routes the key press to the appropriate handler based on the current focus.
func (m Model) routeByFocus(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if m.focus == focusURLBar {
		next, cmd := m.handleURLBarKeyPress(msg)
		return next, cmd, true
	}
	if m.focus == focusGlobal {
		next, cmd, handled := m.handleGlobalFocusKeyPress(msg)
		return next, cmd, handled
	}
	return m, nil, false
}

// handleFocusedPaneKeyPress handles the key press for the focused pane.
func (m Model) handleFocusedPaneKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	if m.focus == focusRequestPane {
		if next, cmd, handled := m.tab().requestPane.Update(msg); handled {
			m.tab().requestPane = next
			return m, cmd, true
		}
	}
	if m.focus == focusResponsePane {
		if next, cmd, handled := m.tab().responsePane.Update(msg); handled {
			m.tab().responsePane = next
			return m, cmd, true
		}
	}
	return m, nil, false
}

// handleGlobalFocusKeyPress handles the key press for the global focus.
func (m Model) handleGlobalFocusKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd, bool) {
	switch msg.String() {
	case "tab":
		m.focus = focusURLBar
		m.syncChildState()
		return m, nil, true
	case "shift+tab":
		m.focus = focusResponsePane
		m.lastPane = responsePane
		m.syncChildState()
		return m, nil, true
	default:
		return m, nil, false
	}
}

// handleURLBarKeyPress handles the key press for the URL bar.
func (m Model) handleURLBarKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.focus = focusRequestPane
		m.lastPane = requestPane
		m.syncChildState()
		return m, nil
	case "esc":
		m.focus = m.focusedPaneTarget()
		m.syncChildState()
		return m, nil
	case "tab":
		m.focusNextMain()
		m.syncChildState()
		return m, nil
	case "shift+tab":
		m.focusPrevMain()
		m.syncChildState()
		return m, nil
	case "ctrl+l":
		m.tab().urlbar.SetURL("")
		return m, nil
	default:
		var cmd tea.Cmd
		m.tab().urlbar, cmd = m.tab().urlbar.Update(msg)
		m.tab().requestPane.SetURL(m.tab().urlbar.URL())
		m.tab().refreshTitle()
		return m, cmd
	}
}

// handleNavigationKey handles the key press for the navigation.
func (m Model) handleNavigationKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Default.URLFocus):
		m.lastPane = m.activePane()
		m.focus = focusURLBar
		m.syncChildState()
	case msg.String() == "tab":
		m.focusNextMain()
		m.syncChildState()
	case msg.String() == "shift+tab":
		m.focusPrevMain()
		m.syncChildState()
	}
	return m, nil
}
