package model

import (
	"path/filepath"

	"github.com/jxdones/ferret/internal/tui/modal"
)

// syncChildState propagates focus and modal child state during transitions.
func (m *Model) syncChildState() {
	m.applyFocus()
	m.syncModalState()
	m.syncTitlebarCollection()
}

// syncTitlebarCollection updates the titlebar collection label from the active tab.
func (m *Model) syncTitlebarCollection() {
	root := m.tab().collectionRoot
	if root == "" {
		m.titlebar.SetCollection("")
		return
	}
	m.titlebar.SetCollection(filepath.Base(root))
}

// syncChildStateWithLayout also recomputes child sizing/layout before syncing.
func (m *Model) syncChildStateWithLayout() {
	m.applySize()
	m.syncChildState()
}

// syncModalState keeps modal child dimensions in sync with current width.
func (m *Model) syncModalState() {
	switch m.activeModal {
	case modalCollection:
		m.collection.SetSize(modal.InnerWidth(m.modalOuterWidth()))
	case modalWorkspace:
		m.workspacePicker.SetSize(modal.InnerWidth(m.modalOuterWidth()))
	case modalMethod:
		m.methods.SetSize(modal.InnerWidth(m.modalOuterWidth()))
	}
}

// applySize propagates the terminal dimensions to all components.
func (m *Model) applySize() {
	mid := m.width / 2
	right := m.width - mid - 1
	contentHeight := m.contentHeight()

	m.titlebar.SetSize(m.width)
	m.tab().urlbar.SetSize(m.width)
	m.tab().requestPane.SetSize(mid, contentHeight)
	m.tab().responsePane.SetSize(right, contentHeight)
	m.statusbar.SetWidth(m.width)
}

// applyFocus propagates focus to panes and the URL bar.
func (m *Model) applyFocus() {
	m.tab().requestPane.SetFocused(m.focus == focusRequestPane)
	m.tab().responsePane.SetFocused(m.focus == focusResponsePane)
	m.tab().urlbar.SetFocused(m.focus == focusURLBar)
}

// contentHeight returns the number of lines available for pane content.
func (m Model) contentHeight() int {
	return max(1, m.height-10-m.optionsHeight())
}

// modalOuterWidth computes the modal outer width clamped to screen.
func (m Model) modalOuterWidth() int {
	w := m.width * 60 / 100
	return max(40, min(w, 70))
}
