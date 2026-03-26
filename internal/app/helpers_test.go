package app

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

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

func TestEditorSizeStatus(t *testing.T) {
	content := "hello"
	if got := editorSizeStatus(content); got != "Size 5 B" {
		t.Fatalf("editorSizeStatus(%q) = %q, want %q", content, got, "Size 5 B")
	}
}

func TestLargeNoteWarning(t *testing.T) {
	below := largeNoteWarning(largeNoteWarningThreshold)
	if below != "" {
		t.Fatalf("largeNoteWarning(%d) = %q, want empty string", largeNoteWarningThreshold, below)
	}

	above := largeNoteWarning(largeNoteWarningThreshold + 1)
	if above == "" {
		t.Fatalf("largeNoteWarning(%d) = empty string, want warning", largeNoteWarningThreshold+1)
	}
	if !strings.Contains(above, "may slow editing and saving") {
		t.Fatalf("largeNoteWarning(%d) = %q, want slowdown warning", largeNoteWarningThreshold+1, above)
	}
}
