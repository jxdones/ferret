package bodyeditor

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jxdones/ferret/internal/tui/tuitest"
)

func keyPress(k tea.Key) tea.KeyPressMsg {
	return tea.KeyPressMsg(k)
}

func TestNew_Defaults(t *testing.T) {
	m := New()
	if m.Syntax() != SyntaxText {
		t.Fatalf("Syntax() = %q, want %q", m.Syntax(), SyntaxText)
	}
}

func TestSetValueAndSyntax(t *testing.T) {
	m := New()
	m.SetSyntax(SyntaxJSON)
	m.SetValue("{\"ok\":true}")
	if got := m.Value(); got != "{\"ok\":true}" {
		t.Fatalf("Value() = %q", got)
	}
	if got := m.Syntax(); got != SyntaxJSON {
		t.Fatalf("Syntax() = %q, want %q", got, SyntaxJSON)
	}
}

func TestUpdate_HandlesTyping(t *testing.T) {
	m := New()
	m.SetFocused(true)

	m, _ = m.Update(keyPress(tea.Key{Code: 'a', Text: "a"}))
	m, _ = m.Update(keyPress(tea.Key{Code: 'b', Text: "b"}))

	if got := m.Value(); got != "ab" {
		t.Fatalf("Value() = %q, want ab", got)
	}
}

func TestView_EmptyShowsPlaceholder(t *testing.T) {
	tuitest.UseStableTheme(t)
	m := New()
	out := tuitest.StripANSI(m.View())
	if !strings.Contains(out, "request body") {
		t.Fatalf("View() = %q, want placeholder", out)
	}
}

// rs builds a []runeStyle from a plain string using a zero Style, for test convenience.
func rs(s string) []runeStyle {
	var out []runeStyle
	for _, r := range s {
		out = append(out, runeStyle{r: r, style: lipgloss.NewStyle()})
	}
	return out
}

// rsText extracts the plain rune text from a []runeStyle row.
func rsText(row []runeStyle) string {
	var sb strings.Builder
	for _, r := range row {
		sb.WriteRune(r.r)
	}
	return sb.String()
}

func TestWrapRuneStyles(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		textWidth int
		wantRows  []string
	}{
		{
			name:      "empty_input_returns_single_empty_row",
			input:     "",
			textWidth: 10,
			wantRows:  []string{""},
		},
		{
			name:      "zero_width_returns_input_unchanged",
			input:     "hello",
			textWidth: 0,
			wantRows:  []string{"hello"},
		},
		{
			name:      "short_text_fits_in_one_row",
			input:     "hello",
			textWidth: 10,
			wantRows:  []string{"hello "},
		},
		{
			name:      "text_wraps_at_word_boundary",
			input:     "hello world",
			textWidth: 8,
			wantRows:  []string{"hello ", "world "},
		},
		{
			// "averylongword" (13 chars) exceeds width 10, so it is split:
			// first 10 chars land on their own row, remainder on the next.
			name:      "word_longer_than_width_splits_across_rows",
			input:     "short averylongword",
			textWidth: 10,
			wantRows:  []string{"short ", "averylongw", "ord "},
		},
		{
			// The trailing-space logic needs one extra column beyond the word,
			// so width must be > len(word) to fit in a single row.
			name:      "word_with_room_for_trailing_space_fits_in_one_row",
			input:     "hello",
			textWidth: 6,
			wantRows:  []string{"hello "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := wrapRuneStyles(rs(tt.input), tt.textWidth)
			if len(rows) != len(tt.wantRows) {
				t.Fatalf("wrapRuneStyles(%q, %d): got %d rows, want %d", tt.input, tt.textWidth, len(rows), len(tt.wantRows))
			}
			for i, row := range rows {
				got := rsText(row)
				if got != tt.wantRows[i] {
					t.Fatalf("row[%d] = %q, want %q", i, got, tt.wantRows[i])
				}
			}
		})
	}
}

func TestBuildLines(t *testing.T) {
	tuitest.UseStableTheme(t)

	plain := func(lines []visualLine) string {
		var sb strings.Builder
		for _, vl := range lines {
			sb.WriteString(tuitest.StripANSI(vl.content))
			sb.WriteString("\n")
		}
		return strings.TrimRight(sb.String(), "\n")
	}

	tests := []struct {
		name            string
		input           string
		syntax          Syntax
		textWidth       int
		cursorHardRow   int
		cursorSubRow    int
		cursorColOffset int
		focused         bool
		wantLines       []string // plain-text content of each visual line
	}{
		{
			name:      "single_line_unfocused",
			input:     "hello",
			syntax:    SyntaxText,
			textWidth: 20,
			focused:   false,
			wantLines: []string{"hello "},
		},
		{
			name:            "cursor_on_first_char",
			input:           "hello",
			syntax:          SyntaxText,
			textWidth:       20,
			cursorHardRow:   0,
			cursorSubRow:    0,
			cursorColOffset: 0,
			focused:         true,
			wantLines:       []string{"hello "},
		},
		{
			name:            "cursor_past_end_appends_space",
			input:           "hi",
			syntax:          SyntaxText,
			textWidth:       20,
			cursorHardRow:   0,
			cursorSubRow:    0,
			cursorColOffset: 2,
			focused:         true,
			wantLines:       []string{"hi "},
		},
		{
			name:      "multiline_produces_one_visual_line_per_hard_line",
			input:     "line one\nline two",
			syntax:    SyntaxText,
			textWidth: 20,
			focused:   false,
			wantLines: []string{"line one ", "line two "},
		},
		{
			name:            "cursor_on_second_hard_line",
			input:           "first\nsecond",
			syntax:          SyntaxText,
			textWidth:       20,
			cursorHardRow:   1,
			cursorSubRow:    0,
			cursorColOffset: 0,
			focused:         true,
			wantLines:       []string{"first ", "second "},
		},
		{
			name:      "long_line_wraps_into_two_visual_lines",
			input:     "hello world",
			syntax:    SyntaxText,
			textWidth: 8,
			focused:   false,
			wantLines: []string{"hello ", "world "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs := tokenize(tt.syntax, tt.input)
			lines := buildLines(segs, tt.textWidth, tt.cursorHardRow, tt.cursorSubRow, tt.cursorColOffset, tt.focused)

			if len(lines) != len(tt.wantLines) {
				t.Fatalf("buildLines: got %d lines, want %d\ngot:  %q\nwant: %q",
					len(lines), len(tt.wantLines), plain(lines), tt.wantLines)
			}
			for i, vl := range lines {
				got := tuitest.StripANSI(vl.content)
				if got != tt.wantLines[i] {
					t.Fatalf("line[%d] = %q, want %q", i, got, tt.wantLines[i])
				}
			}
		})
	}
}
