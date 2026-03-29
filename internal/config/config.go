package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"
)

// Config is the root of ~/.ferret/config.yaml.
type Config struct {
	Workspaces []Workspace `yaml:"workspaces"`
}

// Workspace is a named directory that contains one or more ferret collections
// (subdirectories with .ferret.yaml). Path may be absolute or use a "~"
// prefix for the home directory.
type Workspace struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// UnmarshalYAML accepts either a string (path) or a mapping with name/path.
func (w *Workspace) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		w.Path = s
		if s != "" {
			w.Name = filepath.Base(s)
		}
		return nil
	case yaml.MappingNode:
		type raw Workspace
		if err := value.Decode((*raw)(w)); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("config: workspace must be a string path or a mapping, got YAML kind %v", value.Kind)
	}
}

// DefaultConfig returns config with standard values (e.g. for first-run).
func DefaultConfig() Config {
	return Config{}
}

// ConfigDir returns the ferret config directory (e.g. ~/.ferret).
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ferret"), nil
}

// ConfigPath returns the path to the config file (~/.ferret/config.yaml).
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// EnsureConfigDir creates the config directory if it does not exist.
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o700)
}

// WriteConfig writes cfg to the config file, creating the file if needed.
// A comment header with the config path is written at the top of the file.
func WriteConfig(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	header := "# " + path + "\n\n"
	return os.WriteFile(path, append([]byte(header), data...), 0o600)
}

// LoadConfig loads the config from the config file.
// If the file does not exist, it is created with default values and those are returned.
func LoadConfig() (Config, error) {
	err := EnsureConfigDir()
	if err != nil {
		return Config{}, err
	}

	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = DefaultConfig()
			if writeErr := WriteConfig(cfg); writeErr != nil {
				return Config{}, writeErr
			}
			return cfg, nil
		}
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// ExpandPath resolves path for use on the local filesystem. A leading "~" or
// "~/" is replaced with the user's home directory; otherwise filepath.Clean is
// applied. Empty path returns ("", nil).
func ExpandPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	if path == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Clean(path), nil
}
