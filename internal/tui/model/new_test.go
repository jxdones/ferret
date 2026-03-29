package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_scratchNoWorkspace(t *testing.T) {
	tmp := t.TempDir()
	m, err := New(StartOptions{
		Dir:                 tmp,
		EnvName:             "dev",
		ImplicitDirectory:   true,
		ConfigHasWorkspaces: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.collectionRoot != "" {
		t.Fatalf("collectionRoot = %q, want empty", m.collectionRoot)
	}
	if len(m.collectionDirs) != 0 {
		t.Fatalf("len(collectionDirs) = %d, want 0", len(m.collectionDirs))
	}
	if m.workspaceRoot != tmp {
		t.Fatalf("workspaceRoot = %q, want %q", m.workspaceRoot, tmp)
	}
	if m.envName != "" {
		t.Fatalf("envName = %q, want empty (ignore -e in scratch mode)", m.envName)
	}
	view := m.titlebar.View().Content
	if !strings.Contains(view, "no workspace") {
		t.Fatalf("titlebar should show no workspace, got %q", view)
	}
}

func TestNew_withCollections_resolvesDir(t *testing.T) {
	root := t.TempDir()
	colDir := filepath.Join(root, "api")
	if err := os.MkdirAll(filepath.Join(colDir, "requests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(colDir, ".ferret.yaml"), []byte("name: api\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(colDir, "requests", "ping.yaml"), []byte("name: ping\nmethod: GET\nurl: https://example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := New(StartOptions{
		Dir:                 root,
		ImplicitDirectory:   false,
		ConfigHasWorkspaces: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(m.collectionRoot) != "api" {
		t.Fatalf("collectionRoot base = %q, want api", filepath.Base(m.collectionRoot))
	}
}
