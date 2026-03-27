package collection

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestLoadRequest_ReadsAndParsesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "req.yaml")
	if err := os.WriteFile(path, []byte("name: X\nmethod: GET\nurl: https://example.com\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	req, err := LoadRequest(path)
	if err != nil {
		t.Fatalf("LoadRequest: %v", err)
	}
	if req.Name != "X" || req.Method != "GET" || req.URL != "https://example.com" {
		t.Fatalf("unexpected request: %#v", req)
	}
}

func TestSaveRequest_WritesYAMLAndCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "requests", "nested", "req.yaml")

	in := Request{
		Name:   "X",
		Method: "POST",
		URL:    "https://example.com",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"ok":true}`,
		Auth: "bearer",
	}
	if err := SaveRequest(path, in); err != nil {
		t.Fatalf("SaveRequest: %v", err)
	}

	out, err := LoadRequest(path)
	if err != nil {
		t.Fatalf("LoadRequest: %v", err)
	}
	if out.Name != in.Name || out.Method != in.Method || out.URL != in.URL || out.Body != in.Body || out.Auth != in.Auth {
		t.Fatalf("roundtrip mismatch: in=%#v out=%#v", in, out)
	}
	if out.Headers["Content-Type"] != "application/json" {
		t.Fatalf("expected header to roundtrip, out.Headers=%v", out.Headers)
	}
}

func TestDiscoverCollections(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
		want  func(dir string) []string
	}{
		{
			name:  "when_none_found_returns_input_dir",
			setup: func(t *testing.T, dir string) {},
			want:  func(dir string) []string { return []string{dir} },
		},
		{
			name: "finds_directories_containing_dot_ferret_yaml",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				a := filepath.Join(dir, "a")
				b := filepath.Join(dir, "b", "c")
				if err := os.MkdirAll(a, 0o755); err != nil {
					t.Fatalf("mkdir a: %v", err)
				}
				if err := os.MkdirAll(b, 0o755); err != nil {
					t.Fatalf("mkdir b: %v", err)
				}
				if err := os.WriteFile(filepath.Join(a, ".ferret.yaml"), []byte("name: A\n"), 0o644); err != nil {
					t.Fatalf("write a config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(b, ".ferret.yaml"), []byte("name: C\n"), 0o644); err != nil {
					t.Fatalf("write b config: %v", err)
				}
			},
			want: func(dir string) []string {
				return []string{
					filepath.Join(dir, "a"),
					filepath.Join(dir, "b", "c"),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			got, err := DiscoverCollections(dir)
			if err != nil {
				t.Fatalf("DiscoverCollections: %v", err)
			}

			sort.Strings(got)
			want := tt.want(dir)
			sort.Strings(want)
			if len(got) != len(want) {
				t.Fatalf("expected %v, got %v", want, got)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("expected %v, got %v", want, got)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (dir string)
		want    Config
		wantErr bool
	}{
		{
			name: "loads_nearest_dot_ferret_yaml_up_hierarchy",
			setup: func(t *testing.T) string {
				t.Helper()
				root := t.TempDir()
				if err := os.WriteFile(filepath.Join(root, ".ferret.yaml"), []byte("name: Root\n"), 0o644); err != nil {
					t.Fatalf("write config: %v", err)
				}
				child := filepath.Join(root, "a", "b")
				if err := os.MkdirAll(child, 0o755); err != nil {
					t.Fatalf("mkdir child: %v", err)
				}
				return child
			},
			want:    Config{Name: "Root"},
			wantErr: false,
		},
		{
			name: "when_missing_returns_error",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			got, err := LoadConfig(dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig error = %v, wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Fatalf("expected cfg.Name=%q, got %q", tt.want.Name, got.Name)
			}
		})
	}
}

func TestLoadEntries_LoadsYAMLRequestsAndSkipsDotFerretYAML(t *testing.T) {
	dir := t.TempDir()
	reqDir := filepath.Join(dir, "requests")
	if err := os.MkdirAll(filepath.Join(reqDir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Should be ignored by LoadEntries.
	if err := os.WriteFile(filepath.Join(reqDir, ".ferret.yaml"), []byte("name: ignore\n"), 0o644); err != nil {
		t.Fatalf("write .ferret.yaml: %v", err)
	}
	// Should be ignored (non-yaml).
	if err := os.WriteFile(filepath.Join(reqDir, "README.txt"), []byte("ignore\n"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	if err := os.WriteFile(filepath.Join(reqDir, "one.yaml"), []byte("name: One\nmethod: GET\nurl: https://one\n"), 0o644); err != nil {
		t.Fatalf("write one.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reqDir, "nested", "two.yaml"), []byte("name: Two\nmethod: POST\nurl: https://two\n"), 0o644); err != nil {
		t.Fatalf("write two.yaml: %v", err)
	}

	entries, err := LoadEntries(dir)
	if err != nil {
		t.Fatalf("LoadEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %#v", len(entries), entries)
	}

	paths := []string{entries[0].Path, entries[1].Path}
	sort.Strings(paths)
	wantPaths := []string{"nested/two.yaml", "one.yaml"}
	for i := range wantPaths {
		if paths[i] != wantPaths[i] {
			t.Fatalf("expected paths %v, got %v", wantPaths, paths)
		}
	}
}

func TestLoadEntries_WhenRequestsDirMissing_ReturnsError(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "missing_requests_dir",
			setup: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			_, err := LoadEntries(dir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadEntries error = %v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
