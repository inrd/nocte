package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const sampleTable = `|  VOWEL | AS IN | F1 |
|---|:---:|---:|
| "ee" | leap | 270 |
| "oo" | **loop** | 300 |
Paragraph right after the table.`

func TestParseMarkdownTableStopsAtNonPipeLine(t *testing.T) {
	lines := strings.Split(sampleTable, "\n")

	table, end, ok := parseMarkdownTable(lines, 0)
	if !ok {
		t.Fatalf("parseMarkdownTable() ok = false, want true")
	}
	if end != 4 {
		t.Fatalf("end = %d, want 4", end)
	}
	if got := strings.Join(table.header, ","); got != `VOWEL,AS IN,F1` {
		t.Fatalf("header = %q", got)
	}
	if len(table.rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(table.rows))
	}
	if table.align[1] != tableAlignCenter || table.align[2] != tableAlignRight {
		t.Fatalf("align = %v, want [left center right]", table.align)
	}
}

func TestParseMarkdownTableRequiresSeparatorRow(t *testing.T) {
	cases := []string{
		"| just a pipe line |\nnext line",
		"| a | b |\n| c | d |",
		"| a | b |\n|---|",
		"| a |",
	}
	for _, input := range cases {
		if _, _, ok := parseMarkdownTable(strings.Split(input, "\n"), 0); ok {
			t.Fatalf("parseMarkdownTable(%q) ok = true, want false", input)
		}
	}
}

func TestParseMarkdownTablePadsAndTrimsRows(t *testing.T) {
	lines := []string{"| a | b |", "|---|---|", "| 1 |", "| 1 | 2 | 3 |"}

	table, _, ok := parseMarkdownTable(lines, 0)
	if !ok {
		t.Fatalf("parseMarkdownTable() ok = false")
	}
	for i, row := range table.rows {
		if len(row) != 2 {
			t.Fatalf("row %d has %d cells, want 2", i, len(row))
		}
	}
}

func TestRenderMarkdownPreviewRendersTableColumns(t *testing.T) {
	rendered := ansi.Strip(renderMarkdownPreview(sampleTable, 60))
	lines := strings.Split(rendered, "\n")

	if strings.Contains(rendered, "|---") || strings.Contains(rendered, "| \"ee\"") {
		t.Fatalf("preview still shows raw table syntax:\n%s", rendered)
	}
	if !strings.Contains(lines[0], "VOWEL") || !strings.Contains(lines[0], "│") {
		t.Fatalf("header row = %q, want columns separated by │", lines[0])
	}
	if !strings.Contains(lines[1], "┼") {
		t.Fatalf("separator row = %q, want ┼", lines[1])
	}
	if strings.Contains(rendered, "**loop**") {
		t.Fatalf("inline markdown in cells not rendered:\n%s", rendered)
	}
	if !strings.Contains(lines[len(lines)-1], "Paragraph right after the table.") {
		t.Fatalf("last line = %q, want paragraph after table", lines[len(lines)-1])
	}

	// Columns line up: the first separator sits at the same column on each row.
	col := strings.Index(lines[0], "│")
	for _, line := range lines[2:4] {
		if got := strings.Index(line, "│"); got != col {
			t.Fatalf("misaligned row %q: separator at %d, want %d", line, got, col)
		}
	}
	// Right-aligned column ends flush.
	if !strings.HasSuffix(lines[2], "270") || !strings.HasSuffix(lines[3], "300") {
		t.Fatalf("right-aligned cells not flush: %q / %q", lines[2], lines[3])
	}
}

func TestRenderMarkdownPreviewWrapsWideTableCells(t *testing.T) {
	input := "| name | description |\n|---|---|\n| a | " + strings.Repeat("long words here ", 8) + "|"
	width := 30

	lines, offsets := renderMarkdownPreviewLines(input, width)
	for _, line := range lines {
		if got := ansi.StringWidth(line); got > width {
			t.Fatalf("line width = %d, want <= %d: %q", got, width, ansi.Strip(line))
		}
	}
	if len(lines) <= 3 {
		t.Fatalf("expected wrapped cell rows, got %d lines", len(lines))
	}
	for i, offset := range offsets {
		if offset != 0 {
			t.Fatalf("offsets[%d] = %d, want table lines mapped to table start", i, offset)
		}
	}
}

func TestRenderMarkdownPreviewIgnoresTablesInCodeBlocks(t *testing.T) {
	input := "```\n| a | b |\n|---|---|\n```"
	rendered := ansi.Strip(renderMarkdownPreview(input, 40))

	if !strings.Contains(rendered, "| a | b |") || strings.Contains(rendered, "┼") {
		t.Fatalf("table inside code block should stay raw:\n%s", rendered)
	}
}

func TestRenderMarkdownHTMLBodyRendersTables(t *testing.T) {
	body := renderMarkdownHTMLBody(sampleTable)

	for _, want := range []string{
		"<table>",
		"<thead>\n<tr><th>VOWEL</th><th style=\"text-align:center\">AS IN</th><th style=\"text-align:right\">F1</th></tr>",
		"<tbody>",
		"<td>&#34;ee&#34;</td>",
		"<strong>loop</strong>",
		"</table>\n<p>Paragraph right after the table.</p>",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("renderMarkdownHTMLBody() missing %q:\n%s", want, body)
		}
	}
}
