package common

import "testing"

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
