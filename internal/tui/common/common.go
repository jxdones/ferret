package common

import (
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
