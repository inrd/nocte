package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadOrCreateCreatesDefaultConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, configPath, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}

	wantPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if configPath != wantPath {
		t.Fatalf("configPath = %q, want %q", configPath, wantPath)
	}

	wantNotesPath := filepath.Join(tmpHome, "nocte")
	if cfg.NotesPath != wantNotesPath {
		t.Fatalf("cfg.NotesPath = %q, want %q", cfg.NotesPath, wantNotesPath)
	}
	wantBackupPath := filepath.Join(tmpHome, "nocte_backups")
	if cfg.BackupPath != wantBackupPath {
		t.Fatalf("cfg.BackupPath = %q, want %q", cfg.BackupPath, wantBackupPath)
	}
	if cfg.TabWidth != defaultTabWidth {
		t.Fatalf("cfg.TabWidth = %d, want %d", cfg.TabWidth, defaultTabWidth)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", configPath, err)
	}

	if len(data) == 0 {
		t.Fatalf("config file %q was empty", configPath)
	}
}

func TestLoadBackfillsBlankNotesPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{\n  \"notes_path\": \"   \"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := load(configPath)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	wantNotesPath := filepath.Join(tmpHome, "nocte")
	if cfg.NotesPath != wantNotesPath {
		t.Fatalf("cfg.NotesPath = %q, want %q", cfg.NotesPath, wantNotesPath)
	}
	if cfg.TabWidth != defaultTabWidth {
		t.Fatalf("cfg.TabWidth = %d, want %d", cfg.TabWidth, defaultTabWidth)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", configPath, err)
	}

	if string(saved) == "{\n  \"notes_path\": \"   \"\n}\n" {
		t.Fatalf("load() did not rewrite blank notes_path")
	}
}

func TestLoadBackfillsBlankBackupPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{\n  \"notes_path\": \"~/notes\",\n  \"backup_path\": \"   \"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := load(configPath)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	wantBackupPath := filepath.Join(tmpHome, "nocte_backups")
	if cfg.BackupPath != wantBackupPath {
		t.Fatalf("cfg.BackupPath = %q, want %q", cfg.BackupPath, wantBackupPath)
	}
}

func TestLoadExpandsHomeInBackupPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{\n  \"notes_path\": \"~/notes\",\n  \"backup_path\": \"~/my_backups\"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := load(configPath)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	want := filepath.Join(tmpHome, "my_backups")
	if cfg.BackupPath != want {
		t.Fatalf("cfg.BackupPath = %q, want %q", cfg.BackupPath, want)
	}
}

func TestLoadExpandsHomeInNotesPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{\n  \"notes_path\": \"~/notes\"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := load(configPath)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	want := filepath.Join(tmpHome, "notes")
	if cfg.NotesPath != want {
		t.Fatalf("cfg.NotesPath = %q, want %q", cfg.NotesPath, want)
	}
	if cfg.TabWidth != defaultTabWidth {
		t.Fatalf("cfg.TabWidth = %d, want %d", cfg.TabWidth, defaultTabWidth)
	}
}

func TestLoadBackfillsInvalidTabWidth(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	configPath := filepath.Join(tmpHome, ".config", "nocte", fileName)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{\n  \"notes_path\": \"~/notes\",\n  \"tab_width\": 0\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := load(configPath)
	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.TabWidth != defaultTabWidth {
		t.Fatalf("cfg.TabWidth = %d, want %d", cfg.TabWidth, defaultTabWidth)
	}

	saved, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", configPath, err)
	}

	if !strings.Contains(string(saved), "\"tab_width\": 4") {
		t.Fatalf("saved config = %q, want backfilled tab_width", string(saved))
	}
}

func TestExpandHome(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "bare home", input: "~", want: tmpHome},
		{name: "home child", input: "~/notes", want: filepath.Join(tmpHome, "notes")},
		{name: "unchanged absolute", input: "/tmp/notes", want: "/tmp/notes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandHome(tt.input)
			if err != nil {
				t.Fatalf("expandHome(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
