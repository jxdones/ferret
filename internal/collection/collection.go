package collection

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"go.yaml.in/yaml/v4"
)

// Request represents a single API request.
type Request struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
	Auth    string            `yaml:"auth"`
}

// Entry represents a single entry in the collection.
type Entry struct {
	Path    string  `yaml:"path"`
	Request Request `yaml:"request"`
}

// Config represents the collection configuration.
type Config struct {
	Name string     `yaml:"name"`
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig represents the authentication configuration.
type AuthConfig struct {
	Type         string `yaml:"type"`
	TokenURL     string `yaml:"token_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	InjectAs     string `yaml:"inject_as"`
}

// LoadRequest loads a request from a file.
func LoadRequest(path string) (Request, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return Request{}, fmt.Errorf("collection: read %s: %w", path, err)
	}
	var request Request
	err = yaml.Unmarshal(yamlFile, &request)
	if err != nil {
		return Request{}, fmt.Errorf("collection: parse %s: %w", path, err)
	}
	return request, nil
}

// SaveRequest saves a request to a file.
func SaveRequest(path string, request Request) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("collection: create directory for %s: %w", path, err)
	}
	data, err := yaml.Marshal(request)
	if err != nil {
		return fmt.Errorf("collection: marshal request for %s: %w", path, err)
	}
	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return fmt.Errorf("collection: write request to %s: %w", path, err)
	}
	return nil
}

// DiscoverCollections discovers all collections in a directory.
// A collection is a directory that contains a .ferret.yaml file.
// If no collections are found, the directory itself is considered a collection root.
func DiscoverCollections(dir string) ([]string, error) {
	collections := map[string]struct{}{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == ".ferret.yaml" {
			collections[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collection: discover collections in %s: %w", dir, err)
	}

	if len(collections) == 0 {
		return []string{dir}, nil
	}

	out := make([]string, 0, len(collections))
	for collection := range collections {
		out = append(out, collection)
	}
	sort.Strings(out)
	return out, nil
}

// LoadConfig loads a collection configuration from a file.
// The config is loaded from the nearest .ferret.yaml file in the directory hierarchy.
func LoadConfig(dir string) (Config, error) {
	current := dir
	for {
		path := filepath.Join(current, ".ferret.yaml")
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("collection: parse %s: %w", path, err)
			}
			return cfg, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return Config{}, fmt.Errorf("collection: no .ferret.yaml found in %s or any parent", dir)
		}
		current = parent
	}
}

// LoadEntries loads all entries from a collection directory.
func LoadEntries(dir string) ([]Entry, error) {
	var entries []Entry

	requestsDir := filepath.Join(dir, "requests")
	_, err := os.Stat(requestsDir)
	if err != nil {
		return nil, fmt.Errorf("collection: missing requests directory in %s: %w", dir, err)
	}

	err = filepath.WalkDir(requestsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".yaml" {
			return nil
		}
		if filepath.Base(path) == ".ferret.yaml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("collection: read %s: %w", path, err)
		}

		var request Request
		if err := yaml.Unmarshal(data, &request); err != nil {
			return fmt.Errorf("collection: parse %s: %w", path, err)
		}

		rel, err := filepath.Rel(requestsDir, path)
		if err != nil {
			return fmt.Errorf("collection: get relative path for %s: %w", path, err)
		}

		entries = append(entries, Entry{Path: rel, Request: request})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
