package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inrd/nocte/internal/config"
)

func TestOpenEditorDoesNotTruncateLongNotes(t *testing.T) {
	tmpDir := t.TempDir()
	content := longNoteContent(420)
	notePath := writeTestNote(t, tmpDir, "long.md", content)

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "long.md")

	if model.editor.CharLimit != 0 {
		t.Fatalf("editor.CharLimit = %d, want 0", model.editor.CharLimit)
	}
	if got := model.editor.Value(); got != content {
		t.Fatalf("editor content length = %d, want %d", len(got), len(content))
	}
	if got := model.editorStatusLine(); !strings.Contains(got, "Size") {
		t.Fatalf("editorStatusLine() = %q, want size information", got)
	}
	if got := model.editor.Line(); got != 0 {
		t.Fatalf("editor.Line() = %d, want 0", got)
	}
	if got := model.previewContent(); got == "" || !strings.Contains(got, "I need to make") {
		t.Fatalf("previewContent() = %q, want top of note content", got)
	}
}

func TestUpdateEscWhileEditingSavesAndClosesEditor(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := filepath.Join(tmpDir, "draft.md")
	writeTestNote(t, tmpDir, "draft.md", "before")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.editor.SetValue("after")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	if model.isEditing() {
		t.Fatalf("model should not be editing after escape")
	}
	if model.status != "Saved and closed draft.md" {
		t.Fatalf("status = %q, want %q", model.status, "Saved and closed draft.md")
	}

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", notePath, err)
	}
	if string(data) != "after" {
		t.Fatalf("saved content = %q, want %q", string(data), "after")
	}
}

func TestUpdateEscWithoutChangesClosesEditorWithoutSaveMessage(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "before")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	if model.isEditing() {
		t.Fatalf("model should not be editing after escape")
	}
	if model.status != "Closed draft.md without changes" {
		t.Fatalf("status = %q, want %q", model.status, "Closed draft.md without changes")
	}

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", notePath, err)
	}
	if string(data) != "before" {
		t.Fatalf("saved content = %q, want %q", string(data), "before")
	}
}

func TestUpdateEscDeletesEmptyNewNote(t *testing.T) {
	tmpDir := t.TempDir()
	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("Scratch")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	notePath := filepath.Join(tmpDir, "scratch.md")
	if model.isEditing() {
		t.Fatalf("model should not be editing after escape")
	}
	if model.status != "Discarded empty note scratch.md" {
		t.Fatalf("status = %q, want %q", model.status, "Discarded empty note scratch.md")
	}
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Fatalf("Stat(%q) error = %v, want not exist", notePath, err)
	}
}

func TestUpdateEscKeepsNewNoteWhenEditorHasContent(t *testing.T) {
	tmpDir := t.TempDir()
	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("Scratch")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})
	model.editor.SetValue("hello")
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	notePath := filepath.Join(tmpDir, "scratch.md")
	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", notePath, err)
	}
	if string(data) != "hello" {
		t.Fatalf("saved content = %q, want %q", string(data), "hello")
	}
}

func TestUpdateEscKeepsExistingEmptyNote(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "empty.md", "")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "empty.md")
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	if _, err := os.Stat(notePath); err != nil {
		t.Fatalf("Stat(%q) error = %v, want file to remain", notePath, err)
	}
	if model.status != "Closed empty.md without changes" {
		t.Fatalf("status = %q, want %q", model.status, "Closed empty.md without changes")
	}
}

func TestUpdateCtrlPTogglesPreviewWhileEditing(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "# Title\n\n- item")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 96
	model.openEditor(notePath, "draft.md")

	if !model.previewEnabled {
		t.Fatalf("previewEnabled = false, want true")
	}
	if !model.previewVisible() {
		t.Fatalf("previewVisible() = false, want true")
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlP})

	if model.previewEnabled {
		t.Fatalf("previewEnabled = true after Ctrl+P, want false")
	}
	if model.previewVisible() {
		t.Fatalf("previewVisible() = true after Ctrl+P, want false")
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlP})

	if !model.previewEnabled {
		t.Fatalf("previewEnabled = false after second Ctrl+P, want true")
	}
	if !model.previewVisible() {
		t.Fatalf("previewVisible() = false after second Ctrl+P, want true")
	}
}

func TestUpdateCtrlTTurnsPlainLineIntoOpenTask(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "alpha\nplain text\nomega")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.jumpEditorTo(1, 3)

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlT})

	if got := model.editor.Value(); got != "alpha\n- [ ] plain text\nomega" {
		t.Fatalf("editor.Value() = %q, want %q", got, "alpha\n- [ ] plain text\nomega")
	}
	if got := model.editor.Line(); got != 1 {
		t.Fatalf("editor.Line() = %d, want %d", got, 1)
	}
	if got := model.editor.LineInfo().CharOffset; got != 9 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want %d", got, 9)
	}
}

func TestUpdateCtrlTTogglesExistingTaskState(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "- [ ] ship it")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.jumpEditorTo(0, 7)

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlT})

	if got := model.editor.Value(); got != "- [x] ship it" {
		t.Fatalf("editor.Value() after first Ctrl+T = %q, want %q", got, "- [x] ship it")
	}
	if got := model.editor.LineInfo().CharOffset; got != 7 {
		t.Fatalf("editor.LineInfo().CharOffset after first Ctrl+T = %d, want %d", got, 7)
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlT})

	if got := model.editor.Value(); got != "- [ ] ship it" {
		t.Fatalf("editor.Value() after second Ctrl+T = %q, want %q", got, "- [ ] ship it")
	}
	if got := model.editor.LineInfo().CharOffset; got != 7 {
		t.Fatalf("editor.LineInfo().CharOffset after second Ctrl+T = %d, want %d", got, 7)
	}
}

func TestUpdateCtrlTTurnsBulletIntoOpenTask(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "  * follow up")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.jumpEditorTo(0, 4)

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlT})

	if got := model.editor.Value(); got != "  * [ ] follow up" {
		t.Fatalf("editor.Value() = %q, want %q", got, "  * [ ] follow up")
	}
	if got := model.editor.LineInfo().CharOffset; got != 8 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want %d", got, 8)
	}
}

func TestUpdateTabInsertsDefaultIndentWhileEditing(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "- parent")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.jumpEditorTo(0, 0)

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyTab})

	if got := model.editor.Value(); got != "    - parent" {
		t.Fatalf("editor.Value() = %q, want %q", got, "    - parent")
	}
	if got := model.editor.LineInfo().CharOffset; got != 4 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want %d", got, 4)
	}
}

func TestUpdateTabUsesConfiguredIndentWidth(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "- child")

	model := New(config.Config{NotesPath: tmpDir, TabWidth: 2}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.jumpEditorTo(0, 0)

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyTab})

	if got := model.editor.Value(); got != "  - child" {
		t.Fatalf("editor.Value() = %q, want %q", got, "  - child")
	}
	if got := model.editor.LineInfo().CharOffset; got != 2 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want %d", got, 2)
	}
}

func TestUpdateCtrlERendersHTMLAndOpensIt(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "# Before")
	original := openPathWithSystemApp
	t.Cleanup(func() {
		openPathWithSystemApp = original
	})

	var openedPath string
	openPathWithSystemApp = func(path string) error {
		openedPath = path
		return nil
	}

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 96
	model.openEditor(notePath, "draft.md")
	model.editor.SetValue("# After\n\n[Docs](https://example.com/docs)\n\n![Diagram](./diagram.png)")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlE})

	exportPath := filepath.Join(tmpDir, "html", "draft.html")
	if openedPath != exportPath {
		t.Fatalf("openedPath = %q, want %q", openedPath, exportPath)
	}
	if model.status != "Opened HTML export html/draft.html" {
		t.Fatalf("status = %q, want %q", model.status, "Opened HTML export html/draft.html")
	}
	if !model.isEditing() {
		t.Fatalf("model should stay in the editor after HTML export")
	}

	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", exportPath, err)
	}
	rendered := string(data)
	if !strings.Contains(rendered, "<h1>After</h1>") {
		t.Fatalf("rendered html = %q, want heading", rendered)
	}
	if !strings.Contains(rendered, "<a href=\"https://example.com/docs\">Docs</a>") {
		t.Fatalf("rendered html = %q, want link", rendered)
	}
	if !strings.Contains(rendered, "<img src=\"./diagram.png\" alt=\"Diagram\">") {
		t.Fatalf("rendered html = %q, want image", rendered)
	}
	if !strings.Contains(rendered, "<base href=\"file://") {
		t.Fatalf("rendered html = %q, want file base href", rendered)
	}
}

func TestUpdateCtrlLOpensLinksDialogAndEnterOpensSelectedLink(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "[Docs](https://example.com/docs)\nhttps://example.com/raw")
	original := openURLWithSystemApp
	t.Cleanup(func() {
		openURLWithSystemApp = original
	})

	var openedURL string
	openURLWithSystemApp = func(rawURL string) error {
		openedURL = rawURL
		return nil
	}

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 96
	model.openEditor(notePath, "draft.md")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlL})

	if model.activeDialog != "links" {
		t.Fatalf("activeDialog = %q, want %q", model.activeDialog, "links")
	}
	if len(model.dialogLinks) != 2 {
		t.Fatalf("len(dialogLinks) = %d, want 2", len(model.dialogLinks))
	}
	if model.dialogIndex != 0 {
		t.Fatalf("dialogIndex = %d, want 0", model.dialogIndex)
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if openedURL != "https://example.com/docs" {
		t.Fatalf("openedURL = %q, want %q", openedURL, "https://example.com/docs")
	}
	if model.activeDialog != "" {
		t.Fatalf("activeDialog = %q, want empty after opening link", model.activeDialog)
	}
	if !model.isEditing() {
		t.Fatalf("model should still be editing after opening link")
	}
	if model.status != "Opened link in browser" {
		t.Fatalf("status = %q, want %q", model.status, "Opened link in browser")
	}
}

func TestUpdateCtrlDOpensDeleteConfirmationAndEscCancels(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "before")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlD})

	if model.activeDialog != "delete-confirm" {
		t.Fatalf("activeDialog = %q, want %q", model.activeDialog, "delete-confirm")
	}
	if !model.isEditing() {
		t.Fatalf("model should still be editing while confirming deletion")
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	if model.activeDialog != "" {
		t.Fatalf("activeDialog = %q, want empty after cancel", model.activeDialog)
	}
	if !model.isEditing() {
		t.Fatalf("model should continue editing after cancel")
	}
	if model.status != "Still editing draft.md" {
		t.Fatalf("status = %q, want %q", model.status, "Still editing draft.md")
	}
	if _, err := os.Stat(notePath); err != nil {
		t.Fatalf("Stat(%q) error = %v, want file to remain", notePath, err)
	}
}

func TestUpdateCtrlDEnterDeletesNoteAndReturnsToLauncher(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "before")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openEditor(notePath, "draft.md")
	model.editor.SetValue("after")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyCtrlD})
	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.isEditing() {
		t.Fatalf("model should not be editing after deletion")
	}
	if model.activeDialog != "" {
		t.Fatalf("activeDialog = %q, want empty after deletion", model.activeDialog)
	}
	if model.status != "Deleted draft.md" {
		t.Fatalf("status = %q, want %q", model.status, "Deleted draft.md")
	}
	if model.input.Value() != "" {
		t.Fatalf("input.Value() = %q, want empty", model.input.Value())
	}
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Fatalf("Stat(%q) error = %v, want not exist", notePath, err)
	}
}

func TestExtractNoteLinksFindsMarkdownAndBareURLs(t *testing.T) {
	links := extractNoteLinks("See [Docs](https://example.com/docs).\nRaw https://example.com/raw,\nagain [Docs](https://example.com/docs)")

	if len(links) != 2 {
		t.Fatalf("len(extractNoteLinks()) = %d, want 2", len(links))
	}
	if links[0].label != "Docs" || links[0].url != "https://example.com/docs" {
		t.Fatalf("links[0] = %#v, want Docs markdown link", links[0])
	}
	if links[1].label != "" || links[1].url != "https://example.com/raw" {
		t.Fatalf("links[1] = %#v, want bare url", links[1])
	}
}

func TestEditorViewShowsMarkdownPreviewWhenVisible(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "# Title\n\n- item with `code`\n> quote\n[site](https://example.com)")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 96
	model.openEditor(notePath, "draft.md")

	view := model.editorView()
	preview := model.previewContent()

	if !strings.Contains(view, "# Title") {
		t.Fatalf("editorView() missing editor content: %q", view)
	}
	if !strings.Contains(view, "Title") {
		t.Fatalf("editorView() missing heading preview: %q", view)
	}
	if !strings.Contains(view, "• item") {
		t.Fatalf("editorView() missing bullet preview: %q", view)
	}
	if !strings.Contains(view, "code") {
		t.Fatalf("editorView() missing inline code preview: %q", view)
	}
	if !strings.Contains(view, "│ quote") {
		t.Fatalf("editorView() missing quote preview: %q", view)
	}
	if !strings.Contains(view, "site") {
		t.Fatalf("editorView() missing link preview: %q", view)
	}
	if strings.Contains(preview, "https://example.com") {
		t.Fatalf("previewContent() should render labeled links without showing the raw url: %q", preview)
	}
	if !strings.Contains(view, "Preview (read-only)") {
		t.Fatalf("editorView() missing preview label: %q", view)
	}
	if !strings.Contains(view, "Ctrl+P") || !strings.Contains(view, "Ctrl+T") || !strings.Contains(view, "Ctrl+E") || !strings.Contains(view, "Ctrl+L") || !strings.Contains(view, "Ctrl+D") || !strings.Contains(view, "Esc") || !strings.Contains(view, "Ctrl+C") {
		t.Fatalf("editorView() missing preview shortcut help: %q", view)
	}
	if strings.Contains(view, "Plain text editor with live Markdown preview") {
		t.Fatalf("editorView() help should only show shortcuts: %q", view)
	}
}

func TestRenderMarkdownPreviewKeepsWrappedLinksIntact(t *testing.T) {
	rendered := renderMarkdownPreview("[a fairly long link label](https://example.com/docs/path)", 18)

	if !strings.Contains(rendered, "a fairly long link") || !strings.Contains(rendered, "label") {
		t.Fatalf("renderMarkdownPreview() missing link label: %q", rendered)
	}
	if strings.Contains(rendered, "https://example.com/docs/path") {
		t.Fatalf("renderMarkdownPreview() should not show raw urls for labeled links: %q", rendered)
	}
	if strings.Contains(rendered, "](") {
		t.Fatalf("renderMarkdownPreview() should not expose raw markdown link syntax after wrapping: %q", rendered)
	}
}

func TestRenderMarkdownHTMLDocumentRendersMarkdown(t *testing.T) {
	notePath := filepath.Join("/tmp", "notes", "draft.md")
	rendered := renderMarkdownHTMLDocument("draft.md", notePath, "# Title\n\n- item\n- [x] done\n\n`code` and **bold** and [site](https://example.com)")

	if !strings.Contains(rendered, "<title>draft</title>") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want title", rendered)
	}
	if !strings.Contains(rendered, "<h1>Title</h1>") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want heading", rendered)
	}
	if !strings.Contains(rendered, "<ul>") || !strings.Contains(rendered, "<li>item") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want list", rendered)
	}
	if !strings.Contains(rendered, "<input type=\"checkbox\" checked disabled> done") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want task list checkbox", rendered)
	}
	if !strings.Contains(rendered, "<code>code</code>") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want inline code", rendered)
	}
	if !strings.Contains(rendered, "<strong>bold</strong>") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want bold text", rendered)
	}
	if !strings.Contains(rendered, "<a href=\"https://example.com\">site</a>") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want link", rendered)
	}
	if !strings.Contains(rendered, "<base href=\"file:///tmp/notes/\">") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want base href", rendered)
	}
	if !strings.Contains(rendered, "pre code{display:block;padding:0;border-radius:0;background:transparent;color:inherit;}") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want block code style reset", rendered)
	}
}

func TestRenderMarkdownHTMLDocumentKeepsPlainExclamationText(t *testing.T) {
	rendered := renderMarkdownHTMLDocument("draft.md", "", "Heads up! Plain text stays plain.")

	if !strings.Contains(rendered, "Heads up! Plain text stays plain.") {
		t.Fatalf("renderMarkdownHTMLDocument() = %q, want plain exclamation text", rendered)
	}
}

func TestRenderMarkdownPreviewKeepsWrappedQuotePrefix(t *testing.T) {
	rendered := renderMarkdownPreview("> this is a long quoted line that should stay quoted after wrapping", 18)

	if !strings.Contains(rendered, "│ this is a long") {
		t.Fatalf("renderMarkdownPreview() missing wrapped quote start: %q", rendered)
	}
	if !strings.Contains(rendered, "│ should stay") {
		t.Fatalf("renderMarkdownPreview() missing wrapped quote continuation prefix: %q", rendered)
	}
}

func TestRenderMarkdownPreviewKeepsWrappedInlineCodeStyled(t *testing.T) {
	rendered := renderMarkdownPreview("prefix `this-inline-code-span-is-long` suffix", 16)

	if !strings.Contains(rendered, "this-inline-code") {
		t.Fatalf("renderMarkdownPreview() missing wrapped inline code start: %q", rendered)
	}
	if !strings.Contains(rendered, "span-is-long") {
		t.Fatalf("renderMarkdownPreview() missing wrapped inline code continuation: %q", rendered)
	}
}

func TestRenderMarkdownPreviewRendersBoldAndItalics(t *testing.T) {
	rendered := renderMarkdownPreview("plain *italic* and _also italic_ plus **bold** and __also bold__", 80)

	if !strings.Contains(rendered, "italic") || !strings.Contains(rendered, "also italic") {
		t.Fatalf("renderMarkdownPreview() missing italic text: %q", rendered)
	}
	if !strings.Contains(rendered, "bold") || !strings.Contains(rendered, "also bold") {
		t.Fatalf("renderMarkdownPreview() missing bold text: %q", rendered)
	}
	if strings.Contains(rendered, "*italic*") || strings.Contains(rendered, "_also italic_") {
		t.Fatalf("renderMarkdownPreview() should hide italic markdown markers: %q", rendered)
	}
	if strings.Contains(rendered, "**bold**") || strings.Contains(rendered, "__also bold__") {
		t.Fatalf("renderMarkdownPreview() should hide bold markdown markers: %q", rendered)
	}
}

func TestRenderMarkdownPreviewRendersStrikethrough(t *testing.T) {
	rendered := renderMarkdownPreview("keep ~~crossed out~~ text", 80)

	if !strings.Contains(rendered, "crossed out") {
		t.Fatalf("renderMarkdownPreview() missing strikethrough text: %q", rendered)
	}
	if strings.Contains(rendered, "~~crossed out~~") {
		t.Fatalf("renderMarkdownPreview() should hide strikethrough markdown markers: %q", rendered)
	}
}

func TestRenderMarkdownPreviewKeepsNestedListIndentation(t *testing.T) {
	rendered := renderMarkdownPreview("- parent\n    - child\n        - nested\n- another root", 28)

	if !strings.Contains(rendered, "• parent") {
		t.Fatalf("renderMarkdownPreview() missing root bullet: %q", rendered)
	}
	if !strings.Contains(rendered, "    • child") {
		t.Fatalf("renderMarkdownPreview() missing nested bullet indentation: %q", rendered)
	}
	if !strings.Contains(rendered, "        • nested") {
		t.Fatalf("renderMarkdownPreview() missing deeper nested bullet indentation: %q", rendered)
	}
	if !strings.Contains(rendered, "\n• another root") {
		t.Fatalf("renderMarkdownPreview() missing second root bullet: %q", rendered)
	}
}

func TestRenderMarkdownPreviewKeepsOrderedListIndentationAndWrap(t *testing.T) {
	rendered := renderMarkdownPreview("  12. this ordered item wraps onto the next preview line", 18)

	if !strings.Contains(rendered, "  12. this ordered") {
		t.Fatalf("renderMarkdownPreview() missing ordered list marker: %q", rendered)
	}
	if !strings.Contains(rendered, "\n      item wraps") {
		t.Fatalf("renderMarkdownPreview() missing ordered list continuation alignment: %q", rendered)
	}
}

func TestRenderMarkdownPreviewRendersTaskLists(t *testing.T) {
	rendered := renderMarkdownPreview("- [ ] open item\n  - [x] done item that wraps onto another line", 18)

	if !strings.Contains(rendered, "☐ open item") {
		t.Fatalf("renderMarkdownPreview() missing unchecked task item: %q", rendered)
	}
	if !strings.Contains(rendered, "  ☑ done item") {
		t.Fatalf("renderMarkdownPreview() missing checked task item: %q", rendered)
	}
	if strings.Contains(rendered, "[ ]") || strings.Contains(rendered, "[x]") {
		t.Fatalf("renderMarkdownPreview() should hide raw task list markers: %q", rendered)
	}
}

func TestPreviewContentRendersMarkdownImagesWithChafa(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "draft.md", "![diagram](./diagram.png)")
	imagePath := filepath.Join(tmpDir, "diagram.png")
	if err := os.WriteFile(imagePath, []byte("png"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", imagePath, err)
	}

	originalLookPath := lookPath
	originalRunCommandOutput := runCommandOutput
	t.Cleanup(func() {
		lookPath = originalLookPath
		runCommandOutput = originalRunCommandOutput
		clearPreviewImageCache()
	})

	lookPath = func(file string) (string, error) {
		if file != "chafa" {
			t.Fatalf("LookPath file = %q, want chafa", file)
		}
		return "/opt/homebrew/bin/chafa", nil
	}

	var calls int
	runCommandOutput = func(name string, args ...string) ([]byte, error) {
		calls++
		if name != "/opt/homebrew/bin/chafa" {
			t.Fatalf("command name = %q, want chafa path", name)
		}
		if got := args[len(args)-1]; got != imagePath {
			t.Fatalf("image path arg = %q, want %q", got, imagePath)
		}
		return []byte("img-line-1\nimg-line-2\n"), nil
	}

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 96
	model.openEditor(notePath, "draft.md")

	first := model.previewContent()
	second := model.previewContent()

	if !strings.Contains(first, "img-line-1") || !strings.Contains(first, "img-line-2") {
		t.Fatalf("previewContent() = %q, want chafa image output", first)
	}
	if second != first {
		t.Fatalf("previewContent() second render = %q, want %q", second, first)
	}
	if calls != 1 {
		t.Fatalf("runCommandOutput calls = %d, want 1 due to cache", calls)
	}
}

func TestRenderMarkdownPreviewFallsBackWhenImageCannotBeRendered(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := filepath.Join(tmpDir, "draft.md")

	rendered := strings.Join(renderMarkdownImagePreview(notePath, "![diagram](./missing.png)", 24), "\n")

	if !strings.Contains(rendered, "Image: diagram") {
		t.Fatalf("fallback preview = %q, want image label", rendered)
	}
	if !strings.Contains(rendered, "./missing.png") {
		t.Fatalf("fallback preview = %q, want image path", rendered)
	}
}

func TestRenderMarkdownPreviewFallsBackWhenChafaIsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	imagePath := filepath.Join(tmpDir, "diagram.png")
	if err := os.WriteFile(imagePath, []byte("png"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", imagePath, err)
	}

	originalLookPath := lookPath
	t.Cleanup(func() {
		lookPath = originalLookPath
		clearPreviewImageCache()
	})

	lookPath = func(string) (string, error) {
		return "", os.ErrNotExist
	}

	rendered := strings.Join(renderMarkdownImagePreview(filepath.Join(tmpDir, "draft.md"), "![diagram](./diagram.png)", 24), "\n")

	if !strings.Contains(rendered, "Image: diagram") {
		t.Fatalf("fallback preview = %q, want image label", rendered)
	}
	if !strings.Contains(rendered, "install chafa to preview") {
		t.Fatalf("fallback preview = %q, want chafa hint", rendered)
	}
}

func TestResolveMarkdownImagePathExpandsHomePrefix(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	got, err := resolveMarkdownImagePath(filepath.Join(tmpHome, "note.md"), "~/images/diagram.png")
	if err != nil {
		t.Fatalf("resolveMarkdownImagePath() error = %v", err)
	}

	want := filepath.Join(tmpHome, "images", "diagram.png")
	if got != want {
		t.Fatalf("resolveMarkdownImagePath() = %q, want %q", got, want)
	}
}

func TestResolveMarkdownImagePathExpandsEnvironmentVariables(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	got, err := resolveMarkdownImagePath(filepath.Join(tmpHome, "note.md"), "$HOME/images/diagram.png")
	if err != nil {
		t.Fatalf("resolveMarkdownImagePath() error = %v", err)
	}

	want := filepath.Join(tmpHome, "images", "diagram.png")
	if got != want {
		t.Fatalf("resolveMarkdownImagePath() = %q, want %q", got, want)
	}
}
