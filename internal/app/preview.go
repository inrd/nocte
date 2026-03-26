package app

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) previewVisible() bool {
	if !m.previewEnabled {
		return false
	}

	availableWidth := 64
	if m.width > 0 {
		availableWidth = max(minEditorPaneWidth, m.width-12)
	}

	return availableWidth >= minEditorPaneWidth+minPreviewPaneWidth+editorPaneGap
}

func (m Model) editorPaneWidth() int {
	width := 64
	if m.width > 0 {
		width = max(minEditorPaneWidth, m.width-12)
	}

	if !m.previewVisible() {
		return width
	}

	return max(minEditorPaneWidth, (width-editorPaneGap)/2)
}

func (m Model) previewPaneWidth() int {
	if !m.previewVisible() {
		return 0
	}

	width := 64
	if m.width > 0 {
		width = max(minEditorPaneWidth, m.width-12)
	}

	return max(minPreviewPaneWidth, width-editorPaneGap-m.editorPaneWidth())
}

func (m *Model) togglePreview() {
	m.previewEnabled = !m.previewEnabled
	m.resizeEditor()
}

func (m Model) previewContent() string {
	if !m.previewVisible() {
		return ""
	}

	rendered, offsets := renderMarkdownPreviewLines(m.editor.Value(), m.previewPaneWidth())
	if len(rendered) == 0 {
		return ""
	}

	height := max(1, m.editor.Height())
	start := 0
	line := max(0, min(m.editor.Line(), len(offsets)-1))
	if len(offsets) > 0 {
		start = offsets[line]
	}
	if start > 0 {
		start = max(0, start-height/3)
	}

	end := min(len(rendered), start+height)
	return strings.Join(rendered[start:end], "\n")
}

func renderMarkdownPreview(content string, width int) string {
	lines, _ := renderMarkdownPreviewLines(content, width)
	return strings.Join(lines, "\n")
}

func renderMarkdownPreviewLines(content string, width int) ([]string, []int) {
	if width <= 0 {
		return nil, nil
	}

	if strings.TrimSpace(content) == "" {
		return []string{helpStyle.Render("Preview updates as you write Markdown.")}, []int{0}
	}

	lines := strings.Split(content, "\n")
	rendered := make([]string, 0, len(lines))
	offsets := make([]int, len(lines))
	inCodeBlock := false

	for i, line := range lines {
		offsets[i] = len(rendered)
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "```"):
			inCodeBlock = !inCodeBlock
			rendered = appendWrapped(rendered, previewMutedStyle, trimmed, width)
		case inCodeBlock:
			rendered = appendWrapped(rendered, previewCodeStyle, line, width)
		case trimmed == "":
			rendered = append(rendered, "")
		case headingLevel(trimmed) > 0:
			level := headingLevel(trimmed)
			text := strings.TrimSpace(trimmed[level:])
			if text == "" {
				text = trimmed
			}
			rendered = appendWrapped(rendered, previewHeadingStyle, text, width)
		case strings.HasPrefix(trimmed, ">"):
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			rendered = appendPrefixedWrapped(rendered, previewQuoteStyle, "│ ", "│ ", text, width)
		case isBulletLine(trimmed):
			rendered = appendPrefixedWrapped(rendered, lipgloss.NewStyle(), "• ", "  ", strings.TrimSpace(trimmed[2:]), width)
		case isNumberedLine(trimmed):
			rendered = appendWrapped(rendered, lipgloss.NewStyle(), trimmed, width)
		case isRuleLine(trimmed):
			rendered = append(rendered, previewMutedStyle.Render(strings.Repeat("─", width)))
		default:
			rendered = appendWrapped(rendered, lipgloss.NewStyle(), line, width)
		}
	}

	return rendered, offsets
}

func appendWrapped(lines []string, style lipgloss.Style, content string, width int) []string {
	segments := renderInlineMarkdownSegments(style, content)
	wrapped := wrapPreviewSegments(segments, width)
	lines = append(lines, wrapped...)
	return lines
}

func appendPrefixedWrapped(lines []string, style lipgloss.Style, prefix string, continuationPrefix string, content string, width int) []string {
	availableWidth := max(1, width-ansi.StringWidth(prefix))
	segments := renderInlineMarkdownSegments(style, content)
	wrapped := wrapPreviewSegments(segments, availableWidth)
	for i, line := range wrapped {
		currentPrefix := prefix
		if i > 0 {
			currentPrefix = continuationPrefix
		}
		lines = append(lines, style.Render(currentPrefix)+line)
	}
	return lines
}

func wrapPreviewLine(content string, width int) []string {
	if width <= 0 {
		return nil
	}
	if content == "" {
		return []string{""}
	}

	wrapped := ansi.Wordwrap(content, width, "")
	wrapped = ansi.Hardwrap(wrapped, width, true)
	return strings.Split(wrapped, "\n")
}

func renderInlineMarkdown(base lipgloss.Style, content string) string {
	segments := renderInlineMarkdownSegments(base, content)
	var b strings.Builder
	for _, segment := range segments {
		b.WriteString(segment.style.Render(segment.text))
	}
	return b.String()
}

type previewSegment struct {
	text  string
	style lipgloss.Style
}

func renderInlineMarkdownSegments(base lipgloss.Style, content string) []previewSegment {
	segments := make([]previewSegment, 0)

	appendSegment := func(style lipgloss.Style, text string) {
		if text == "" {
			return
		}
		segments = append(segments, previewSegment{text: text, style: style})
	}

	for len(content) > 0 {
		switch {
		case strings.HasPrefix(content, "`"):
			end := strings.Index(content[1:], "`")
			if end < 0 {
				appendSegment(base, content)
				return segments
			}

			code := content[1 : end+1]
			appendSegment(previewInlineCodeStyle.Inherit(base), code)
			content = content[end+2:]
		case strings.HasPrefix(content, "["):
			labelEnd := strings.Index(content, "](")
			if labelEnd < 0 {
				appendSegment(base, content[:1])
				content = content[1:]
				continue
			}

			urlEnd := strings.Index(content[labelEnd+2:], ")")
			if urlEnd < 0 {
				appendSegment(base, content[:1])
				content = content[1:]
				continue
			}

			label := content[1:labelEnd]
			appendSegment(previewLinkLabelStyle.Inherit(base), label)
			content = content[labelEnd+3+urlEnd:]
		default:
			nextSpecial := nextInlineSpecial(content)
			appendSegment(base, content[:nextSpecial])
			content = content[nextSpecial:]
		}
	}

	return segments
}

func nextInlineSpecial(content string) int {
	for i, r := range content {
		if r == '`' || r == '[' {
			return i
		}
	}

	return len(content)
}

func wrapPreviewSegments(segments []previewSegment, width int) []string {
	if width <= 0 {
		return nil
	}
	if len(segments) == 0 {
		return []string{""}
	}

	tokens := make([]previewSegment, 0, len(segments))
	for _, segment := range segments {
		tokens = append(tokens, splitPreviewSegment(segment)...)
	}

	lines := make([]string, 0, 1)
	var line strings.Builder
	lineWidth := 0

	flush := func() {
		lines = append(lines, line.String())
		line.Reset()
		lineWidth = 0
	}

	for _, token := range tokens {
		if token.text == "" {
			continue
		}

		tokenWidth := ansi.StringWidth(token.text)
		if tokenWidth == 0 {
			line.WriteString(token.style.Render(token.text))
			continue
		}

		if lineWidth == 0 && strings.TrimSpace(token.text) == "" {
			continue
		}

		if lineWidth > 0 && lineWidth+tokenWidth > width {
			flush()
			if strings.TrimSpace(token.text) == "" {
				continue
			}
		}

		if tokenWidth <= width {
			line.WriteString(token.style.Render(token.text))
			lineWidth += tokenWidth
			continue
		}

		chunks := hardWrapSegment(token, width)
		for i, chunk := range chunks {
			if i > 0 {
				flush()
			}
			line.WriteString(chunk.style.Render(chunk.text))
			lineWidth += ansi.StringWidth(chunk.text)
		}
	}

	if line.Len() > 0 || len(lines) == 0 {
		flush()
	}

	return lines
}

func splitPreviewSegment(segment previewSegment) []previewSegment {
	if segment.text == "" {
		return nil
	}

	pieces := make([]previewSegment, 0)
	var current strings.Builder
	currentSpace := unicode.IsSpace([]rune(segment.text)[0])

	flush := func() {
		if current.Len() == 0 {
			return
		}
		pieces = append(pieces, previewSegment{text: current.String(), style: segment.style})
		current.Reset()
	}

	for _, r := range segment.text {
		isSpace := unicode.IsSpace(r)
		if current.Len() > 0 && isSpace != currentSpace {
			flush()
		}
		current.WriteRune(r)
		currentSpace = isSpace
	}
	flush()

	return pieces
}

func hardWrapSegment(segment previewSegment, width int) []previewSegment {
	if width <= 0 {
		return nil
	}

	chunks := make([]previewSegment, 0)
	var current strings.Builder
	currentWidth := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, previewSegment{text: current.String(), style: segment.style})
		current.Reset()
		currentWidth = 0
	}

	for _, r := range segment.text {
		rw := ansi.StringWidth(string(r))
		if currentWidth > 0 && currentWidth+rw > width {
			flush()
		}
		current.WriteRune(r)
		currentWidth += rw
	}
	flush()

	return chunks
}

func headingLevel(line string) int {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level >= len(line) || line[level] != ' ' {
		return 0
	}
	return level
}

func isBulletLine(line string) bool {
	return strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ")
}

func isNumberedLine(line string) bool {
	parts := strings.SplitN(line, ". ", 2)
	if len(parts) != 2 || parts[0] == "" {
		return false
	}

	_, err := strconv.Atoi(parts[0])
	return err == nil
}

func isRuleLine(line string) bool {
	if len(line) < 3 {
		return false
	}

	return line == strings.Repeat("-", len(line)) || line == strings.Repeat("*", len(line))
}
