package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func backupNotes(notesPath, backupPath string) (string, error) {
	if err := os.MkdirAll(backupPath, 0o755); err != nil {
		return "", fmt.Errorf("could not create backup dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	zipName := fmt.Sprintf("nocte_backup_%s.zip", timestamp)
	zipPath := filepath.Join(backupPath, zipName)

	f, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("could not create backup file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	err = filepath.Walk(notesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(notesPath, path)
		if err != nil {
			return err
		}

		w, err := zw.Create(rel)
		if err != nil {
			return fmt.Errorf("could not add %s to zip: %w", rel, err)
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not read %s: %w", rel, err)
		}
		defer src.Close()

		if _, err := io.Copy(w, src); err != nil {
			return fmt.Errorf("could not write %s to zip: %w", rel, err)
		}

		return nil
	})
	if err != nil {
		os.Remove(zipPath)
		return "", fmt.Errorf("backup failed: %w", err)
	}

	return zipPath, nil
}

func (m *Model) runBackup() error {
	zipPath, err := backupNotes(m.config.NotesPath, m.config.BackupPath)
	if err != nil {
		return err
	}

	m.status = fmt.Sprintf("Backup saved to %s", zipPath)
	m.isError = false
	return nil
}
