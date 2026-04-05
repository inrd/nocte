package app

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type listState struct {
	indent   int
	ordered  bool
	liOpened bool
}

func (m *Model) exportEditorHTML() error {
	exportDir := filepath.Join(m.config.NotesPath, "html")
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("could not prepare HTML export dir: %w", err)
	}

	fileName := strings.TrimSuffix(m.editorName, filepath.Ext(m.editorName)) + ".html"
	exportPath := filepath.Join(exportDir, fileName)
	rendered := renderMarkdownHTMLDocument(m.editorName, m.editorPath, m.editor.Value())

	if err := os.WriteFile(exportPath, []byte(rendered), 0o644); err != nil {
		return fmt.Errorf("could not write HTML export: %w", err)
	}

	if err := openPathWithSystemApp(exportPath); err != nil {
		return fmt.Errorf("rendered HTML but could not open it: %w", err)
	}

	m.status = fmt.Sprintf("Opened HTML export html/%s", fileName)
	m.isError = false
	return nil
}

func (m *Model) exportAllNotesHTML() error {
	if err := os.MkdirAll(m.config.NotesPath, 0o755); err != nil {
		return fmt.Errorf("could not prepare notes dir: %w", err)
	}

	exportDir := filepath.Join(m.config.NotesPath, "html")
	if err := os.RemoveAll(exportDir); err != nil {
		return fmt.Errorf("could not clean HTML export dir: %w", err)
	}
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("could not prepare HTML export dir: %w", err)
	}

	notes := m.listNotes()
	for _, note := range notes {
		content, err := os.ReadFile(note.path)
		if err != nil {
			return fmt.Errorf("could not read %s: %w", note.name, err)
		}

		fileName := strings.TrimSuffix(note.name, filepath.Ext(note.name)) + ".html"
		exportPath := filepath.Join(exportDir, fileName)
		rendered := renderMarkdownHTMLDocument(note.name, note.path, string(content))
		if err := os.WriteFile(exportPath, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("could not write HTML export for %s: %w", note.name, err)
		}
	}

	m.status = fmt.Sprintf("Rendered %d notes to html", len(notes))
	m.isError = false
	return nil
}

func renderMarkdownHTMLDocument(noteName string, notePath string, content string) string {
	title := strings.TrimSuffix(noteName, filepath.Ext(noteName))
	if strings.TrimSpace(title) == "" {
		title = noteName
	}
	if strings.TrimSpace(title) == "" {
		title = "nocte note"
	}

	baseHref := noteBaseHref(notePath)
	body := renderMarkdownHTMLBody(content)

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n")
	b.WriteString("<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(title))
	b.WriteString("</title>\n")
	if baseHref != "" {
		b.WriteString("<base href=\"")
		b.WriteString(html.EscapeString(baseHref))
		b.WriteString("\">\n")
	}
	b.WriteString("<style>\n")
	b.WriteString("body{margin:0 auto;max-width:56rem;padding:2rem 1.25rem 4rem;font:16px/1.6 -apple-system,BlinkMacSystemFont,\"Segoe UI\",sans-serif;color:#1f2933;background:#f8fafc;}h1,h2,h3,h4,h5,h6{line-height:1.25;color:#102a43;}a{color:#0f766e;}pre,code{font-family:\"SFMono-Regular\",Consolas,\"Liberation Mono\",Menlo,monospace;}pre{overflow-x:auto;padding:1rem;border-radius:.75rem;background:#0f172a;color:#e2e8f0;}pre code{display:block;padding:0;border-radius:0;background:transparent;color:inherit;}code{padding:.1rem .3rem;border-radius:.35rem;background:#e2e8f0;color:#0f172a;}blockquote{margin:1.25rem 0;padding:0 1rem;border-left:.3rem solid #94a3b8;color:#475569;}hr{border:0;border-top:1px solid #cbd5e1;margin:2rem 0;}img{max-width:100%;height:auto;border-radius:.5rem;}table{border-collapse:collapse;}ul,ol{padding-left:1.5rem;}li+li{margin-top:.35rem;}input[type=checkbox]{margin-right:.45rem;}li.task-done{color:#64748b;text-decoration:line-through;}li.task-done input[type=checkbox]{accent-color:#64748b;}main{background:#ffffff;border-radius:1rem;padding:1.5rem;box-shadow:0 10px 30px rgba(15,23,42,.08);}\n")
	b.WriteString("</style>\n")
	b.WriteString("</head>\n<body>\n<main>\n")
	b.WriteString(body)
	b.WriteString("\n</main>\n</body>\n</html>\n")
	return b.String()
}

func noteBaseHref(notePath string) string {
	if notePath == "" {
		return ""
	}

	dir := filepath.Dir(notePath)
	if dir == "." || dir == "" {
		return ""
	}

	slashed := filepath.ToSlash(dir)
	if !strings.HasSuffix(slashed, "/") {
		slashed += "/"
	}

	return (&url.URL{Scheme: "file", Path: slashed}).String()
}

func renderMarkdownHTMLBody(content string) string {
	if strings.TrimSpace(content) == "" {
		return "<p></p>"
	}

	lines := strings.Split(content, "\n")
	var b strings.Builder
	paragraph := make([]string, 0)
	codeLines := make([]string, 0)
	lists := make([]listState, 0)
	inCodeBlock := false
	codeLanguage := ""
	completedParentIndent := -1

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		text := strings.Join(paragraph, " ")
		b.WriteString("<p>")
		b.WriteString(renderInlineMarkdownHTML(text))
		b.WriteString("</p>\n")
		paragraph = paragraph[:0]
	}

	closeListsTo := func(indent int) {
		for len(lists) > 0 && lists[len(lists)-1].indent >= indent {
			if lists[len(lists)-1].liOpened {
				b.WriteString("</li>\n")
			}
			if lists[len(lists)-1].ordered {
				b.WriteString("</ol>\n")
			} else {
				b.WriteString("</ul>\n")
			}
			lists = lists[:len(lists)-1]
		}
	}

	closeAllLists := func() {
		closeListsTo(-1)
	}

	flushCodeBlock := func() {
		if !inCodeBlock {
			return
		}
		b.WriteString("<pre><code")
		if codeLanguage != "" {
			b.WriteString(" class=\"language-")
			b.WriteString(html.EscapeString(codeLanguage))
			b.WriteString("\"")
		}
		b.WriteString(">")
		b.WriteString(html.EscapeString(strings.Join(codeLines, "\n")))
		b.WriteString("</code></pre>\n")
		codeLines = codeLines[:0]
		codeLanguage = ""
		inCodeBlock = false
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flushParagraph()
			closeAllLists()
			completedParentIndent = -1
			if inCodeBlock {
				flushCodeBlock()
				continue
			}
			inCodeBlock = true
			codeLanguage = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			codeLines = codeLines[:0]
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			flushParagraph()
			closeAllLists()
			completedParentIndent = -1
			block := make([]string, 0)
			for i < len(lines) {
				current := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(current, ">") {
					break
				}
				block = append(block, strings.TrimSpace(strings.TrimPrefix(current, ">")))
				i++
			}
			i--
			b.WriteString("<blockquote>\n")
			b.WriteString(renderMarkdownHTMLBody(strings.Join(block, "\n")))
			b.WriteString("\n</blockquote>\n")
			continue
		}

		if trimmed == "" {
			flushParagraph()
			closeAllLists()
			completedParentIndent = -1
			continue
		}

		if level := headingLevel(trimmed); level > 0 {
			flushParagraph()
			closeAllLists()
			completedParentIndent = -1
			text := strings.TrimSpace(trimmed[level:])
			if text == "" {
				text = trimmed
			}
			b.WriteString(fmt.Sprintf("<h%d>%s</h%d>\n", level, renderInlineMarkdownHTML(text), level))
			continue
		}

		if isRuleLine(trimmed) {
			flushParagraph()
			closeAllLists()
			completedParentIndent = -1
			b.WriteString("<hr>\n")
			continue
		}

		if isTaskListLine(line) || isBulletLine(line) || isNumberedLine(line) {
			flushParagraph()

			var (
				indent  int
				content string
				ordered bool
				prefix  string
			)

			switch {
			case isTaskListLine(line):
				var marker string
				indent, marker, content = parseTaskListLine(line)
				prefix = taskCheckboxHTML(marker)
				if isCheckedTaskMarker(marker) {
					completedParentIndent = indent
					prefix = "<span class=\"task-checkbox\">" + prefix + "</span>"
				} else {
					completedParentIndent = -1
				}
			case isBulletLine(line):
				var markerWidth int
				indent, markerWidth, content = parseBulletLine(line)
				_ = markerWidth
			default:
				var marker string
				indent, marker, content = parseNumberedLine(line)
				ordered = marker != ""
			}

			childOfCompleted := !isTaskListLine(line) && completedParentIndent >= 0 && indent > completedParentIndent

			for len(lists) > 0 && indent < lists[len(lists)-1].indent {
				closeListsTo(lists[len(lists)-1].indent)
			}
			if len(lists) > 0 && indent == lists[len(lists)-1].indent && lists[len(lists)-1].ordered != ordered {
				closeListsTo(indent)
			}
			if len(lists) == 0 || indent > lists[len(lists)-1].indent {
				if ordered {
					b.WriteString("<ol>\n")
				} else {
					b.WriteString("<ul>\n")
				}
				lists = append(lists, listState{indent: indent, ordered: ordered})
			} else if lists[len(lists)-1].liOpened {
				b.WriteString("</li>\n")
				lists[len(lists)-1].liOpened = false
			}

			if strings.Contains(prefix, "task-checkbox") || childOfCompleted {
				b.WriteString("<li class=\"task-done\">")
			} else {
				b.WriteString("<li>")
			}
			b.WriteString(prefix)
			b.WriteString(renderInlineMarkdownHTML(content))
			lists[len(lists)-1].liOpened = true
			continue
		}

		closeAllLists()
		completedParentIndent = -1
		paragraph = append(paragraph, trimmed)
	}

	flushParagraph()
	closeAllLists()
	flushCodeBlock()

	return strings.TrimSpace(b.String())
}

func taskCheckboxHTML(marker string) string {
	checked := strings.Contains(marker, "☑")
	if checked {
		return "<input type=\"checkbox\" checked disabled> "
	}
	return "<input type=\"checkbox\" disabled> "
}

func renderInlineMarkdownHTML(content string) string {
	var b strings.Builder

	for len(content) > 0 {
		switch {
		case strings.HasPrefix(content, "!["):
			alt, target, consumed, ok := parseInlineLinkLike(content, "![")
			if !ok {
				b.WriteString(html.EscapeString(content[:1]))
				content = content[1:]
				continue
			}
			b.WriteString("<img src=\"")
			b.WriteString(html.EscapeString(target))
			b.WriteString("\" alt=\"")
			b.WriteString(html.EscapeString(alt))
			b.WriteString("\">")
			content = content[consumed:]
		case strings.HasPrefix(content, "["):
			label, target, consumed, ok := parseInlineLinkLike(content, "[")
			if !ok {
				b.WriteString(html.EscapeString(content[:1]))
				content = content[1:]
				continue
			}
			b.WriteString("<a href=\"")
			b.WriteString(html.EscapeString(target))
			b.WriteString("\">")
			b.WriteString(renderInlineMarkdownHTML(label))
			b.WriteString("</a>")
			content = content[consumed:]
		case strings.HasPrefix(content, "!"):
			b.WriteString(html.EscapeString(content[:1]))
			content = content[1:]
		case strings.HasPrefix(content, "`"):
			end := strings.Index(content[1:], "`")
			if end < 0 {
				b.WriteString(html.EscapeString(content))
				return b.String()
			}
			b.WriteString("<code>")
			b.WriteString(html.EscapeString(content[1 : end+1]))
			b.WriteString("</code>")
			content = content[end+2:]
		case strings.HasPrefix(content, "~~"):
			inner, consumed, ok := parseInlineDelimitedContent(content, "~~")
			if !ok {
				b.WriteString(html.EscapeString(content[:1]))
				content = content[1:]
				continue
			}
			b.WriteString("<del>")
			b.WriteString(renderInlineMarkdownHTML(inner))
			b.WriteString("</del>")
			content = content[consumed:]
		case strings.HasPrefix(content, "**"), strings.HasPrefix(content, "__"):
			delimiter := content[:2]
			inner, consumed, ok := parseInlineDelimitedContent(content, delimiter)
			if !ok {
				b.WriteString(html.EscapeString(content[:1]))
				content = content[1:]
				continue
			}
			b.WriteString("<strong>")
			b.WriteString(renderInlineMarkdownHTML(inner))
			b.WriteString("</strong>")
			content = content[consumed:]
		case strings.HasPrefix(content, "*"), strings.HasPrefix(content, "_"):
			delimiter := content[:1]
			inner, consumed, ok := parseInlineDelimitedContent(content, delimiter)
			if !ok {
				b.WriteString(html.EscapeString(content[:1]))
				content = content[1:]
				continue
			}
			b.WriteString("<em>")
			b.WriteString(renderInlineMarkdownHTML(inner))
			b.WriteString("</em>")
			content = content[consumed:]
		default:
			next := nextInlineSpecialHTML(content)
			b.WriteString(html.EscapeString(content[:next]))
			content = content[next:]
		}
	}

	return b.String()
}

func nextInlineSpecialHTML(content string) int {
	for i, r := range content {
		if r == '`' || r == '[' || r == '*' || r == '_' || r == '~' {
			return i
		}
	}

	return len(content)
}

func parseInlineLinkLike(content string, prefix string) (label string, target string, consumed int, ok bool) {
	start := len(prefix)
	labelEnd := strings.Index(content[start:], "](")
	if labelEnd < 0 {
		return "", "", 0, false
	}
	labelEnd += start

	targetStart := labelEnd + 2
	targetEnd := strings.Index(content[targetStart:], ")")
	if targetEnd < 0 {
		return "", "", 0, false
	}
	targetEnd += targetStart

	return content[start:labelEnd], content[targetStart:targetEnd], targetEnd + 1, true
}

func parseInlineDelimitedContent(content string, delimiter string) (string, int, bool) {
	if len(content) <= len(delimiter) {
		return "", 0, false
	}

	innerStart := len(delimiter)
	if strings.TrimSpace(content[innerStart:innerStart+1]) == "" {
		return "", 0, false
	}

	closing := strings.Index(content[innerStart:], delimiter)
	if closing < 0 {
		return "", 0, false
	}

	inner := content[innerStart : innerStart+closing]
	if strings.TrimSpace(inner) == "" || strings.TrimSpace(inner[len(inner)-1:]) == "" {
		return "", 0, false
	}

	return inner, innerStart + closing + len(delimiter), true
}
