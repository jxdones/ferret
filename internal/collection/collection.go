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
	Name       string            `yaml:"name"`
	Method     string            `yaml:"method"`
	URL        string            `yaml:"url"`
	Headers    map[string]string `yaml:"headers"`
	Body       string            `yaml:"body"`
	Auth       *Auth             `yaml:"auth,omitempty"`
	PreRequest string            `yaml:"pre_request,omitempty"`
}

// Entry represents a single entry in the collection.
type Entry struct {
	Path    string  `yaml:"path"`
	Request Request `yaml:"request"`
}

// Config represents the collection configuration.
type Config struct {
	Name       string `yaml:"name"`
	PreRequest string `yaml:"pre_request,omitempty"`
	Auth       *Auth  `yaml:"auth,omitempty"`
}

type Auth struct {
	Type string `yaml:"type"`

	// bearer auth
	Token string `yaml:"token,omitempty"`

	// basic auth
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`

	// api key auth
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
	In    string `yaml:"in,omitempty"` // header or query

}

// LoadRequest loads a request from a file.
func LoadRequest(path string) (Request, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return Request{}, fmt.Errorf("collection: read %s: %w", shortFilePath(path), err)
	}
	var request Request
	err = yaml.Unmarshal(yamlFile, &request)
	if err != nil {
		return Request{}, fmt.Errorf("collection: parse %s: %w", shortFilePath(path), err)
	}
	return request, nil
}

// SaveRequest saves a request to a file.
func SaveRequest(path string, request Request) error {
	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		return fmt.Errorf("collection: create directory for %s: %w", shortFilePath(path), err)
	}
	data, err := yaml.Marshal(request)
	if err != nil {
		return fmt.Errorf("collection: marshal %s: %w", shortFilePath(path), err)
	}
	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		return fmt.Errorf("collection: write %s: %w", shortFilePath(path), err)
	}
	return nil
}

// DiscoverCollections finds collection roots under dir without recursing into
// subdirectories. It looks for .ferret.yaml in dir itself and in each immediate
// child directory only (so scanning a large folder like $HOME stays cheap).
// Deeper layouts (e.g. dir/a/b/.ferret.yaml with no marker at dir/a) are not
// discovered here. If no marker is found, dir is treated as a single collection
// root (same as before).
func DiscoverCollections(dir string) ([]string, error) {
	dir = filepath.Clean(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("collection: could not read workspace: %w", err)
	}

	collections := map[string]struct{}{}

	if ok, err := hasFerretYAML(dir); err != nil {
		return nil, fmt.Errorf("collection: could not read workspace: %w", err)
	} else if ok {
		collections[dir] = struct{}{}
	}

	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		sub := filepath.Join(dir, ent.Name())
		ok, err := hasFerretYAML(sub)
		if err != nil {
			return nil, fmt.Errorf("collection: could not read workspace: %w", err)
		}
		if ok {
			collections[sub] = struct{}{}
		}
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

func hasFerretYAML(dir string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, ".ferret.yaml"))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) || os.IsPermission(err) {
		return false, nil
	}
	return false, err
}

// LoadConfig loads a collection configuration from a file.
// The config is loaded from the nearest .ferret.yaml file in the directory
// hierarchy.
func LoadConfig(dir string) (Config, error) {
	current := dir
	for {
		path := filepath.Join(current, ".ferret.yaml")
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("collection: parse .ferret.yaml: %w", err)
			}
			return cfg, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return Config{}, fmt.Errorf("collection: no .ferret.yaml found in current directory or any parent")
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
		return nil, fmt.Errorf("collection: missing requests/ directory: %w", err)
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
			return fmt.Errorf("collection: read %s: %w", shortFilePath(path), err)
		}

		var request Request
		if err := yaml.Unmarshal(data, &request); err != nil {
			return fmt.Errorf("collection: parse %s: %w", shortFilePath(path), err)
		}

		rel, err := filepath.Rel(requestsDir, path)
		if err != nil {
			return fmt.Errorf("collection: get relative path for %s: %w", shortFilePath(path), err)
		}

		entries = append(entries, Entry{Path: rel, Request: request})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// shortFilePath returns a short, user-facing path for error messages.
// Direct children of requests/ use the collection name as prefix.
// Nested files use their immediate parent directory as prefix.
func shortFilePath(path string) string {
	parent := filepath.Base(filepath.Dir(path))
	if parent == "requests" {
		return filepath.Base(filepath.Dir(filepath.Dir(path))) + "/" + filepath.Base(path)
	}
	return parent + "/" + filepath.Base(path)
}
