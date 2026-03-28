package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/inrd/nocte/internal/config"
)

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

func TestListDialogKeepsTimestampOnOneLineWhenWideEnough(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 120
	model.dialogNotes = []noteMatch{
		{
			name:      "windows audio dev.md",
			wordCount: 51,
			sizeBytes: 410,
			modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
		},
	}
	model.dialogIndex = 0

	rendered := model.listDialog()

	if strings.Contains(rendered, "Apr 29 2024\n16:35") {
		t.Fatalf("listDialog() wrapped the timestamp:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Apr 29 2024 16:35") {
		t.Fatalf("listDialog() = %q, want single-line timestamp", rendered)
	}
}

func TestListDialogShowsTaskProgressForNotesWithTasks(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 120
	model.dialogNotes = []noteMatch{
		{
			name:      "todo.md",
			wordCount: 12,
			sizeBytes: 128,
			modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
			taskDone:  1,
			taskTotal: 2,
		},
	}
	model.dialogIndex = 0

	rendered := model.listDialog()

	if !strings.Contains(rendered, "50%") {
		t.Fatalf("listDialog() = %q, want task percentage", rendered)
	}
	if !strings.Contains(rendered, renderListTaskProgressBar(50)) {
		t.Fatalf("listDialog() = %q, want compact task progress bar", rendered)
	}
}

func TestListDialogOmitsTaskProgressForNotesWithoutTasks(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 120
	model.dialogNotes = []noteMatch{
		{
			name:      "plain.md",
			wordCount: 12,
			sizeBytes: 128,
			modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
		},
	}
	model.dialogIndex = 0

	rendered := model.listDialog()

	if strings.Contains(rendered, "%") {
		t.Fatalf("listDialog() = %q, want no task percentage", rendered)
	}
	if strings.Contains(rendered, "█") || strings.Contains(rendered, "░") {
		t.Fatalf("listDialog() = %q, want no task progress bar", rendered)
	}
}

func TestListDialogLineReservesProgressColumnWidth(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 120

	withTasks := model.listDialogLine(noteMatch{
		name:      "todo.md",
		wordCount: 12,
		sizeBytes: 128,
		modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
		taskDone:  1,
		taskTotal: 2,
	}, false)
	withoutTasks := model.listDialogLine(noteMatch{
		name:      "plain.md",
		wordCount: 12,
		sizeBytes: 128,
		modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
	}, false)

	if ansi.StringWidth(withTasks) != ansi.StringWidth(withoutTasks) {
		t.Fatalf("listDialogLine() widths = %d and %d, want equal reserved column width", ansi.StringWidth(withTasks), ansi.StringWidth(withoutTasks))
	}
}

func TestListDialogSelectedRowKeepsMetadataReadable(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 120

	line := model.listDialogLine(noteMatch{
		name:      "todo.md",
		wordCount: 12,
		sizeBytes: 128,
		modTime:   time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local),
		taskDone:  1,
		taskTotal: 2,
	}, true)

	if !strings.Contains(line, formatNoteUpdatedAt(time.Date(2024, time.April, 29, 16, 35, 0, 0, time.Local))) {
		t.Fatalf("selected listDialogLine() = %q, want readable updated-at text", line)
	}
	if !strings.Contains(line, noteMeta(noteMatch{wordCount: 12, sizeBytes: 128})) {
		t.Fatalf("selected listDialogLine() = %q, want readable metadata text", line)
	}
	if !strings.Contains(line, "50%") {
		t.Fatalf("selected listDialogLine() = %q, want readable task percentage", line)
	}
}

func TestLinksDialogShowsLabelAndURL(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.width = 100
	model.activeDialog = "links"
	model.dialogLinks = []noteLink{
		{label: "Project board", url: "https://example.com/board"},
	}
	model.dialogIndex = 0

	rendered := model.linksDialog()

	if !strings.Contains(rendered, "Project board") {
		t.Fatalf("linksDialog() missing label: %q", rendered)
	}
	if !strings.Contains(rendered, "https://example.com/board") {
		t.Fatalf("linksDialog() missing url: %q", rendered)
	}
}

func TestDeleteConfirmDialogShowsNoteName(t *testing.T) {
	model := New(config.Config{}, "", "test")
	model.editorName = "draft.md"

	rendered := model.deleteConfirmDialog()

	if !strings.Contains(rendered, "Delete draft.md?") {
		t.Fatalf("deleteConfirmDialog() missing note name: %q", rendered)
	}
	if !strings.Contains(rendered, "Press Enter to delete this note.") {
		t.Fatalf("deleteConfirmDialog() missing confirm instructions: %q", rendered)
	}
}
