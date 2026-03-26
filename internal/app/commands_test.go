package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inrd/nocte/internal/config"
)

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

func TestHandleCommandFilesLaunchesNotesFolder(t *testing.T) {
	tmpDir := t.TempDir()
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
	model.input.SetValue(":files")
	model.syncLauncherState()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if openedPath != tmpDir {
		t.Fatalf("openedPath = %q, want %q", openedPath, tmpDir)
	}
	if model.input.Value() != "" {
		t.Fatalf("input.Value() = %q, want empty string", model.input.Value())
	}
	if model.status != "Opened notes folder" {
		t.Fatalf("status = %q, want %q", model.status, "Opened notes folder")
	}
	if model.isError {
		t.Fatalf("isError = true, want false")
	}
}

func TestHandleCommandFilesCreatesMissingNotesFolder(t *testing.T) {
	baseDir := t.TempDir()
	notesDir := filepath.Join(baseDir, "notes")
	original := openPathWithSystemApp
	t.Cleanup(func() {
		openPathWithSystemApp = original
	})

	openPathWithSystemApp = func(path string) error {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("Stat(%q) error = %v, want created directory", path, err)
		}
		return nil
	}

	model := New(config.Config{NotesPath: notesDir}, "", "test")
	model.input.SetValue(":files")
	model.syncLauncherState()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	info, err := os.Stat(notesDir)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", notesDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", notesDir)
	}
}

func TestHandleCommandFilesShowsLaunchError(t *testing.T) {
	tmpDir := t.TempDir()
	original := openPathWithSystemApp
	t.Cleanup(func() {
		openPathWithSystemApp = original
	})

	openPathWithSystemApp = func(path string) error {
		return exec.ErrNotFound
	}

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue(":files")
	model.syncLauncherState()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if !model.isError {
		t.Fatalf("isError = false, want true")
	}
	if !strings.Contains(model.status, "Could not open notes dir") {
		t.Fatalf("status = %q, want open-notes-dir error", model.status)
	}
}

func TestFilteredCommandsPrefersNamePrefixMatches(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.input.SetValue(":l")

	got := commandNames(model.filteredCommands())
	want := []string{":list", ":help", ":files"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filteredCommands() = %v, want %v", got, want)
	}
}

func TestFilteredCommandsPrefersNameMatchesOverDescriptionMatches(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.input.SetValue(":inf")

	got := commandNames(model.filteredCommands())
	want := []string{":info"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filteredCommands() = %v, want %v", got, want)
	}
}

func commandNames(commands []command) []string {
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.name)
	}

	return names
}
