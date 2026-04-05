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
