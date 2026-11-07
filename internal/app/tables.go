package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type tableAlign int

const (
	tableAlignLeft tableAlign = iota
	tableAlignCenter
	tableAlignRight
)

type markdownTable struct {
	header []string
	align  []tableAlign
	rows   [][]string
}

const previewTableColumnGap = " │ "

// parseMarkdownTable detects a pipe table starting at lines[start]. A table
// needs a `|`-prefixed header row followed by a separator row like `|---|:-:|`,
// and continues while lines start with `|`. It returns the index just past the
// last table line.
func parseMarkdownTable(lines []string, start int) (markdownTable, int, bool) {
	if start+1 >= len(lines) || !isTableRowLine(lines[start]) {
		return markdownTable{}, start, false
	}

	align, ok := parseTableSeparator(lines[start+1])
	if !ok {
		return markdownTable{}, start, false
	}

	header := splitTableRow(lines[start])
	if len(header) != len(align) {
		return markdownTable{}, start, false
	}

	table := markdownTable{header: header, align: align}
	end := start + 2
	for end < len(lines) && isTableRowLine(lines[end]) {
		table.rows = append(table.rows, normalizeTableRow(splitTableRow(lines[end]), len(header)))
		end++
	}

	return table, end, true
}

func isTableRowLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "|")
}

func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")

	cells := strings.Split(trimmed, "|")
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}
	return cells
}

func parseTableSeparator(line string) ([]tableAlign, bool) {
	if !isTableRowLine(line) {
		return nil, false
	}

	cells := splitTableRow(line)
	align := make([]tableAlign, len(cells))
	for i, cell := range cells {
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
		dashes := strings.Trim(cell, ":")
		if len(dashes) == 0 || strings.Trim(dashes, "-") != "" {
			return nil, false
		}

		switch {
		case left && right:
			align[i] = tableAlignCenter
		case right:
			align[i] = tableAlignRight
		default:
			align[i] = tableAlignLeft
		}
	}

	return align, true
}

func normalizeTableRow(cells []string, columns int) []string {
	if len(cells) >= columns {
		return cells[:columns]
	}

	padded := make([]string, columns)
	copy(padded, cells)
	return padded
}

// renderTablePreview renders a table as aligned columns. When the natural
// layout is wider than the pane, the widest columns shrink and their cells
// wrap. Tables too narrow to lay out fall back to their raw lines.
func renderTablePreview(table markdownTable, rawLines []string, width int) []string {
	columns := len(table.header)
	gapWidth := ansi.StringWidth(previewTableColumnGap)
	available := width - gapWidth*(columns-1)
	if available < columns*3 {
		rendered := make([]string, 0, len(rawLines))
		for _, line := range rawLines {
			rendered = appendWrapped(rendered, lipgloss.NewStyle(), line, width)
		}
		return rendered
	}

	widths := make([]int, columns)
	measure := func(cells []string) {
		for i, cell := range cells {
			widths[i] = max(widths[i], previewSegmentsWidth(renderInlineMarkdownSegments(lipgloss.NewStyle(), cell)))
		}
	}
	measure(table.header)
	for _, row := range table.rows {
		measure(row)
	}
	for i := range widths {
		widths[i] = max(1, widths[i])
	}

	for sumInts(widths) > available {
		widest := 0
		for i := range widths {
			if widths[i] > widths[widest] {
				widest = i
			}
		}
		widths[widest]--
	}

	rendered := renderTablePreviewRow(nil, table.header, table.align, widths, previewBoldStyle)

	separators := make([]string, columns)
	for i, w := range widths {
		separators[i] = strings.Repeat("─", w)
	}
	rendered = append(rendered, previewMutedStyle.Render(strings.Join(separators, "─┼─")))

	for _, row := range table.rows {
		rendered = renderTablePreviewRow(rendered, row, table.align, widths, lipgloss.NewStyle())
	}

	return rendered
}

func renderTablePreviewRow(lines []string, cells []string, align []tableAlign, widths []int, style lipgloss.Style) []string {
	wrapped := make([][]string, len(cells))
	height := 1
	for i, cell := range cells {
		wrapped[i] = wrapPreviewSegments(renderInlineMarkdownSegments(style, cell), widths[i])
		height = max(height, len(wrapped[i]))
	}

	gap := previewMutedStyle.Render(previewTableColumnGap)
	for lineIndex := 0; lineIndex < height; lineIndex++ {
		parts := make([]string, len(cells))
		for i := range cells {
			text := ""
			if lineIndex < len(wrapped[i]) {
				text = wrapped[i][lineIndex]
			}
			parts[i] = padTableCell(text, widths[i], align[i])
		}
		lines = append(lines, strings.TrimRight(strings.Join(parts, gap), " "))
	}

	return lines
}

func padTableCell(text string, width int, align tableAlign) string {
	padding := max(0, width-ansi.StringWidth(text))
	switch align {
	case tableAlignRight:
		return strings.Repeat(" ", padding) + text
	case tableAlignCenter:
		left := padding / 2
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", padding-left)
	default:
		return text + strings.Repeat(" ", padding)
	}
}

func previewSegmentsWidth(segments []previewSegment) int {
	width := 0
	for _, segment := range segments {
		width += ansi.StringWidth(segment.text)
	}
	return width
}

func sumInts(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func renderTableHTML(table markdownTable) string {
	var b strings.Builder
	b.WriteString("<table>\n<thead>\n")
	writeTableHTMLRow(&b, "th", table.header, table.align)
	b.WriteString("</thead>\n")
	if len(table.rows) > 0 {
		b.WriteString("<tbody>\n")
		for _, row := range table.rows {
			writeTableHTMLRow(&b, "td", row, table.align)
		}
		b.WriteString("</tbody>\n")
	}
	b.WriteString("</table>\n")
	return b.String()
}

func writeTableHTMLRow(b *strings.Builder, tag string, cells []string, align []tableAlign) {
	b.WriteString("<tr>")
	for i, cell := range cells {
		b.WriteString("<")
		b.WriteString(tag)
		switch align[i] {
		case tableAlignCenter:
			b.WriteString(` style="text-align:center"`)
		case tableAlignRight:
			b.WriteString(` style="text-align:right"`)
		}
		b.WriteString(">")
		b.WriteString(renderInlineMarkdownHTML(cell))
		b.WriteString("</")
		b.WriteString(tag)
		b.WriteString(">")
	}
	b.WriteString("</tr>\n")
}
