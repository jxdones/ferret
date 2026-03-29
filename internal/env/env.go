// Package env loads collection environment YAML and merges it with the process
// environment. The TUI uses ResolveStartEnv at startup; ferret run uses Load
// with an explicit name.
package env

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Env holds environment variables for a collection from three sources with
// decreasing priority:
//
//   - Shell: a snapshot of the process environment from os.Environ at the time
//     the Env value was built (see NewFromShell and Load). Has highest priority.
//     The full process environment is stored so {{VAR}} resolution matches what
//     you exported in your shell; ferret does not print or list these vars in
//     the UI.
//   - Session: Environment variables set in the current session. Resets on
//     exit.
//   - File: Environment variables set in the collection file. Has lowest
//     priority.
//     A good place to store base URLs, and default values.
type Env struct {
	Shell   map[string]string
	Session map[string]string
	File    map[string]string
}

// Get returns the value of the environment variable for the given key.
func (e *Env) Get(key string) (string, bool) {
	if value, ok := e.Shell[key]; ok {
		return value, true
	}
	if value, ok := e.Session[key]; ok {
		return value, true
	}
	if value, ok := e.File[key]; ok {
		return value, true
	}
	return "", false
}

// Set writes a value into the session layer.
func (e *Env) Set(key, value string) {
	if e.Session == nil {
		e.Session = make(map[string]string)
	}
	e.Session[key] = value
}

// ResolveStartEnv selects which environments/<name>.yaml to apply for a
// collection directory when the TUI starts.
//
// Rules:
//   - If name is non-empty: load that file only (same as --env / -e on the CLI).
//   - If name is empty and there is at least one *.yaml under environments/
//     (including nested paths, see ListNames): load the lexicographically first
//     name and return it as the second value.
//   - If name is empty and there are no such files: return NewFromShell() and
//     an empty name (title bar shows "shell only"; still uses OS env for vars).
func ResolveStartEnv(dir, name string) (*Env, string, error) {
	if name != "" {
		e, err := Load(dir, name)
		if err != nil {
			return nil, "", err
		}
		return e, name, nil
	}
	names, err := ListNames(dir)
	if err != nil {
		return nil, "", err
	}
	if len(names) == 0 {
		return NewFromShell(), "", nil
	}
	sort.Strings(names)
	e, err := Load(dir, names[0])
	if err != nil {
		return nil, "", err
	}
	return e, names[0], nil
}

// NewFromShell returns an Env with only the Shell layer filled from os.Environ.
// Used when no file-backed env is active (e.g. after ResolveStartEnv finds no
// YAML, or when the user cycles to "shell only" in the TUI).
func NewFromShell() *Env {
	shell := make(map[string]string)
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		shell[k] = v
	}
	return &Env{
		Shell: shell,
	}
}

// Load reads environments/<name>.yaml under dir. The process environment is
// always copied into the Shell layer first. name must be the file stem (e.g.
// "dev" for dev.yaml); empty name is invalid.
func Load(dir, name string) (*Env, error) {
	shell := make(map[string]string)
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		shell[k] = v
	}

	path := filepath.Join(dir, "environments", name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("env: read %s: %w", path, err)
	}

	var file map[string]string
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("env: parse %s: %w", path, err)
	}

	return &Env{
		Shell: shell,
		File:  file,
	}, nil
}

// ListNames returns the stem of each *.yaml under environments/ (recursive).
// Order follows filepath.WalkDir (not sorted). Callers that need a stable pick
// (e.g. ResolveStartEnv) must sort the slice.
// If environments/ is missing, returns (nil, nil).
func ListNames(dir string) ([]string, error) {
	envDir := filepath.Join(dir, "environments")
	var names []string

	err := filepath.WalkDir(envDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		base := filepath.Base(path)
		names = append(names, strings.TrimSuffix(base, ".yaml"))
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("env: list in %s: %w", envDir, err)
	}
	return names, nil
}
