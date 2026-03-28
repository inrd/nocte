package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inrd/nocte/internal/config"
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

func TestFindNoteMatchesFiltersSortsAndAddsMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "meeting-notes.md", "hello there\nsecond line")
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
	if matches[1].wordCount != 4 {
		t.Fatalf("matches[1].wordCount = %d, want 4", matches[1].wordCount)
	}
	if matches[0].sizeBytes <= 0 {
		t.Fatalf("matches[0].sizeBytes = %d, want > 0", matches[0].sizeBytes)
	}
	if got := matches[1].preview; len(got) != 2 || got[0] != "hello there" || got[1] != "second line" {
		t.Fatalf("matches[1].preview = %v, want [hello there second line]", got)
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

func TestNotePaletteViewUsesScrollableViewport(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.height = 16
	model.noteMatches = []noteMatch{
		{name: "one.md", preview: []string{"one a", "one b"}},
		{name: "two.md", preview: []string{"two a", "two b"}},
		{name: "three.md", preview: []string{"three a", "three b"}},
		{name: "four.md", preview: []string{"four a", "four b"}},
		{name: "five.md", preview: []string{"five a", "five b"}},
		{name: "six.md", preview: []string{"six a", "six b"}},
	}
	model.noteIndex = len(model.noteMatches) - 1
	model.syncNoteOffset()

	rendered := model.notePaletteView()

	if !strings.Contains(rendered, "Showing") {
		t.Fatalf("notePaletteView() missing viewport footer: %q", rendered)
	}
	if !strings.Contains(rendered, "six.md") {
		t.Fatalf("notePaletteView() missing tail item: %q", rendered)
	}
	if !strings.Contains(rendered, "six a") {
		t.Fatalf("notePaletteView() missing preview line: %q", rendered)
	}
	if strings.Contains(rendered, "one.md") {
		t.Fatalf("notePaletteView() should scroll past early items: %q", rendered)
	}
}

func TestNotePaletteViewShowsCreateHintWhenNoMatches(t *testing.T) {
	model := New(config.Config{}, "", "test")

	rendered := model.notePaletteView()

	if !strings.Contains(rendered, "No matching notes") {
		t.Fatalf("notePaletteView() missing empty message: %q", rendered)
	}
	if !strings.Contains(rendered, "Press Enter to create a new note") {
		t.Fatalf("notePaletteView() missing create hint: %q", rendered)
	}
}

func TestLauncherViewReservesPaletteSpaceWhenInputIsEmpty(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 80
	model.height = 24

	rendered := model.View()

	if !strings.Contains(rendered, "Start typing to search notes, create one, or") {
		t.Fatalf("View() missing empty launcher placeholder: %q", rendered)
	}
	if !strings.Contains(rendered, "run a command") {
		t.Fatalf("View() missing empty launcher placeholder: %q", rendered)
	}
}
