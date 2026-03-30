// Package bodyeditor provides a textarea with syntax-aware rendering for request bodies.
package bodyeditor

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/rivo/uniseg"

	"github.com/jxdones/ferret/internal/tui/theme"
)

// Syntax controls which lexer is used to render the body.
type Syntax string

const (
	SyntaxText    Syntax = "text"
	SyntaxJSON    Syntax = "json"
	SyntaxYAML    Syntax = "yaml"
	SyntaxGraphQL Syntax = "graphql"

	// maxBodyChars caps paste/typing size (~1 MiB of ASCII) for the request body editor.
	maxBodyChars = 1 << 20

	// contentLeftMargin is the fixed left gutter for the body editor.
	contentLeftMargin = 0
)

type segment struct {
	text  string
	style lipgloss.Style
}

type runeStyle struct {
	r     rune
	style lipgloss.Style
}

type visualLine struct {
	content  string
	hardLine int
	subRow   int
}

// Model wraps textarea.Model with syntax-aware rendering.
type Model struct {
	input  textarea.Model
	syntax Syntax
}

// New creates a body editor with the default textarea configuration.
func New() Model {
	ta := textarea.New()
	ta.Prompt = ""
	ta.Placeholder = "request body..."
	ta.ShowLineNumbers = true
	ta.CharLimit = maxBodyChars
	ta.MaxHeight = 999
	ta.SetStyles(themedStyles(ta.Styles()))

	return Model{
		input:  ta,
		syntax: SyntaxText,
	}
}

// Value returns the current body text.
func (m Model) Value() string {
	return m.input.Value()
}

// SetValue updates the body text without changing the syntax mode.
func (m *Model) SetValue(value string) {
	m.input.SetValue(value)
}

// Syntax returns the current syntax mode.
func (m Model) Syntax() Syntax {
	return m.syntax
}

// SetSyntax updates the syntax mode used for rendering.
func (m *Model) SetSyntax(syntax Syntax) {
	m.syntax = syntax
}

// SetSize sets the body editor width and height.
func (m *Model) SetSize(width, height int) {
	inner := max(1, width-contentLeftMargin)
	m.input.SetWidth(inner)
	m.input.SetHeight(max(1, height))
}

// InsertString inserts a string at the current cursor position.
func (m *Model) InsertString(s string) {
	m.input.InsertString(s)
}

// SetFocused focuses or blurs the underlying textarea.
func (m *Model) SetFocused(focused bool) {
	if focused {
		m.input.Focus()
		return
	}
	m.input.Blur()
}

// Focused reports whether the body editor is focused.
func (m Model) Focused() bool {
	return m.input.Focused()
}

// Update delegates input handling to the textarea.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the textarea with syntax highlighting.
func (m Model) View() string {
	lead := ""
	value := m.input.Value()
	if value == "" {
		return prefixEachLine(m.input.View(), lead)
	}

	textWidth := max(1, m.input.Width())
	prompt := m.input.Prompt
	showLineNumbers := m.input.ShowLineNumbers
	lineNumDigits := len(strconv.Itoa(max(1, m.input.MaxHeight)))

	li := m.input.LineInfo()
	segments := tokenize(m.syntax, value)
	visualLines := buildLines(segments, textWidth, m.input.Line(), li.RowOffset, li.ColumnOffset, m.input.Focused())

	height := max(1, m.input.Height())
	scrollOffset := m.input.ScrollYOffset()
	start := min(scrollOffset, len(visualLines))
	end := min(start+height, len(visualLines))
	visible := visualLines[start:end]

	promptStyle := lipgloss.NewStyle().Foreground(theme.Current.TextAccent)
	lineNumStyle := lipgloss.NewStyle().Foreground(theme.Current.TextMuted)

	var sb strings.Builder
	for i, vl := range visible {
		sb.WriteString(lead)
		sb.WriteString(promptStyle.Render(prompt))

		if showLineNumbers {
			var lineNumStr string
			if vl.subRow == 0 {
				lineNumStr = fmt.Sprintf("%*d ", lineNumDigits, vl.hardLine+1)
			} else {
				lineNumStr = fmt.Sprintf(" %*s ", lineNumDigits, "")
			}
			sb.WriteString(lineNumStyle.Render(lineNumStr))
		}

		sb.WriteString(lipgloss.NewStyle().Width(textWidth).Render(vl.content))

		if i < len(visible)-1 {
			sb.WriteRune('\n')
		}
	}

	for i := len(visible); i < height; i++ {
		if i > 0 {
			sb.WriteRune('\n')
		}
		sb.WriteString(lead)
		sb.WriteString(promptStyle.Render(prompt))
		if showLineNumbers {
			sb.WriteString(lineNumStyle.Render(fmt.Sprintf(" %*s ", lineNumDigits, "")))
		}
	}

	return sb.String()
}

func prefixEachLine(s, prefix string) string {
	if s == "" {
		return prefix
	}
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

// themedStyles returns the styles for the body editor.
func themedStyles(styles textarea.Styles) textarea.Styles {
	styles.Focused.Prompt = styles.Focused.Prompt.Foreground(theme.Current.TextAccent)
	styles.Focused.Text = styles.Focused.Text.Foreground(theme.Current.TextPrimary)
	styles.Focused.CursorLine = styles.Focused.CursorLine.Foreground(theme.Current.TextPrimary)
	styles.Focused.CursorLineNumber = styles.Focused.CursorLineNumber.Foreground(theme.Current.TextMuted)
	styles.Focused.LineNumber = styles.Focused.LineNumber.Foreground(theme.Current.TextMuted)
	styles.Focused.Placeholder = styles.Focused.Placeholder.Foreground(theme.Current.TextMuted)
	styles.Blurred.Prompt = styles.Blurred.Prompt.Foreground(theme.Current.TextMuted)
	styles.Blurred.Text = styles.Blurred.Text.Foreground(theme.Current.TextPrimary)
	styles.Blurred.CursorLine = styles.Blurred.CursorLine.Foreground(theme.Current.TextPrimary)
	styles.Blurred.CursorLineNumber = styles.Blurred.CursorLineNumber.Foreground(theme.Current.TextMuted)
	styles.Blurred.LineNumber = styles.Blurred.LineNumber.Foreground(theme.Current.TextMuted)
	styles.Blurred.Placeholder = styles.Blurred.Placeholder.Foreground(theme.Current.TextMuted)
	return styles
}

// tokenStyle returns the style for a chroma token type.
func tokenStyle(t chroma.TokenType) lipgloss.Style {
	switch {
	case t.InCategory(chroma.Keyword):
		return lipgloss.NewStyle().Foreground(theme.Current.SyntaxKeyword)
	case t.InCategory(chroma.LiteralString):
		return lipgloss.NewStyle().Foreground(theme.Current.SyntaxString)
	case t.InCategory(chroma.LiteralNumber):
		return lipgloss.NewStyle().Foreground(theme.Current.SyntaxNumber)
	case t.InCategory(chroma.Comment):
		return lipgloss.NewStyle().Foreground(theme.Current.SyntaxComment)
	case t.InCategory(chroma.Operator), t.InCategory(chroma.Punctuation):
		return lipgloss.NewStyle().Foreground(theme.Current.SyntaxOperator)
	default:
		return lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)
	}
}

// tokenize tokenizes the content into segments.
func tokenize(syntax Syntax, content string) []segment {
	if syntax == SyntaxText {
		return []segment{{text: content, style: lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)}}
	}

	lexer := lexers.Get(string(syntax))
	if lexer == nil {
		lexer = lexers.Fallback
	}
	iter, err := lexer.Tokenise(nil, content)
	if err != nil {
		return []segment{{text: content, style: lipgloss.NewStyle().Foreground(theme.Current.TextPrimary)}}
	}

	var segments []segment
	for tok := iter(); tok != chroma.EOF; tok = iter() {
		segments = append(segments, segment{
			text:  tok.Value,
			style: tokenStyle(tok.Type),
		})
	}
	return segments
}

// rsDisplayWidth calculates the display width of a slice of runeStyles.
func rsDisplayWidth(rs []runeStyle) int {
	var w int
	for _, r := range rs {
		w += uniseg.StringWidth(string(r.r))
	}
	return w
}

// wrapRuneStyles wraps the runeStyles into rows of the given width.
func wrapRuneStyles(rs []runeStyle, textWidth int) [][]runeStyle {
	if textWidth <= 0 || len(rs) == 0 {
		return [][]runeStyle{rs}
	}

	var (
		rows   = [][]runeStyle{{}}
		word   []runeStyle
		row    int
		spaces int
	)

	spaceStyle := func() lipgloss.Style {
		if len(word) > 0 {
			return word[len(word)-1].style
		}
		if len(rows[row]) > 0 {
			return rows[row][len(rows[row])-1].style
		}
		return lipgloss.NewStyle()
	}

	for _, rr := range rs {
		if unicode.IsSpace(rr.r) {
			spaces++
		} else {
			word = append(word, rr)
		}

		if spaces > 0 {
			ss := spaceStyle()
			if rsDisplayWidth(rows[row])+rsDisplayWidth(word)+spaces > textWidth {
				row++
				rows = append(rows, []runeStyle{})
				rows[row] = append(rows[row], word...)
				for range spaces {
					rows[row] = append(rows[row], runeStyle{' ', ss})
				}
			} else {
				rows[row] = append(rows[row], word...)
				for range spaces {
					rows[row] = append(rows[row], runeStyle{' ', ss})
				}
			}
			spaces = 0
			word = nil
		} else if len(word) > 0 {
			lastW := uniseg.StringWidth(string(word[len(word)-1].r))
			if rsDisplayWidth(word)+lastW > textWidth {
				if len(rows[row]) > 0 {
					row++
					rows = append(rows, []runeStyle{})
				}
				rows[row] = append(rows[row], word...)
				word = nil
			}
		}
	}

	ss := spaceStyle()
	if rsDisplayWidth(rows[row])+rsDisplayWidth(word)+spaces >= textWidth {
		row++
		rows = append(rows, []runeStyle{})
		rows[row] = append(rows[row], word...)
		spaces++
		for range spaces {
			rows[row] = append(rows[row], runeStyle{' ', ss})
		}
	} else {
		rows[row] = append(rows[row], word...)
		spaces++
		for range spaces {
			rows[row] = append(rows[row], runeStyle{' ', ss})
		}
	}

	return rows
}

// buildLines builds the visual lines for the body editor.
func buildLines(segments []segment, textWidth, cursorHardRow, cursorSubRow, cursorColOffset int, focused bool) []visualLine {
	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(theme.Current.TextPrimary)

	hardLines := [][]runeStyle{{}}
	lineIdx := 0
	for _, seg := range segments {
		for _, r := range seg.text {
			if r == '\n' {
				lineIdx++
				hardLines = append(hardLines, []runeStyle{})
				continue
			}
			hardLines[lineIdx] = append(hardLines[lineIdx], runeStyle{r: r, style: seg.style})
		}
	}

	var result []visualLine
	for hardIdx, hl := range hardLines {
		subRows := wrapRuneStyles(hl, textWidth)
		isCursorHardLine := hardIdx == cursorHardRow

		for subIdx, subRow := range subRows {
			isCursorRow := focused && isCursorHardLine && subIdx == cursorSubRow

			var sb strings.Builder
			if isCursorRow {
				colPos := 0
				cursorPlaced := false
				for _, rr := range subRow {
					if !cursorPlaced && colPos == cursorColOffset {
						sb.WriteString(cursorStyle.Render(string(rr.r)))
						cursorPlaced = true
					} else {
						sb.WriteString(rr.style.Render(string(rr.r)))
					}
					colPos++
				}
				if !cursorPlaced {
					sb.WriteString(cursorStyle.Render(" "))
				}
			} else {
				for _, rr := range subRow {
					sb.WriteString(rr.style.Render(string(rr.r)))
				}
			}

			result = append(result, visualLine{
				content:  sb.String(),
				hardLine: hardIdx,
				subRow:   subIdx,
			})
		}
	}

	return result
}
