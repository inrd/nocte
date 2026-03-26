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
	if !strings.Contains(view, "Ctrl+P") || !strings.Contains(view, "Esc") || !strings.Contains(view, "Ctrl+C") {
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
