package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jxdones/ferret/internal/collection"
)

func requireCmd(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not available in PATH", name)
	}
}

func assertVars(t *testing.T, got *HookResult) {
	t.Helper()
	if got == nil || got.Vars == nil {
		t.Fatalf("got %#v, want vars", got)
	}
	if got.Vars["FOO"] != "bar" {
		t.Fatalf("Vars[FOO] = %q, want %q (vars=%v)", got.Vars["FOO"], "bar", got.Vars)
	}
	if got.Vars["TOKEN"] != "a=b=c" {
		t.Fatalf("Vars[TOKEN] = %q, want %q (vars=%v)", got.Vars["TOKEN"], "a=b=c", got.Vars)
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name string
		req  *collection.Request
		cfg  *collection.Config
		want string
	}{
		{
			name: "when_req_pre_request_inherit_returns_cfg_pre_request",
			req:  &collection.Request{PreRequest: "inherit"},
			cfg:  &collection.Config{PreRequest: "cfg-hook"},
			want: "cfg-hook",
		},
		{
			name: "when_req_pre_request_is_explicit_returns_req_pre_request",
			req:  &collection.Request{PreRequest: "req-hook"},
			cfg:  &collection.Config{PreRequest: "cfg-hook"},
			want: "req-hook",
		},
		{
			name: "when_req_pre_request_empty_returns_empty",
			req:  &collection.Request{PreRequest: ""},
			cfg:  &collection.Config{PreRequest: "cfg-hook"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.req, tt.cfg)
			if got != tt.want {
				t.Fatalf("Resolve(%#v, %#v) = %q, want %q", tt.req, tt.cfg, got, tt.want)
			}
		})
	}
}

func TestParseVars(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want map[string]string
	}{
		{
			name: "empty_output_returns_empty_map",
			in:   nil,
			want: map[string]string{},
		},
		{
			name: "parses_key_value_pairs",
			in:   []byte("FOO=bar\nBAR=baz\n"),
			want: map[string]string{"FOO": "bar", "BAR": "baz"},
		},
		{
			name: "ignores_non_matching_lines_and_blank_lines",
			in:   []byte("\nnotvar\nOK=yes\n"),
			want: map[string]string{"OK": "yes"},
		},
		{
			name: "value_can_contain_equals_signs",
			in:   []byte("TOKEN=a=b=c\n"),
			want: map[string]string{"TOKEN": "a=b=c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVars(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("parseVars(%q) len=%d, want %d (got=%v)", string(tt.in), len(got), len(tt.want), got)
			}
			for k, wantV := range tt.want {
				if gotV, ok := got[k]; !ok || gotV != wantV {
					t.Fatalf("parseVars(%q)[%q] = %q, want %q (got=%v)", string(tt.in), k, gotV, wantV, got)
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Run("empty_hook_returns_nil_result_and_nil_error", func(t *testing.T) {
		got, err := Run(context.Background(), "", nil)
		if err != nil {
			t.Fatalf("Run(\"\") err = %v, want nil", err)
		}
		if got != nil {
			t.Fatalf("Run(\"\") = %#v, want nil", got)
		}
	})

	t.Run("go_binary_hook_parses_vars_from_stdout", func(t *testing.T) {
		dir := t.TempDir()
		mainPath := filepath.Join(dir, "main.go")
		prog := `package main

import "fmt"

func main() {
	fmt.Println("FOO=bar")
	fmt.Println("TOKEN=a=b=c")
}
`
		if err := os.WriteFile(mainPath, []byte(prog), 0o644); err != nil {
			t.Fatalf("write hook program: %v", err)
		}

		hookPath := filepath.Join(dir, "hook")
		out, err := exec.Command("go", "build", "-o", hookPath, mainPath).CombinedOutput()
		if err != nil {
			t.Fatalf("build hook program: %v\n%s", err, string(out))
		}

		got, err := Run(context.Background(), hookPath, nil)
		if err != nil {
			t.Fatalf("Run(%q) err = %v, want nil", hookPath, err)
		}
		assertVars(t, got)
	})

	t.Run("polyglot_hooks", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("shebang-based hooks are not portable on Windows")
		}

		tests := []struct {
			name        string
			interpreter string
			ext         string
			body        string
		}{
			{
				name:        "sh",
				interpreter: "sh",
				ext:         ".sh",
				body:        "echo 'FOO=bar'\necho 'TOKEN=a=b=c'\n",
			},
			{
				name:        "python",
				interpreter: "python3",
				ext:         ".py",
				body:        "print('FOO=bar')\nprint('TOKEN=a=b=c')\n",
			},
			{
				name:        "node",
				interpreter: "node",
				ext:         ".js",
				body:        "console.log('FOO=bar');\nconsole.log('TOKEN=a=b=c');\n",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				requireCmd(t, tt.interpreter)

				dir := t.TempDir()
				hookPath := filepath.Join(dir, "hook"+tt.ext)
				script := "#!/usr/bin/env " + tt.interpreter + "\n" + tt.body
				if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
					t.Fatalf("write hook: %v", err)
				}

				got, err := Run(context.Background(), hookPath, nil)
				if err != nil {
					t.Fatalf("Run(%q) err = %v, want nil", hookPath, err)
				}
				assertVars(t, got)
			})
		}
	})

	t.Run("missing_hook_returns_error", func(t *testing.T) {
		_, err := Run(context.Background(), filepath.Join(t.TempDir(), "does-not-exist"), nil)
		if err == nil {
			t.Fatal("Run(missing) err = nil, want error")
		}
	})

	t.Run("timeout_returns_clear_error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("shebang-based hooks are not portable on Windows")
		}

		dir := t.TempDir()
		hookPath := filepath.Join(dir, "slow.sh")
		script := "#!/usr/bin/env sh\nsleep 10\n"
		if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := Run(ctx, hookPath, nil)
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "timed out") {
			t.Fatalf("err = %q, want message containing 'timed out'", err.Error())
		}
	})

	t.Run("cancellation_returns_clear_error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("shebang-based hooks are not portable on Windows")
		}

		dir := t.TempDir()
		hookPath := filepath.Join(dir, "slow.sh")
		script := "#!/usr/bin/env sh\nsleep 10\n"
		if err := os.WriteFile(hookPath, []byte(script), 0o755); err != nil {
			t.Fatalf("write hook: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		_, err := Run(ctx, hookPath, nil)
		if err == nil {
			t.Fatal("expected cancellation error, got nil")
		}
		if !strings.Contains(err.Error(), "cancelled") {
			t.Fatalf("err = %q, want message containing 'cancelled'", err.Error())
		}
	})
}
