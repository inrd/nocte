package app

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/thomas/nocte/internal/config"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normalizes spaces and case", input: "Project Notes", want: "project-notes"},
		{name: "keeps safe punctuation", input: "build_v1.2", want: "build_v1.2"},
		{name: "replaces unsupported characters", input: "Ideas/Today!", want: "ideas-today"},
		{name: "rejects blank", input: "   ", wantErr: true},
		{name: "rejects punctuation only", input: "!!!", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeFilename(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("sanitizeFilename(%q) error = nil, want error", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("sanitizeFilename(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFuzzyScore(t *testing.T) {
	scoreSubstring, ok := fuzzyScore("meeting-notes", "meet")
	if !ok {
		t.Fatalf("fuzzyScore() did not match substring candidate")
	}

	scoreSpread, ok := fuzzyScore("mxxexxexxting", "meet")
	if !ok {
		t.Fatalf("fuzzyScore() did not match spread candidate")
	}

	if scoreSubstring >= scoreSpread {
		t.Fatalf("substring match score = %d, want less than spread match score = %d", scoreSubstring, scoreSpread)
	}

	if _, ok := fuzzyScore("meeting-notes", "zz"); ok {
		t.Fatalf("fuzzyScore() unexpectedly matched missing query")
	}

	if _, ok := fuzzyScore("meeting-notes", "   "); ok {
		t.Fatalf("fuzzyScore() unexpectedly matched blank query")
	}
}

func TestFindNoteMatchesFiltersSortsAndAddsMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "meeting-notes.md", "hello")
	writeTestNote(t, tmpDir, "math-notes.md", "123456789")
	writeTestNote(t, tmpDir, "zebra.md", "z")
	writeTestNote(t, tmpDir, "ignore.txt", "skip")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")

	matches := model.findNoteMatches("mn")
	if len(matches) != 2 {
		t.Fatalf("len(matches) = %d, want 2", len(matches))
	}

	if matches[0].name != "math-notes.md" {
		t.Fatalf("matches[0].name = %q, want %q", matches[0].name, "math-notes.md")
	}
	if matches[1].name != "meeting-notes.md" {
		t.Fatalf("matches[1].name = %q, want %q", matches[1].name, "meeting-notes.md")
	}

	if matches[0].charCount != 9 {
		t.Fatalf("matches[0].charCount = %d, want 9", matches[0].charCount)
	}
	if matches[1].charCount != 5 {
		t.Fatalf("matches[1].charCount = %d, want 5", matches[1].charCount)
	}
	if matches[0].sizeBytes <= 0 {
		t.Fatalf("matches[0].sizeBytes = %d, want > 0", matches[0].sizeBytes)
	}
}

func TestFindNoteMatchesEmptyQueryReturnsAlphabeticalList(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "zebra.md", "z")
	writeTestNote(t, tmpDir, "alpha.md", "a")
	writeTestNote(t, tmpDir, "middle.md", "m")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")

	matches := model.findNoteMatches("")
	if len(matches) != 3 {
		t.Fatalf("len(matches) = %d, want 3", len(matches))
	}

	wantOrder := []string{"alpha.md", "middle.md", "zebra.md"}
	for i, want := range wantOrder {
		if matches[i].name != want {
			t.Fatalf("matches[%d].name = %q, want %q", i, matches[i].name, want)
		}
	}
}

func TestMoveNoteSelectionStartsUnselected(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.noteMatches = []noteMatch{{name: "one.md"}, {name: "two.md"}}

	model.moveNoteSelection(1)
	if model.noteIndex != 0 {
		t.Fatalf("noteIndex after move down = %d, want 0", model.noteIndex)
	}

	model.noteIndex = -1
	model.moveNoteSelection(-1)
	if model.noteIndex != 1 {
		t.Fatalf("noteIndex after move up = %d, want 1", model.noteIndex)
	}
}

func TestDialogRangeHonorsOffsetAndHeight(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.height = 15
	model.dialogOffset = 2
	model.dialogNotes = []noteMatch{
		{name: "1.md"},
		{name: "2.md"},
		{name: "3.md"},
		{name: "4.md"},
		{name: "5.md"},
		{name: "6.md"},
		{name: "7.md"},
	}

	start, end := model.dialogRange()
	if start != 2 || end != 7 {
		t.Fatalf("dialogRange() = (%d, %d), want (2, 7)", start, end)
	}
}

func TestUpdateEnterCreatesNewNoteAndOpensEditor(t *testing.T) {
	tmpDir := t.TempDir()
	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("Project Notes")

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	wantPath := filepath.Join(tmpDir, "project-notes.md")
	if model.editorPath != wantPath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, wantPath)
	}
	if model.editorName != "project-notes.md" {
		t.Fatalf("editorName = %q, want %q", model.editorName, "project-notes.md")
	}
	if !model.isEditing() {
		t.Fatalf("model should be editing after creating a note")
	}
	if model.status != "Created project-notes.md" {
		t.Fatalf("status = %q, want %q", model.status, "Created project-notes.md")
	}

	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", wantPath, err)
	}
	if string(data) != "" {
		t.Fatalf("new note content = %q, want empty file", string(data))
	}
}

func TestUpdateEnterOpensSelectedExistingNote(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := filepath.Join(tmpDir, "meeting-notes.md")
	writeTestNote(t, tmpDir, "meeting-notes.md", "agenda")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("meeting")
	model.syncLauncherState()
	model.noteIndex = 0

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.editorPath != notePath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, notePath)
	}
	if model.editor.Value() != "agenda" {
		t.Fatalf("editor content = %q, want %q", model.editor.Value(), "agenda")
	}
	if model.status != "Editing meeting-notes.md" {
		t.Fatalf("status = %q, want %q", model.status, "Editing meeting-notes.md")
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

func TestUpdateListDialogCanOpenSelectedNote(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "alpha.md", "a")
	secondPath := filepath.Join(tmpDir, "middle.md")
	writeTestNote(t, tmpDir, "middle.md", "m")
	writeTestNote(t, tmpDir, "zebra.md", "z")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue(":list")
	model.syncLauncherState()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.activeDialog != "list" {
		t.Fatalf("activeDialog = %q, want %q", model.activeDialog, "list")
	}
	if model.dialogIndex != 0 {
		t.Fatalf("dialogIndex = %d, want 0", model.dialogIndex)
	}
	if len(model.dialogNotes) != 3 {
		t.Fatalf("len(dialogNotes) = %d, want 3", len(model.dialogNotes))
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyDown})
	if model.dialogIndex != 1 {
		t.Fatalf("dialogIndex after down = %d, want 1", model.dialogIndex)
	}

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.editorPath != secondPath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, secondPath)
	}
	if model.editorName != "middle.md" {
		t.Fatalf("editorName = %q, want %q", model.editorName, "middle.md")
	}
	if model.editor.Value() != "m" {
		t.Fatalf("editor content = %q, want %q", model.editor.Value(), "m")
	}
}

func updateModel(t *testing.T, model Model, msg tea.Msg) Model {
	t.Helper()

	updated, _ := model.Update(msg)
	next, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update() returned %T, want app.Model", updated)
	}

	return next
}

func writeTestNote(t *testing.T, dir string, name string, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
