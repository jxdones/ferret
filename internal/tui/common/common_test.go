package common

import (
	"testing"
)

func TestClampMin(t *testing.T) {
	tests := []struct {
		name  string
		value int
		min   int
		want  int
	}{
		{
			name:  "value_below_min_clamps",
			value: 0,
			min:   1,
			want:  1,
		},
		{
			name:  "value_equal_min_keeps_value",
			value: 5,
			min:   5,
			want:  5,
		},
		{
			name:  "value_above_min_keeps_value",
			value: 10,
			min:   5,
			want:  10,
		},
		{
			name:  "negative_min_allows_negative",
			value: -5,
			min:   -10,
			want:  -5,
		},
		{
			name:  "negative_value_clamps_to_min",
			value: -11,
			min:   -10,
			want:  -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampMin(tt.value, tt.min)
			if got != tt.want {
				t.Fatalf("ClampMin(%d, %d) = %d, want %d", tt.value, tt.min, got, tt.want)
			}
		})
	}
}

func TestTruncatePad(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		want  string
	}{
		{
			name:  "pads_short_string",
			in:    "abc",
			width: 5,
			want:  "abc  ",
		},
		{
			name:  "truncates_long_string",
			in:    "abcdef",
			width: 3,
			want:  "abc",
		},
		{
			name:  "exact_width_unchanged",
			in:    "abc",
			width: 3,
			want:  "abc",
		},
		{
			name:  "zero_width_returns_empty",
			in:    "abcdef",
			width: 0,
			want:  "",
		},
		{
			name:  "negative_width_returns_empty",
			in:    "abc",
			width: -1,
			want:  "",
		},
		{
			name:  "empty_string_pads_to_width",
			in:    "",
			width: 3,
			want:  "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncatePad(tt.in, tt.width)
			if got != tt.want {
				t.Fatalf("TruncatePad(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
		})
	}
}

func TestDetectSyntax(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		want        string
	}{
		// Content-Type header takes priority.
		{
			name:        "json_content_type",
			contentType: "application/json",
			body:        "",
			want:        "json",
		},
		{
			name:        "yaml_content_type",
			contentType: "application/yaml",
			body:        "",
			want:        "yaml",
		},
		{
			name:        "xml_content_type",
			contentType: "application/xml",
			body:        "",
			want:        "xml",
		},
		{
			name:        "html_content_type",
			contentType: "text/html",
			body:        "",
			want:        "html",
		},
		{
			name:        "graphql_content_type",
			contentType: "application/graphql",
			body:        "",
			want:        "graphql",
		},
		{
			name:        "content_type_case_insensitive",
			contentType: "Application/JSON; charset=utf-8",
			body:        "",
			want:        "json",
		},
		// Body sniffing when no content-type is present.
		{
			name:        "json_object_body",
			contentType: "",
			body:        `{"key": "value"}`,
			want:        "json",
		},
		{
			name:        "json_array_body",
			contentType: "",
			body:        `[1, 2, 3]`,
			want:        "json",
		},
		{
			name:        "xml_body",
			contentType: "",
			body:        `<root><child/></root>`,
			want:        "xml",
		},
		{
			name:        "yaml_body_with_separator",
			contentType: "",
			body:        "---\nkey: value",
			want:        "yaml",
		},
		{
			name:        "graphql_query_body",
			contentType: "",
			body:        "query { user { id } }",
			want:        "graphql",
		},
		{
			name:        "graphql_mutation_body",
			contentType: "",
			body:        "mutation CreateUser { ... }",
			want:        "graphql",
		},
		{
			name:        "empty_body_returns_text",
			contentType: "",
			body:        "",
			want:        "text",
		},
		{
			name:        "unrecognised_body_returns_text",
			contentType: "",
			body:        "hello world",
			want:        "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectSyntax(tt.contentType, tt.body)
			if got != tt.want {
				t.Fatalf("DetectSyntax(%q, %q) = %q, want %q", tt.contentType, tt.body, got, tt.want)
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
			in:   512,
			want: "512B",
		},
		{
			name: "zero_uses_bytes",
			in:   0,
			want: "0B",
		},
		{
			name: "exactly_one_kb_uses_kb",
			in:   1024,
			want: "1.0KB",
		},
		{
			name: "mid_kb_range_uses_kb",
			in:   2048,
			want: "2.0KB",
		},
		{
			name: "just_below_mb_uses_kb",
			in:   1024*1024 - 1,
			want: "1024.0KB",
		},
		{
			name: "exactly_one_mb_uses_mb",
			in:   1024 * 1024,
			want: "1.0MB",
		},
		{
			name: "large_mb_value",
			in:   10 * 1024 * 1024,
			want: "10.0MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSize(tt.in)
			if got != tt.want {
				t.Fatalf("FormatSize(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
