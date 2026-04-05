package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// NormalizeCanvas clips/pads content to an exact width x height rectangle.
// This prevents section-overflow artifacts when terminal space is tight.
func NormalizeCanvas(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	for i := range lines {
		line := ansi.Truncate(lines[i], width, "")
		lineWidth := ansi.StringWidth(line)
		if lineWidth < width {
			line += strings.Repeat(" ", width-lineWidth)
		}
		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

// ClampMin enforces a lower bound on a value.
// It ensures that a value is not less than a minimum.
func ClampMin(value, min int) int {
	if value < min {
		return min
	}
	return value
}

// FormatSize formats a byte count as a human-readable string.
// It selects the largest unit (B, KB, MB) that keeps the value >= 1,
// and formats KB/MB with one decimal place.
func FormatSize(b int64) string {
	const kb = 1024
	const mb = 1024 * 1024
	switch {
	case b >= mb:
		return fmt.Sprintf("%.1fMB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1fKB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%dB", b)
	}
}
