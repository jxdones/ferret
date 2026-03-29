package env

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestEnv_Get_Priority(t *testing.T) {
	tests := []struct {
		name   string
		env    *Env
		key    string
		want   string
		wantOK bool
	}{
		{
			name: "shell_overrides_session_and_file",
			env: &Env{
				Shell:   map[string]string{"K": "shell"},
				Session: map[string]string{"K": "session"},
				File:    map[string]string{"K": "file"},
			},
			key:    "K",
			want:   "shell",
			wantOK: true,
		},
		{
			name: "session_overrides_file",
			env: &Env{
				Session: map[string]string{"K": "session"},
				File:    map[string]string{"K": "file"},
			},
			key:    "K",
			want:   "session",
			wantOK: true,
		},
		{
			name: "file_used_when_others_missing",
			env: &Env{
				File: map[string]string{"K": "file"},
			},
			key:    "K",
			want:   "file",
			wantOK: true,
		},
		{
			name:   "missing_key",
			env:    &Env{},
			key:    "K",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.env.Get(tt.key)
			if ok != tt.wantOK || got != tt.want {
				t.Fatalf("Get(%q) = (%q, %v), want (%q, %v)", tt.key, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestEnv_Set_InitializesSession(t *testing.T) {
	e := &Env{}
	e.Set("A", "1")
	if e.Session == nil {
		t.Fatalf("expected Session to be initialized")
	}
	if got := e.Session["A"]; got != "1" {
		t.Fatalf("expected Session[A]=%q, got %q", "1", got)
	}
}

func TestNewFromShell_IncludesProcessEnv(t *testing.T) {
	t.Setenv("FERRET_TEST_SHELL_KEY", "from-shell")
	e := NewFromShell()
	got, ok := e.Get("FERRET_TEST_SHELL_KEY")
	if !ok || got != "from-shell" {
		t.Fatalf("expected shell env variable, got (%q, %v)", got, ok)
	}
}

func TestLoad_ReadsFileAndPrepopulatesShell(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		contents  string
		wantErr   bool
		checkFunc func(t *testing.T, e *Env)
	}{
		{
			name:     "reads_file_and_prepopulates_shell",
			filename: "dev.yaml",
			contents: "A: one\nB: two\n",
			wantErr:  false,
			checkFunc: func(t *testing.T, e *Env) {
				t.Helper()
				if e == nil {
					t.Fatalf("expected non-nil Env")
				}
				if got, ok := e.Get("FERRET_TEST_SHELL_KEY"); !ok || got != "from-shell" {
					t.Fatalf("expected shell value, got (%q, %v)", got, ok)
				}
				if got, ok := e.Get("A"); !ok || got != "one" {
					t.Fatalf("expected file value for A, got (%q, %v)", got, ok)
				}
			},
		},
		{
			name:     "parse_error",
			filename: "bad.yaml",
			contents: ":\n  - not-a-map\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FERRET_TEST_SHELL_KEY", "from-shell")
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, "environments"), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			path := filepath.Join(dir, "environments", tt.filename)
			if err := os.WriteFile(path, []byte(tt.contents), 0o644); err != nil {
				t.Fatalf("write file: %v", err)
			}

			name := tt.filename[:len(tt.filename)-len(filepath.Ext(tt.filename))]
			e, err := Load(dir, name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load error = %v, wantErr=%v", err, tt.wantErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, e)
			}
		})
	}
}

func TestListNames_MissingDirReturnsNil(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		want    []string
		wantNil bool
	}{
		{
			name:    "missing_dir_returns_nil",
			setup:   func(t *testing.T, dir string) {},
			wantNil: true,
		},
		{
			name: "filters_and_trims_suffix_includes_nested",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				envDir := filepath.Join(dir, "environments")
				if err := os.MkdirAll(filepath.Join(envDir, "subdir"), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte("A: one\n"), 0o644); err != nil {
					t.Fatalf("write dev: %v", err)
				}
				if err := os.WriteFile(filepath.Join(envDir, "prod.yaml"), []byte("A: one\n"), 0o644); err != nil {
					t.Fatalf("write prod: %v", err)
				}
				if err := os.WriteFile(filepath.Join(envDir, "README.txt"), []byte("ignore\n"), 0o644); err != nil {
					t.Fatalf("write readme: %v", err)
				}
				if err := os.WriteFile(filepath.Join(envDir, "subdir", "nested.yaml"), []byte("A: one\n"), 0o644); err != nil {
					t.Fatalf("write nested: %v", err)
				}
			},
			want: []string{"dev", "nested", "prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			names, err := ListNames(dir)
			if err != nil {
				t.Fatalf("ListNames: %v", err)
			}
			if tt.wantNil {
				if names != nil {
					t.Fatalf("expected nil names, got %#v", names)
				}
				return
			}

			sort.Strings(names)
			if len(names) != len(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, names)
			}
			for i := range tt.want {
				if names[i] != tt.want[i] {
					t.Fatalf("expected %v, got %v", tt.want, names)
				}
			}
		})
	}
}

func TestResolveStartEnv(t *testing.T) {
	t.Run("explicit_name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "environments"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "environments", "dev.yaml"), []byte("BASE: https://a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		e, name, err := ResolveStartEnv(dir, "dev")
		if err != nil {
			t.Fatal(err)
		}
		if name != "dev" {
			t.Fatalf("name = %q, want dev", name)
		}
		got, ok := e.Get("BASE")
		if !ok || got != "https://a" {
			t.Fatalf("Get(BASE) = (%q, %v)", got, ok)
		}
	})

	t.Run("empty_flag_no_files_is_shell_only", func(t *testing.T) {
		dir := t.TempDir()
		e, name, err := ResolveStartEnv(dir, "")
		if err != nil {
			t.Fatal(err)
		}
		if name != "" {
			t.Fatalf("name = %q, want empty", name)
		}
		if e.File != nil && len(e.File) > 0 {
			t.Fatalf("expected no file layer, got %#v", e.File)
		}
	})

	t.Run("empty_flag_picks_first_sorted_yaml", func(t *testing.T) {
		dir := t.TempDir()
		envDir := filepath.Join(dir, "environments")
		if err := os.MkdirAll(envDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(envDir, "prod.yaml"), []byte("N: prod\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), []byte("N: dev\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		e, name, err := ResolveStartEnv(dir, "")
		if err != nil {
			t.Fatal(err)
		}
		if name != "dev" {
			t.Fatalf("name = %q, want dev", name)
		}
		got, _ := e.Get("N")
		if got != "dev" {
			t.Fatalf("Get(N) = %q, want dev", got)
		}
	})
}
