package app

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupNotesCreatesZip(t *testing.T) {
	notesDir := t.TempDir()
	backupDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(notesDir, "note1.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notesDir, "note2.md"), []byte("world"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	zipPath, err := backupNotes(notesDir, backupDir)
	if err != nil {
		t.Fatalf("backupNotes() error = %v", err)
	}

	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("zip file not found at %s: %v", zipPath, err)
	}

	if !strings.HasPrefix(filepath.Base(zipPath), "nocte_backup_") {
		t.Fatalf("zip file name %q does not start with nocte_backup_", filepath.Base(zipPath))
	}
	if !strings.HasSuffix(zipPath, ".zip") {
		t.Fatalf("zip file name %q does not end with .zip", zipPath)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("zip.OpenReader: %v", err)
	}
	defer zr.Close()

	found := make(map[string]bool)
	for _, f := range zr.File {
		found[f.Name] = true
	}
	for _, want := range []string{"note1.md", "note2.md"} {
		if !found[want] {
			t.Errorf("zip missing entry %q", want)
		}
	}
}

func TestBackupNotesCreatesBackupDir(t *testing.T) {
	notesDir := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "nested", "backups")

	if err := os.WriteFile(filepath.Join(notesDir, "note.md"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := backupNotes(notesDir, backupDir); err != nil {
		t.Fatalf("backupNotes() error = %v", err)
	}

	if _, err := os.Stat(backupDir); err != nil {
		t.Fatalf("backup dir not created: %v", err)
	}
}

func TestBackupNotesEmptyDir(t *testing.T) {
	notesDir := t.TempDir()
	backupDir := t.TempDir()

	zipPath, err := backupNotes(notesDir, backupDir)
	if err != nil {
		t.Fatalf("backupNotes() error = %v", err)
	}

	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("zip.OpenReader: %v", err)
	}
	defer zr.Close()

	if len(zr.File) != 0 {
		t.Fatalf("expected empty zip, got %d entries", len(zr.File))
	}
}
