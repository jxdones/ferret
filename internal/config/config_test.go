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
