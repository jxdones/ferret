package model

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/keys"
)

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusbar.ExpiredMsg:
		m.statusbar.HandleExpired(msg)
		return m, nil
	case RequestStartedMsg:
		next, cmd := m.onRequestStarted()
		return next, cmd
	case RequestFinishedMsg:
		next, cmd := m.onRequestFinished(msg)
		return next, cmd
	case RequestFailedMsg:
		next, cmd := m.onRequestFailed(msg)
		return next, cmd
	case tea.WindowSizeMsg:
		next, cmd := m.handleWindowSize(msg)
		return next, batch(cmd, next.statusbar.Update(msg))
	case tea.KeyPressMsg:
		next, cmd := m.handleKeyPress(msg)
		return next, batch(cmd, next.statusbar.Update(msg))
	case tea.PasteMsg:
		if m.focus == focusURLBar {
			m.tab().urlbar, _ = m.tab().urlbar.Update(msg)
			m.tab().requestPane.SetURL(m.tab().urlbar.URL())
			m.tab().refreshTitle()
		} else if m.focus == focusRequestPane && m.tab().requestPane.BodyFocused() {
			next, _, _ := m.tab().requestPane.Update(msg)
			m.tab().requestPane = next
		}
		return m, nil
	default:
		return m, m.statusbar.Update(msg)
	}
}

// handleWindowSize handles the window size message.
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.syncChildStateWithLayout()
	return m, nil
}

// handleKeyPress handles Global, Modal, and Focused pane key presses.
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if m.activeModal != modalNone {
		return m.handleModalKey(msg)
	}
	if next, cmd, handled := m.routeByFocus(msg); handled {
		return next, cmd
	}
	if next, cmd, handled := m.handleFocusedPaneKeyPress(msg); handled {
		return next, cmd
	}
	return m.handleGlobalKeyPress(msg)
}

// handleModalKey handles the modal key press message.
func (m Model) handleModalKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.activeModal = modalNone
		m.focus = m.focusedPaneTarget()
		m.syncChildState()
		return m, nil
	}
	switch m.activeModal {
	case modalCollection:
		return m.handleCollectionModalKey(msg)
	case modalWorkspace:
		return m.handleWorkspaceModalKey(msg)
	case modalMethod:
		return m.handleMethodModalKey(msg)
	}
	return m, nil
}

// handleGlobalKeyPress handles global key presses.
func (m Model) handleGlobalKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case msg.String() == "ctrl+c", msg.String() == "q":
		return m, tea.Quit
	case msg.String() == "esc":
		m.helpExpanded = false
		m.focus = focusGlobal
		m.syncChildStateWithLayout()
	case key.Matches(msg, keys.Default.Help):
		m.helpExpanded = !m.helpExpanded
		m.syncChildStateWithLayout()
	case key.Matches(msg, keys.Default.NextTab):
		m.switchTab(m.activeTab + 1)
	case key.Matches(msg, keys.Default.PrevTab):
		m.switchTab(m.activeTab - 1)
	case key.Matches(msg, keys.Default.NewTab):
		m.newTab()
	case key.Matches(msg, keys.Default.CloseTab):
		m.closeTab()
	case key.Matches(msg, keys.Default.URLFocus),
		msg.String() == "tab",
		msg.String() == "shift+tab":
		return m.handleNavigationKey(msg)
	case key.Matches(msg, keys.Default.Send),
		key.Matches(msg, keys.Default.NewRequest),
		key.Matches(msg, keys.Default.MethodCycle),
		key.Matches(msg, keys.Default.MethodPicker):
		return m.handleRequestKey(msg)
	case key.Matches(msg, keys.Default.Collection),
		key.Matches(msg, keys.Default.CollectionCycle),
		key.Matches(msg, keys.Default.WorkspacePick):
		return m.handleCollectionKey(msg)
	case key.Matches(msg, keys.Default.EnvCycle):
		return m.handleEnvKey(msg)
	}
	return m, nil
}

// batch batches commands and returns nil if all are nil.
func batch(cmds ...tea.Cmd) tea.Cmd {
	filtered := make([]tea.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd != nil {
			filtered = append(filtered, cmd)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return tea.Batch(filtered...)
}
