package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/inrd/nocte/internal/config"
)

func TestFindSearchMatchesReturnsOneRowPerMatchWithSnippets(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "alpha.md", "before\nFind the needle here\nafter")
	writeTestNote(t, tmpDir, "beta.md", "needle and another needle")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")

	matches := model.findSearchMatches("needle")
	if len(matches) != 3 {
		t.Fatalf("len(matches) = %d, want 3", len(matches))
	}

	if matches[0].name != "alpha.md" {
		t.Fatalf("matches[0].name = %q, want %q", matches[0].name, "alpha.md")
	}
	if matches[0].lineNumber != 2 {
		t.Fatalf("matches[0].lineNumber = %d, want 2", matches[0].lineNumber)
	}
	if matches[0].column != 9 {
		t.Fatalf("matches[0].column = %d, want 9", matches[0].column)
	}

	wantSnippet := []string{"before", "Find the needle here", "after"}
	if !reflect.DeepEqual(matches[0].snippetLines, wantSnippet) {
		t.Fatalf("matches[0].snippetLines = %v, want %v", matches[0].snippetLines, wantSnippet)
	}

	if matches[1].name != "beta.md" || matches[1].column != 0 {
		t.Fatalf("matches[1] = %+v, want beta.md at column 0", matches[1])
	}
	if matches[2].name != "beta.md" || matches[2].column != 19 {
		t.Fatalf("matches[2] = %+v, want beta.md at column 19", matches[2])
	}
}

func TestFindTodoMatchesReturnsOneRowPerTask(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "alpha.md", "- [ ] first\nnot a task\n  * [x] second")
	writeTestNote(t, tmpDir, "beta.md", "plain\n+ [ ] third")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")

	matches := model.findTodoMatches()
	if len(matches) != 2 {
		t.Fatalf("len(matches) = %d, want 2", len(matches))
	}
	if matches[0].name != "alpha.md" || matches[0].lineNumber != 1 {
		t.Fatalf("matches[0] = %+v, want alpha.md line 1", matches[0])
	}
	if !reflect.DeepEqual(matches[1].snippetLines, []string{"+ [ ] third"}) {
		t.Fatalf("matches[1].snippetLines = %v, want %v", matches[1].snippetLines, []string{"+ [ ] third"})
	}
	if matches[1].name != "beta.md" || matches[1].lineNumber != 2 {
		t.Fatalf("matches[1] = %+v, want beta.md line 2", matches[1])
	}
}

func TestUpdateSearchModeEnterOpensSelectedMatchAtLineAndColumn(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "meeting.md", "alpha line\nbeta target here\ngamma line")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("/target")
	model.syncLauncherState()
	model.searchIndex = 0

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.editorPath != notePath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, notePath)
	}
	if model.editor.Line() != 1 {
		t.Fatalf("editor.Line() = %d, want 1", model.editor.Line())
	}
	if model.editor.LineInfo().CharOffset != 5 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want 5", model.editor.LineInfo().CharOffset)
	}
	if model.status != "Editing meeting.md at line 2" {
		t.Fatalf("status = %q, want %q", model.status, "Editing meeting.md at line 2")
	}
}

func TestUpdateSearchModeEnterJumpsPastWrappedVisualRows(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "meeting.md", "this is a very long first line that will wrap in a narrow editor width\nbeta target here\nthird line")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.width = 40
	model.height = 24
	model.input.SetValue("/target")
	model.syncLauncherState()
	model.searchIndex = 0

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.editorPath != notePath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, notePath)
	}
	if model.editor.Line() != 1 {
		t.Fatalf("editor.Line() = %d, want 1", model.editor.Line())
	}
	if model.editor.LineInfo().CharOffset != 5 {
		t.Fatalf("editor.LineInfo().CharOffset = %d, want 5", model.editor.LineInfo().CharOffset)
	}
}

func TestUpdateSearchModeEnterWithoutSelectionDoesNotCreateNote(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "meeting.md", "beta target here")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.input.SetValue("/target")
	model.syncLauncherState()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.isEditing() {
		t.Fatalf("model should not enter editor without a selected search result")
	}
	if model.status != "Select a search result with Up or Down" {
		t.Fatalf("status = %q, want %q", model.status, "Select a search result with Up or Down")
	}

	createdPath := filepath.Join(tmpDir, "target.md")
	if _, err := os.Stat(createdPath); err == nil {
		t.Fatalf("Stat(%q) = nil, want no created file", createdPath)
	}
}

func TestUpdateTodoModeEnterOpensSelectedTaskAtLine(t *testing.T) {
	tmpDir := t.TempDir()
	notePath := writeTestNote(t, tmpDir, "meeting.md", "alpha\n- [ ] beta task\ngamma")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openTodoPalette()
	model.searchIndex = 0

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.editorPath != notePath {
		t.Fatalf("editorPath = %q, want %q", model.editorPath, notePath)
	}
	if model.editor.Line() != 1 {
		t.Fatalf("editor.Line() = %d, want 1", model.editor.Line())
	}
	if model.status != "Editing meeting.md at line 2" {
		t.Fatalf("status = %q, want %q", model.status, "Editing meeting.md at line 2")
	}
}

func TestUpdateTodoModeEnterWithoutSelectionPromptsForChoice(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestNote(t, tmpDir, "meeting.md", "- [ ] beta task")

	model := New(config.Config{NotesPath: tmpDir}, "", "test")
	model.openTodoPalette()

	model = updateModel(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	if model.isEditing() {
		t.Fatalf("model should not enter editor without a selected todo result")
	}
	if model.status != "Select a task with Up or Down" {
		t.Fatalf("status = %q, want %q", model.status, "Select a task with Up or Down")
	}
}

func TestSearchVisibleRangeKeepsResultsWithinViewportBudget(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.height = 20
	model.searchMatches = make([]searchMatch, 6)
	for i := range model.searchMatches {
		model.searchMatches[i] = searchMatch{
			name:         "note.md",
			lineNumber:   i + 1,
			snippetLines: []string{"one", "two", "three"},
		}
	}

	start, end := model.searchVisibleRange()
	if start != 0 {
		t.Fatalf("start = %d, want 0", start)
	}
	if end != 1 {
		t.Fatalf("end = %d, want 1 for capped viewport", end)
	}

	model.searchIndex = 3
	model.syncSearchOffset()
	start, end = model.searchVisibleRange()
	if model.searchOffset != 3 {
		t.Fatalf("searchOffset = %d, want 3", model.searchOffset)
	}
	if start != 3 || end != 4 {
		t.Fatalf("visible range = %d..%d, want 3..4", start, end)
	}
}
