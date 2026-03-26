package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
