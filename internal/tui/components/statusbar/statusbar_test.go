package statusbar

import (
	"strings"
	"testing"
	"time"

	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{
			name: "subsecond_uses_ms",
			in:   500 * time.Millisecond,
			want: "500ms",
		},
		{
			name: "one_second_or_more_uses_seconds_with_one_decimal",
			in:   1500 * time.Millisecond,
			want: "1.5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.in)
			if got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{
			name: "sub_kb_uses_bytes",
			in:   12,
			want: "12B",
		},
		{
			name: "one_kb_or_more_uses_kb_with_one_decimal",
			in:   2048,
			want: "2.0KB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.in)
			if got != tt.want {
				t.Fatalf("formatSize(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestModel_RenderRight(t *testing.T) {
	tuitest.UseStableTheme(t)

	tests := []struct {
		name    string
		model   Model
		want    []string
		wantNot []string
	}{
		{
			name:  "nil_response_returns_empty_string",
			model: Model{response: nil},
			want:  []string{""},
		},
		{
			name: "falls_back_to_http_status_text_when_missing",
			model: Model{
				response: &Response{
					StatusCode: 404,
					StatusText: "",
					Duration:   10 * time.Millisecond,
					Size:       12,
				},
			},
			want: []string{"404 Not Found", "10ms", "12B"},
		},
		{
			name: "includes_format_when_present",
			model: Model{
				response: &Response{
					StatusCode: 200,
					StatusText: "200 OK",
					Duration:   1500 * time.Millisecond,
					Size:       2048,
					Format:     "json",
				},
			},
			want: []string{"200 OK", "1.5s", "2.0KB", "json"},
		},
		{
			name: "trims_code_prefix_from_status_text",
			model: Model{
				response: &Response{
					StatusCode: 200,
					StatusText: "200 OK",
					Duration:   10 * time.Millisecond,
					Size:       1,
				},
			},
			want:    []string{"200 OK"},
			wantNot: []string{"200 200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tuitest.StripANSI(tt.model.renderRight())

			// Special case: when response is nil we expect exact empty.
			if tt.model.response == nil {
				if got != "" {
					t.Fatalf("renderRight() = %q, want empty string", got)
				}
				return
			}

			for _, w := range tt.want {
				if w == "" {
					continue
				}
				if !strings.Contains(got, w) {
					t.Fatalf("renderRight() = %q, want to contain %q", got, w)
				}
			}
			for _, w := range tt.wantNot {
				if strings.Contains(got, w) {
					t.Fatalf("renderRight() = %q, want not to contain %q", got, w)
				}
			}
		})
	}
}
