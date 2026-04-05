package model

import "testing"

func TestClampTabTitle(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "short_ascii_unchanged",
			in:   "hello",
			want: "hello",
		},
		{
			name: "exact_10_ascii_unchanged",
			in:   "1234567890",
			want: "1234567890",
		},
		{
			name: "over_10_ascii_truncated",
			in:   "12345678901",
			want: "1234567890…",
		},
		{
			name: "accented_chars_count_as_one_column",
			in:   "café-résumé-extra",
			want: "café-résum…",
		},
		{
			name: "cjk_wide_chars_count_as_two_columns",
			// Each CJK char is 2 columns, so 5 chars = 10 cols exactly.
			in:   "你好世界吧",
			want: "你好世界吧",
		},
		{
			name: "cjk_wide_chars_over_limit",
			// 6 CJK chars = 12 cols — truncates after 5 (10 cols).
			in:   "你好世界吧啊",
			want: "你好世界吧…",
		},
		{
			name: "emoji_wide_chars_truncated",
			// Each emoji is typically 2 columns; 6 emoji = 12 cols.
			in:   "😀😁😂😃😄😅",
			want: "😀😁😂😃😄…",
		},
		{
			name: "empty_string_unchanged",
			in:   "",
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clampTabTitle(tc.in)
			if got != tc.want {
				t.Fatalf("clampTabTitle(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTitleFromMethodAndURL(t *testing.T) {
	tests := []struct {
		name   string
		method string
		rawURL string
		want   string
	}{
		{
			name:   "empty_url_returns_new_request",
			method: "GET",
			rawURL: "",
			want:   "new request",
		},
		{
			name:   "https_scheme_stripped",
			method: "GET",
			rawURL: "https://api.example.com/users",
			want:   "GET api.exampl…",
		},
		{
			name:   "http_scheme_stripped",
			method: "POST",
			rawURL: "http://api.example.com/users",
			want:   "POST api.exampl…",
		},
		{
			name:   "no_scheme_left_as_is",
			method: "GET",
			rawURL: "api.example.com/users",
			want:   "GET api.exampl…",
		},
		{
			name:   "short_url_not_truncated",
			method: "GET",
			rawURL: "https://foo.io",
			want:   "GET foo.io",
		},
		{
			name:   "cjk_url_truncated_by_display_width",
			method: "GET",
			rawURL: "https://你好世界吧啊",
			want:   "GET 你好世界吧…",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := titleFromMethodAndURL(tc.method, tc.rawURL)
			if got != tc.want {
				t.Fatalf("titleFromMethodAndURL(%q, %q) = %q, want %q", tc.method, tc.rawURL, got, tc.want)
			}
		})
	}
}
