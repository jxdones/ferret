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

// DetectSyntax infers a syntax name from a Content-Type header value and body content.
// It checks the Content-Type value first, then falls back to body sniffing.
// Returns a canonical name: "json", "yaml", "xml", "html", "graphql", or "text".
func DetectSyntax(contentType, body string) string {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "json"):
		return "json"
	case strings.Contains(ct, "yaml"):
		return "yaml"
	case strings.Contains(ct, "xml"):
		return "xml"
	case strings.Contains(ct, "html"):
		return "html"
	case strings.Contains(ct, "graphql"):
		return "graphql"
	}

	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return "text"
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return "json"
	}
	if strings.HasPrefix(trimmed, "<") {
		return "xml"
	}
	if strings.HasPrefix(trimmed, "---") || strings.Contains(trimmed, ": ") {
		return "yaml"
	}
	first := strings.Fields(trimmed)
	if len(first) > 0 {
		switch first[0] {
		case "query", "mutation", "subscription", "fragment":
			return "graphql"
		}
	}
	return "text"
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
