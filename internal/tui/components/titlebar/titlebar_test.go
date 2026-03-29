package titlebar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestModelView(t *testing.T) {
	tuitest.UseStableTheme(t)
	intPtr := func(v int) *int { return &v }

	tests := []struct {
		name       string
		width      int
		workspace  string
		collection string
		entry      string
		env        string
		want       []string
		wantWidth  *int
	}{
		{
			name:      "no_workspace_shell_only",
			width:     50,
			workspace: "no workspace",
			env:       "",
			want:      []string{"no workspace", "shell only"},
			wantWidth: intPtr(50),
		},
		{
			name:       "collection_shell_only_when_env_blank",
			width:      40,
			collection: "api",
			env:        "   ",
			want:       []string{"api", "shell only"},
			wantWidth:  intPtr(40),
		},
		{
			name:       "shows_collection_entry_and_file_env",
			width:      60,
			collection: "my-collection",
			entry:      "List users",
			env:        "dev",
			want:       []string{"my-collection", " / ", "List users", "dev"},
			wantWidth:  intPtr(60),
		},
		{
			name:       "workspace_collection_entry_uses_path_separators",
			width:      80,
			workspace:  "main workspace",
			collection: "pokeapi",
			entry:      "get ditto",
			env:        "dev",
			want:       []string{"main workspace", " / ", "pokeapi", " / ", "get ditto", "dev"},
			wantWidth:  intPtr(80),
		},
		{
			name:      "small_width_does_not_panic",
			width:     0,
			env:       "",
			want:      nil,
			wantWidth: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.SetSize(tt.width)
			m.SetWorkspace(tt.workspace)
			m.SetCollection(tt.collection)
			m.SetEntry(tt.entry)
			m.SetEnv(tt.env)

			got := tuitest.StripANSI(m.View().Content)
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Fatalf("View() = %q, want to contain %q", got, w)
				}
			}
			if tt.wantWidth != nil {
				if gotW := ansi.StringWidth(got); gotW != *tt.wantWidth {
					t.Fatalf("View() width = %d, want %d (output=%q)", gotW, *tt.wantWidth, got)
				}
			}
		})
	}
}
