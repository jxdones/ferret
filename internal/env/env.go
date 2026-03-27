package env

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Env holds environment variables for a collections from three sources with
// decreasing priority:
//
//   - Shell: OS process environment (os.Environ). Has highest priority.
//     a variable set in the terminal before running ferret always overrides everything else.
//   - Session: Environment variables set in the current session. Resets on exit.
//   - File: Environment variables set in the collection file. Has lowest priority.
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

// NewFromShell returns an Env initialized with the current shell environment.
// Use when no environment file is present. (e.g. ferret starts without --env flag)
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

// Load reads environments/<name>.yaml from dir and returns an Env
// with shell variables pre-populated from the process environment.
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

// ListNames lists all environment names in the environments directory.
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
