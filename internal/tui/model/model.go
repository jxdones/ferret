package model

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	collectiondata "github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
	"github.com/jxdones/ferret/internal/tui/components/collection"
	"github.com/jxdones/ferret/internal/tui/components/methodpicker"
	"github.com/jxdones/ferret/internal/tui/components/requestpane"
	"github.com/jxdones/ferret/internal/tui/components/responsepane"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/components/titlebar"
	"github.com/jxdones/ferret/internal/tui/components/urlbar"
	"github.com/jxdones/ferret/internal/tui/components/workspacepicker"
)

// StartOptions configures TUI startup from the CLI.
type StartOptions struct {
	Dir                 string
	EnvName             string
	ImplicitDirectory   bool // true when -d was not passed (flag default ".")
	ConfigHasWorkspaces bool // true when ~/.ferret/config.yaml lists at least one workspace
	// WorkspaceName is the config workspace name shown in the title bar when Dir
	// comes from the first entry in config (implicit -d); empty otherwise.
	WorkspaceName string
}

// requestTab holds the per-tab request/response context.
type requestTab struct {
	id             int
	title          string
	requestName    string // non-empty when loaded from a collection entry
	collectionRoot string
	urlbar         urlbar.Model
	requestPane    requestpane.Model
	responsePane   responsepane.Model
	cancel         context.CancelFunc
	lastResponse   *statusbar.Response

	isLoading bool
}

// Model is the main application model for the ferret TUI.
type Model struct {
	// dimensions
	width  int
	height int

	// components
	titlebar        titlebar.Model
	collection      collection.Model
	workspacePicker workspacepicker.Model
	statusbar       statusbar.Model
	methods         methodpicker.Model

	// tabs
	tabs      []requestTab
	activeTab int

	// data references
	workspaceRoot  string
	workspaceName  string // config name for title bar; empty if none
	collectionDirs []string
	nextTabID      int
	env            *env.Env
	// envName is the active file env stem (e.g. "dev") or empty for shell-only.
	envName string

	// states
	focus        focusedTarget
	lastPane     pane
	activeModal  modalKind
	helpExpanded bool
}

// Start runs the TUI.
func Start(opts StartOptions) error {
	m, err := New(opts)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m).Run()
	return err
}

// New creates the root TUI model from StartOptions.
func New(opts StartOptions) (Model, error) {
	workspaceRoot, err := resolveWorkspaceRoot(opts.Dir)
	if err != nil {
		return Model{}, err
	}

	scratchNoWorkspace := opts.ImplicitDirectory && !opts.ConfigHasWorkspaces
	cliEnv := opts.EnvName
	if scratchNoWorkspace {
		cliEnv = ""
	}

	if scratchNoWorkspace {
		e := env.NewFromShell()
		u := urlbar.New()
		u.SetMethod("GET")
		rp := requestpane.New()
		m := Model{
			titlebar: titlebar.New(),
			tabs: []requestTab{{
				id:             1,
				title:          "new request",
				collectionRoot: "",
				urlbar:         u,
				requestPane:    rp,
				responsePane:   responsepane.New(),
			}},
			nextTabID:       2,
			activeTab:       0,
			collection:      collection.New(),
			workspacePicker: workspacepicker.New(),
			statusbar:       statusbar.New(),
			methods:         methodpicker.New(),
			workspaceRoot:   workspaceRoot,
			workspaceName:   "",
			collectionDirs:  nil,
			env:             e,
			envName:         "",
			focus:           focusRequestPane,
			lastPane:        requestPane,
		}
		m.titlebar.SetWorkspace("no workspace")
		m.titlebar.SetCollection("")
		m.titlebar.SetEnv("")
		m.syncChildStateWithLayout()
		return m, nil
	}

	collectionDirs, err := collectiondata.DiscoverCollections(workspaceRoot)
	if err != nil {
		return Model{}, fmt.Errorf("model: discover collections in %s: %w", workspaceRoot, err)
	}
	e, resolvedEnv, err := env.ResolveStartEnvFromAll(collectionDirs, cliEnv)
	if err != nil {
		return Model{}, fmt.Errorf("model: environment: %w", err)
	}

	u := urlbar.New()
	u.SetMethod("GET")
	rp := requestpane.New()
	m := Model{
		titlebar: titlebar.New(),
		tabs: []requestTab{{
			id:           1,
			title:        "new request",
			urlbar:       u,
			requestPane:  rp,
			responsePane: responsepane.New(),
		}},
		activeTab:       0,
		nextTabID:       2,
		collection:      collection.New(),
		workspacePicker: workspacepicker.New(),
		statusbar:       statusbar.New(),
		methods:         methodpicker.New(),
		workspaceRoot:   workspaceRoot,
		workspaceName:   opts.WorkspaceName,
		collectionDirs:  collectionDirs,
		env:             e,
		envName:         resolvedEnv,
		focus:           focusRequestPane,
		lastPane:        requestPane,
	}

	m.titlebar.SetWorkspace(m.workspaceName)
	m.titlebar.SetEnv(resolvedEnv)
	m.syncChildStateWithLayout()
	return m, nil
}

// tab returns a pointer to the active request tab.
func (m Model) tab() *requestTab {
	return &m.tabs[m.activeTab]
}

// newTab opens a new empty request tab and focuses its URL bar.
func (m *Model) newTab() tea.Cmd {
	// increment the next tab ID and use it for the new tab
	m.nextTabID++
	u := urlbar.New()
	u.SetMethod("GET")
	m.tabs = append(m.tabs, requestTab{
		id:           m.nextTabID,
		title:        "new request",
		urlbar:       u,
		requestPane:  requestpane.New(),
		responsePane: responsepane.New(),
	})
	m.activeTab = len(m.tabs) - 1
	m.focus = focusURLBar
	m.syncChildStateWithLayout()
	return m.syncStatusbarToActiveTab()
}

// closeTab removes the active tab. Does nothing if only one tab is open.
func (m *Model) closeTab() tea.Cmd {
	if len(m.tabs) <= 1 {
		return nil
	}
	if t := m.tabs[m.activeTab]; t.cancel != nil {
		t.cancel()
	}
	m.tabs = append(m.tabs[:m.activeTab], m.tabs[m.activeTab+1:]...)
	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}
	m.syncChildStateWithLayout()
	return m.syncStatusbarToActiveTab()
}

// switchTab switches to the tab at index i, wrapping around.
func (m *Model) switchTab(i int) tea.Cmd {
	n := len(m.tabs)
	m.activeTab = ((i % n) + n) % n
	m.syncChildStateWithLayout()
	return m.syncStatusbarToActiveTab()
}

// resolveWorkspaceRoot resolves the workspace root from the given directory.
func resolveWorkspaceRoot(dir string) (string, error) {
	inputDir := dir
	if inputDir == "" || inputDir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("model: get current working directory: %w", err)
		}
		return wd, nil
	}
	abs, err := filepath.Abs(filepath.Clean(inputDir))
	if err != nil {
		return "", fmt.Errorf("model: resolve directory %q: %w", inputDir, err)
	}
	return abs, nil
}

// syncStatusbarToActiveTab synchronizes the status bar to the active tab.
func (m *Model) syncStatusbarToActiveTab() tea.Cmd {
	tab := m.tab()
	if tab.isLoading {
		return m.statusbar.SetSending()
	}
	if tab.lastResponse != nil {
		m.statusbar.SetResponse(*tab.lastResponse)
		return nil
	}
	m.statusbar.SetIdle()
	return nil
}
