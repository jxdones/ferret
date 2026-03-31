package model

import (
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	collectiondata "github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/keys"
)

// handleCollectionKey handles the key press for the collection.
func (m Model) handleCollectionKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Default.Collection):
		m.collection.Reset()
		if entries, err := collectiondata.LoadEntries(m.tab().collectionRoot); err == nil {
			m.collection.Load(entries)
		}
		m.activeModal = modalCollection
		m.lastPane = m.activePane()
		m.focus = focusModalCollection
		m.syncChildState()
	case key.Matches(msg, keys.Default.CollectionCycle):
		cmd := m.cycleCollection()
		return m, cmd
	case key.Matches(msg, keys.Default.WorkspacePick):
		m.workspacePicker.Reset()
		m.workspacePicker.Load(m.collectionDirs)
		m.workspacePicker.SetActive(m.tab().collectionRoot)
		m.activeModal = modalWorkspace
		m.lastPane = m.activePane()
		m.focus = focusModalWorkspace
		m.syncChildState()
	}
	return m, nil
}

// handleCollectionModalKey handles the key press for the collection modal.
func (m Model) handleCollectionModalKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "ctrl+p":
		m.collection.MoveCursor(-1)
	case "down", "ctrl+n":
		m.collection.MoveCursor(1)
	case "enter":
		if entry, ok := m.collection.Selected(); ok {
			m.tab().urlbar.SetMethod(entry.Request.Method)
			m.tab().urlbar.SetURL(entry.Request.URL)
			m.tab().requestPane.SetURL(entry.Request.URL)
			m.tab().requestPane.SetHeaders(entry.Request.Headers)
			m.tab().requestPane.SetBody(entry.Request.Body)
			m.tab().requestPane.ResetBodyFocus()
			m.tab().responsePane.Reset()
			m.tab().requestName = entryDisplayTitle(entry)
			m.tab().refreshTitle()
			m.titlebar.SetEntry(entryDisplayTitle(entry))
			m.statusbar.SetIdle()
		}
		m.activeModal = modalNone
		m.focus = m.focusedPaneTarget()
		m.syncChildState()
	default:
		var cmd tea.Cmd
		m.collection, cmd = m.collection.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleWorkspaceModalKey handles the key press for the workspace modal.
func (m Model) handleWorkspaceModalKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "ctrl+p":
		m.workspacePicker.MoveCursor(-1)
	case "down", "ctrl+n":
		m.workspacePicker.MoveCursor(1)
	case "enter":
		var cmd tea.Cmd
		if path, ok := m.workspacePicker.Selected(); ok {
			cmd = m.switchToCollection(path)
		}
		m.activeModal = modalNone
		m.focus = m.focusedPaneTarget()
		m.syncChildState()
		return m, cmd
	default:
		var cmd tea.Cmd
		m.workspacePicker, cmd = m.workspacePicker.Update(msg)
		return m, cmd
	}
	return m, nil
}

// cycleCollection cycles through the collection directories for the active tab.
func (m *Model) cycleCollection() tea.Cmd {
	if len(m.collectionDirs) == 0 {
		return m.statusbar.SetStatusWithTTL("no collections", statusbar.Info, 2*time.Second)
	}
	if len(m.collectionDirs) <= 1 {
		return m.statusbar.SetStatusWithTTL("single collection workspace", statusbar.Info, 2*time.Second)
	}
	current := m.tab().collectionRoot
	currentIdx := -1
	for i, d := range m.collectionDirs {
		if d == current {
			currentIdx = i
			break
		}
	}
	next := currentIdx + 1
	if next >= len(m.collectionDirs) {
		next = 0
	}
	return m.activateCollectionAtIndex(next)
}

// activateCollectionAtIndex activates the collection at the given index for the active tab.
func (m *Model) activateCollectionAtIndex(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.collectionDirs) {
		return nil
	}
	m.tab().collectionRoot = m.collectionDirs[idx]
	m.titlebar.SetWorkspace(m.workspaceName)
	m.titlebar.SetCollection(filepath.Base(m.tab().collectionRoot))
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

	if m.envName == "" {
		m.env = env.NewFromShell()
		return m.statusbar.SetStatusWithTTL("collection -> "+filepath.Base(m.tab().collectionRoot), statusbar.Info, 2*time.Second)
	}

	loaded, err := env.Load(m.tab().collectionRoot, m.envName)
	if err != nil {
		m.env = env.NewFromShell()
		m.env.Session = m.copySessionVars()
		m.titlebar.SetEnv("")
		return m.statusbar.SetStatusWithTTL("collection -> "+filepath.Base(m.tab().collectionRoot)+" (env unavailable, shell only)", statusbar.Warning, 3*time.Second)
	}
	loaded.Session = m.copySessionVars()
	m.env = loaded
	return m.statusbar.SetStatusWithTTL("collection -> "+filepath.Base(m.tab().collectionRoot), statusbar.Info, 2*time.Second)
}

// switchToCollection switches to the collection at the given path.
func (m *Model) switchToCollection(path string) tea.Cmd {
	for i, d := range m.collectionDirs {
		if d == path {
			return m.activateCollectionAtIndex(i)
		}
	}
	return m.statusbar.SetError("unknown collection: " + path)
}

// entryDisplayTitle returns the display title for a collection entry.
func entryDisplayTitle(e collectiondata.Entry) string {
	name := strings.TrimSpace(e.Request.Name)
	if name != "" {
		return name
	}
	return e.Path
}
