package app

import (
	"strings"
	"testing"
	"time"

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
