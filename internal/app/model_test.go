package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
	writeTestNote(t, tmpDir, "meeting-notes.md", "hello there")
	writeTestNote(t, tmpDir, "math-notes.md", "123 456 789")
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

	if matches[0].wordCount != 3 {
		t.Fatalf("matches[0].wordCount = %d, want 3", matches[0].wordCount)
	}
	if matches[1].wordCount != 2 {
		t.Fatalf("matches[1].wordCount = %d, want 2", matches[1].wordCount)
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

func TestListNotesReturnsMostRecentlyUpdatedFirst(t *testing.T) {
	tmpDir := t.TempDir()
	oldestPath := writeTestNote(t, tmpDir, "oldest.md", "a")
	middlePath := writeTestNote(t, tmpDir, "middle.md", "bb")
	newestPath := writeTestNote(t, tmpDir, "newest.md", "ccc")

	now := time.Now()
	mustSetModTime(t, oldestPath, now.Add(-48*time.Hour))
	mustSetModTime(t, middlePath, now.Add(-6*time.Hour))
	mustSetModTime(t, newestPath, now.Add(-1*time.Hour))

	model := New(config.Config{NotesPath: tmpDir}, "", "test")

	notes := model.listNotes()
	if len(notes) != 3 {
		t.Fatalf("len(notes) = %d, want 3", len(notes))
	}

	wantOrder := []string{"newest.md", "middle.md", "oldest.md"}
	for i, want := range wantOrder {
		if notes[i].name != want {
			t.Fatalf("notes[%d].name = %q, want %q", i, notes[i].name, want)
		}
	}
}

func TestFormatNoteUpdatedAt(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 14, 5, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	sameYear := time.Date(now.Year(), now.Month(), 1, 15, 4, 0, 0, now.Location())
	if sameDay(sameYear, today) || sameDay(sameYear, yesterday) {
		sameYear = sameYear.AddDate(0, -1, 0)
	}
	otherYear := time.Date(now.Year()-1, time.January, 2, 15, 4, 0, 0, now.Location())

	tests := []struct {
		name    string
		modTime time.Time
		want    string
	}{
		{name: "today", modTime: today, want: "Today 14:05"},
		{name: "yesterday", modTime: yesterday, want: "Yest 14:05"},
		{name: "same year", modTime: sameYear, want: sameYear.Format("Jan 2 15:04")},
		{name: "other year", modTime: otherYear, want: fmt.Sprintf("Jan 2 %d 15:04", now.Year()-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatNoteUpdatedAt(tt.modTime); got != tt.want {
				t.Fatalf("formatNoteUpdatedAt(%v) = %q, want %q", tt.modTime, got, tt.want)
			}
		})
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{name: "small bytes", size: 12, want: "12 B"},
		{name: "promotes to kilobytes", size: 425, want: "0.4 KB"},
		{name: "kilobytes", size: 1536, want: "1.5 KB"},
		{name: "promotes to megabytes", size: 60 * 1024, want: "0.1 MB"},
		{name: "megabytes", size: 2 * 1024 * 1024, want: "2.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanSize(tt.size); got != tt.want {
				t.Fatalf("humanSize(%d) = %q, want %q", tt.size, got, tt.want)
			}
		})
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
	oldestPath := writeTestNote(t, tmpDir, "alpha.md", "a")
	secondPath := writeTestNote(t, tmpDir, "middle.md", "m")
	newestPath := writeTestNote(t, tmpDir, "zebra.md", "z")
	now := time.Now()
	mustSetModTime(t, oldestPath, now.Add(-72*time.Hour))
	mustSetModTime(t, secondPath, now.Add(-24*time.Hour))
	mustSetModTime(t, newestPath, now.Add(-1*time.Hour))

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

func writeTestNote(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	return path
}

func mustSetModTime(t *testing.T, path string, modTime time.Time) {
	t.Helper()

	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(%q) error = %v", path, err)
	}
}

func longNoteContent(length int) string {
	if length <= 0 {
		return ""
	}

	const pattern = "I need to make this not grow so I will copy paste that "
	var b []byte
	for len(b) < length {
		remaining := length - len(b)
		if remaining >= len(pattern) {
			b = append(b, pattern...)
			continue
		}
		b = append(b, pattern[:remaining]...)
	}

	return string(b)
}
