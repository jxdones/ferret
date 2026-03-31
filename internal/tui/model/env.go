package model

import (
	"maps"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/jxdones/ferret/internal/env"
	"github.com/jxdones/ferret/internal/tui/components/statusbar"
	"github.com/jxdones/ferret/internal/tui/keys"
)

// handleEnvKey handles the key press for the environment.
func (m Model) handleEnvKey(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Default.EnvCycle):
		cmd := m.cycleEnv()
		return m, cmd
	}
	return m, nil
}

// cycleEnv cycles through the environment variables across all workspace collections.
func (m *Model) cycleEnv() tea.Cmd {
	if len(m.collectionDirs) == 0 {
		m.env = env.NewFromShell()
		m.env.Session = m.copySessionVars()
		m.envName = ""
		m.titlebar.SetEnv("")
		return m.statusbar.SetStatusWithTTL("env -> shell only", statusbar.Info, 2*time.Second)
	}

	names, err := env.ListNamesFromAll(m.collectionDirs)
	if err != nil {
		return m.statusbar.SetError(err.Error())
	}

	options := append([]string{""}, names...)
	next := nextEnvOption(options, m.envName)

	if next == "" {
		m.env = env.NewFromShell()
		m.env.Session = m.copySessionVars()
		m.envName = ""
		m.titlebar.SetEnv("")
		return m.statusbar.SetStatusWithTTL("env -> shell only", statusbar.Info, 2*time.Second)
	}

	loaded, collisions, err := env.LoadMerged(m.collectionDirs, next)
	if err != nil {
		return m.statusbar.SetError(err.Error())
	}
	loaded.Session = m.copySessionVars()
	m.env = loaded
	m.envName = next
	m.titlebar.SetEnv(next)
	msg := "env -> " + next
	if collisions {
		msg += " (key collisions)"
	}
	return m.statusbar.SetStatusWithTTL(msg, statusbar.Info, 2*time.Second)
}

// nextEnvOption returns the next environment variable in the order of the options.
func nextEnvOption(options []string, current string) string {
	if len(options) == 0 {
		return ""
	}

	currentIdx := -1
	for i, name := range options {
		if strings.EqualFold(name, current) {
			currentIdx = i
			break
		}
	}

	if currentIdx < 0 {
		if len(options) > 1 {
			return options[1]
		}
		return options[0]
	}

	next := currentIdx + 1
	if next >= len(options) {
		next = 0
	}
	return options[next]
}

// copySessionVars copies the session variables from the environment.
func (m Model) copySessionVars() map[string]string {
	if m.env == nil || len(m.env.Session) == 0 {
		return nil
	}
	out := make(map[string]string, len(m.env.Session))
	maps.Copy(out, m.env.Session)
	return out
}
