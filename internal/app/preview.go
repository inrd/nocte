package app

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

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

func (m Model) previewPaneContentWidth() int {
	return max(1, m.previewPaneWidth()-inputStyle.GetHorizontalFrameSize())
}

func (m *Model) togglePreview() {
	m.previewEnabled = !m.previewEnabled
	m.resizeEditor()
}

func (m Model) previewContent() string {
	if !m.previewVisible() {
		return ""
	}

	rendered, offsets := renderMarkdownPreviewLinesForNote(m.editor.Value(), m.editorPath, m.previewPaneContentWidth())
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
	return renderMarkdownPreviewLinesForNote(content, "", width)
}

func renderMarkdownPreviewLinesForNote(content string, notePath string, width int) ([]string, []int) {
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
	completedParentIndent := -1

	for i, line := range lines {
		offsets[i] = len(rendered)
		trimmed := strings.TrimSpace(line)
		lineIndent := leadingVisualIndent(line)

		switch {
		case strings.HasPrefix(trimmed, "```"):
			inCodeBlock = !inCodeBlock
			completedParentIndent = -1
			rendered = appendWrapped(rendered, previewMutedStyle, trimmed, width)
		case inCodeBlock:
			rendered = appendWrapped(rendered, previewCodeStyle, line, width)
		case isMarkdownImageLine(trimmed):
			completedParentIndent = -1
			rendered = append(rendered, renderMarkdownImagePreview(notePath, trimmed, width)...)
		case trimmed == "":
			completedParentIndent = -1
			rendered = append(rendered, "")
		case headingLevel(trimmed) > 0:
			completedParentIndent = -1
			level := headingLevel(trimmed)
			text := strings.TrimSpace(trimmed[level:])
			if text == "" {
				text = trimmed
			}
			rendered = appendWrapped(rendered, previewHeadingStyle, text, width)
		case strings.HasPrefix(trimmed, ">"):
			completedParentIndent = -1
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
			rendered = appendPrefixedWrapped(rendered, previewQuoteStyle, "│ ", "│ ", text, width)
		case isTaskListLine(line):
			indent, marker, content := parseTaskListLine(line)
			if isCheckedTaskMarker(marker) {
				completedParentIndent = indent
			} else {
				completedParentIndent = -1
			}
			prefix := strings.Repeat(" ", indent) + marker
			continuationPrefix := strings.Repeat(" ", indent+ansi.StringWidth(marker))
			style := lipgloss.NewStyle()
			if isCheckedTaskMarker(marker) {
				style = previewCompletedTaskStyle
			}
			rendered = appendPrefixedWrapped(rendered, style, prefix, continuationPrefix, content, width)
		case isBulletLine(line):
			childOfCompleted := completedParentIndent >= 0 && lineIndent > completedParentIndent
			if !childOfCompleted {
				completedParentIndent = -1
			}
			indent, markerWidth, content := parseBulletLine(line)
			prefix := strings.Repeat(" ", indent) + "• "
			continuationPrefix := strings.Repeat(" ", indent+markerWidth)
			style := lipgloss.NewStyle()
			if childOfCompleted {
				style = previewCompletedTaskStyle
			}
			rendered = appendPrefixedWrapped(rendered, style, prefix, continuationPrefix, content, width)
		case isNumberedLine(line):
			childOfCompleted := completedParentIndent >= 0 && lineIndent > completedParentIndent
			if !childOfCompleted {
				completedParentIndent = -1
			}
			indent, marker, content := parseNumberedLine(line)
			prefix := strings.Repeat(" ", indent) + marker
			continuationPrefix := strings.Repeat(" ", indent+ansi.StringWidth(marker))
			style := lipgloss.NewStyle()
			if childOfCompleted {
				style = previewCompletedTaskStyle
			}
			rendered = appendPrefixedWrapped(rendered, style, prefix, continuationPrefix, content, width)
		case isRuleLine(trimmed):
			completedParentIndent = -1
			rendered = append(rendered, previewMutedStyle.Render(strings.Repeat("─", width)))
		default:
			childOfCompleted := completedParentIndent >= 0 && lineIndent > completedParentIndent
			if !childOfCompleted {
				completedParentIndent = -1
			}
			style := lipgloss.NewStyle()
			if childOfCompleted {
				style = previewCompletedTaskStyle
			}
			rendered = appendWrapped(rendered, style, line, width)
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

func isCheckedTaskMarker(marker string) bool {
	return strings.Contains(marker, "☑")
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
		case strings.HasPrefix(content, "~~"):
			strikethroughStyle, inner, consumed, ok := parseInlineDelimitedStyle(content, "~~", previewStrikethroughStyle)
			if !ok {
				appendSegment(base, content[:1])
				content = content[1:]
				continue
			}

			segments = append(segments, renderInlineMarkdownSegments(strikethroughStyle.Inherit(base), inner)...)
			content = content[consumed:]
		case strings.HasPrefix(content, "**"), strings.HasPrefix(content, "__"), strings.HasPrefix(content, "*"), strings.HasPrefix(content, "_"):
			emphasisStyle, inner, consumed, ok := parseInlineEmphasis(content)
			if !ok {
				appendSegment(base, content[:1])
				content = content[1:]
				continue
			}

			segments = append(segments, renderInlineMarkdownSegments(emphasisStyle.Inherit(base), inner)...)
			content = content[consumed:]
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
		if r == '`' || r == '[' || r == '*' || r == '_' || r == '~' {
			return i
		}
	}

	return len(content)
}

func parseInlineEmphasis(content string) (lipgloss.Style, string, int, bool) {
	switch {
	case strings.HasPrefix(content, "**"):
		return parseInlineDelimitedStyle(content, "**", previewBoldStyle)
	case strings.HasPrefix(content, "__"):
		return parseInlineDelimitedStyle(content, "__", previewBoldStyle)
	case strings.HasPrefix(content, "*"):
		return parseInlineDelimitedStyle(content, "*", previewItalicStyle)
	case strings.HasPrefix(content, "_"):
		return parseInlineDelimitedStyle(content, "_", previewItalicStyle)
	default:
		return lipgloss.Style{}, "", 0, false
	}
}

func parseInlineDelimitedStyle(content string, delimiter string, style lipgloss.Style) (lipgloss.Style, string, int, bool) {
	if len(content) <= len(delimiter) {
		return lipgloss.Style{}, "", 0, false
	}

	innerStart := len(delimiter)
	if unicode.IsSpace(rune(content[innerStart])) {
		return lipgloss.Style{}, "", 0, false
	}

	closing := strings.Index(content[innerStart:], delimiter)
	if closing < 0 {
		return lipgloss.Style{}, "", 0, false
	}

	inner := content[innerStart : innerStart+closing]
	if strings.TrimSpace(inner) == "" || unicode.IsSpace(rune(inner[len(inner)-1])) {
		return lipgloss.Style{}, "", 0, false
	}

	return style, inner, innerStart + closing + len(delimiter), true
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
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	return strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ")
}

func isTaskListLine(line string) bool {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	return strings.HasPrefix(trimmed, "- [ ] ") ||
		strings.HasPrefix(trimmed, "- [x] ") ||
		strings.HasPrefix(trimmed, "- [X] ") ||
		strings.HasPrefix(trimmed, "* [ ] ") ||
		strings.HasPrefix(trimmed, "* [x] ") ||
		strings.HasPrefix(trimmed, "* [X] ") ||
		strings.HasPrefix(trimmed, "+ [ ] ") ||
		strings.HasPrefix(trimmed, "+ [x] ") ||
		strings.HasPrefix(trimmed, "+ [X] ")
}

func isOpenTaskListLine(line string) bool {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	return strings.HasPrefix(trimmed, "- [ ] ") ||
		strings.HasPrefix(trimmed, "* [ ] ") ||
		strings.HasPrefix(trimmed, "+ [ ] ")
}

func isNumberedLine(line string) bool {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	parts := strings.SplitN(trimmed, ". ", 2)
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

func parseBulletLine(line string) (indent int, markerWidth int, content string) {
	indent = leadingVisualIndent(line)
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	if trimmed == "" {
		return indent, 2, ""
	}

	_, size := utf8.DecodeRuneInString(trimmed)
	content = strings.TrimSpace(trimmed[size:])
	return indent, 2, content
}

func parseTaskListLine(line string) (indent int, marker string, content string) {
	indent = leadingVisualIndent(line)
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	if len(trimmed) < 6 {
		return indent, "• ", trimmed
	}

	checked := trimmed[3] == 'x' || trimmed[3] == 'X'
	if checked {
		marker = "☑ "
	} else {
		marker = "☐ "
	}

	content = strings.TrimSpace(trimmed[6:])
	return indent, marker, content
}

func parseNumberedLine(line string) (indent int, marker string, content string) {
	indent = leadingVisualIndent(line)
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	parts := strings.SplitN(trimmed, ". ", 2)
	if len(parts) != 2 {
		return indent, "", trimmed
	}

	marker = parts[0] + ". "
	content = strings.TrimSpace(parts[1])
	return indent, marker, content
}

func leadingVisualIndent(line string) int {
	width := 0
	for _, r := range line {
		if !unicode.IsSpace(r) {
			break
		}
		if r == '\t' {
			width += 4
			continue
		}
		width++
	}
	return width
}
