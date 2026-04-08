package hook

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jxdones/ferret/internal/collection"
	"github.com/jxdones/ferret/internal/env"
)

var varPattern = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)=(.*)$`)

type HookResult struct {
	Vars map[string]string
}

// Resolve resolves the pre-request hook for a request.
func Resolve(req *collection.Request, cfg *collection.Config) string {
	if req.PreRequest == "inherit" {
		return cfg.PreRequest
	}
	return req.PreRequest
}

// Run runs a hook script and returns the result.
func Run(ctx context.Context, scriptPath string, e *env.Env) (*HookResult, error) {
	if scriptPath == "" {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, scriptPath)
	cmd.Env = buildEnv(e)
	output, err := cmd.Output()
	if err != nil {
		short := shortPath(scriptPath)
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("hook timed out after 10s: %s", short)
		}
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("hook cancelled: %s", short)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("hook script is not executable: %s (run chmod +x %s)", short, short)
		}
		return nil, fmt.Errorf("hook: run %s: %w", short, err)
	}

	return &HookResult{
		Vars: parseVars(output),
	}, nil
}

// buildEnv merges the three env layers into a []string suitable for cmd.Env.
// Priority: Shell > Session > File (Shell wins on conflict).
func buildEnv(e *env.Env) []string {
	if e == nil {
		return nil
	}
	merged := make(map[string]string)
	maps.Copy(merged, e.File)
	maps.Copy(merged, e.Session)
	maps.Copy(merged, e.Shell)
	out := make([]string, 0, len(merged))
	for k, v := range merged {
		out = append(out, k+"="+v)
	}
	return out
}

// shortPath returns the last two components of a path (parent/file) for display.
func shortPath(p string) string {
	return filepath.Join(filepath.Base(filepath.Dir(p)), filepath.Base(p))
}

// parseVars parses the output of a hook script into a map of variables.
// The variable name and value are separated by an equals sign.
// e.g.: FOO=bar, BAR=baz, token=1234567890, etc.
func parseVars(output []byte) map[string]string {
	vars := make(map[string]string)
	for line := range strings.SplitSeq(string(output), "\n") {
		match := varPattern.FindStringSubmatch(line)
		if match != nil {
			vars[match[1]] = match[2]
		}
	}
	return vars
}
