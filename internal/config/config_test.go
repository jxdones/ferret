package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("empty", func(t *testing.T) {
		got, err := ExpandPath("")
		if err != nil || got != "" {
			t.Fatalf("ExpandPath(\"\") = %q, %v; want \"\", nil", got, err)
		}
	})
	t.Run("tilde", func(t *testing.T) {
		got, err := ExpandPath("~")
		if err != nil || got != home {
			t.Fatalf("ExpandPath(\"~\") = %q, %v; want %q", got, err, home)
		}
	})
	t.Run("tilde slash", func(t *testing.T) {
		got, err := ExpandPath("~/foo/bar")
		want := filepath.Join(home, "foo", "bar")
		if err != nil || got != want {
			t.Fatalf("ExpandPath(\"~/foo/bar\") = %q, %v; want %q", got, err, want)
		}
	})
	t.Run("relative", func(t *testing.T) {
		got, err := ExpandPath("foo/bar")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Clean("foo/bar")
		if got != want {
			t.Fatalf("ExpandPath(\"foo/bar\") = %q; want %q", got, want)
		}
	})
}

func TestEnsureConfigExists(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, home string)
		wantFile  bool
		checkFile func(t *testing.T, data []byte)
	}{
		{
			name:     "creates_dir_and_file_when_missing",
			setup:    func(t *testing.T, home string) {},
			wantFile: true,
			checkFile: func(t *testing.T, data []byte) {
				t.Helper()
				var cfg Config
				if err := yaml.Unmarshal(data, &cfg); err != nil {
					t.Fatalf("unmarshal written file: %v", err)
				}
			},
		},
		{
			name: "does_not_overwrite_existing_file",
			setup: func(t *testing.T, home string) {
				t.Helper()
				dir := filepath.Join(home, ".ferret")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				content := "# existing\nworkspaces:\n  - /some/path\n"
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantFile: true,
			checkFile: func(t *testing.T, data []byte) {
				t.Helper()
				if !strings.Contains(string(data), "/some/path") {
					t.Fatalf("existing file was overwritten; got: %s", data)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			tt.setup(t, home)

			if err := EnsureConfigExists(); err != nil {
				t.Fatalf("EnsureConfigExists: %v", err)
			}

			path := filepath.Join(home, ".ferret", "config.yaml")
			data, err := os.ReadFile(path)
			if (err == nil) != tt.wantFile {
				t.Fatalf("file exists = %v, want %v (err: %v)", err == nil, tt.wantFile, err)
			}
			if tt.checkFile != nil && err == nil {
				tt.checkFile(t, data)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, home string)
		wantErr   bool
		checkFunc func(t *testing.T, cfg Config)
	}{
		{
			name:  "returns_default_when_file_missing",
			setup: func(t *testing.T, home string) {},
			checkFunc: func(t *testing.T, cfg Config) {
				t.Helper()
				want := DefaultConfig()
				if len(cfg.Workspaces) != len(want.Workspaces) {
					t.Fatalf("got %+v, want %+v", cfg, want)
				}
			},
		},
		{
			name: "reads_existing_file",
			setup: func(t *testing.T, home string) {
				t.Helper()
				dir := filepath.Join(home, ".ferret")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				content := "workspaces:\n  - name: work\n    path: /opt/work\n"
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			checkFunc: func(t *testing.T, cfg Config) {
				t.Helper()
				if len(cfg.Workspaces) != 1 {
					t.Fatalf("expected 1 workspace, got %d", len(cfg.Workspaces))
				}
				if cfg.Workspaces[0].Name != "work" || cfg.Workspaces[0].Path != "/opt/work" {
					t.Fatalf("unexpected workspace: %+v", cfg.Workspaces[0])
				}
			},
		},
		{
			name: "reads_theme_field",
			setup: func(t *testing.T, home string) {
				t.Helper()
				dir := filepath.Join(home, ".ferret")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				content := "theme: princess\nworkspaces:\n  - name: work\n    path: /opt/work\n"
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			checkFunc: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Theme != "princess" {
					t.Fatalf("Theme = %q, want %q", cfg.Theme, "princess")
				}
			},
		},
		{
			name: "theme_defaults_to_empty_when_absent",
			setup: func(t *testing.T, home string) {
				t.Helper()
				dir := filepath.Join(home, ".ferret")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				content := "workspaces:\n  - name: work\n    path: /opt/work\n"
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			checkFunc: func(t *testing.T, cfg Config) {
				t.Helper()
				if cfg.Theme != "" {
					t.Fatalf("Theme = %q, want empty string", cfg.Theme)
				}
			},
		},
		{
			name: "returns_error_on_invalid_yaml",
			setup: func(t *testing.T, home string) {
				t.Helper()
				dir := filepath.Join(home, ".ferret")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(":\n  - bad\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			tt.setup(t, home)

			cfg, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.checkFunc != nil && err == nil {
				tt.checkFunc(t, cfg)
			}
		})
	}
}

func TestWorkspaceUnmarshalYAML_stringOrMapping(t *testing.T) {
	t.Run("path_only", func(t *testing.T) {
		const in = `
workspaces:
  - ~/development/my-saas
`
		var cfg Config
		if err := yaml.Unmarshal([]byte(in), &cfg); err != nil {
			t.Fatal(err)
		}
		if len(cfg.Workspaces) != 1 {
			t.Fatalf("len = %d", len(cfg.Workspaces))
		}
		w := cfg.Workspaces[0]
		if w.Path != "~/development/my-saas" {
			t.Fatalf("Path = %q", w.Path)
		}
		if w.Name != "my-saas" {
			t.Fatalf("Name = %q, want my-saas", w.Name)
		}
	})
	t.Run("name_and_path", func(t *testing.T) {
		const in = `
workspaces:
  - name: SaaS
    path: /opt/projects/api
`
		var cfg Config
		if err := yaml.Unmarshal([]byte(in), &cfg); err != nil {
			t.Fatal(err)
		}
		w := cfg.Workspaces[0]
		if w.Name != "SaaS" || w.Path != "/opt/projects/api" {
			t.Fatalf("got %#v", w)
		}
	})
	t.Run("comment_header", func(t *testing.T) {
		in := strings.TrimSpace(`
# ~/.ferret/config.yaml

workspaces:
  - ~/devel/foo
`)
		var cfg Config
		if err := yaml.Unmarshal([]byte(in), &cfg); err != nil {
			t.Fatal(err)
		}
		if len(cfg.Workspaces) != 1 || cfg.Workspaces[0].Path != "~/devel/foo" {
			t.Fatalf("got %#v", cfg.Workspaces)
		}
	})
}
