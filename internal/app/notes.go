package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func sanitizeFilename(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("Note name cannot be blank")
	}

	normalized := strings.ToLower(trimmed)
	normalized = strings.ReplaceAll(normalized, " ", "-")
	normalized = invalidFileChars.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-.")

	if normalized == "" {
		return "", fmt.Errorf("Note name must contain letters or numbers")
	}

	return normalized, nil
}

func (m Model) shouldShowNotePalette() bool {
	return !m.isCommandMode() && !m.isSearchMode() && strings.TrimSpace(m.input.Value()) != ""
}

func (m Model) hasNotePalette() bool {
	return m.shouldShowNotePalette() && len(m.noteMatches) > 0
}

func (m *Model) syncLauncherState() {
	m.todoMode = false

	if m.isCommandMode() {
		m.syncCommandSelection()
		m.noteMatches = nil
		m.noteIndex = -1
		m.noteOffset = 0
		m.searchMatches = nil
		m.searchIndex = -1
		m.searchOffset = 0
		return
	}

	if m.isSearchMode() {
		m.commandIndex = 0
		m.commandOffset = 0
		m.noteMatches = nil
		m.noteIndex = -1
		m.noteOffset = 0
		m.searchMatches = m.findSearchMatches(m.searchQuery())
		m.searchIndex = -1
		m.searchOffset = 0
		return
	}

	m.commandIndex = 0
	m.commandOffset = 0
	m.noteMatches = m.findNoteMatches(strings.TrimSpace(m.input.Value()))
	m.noteIndex = -1
	m.noteOffset = 0
	m.searchMatches = nil
	m.searchIndex = -1
	m.searchOffset = 0
}

func (m *Model) moveNoteSelection(delta int) {
	if len(m.noteMatches) == 0 {
		m.noteIndex = -1
		m.noteOffset = 0
		return
	}

	if m.noteIndex == -1 {
		if delta > 0 {
			m.noteIndex = 0
			m.syncNoteOffset()
			return
		}

		m.noteIndex = len(m.noteMatches) - 1
		m.syncNoteOffset()
		return
	}

	m.noteIndex = (m.noteIndex + delta + len(m.noteMatches)) % len(m.noteMatches)
	m.syncNoteOffset()
}

func (m *Model) openListDialog() {
	m.activeDialog = "list"
	m.dialogNotes = m.listNotes()
	m.dialogIndex = -1
	m.dialogOffset = 0
	if len(m.dialogNotes) > 0 {
		m.dialogIndex = 0
	}
	m.input.SetValue(":list")
	m.input.CursorEnd()
	m.input.Blur()
	m.status = ""
	m.isError = false
}

func (m *Model) openTodoPalette() {
	m.todoMode = true
	m.input.SetValue(":todo")
	m.input.CursorEnd()
	m.input.Focus()
	m.commandIndex = 0
	m.commandOffset = 0
	m.noteMatches = nil
	m.noteIndex = -1
	m.noteOffset = 0
	m.searchMatches = m.findTodoMatches()
	m.searchIndex = -1
	m.searchOffset = 0
	m.status = ""
	m.isError = false
}

func (m *Model) moveDialogSelection(delta int) {
	total := m.dialogItems()
	if total == 0 {
		m.dialogIndex = -1
		m.dialogOffset = 0
		return
	}

	m.dialogIndex = (m.dialogIndex + delta + total) % total
	m.syncDialogOffset()
}

func (m *Model) openSelectedDialogNote() error {
	if len(m.dialogNotes) == 0 {
		m.closeDialog()
		return nil
	}

	if m.dialogIndex < 0 || m.dialogIndex >= len(m.dialogNotes) {
		m.dialogIndex = 0
	}

	return m.openExistingNote(m.dialogNotes[m.dialogIndex])
}

func (m *Model) syncDialogOffset() {
	visible := m.dialogVisibleCount()
	if visible <= 0 {
		m.dialogOffset = 0
		return
	}

	maxOffset := max(0, len(m.dialogNotes)-visible)
	if m.activeDialog == "links" {
		maxOffset = max(0, len(m.dialogLinks)-visible)
	}
	if m.dialogOffset > maxOffset {
		m.dialogOffset = maxOffset
	}

	if m.dialogIndex < m.dialogOffset {
		m.dialogOffset = m.dialogIndex
	}

	if m.dialogIndex >= m.dialogOffset+visible {
		m.dialogOffset = m.dialogIndex - visible + 1
	}
}

func (m Model) findNoteMatches(query string) []noteMatch {
	entries, err := os.ReadDir(m.config.NotesPath)
	if err != nil {
		return nil
	}

	matches := make([]noteMatch, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(m.config.NotesPath, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		taskDone, taskTotal := countTaskProgress(string(content))

		matches = append(matches, noteMatch{
			name:      name,
			path:      path,
			wordCount: len(strings.Fields(string(content))),
			sizeBytes: info.Size(),
			modTime:   info.ModTime(),
			taskDone:  taskDone,
			taskTotal: taskTotal,
			preview:   notePreviewLines(string(content)),
		})
	}

	if query == "" {
		sort.Slice(matches, func(i int, j int) bool {
			return matches[i].name < matches[j].name
		})
		return matches
	}

	filtered := make([]noteMatch, 0, len(matches))
	for _, note := range matches {
		score, ok := fuzzyScore(strings.TrimSuffix(note.name, filepath.Ext(note.name)), query)
		if !ok {
			continue
		}

		note.score = score
		filtered = append(filtered, note)
	}

	sort.Slice(filtered, func(i int, j int) bool {
		if filtered[i].score == filtered[j].score {
			return filtered[i].name < filtered[j].name
		}
		return filtered[i].score < filtered[j].score
	})

	return filtered
}

func (m *Model) syncNoteOffset() {
	if len(m.noteMatches) == 0 {
		m.noteOffset = 0
		return
	}

	budget := m.launcherPaletteContentBudget()
	if len(m.noteMatches) > 1 {
		budget--
	}
	if budget <= 0 {
		m.noteOffset = 0
		return
	}

	m.noteOffset = clampInt(m.noteOffset, 0, len(m.noteMatches)-1)
	if m.noteIndex < 0 {
		return
	}

	for m.noteIndex < m.noteOffset {
		m.noteOffset = m.noteIndex
	}

	for {
		start, end := m.noteVisibleRangeFrom(m.noteOffset)
		if m.noteIndex >= start && m.noteIndex < end {
			return
		}
		m.noteOffset++
		if m.noteOffset >= len(m.noteMatches) {
			m.noteOffset = len(m.noteMatches) - 1
			return
		}
	}
}

func (m Model) listNotes() []noteMatch {
	notes := m.findNoteMatches("")
	sort.Slice(notes, func(i int, j int) bool {
		if notes[i].modTime.Equal(notes[j].modTime) {
			return notes[i].name < notes[j].name
		}

		return notes[i].modTime.After(notes[j].modTime)
	})

	return notes
}

func noteMeta(note noteMatch) string {
	return fmt.Sprintf("%d words | %s", note.wordCount, humanSize(note.sizeBytes))
}

func notePreviewLines(content string) []string {
	lines := strings.Split(content, "\n")
	preview := make([]string, 0, 2)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		preview = append(preview, line)
		if len(preview) == 2 {
			return preview
		}
	}

	if len(preview) == 0 {
		return []string{"Empty note"}
	}

	return preview
}

func (m *Model) closeDialog() {
	if m.activeDialog == "save-error" {
		m.activeDialog = ""
		m.editorAction = ""
		m.status = fmt.Sprintf("Still editing %s", m.editorName)
		m.isError = false
		return
	}

	if m.activeDialog == "delete-confirm" {
		m.activeDialog = ""
		m.status = fmt.Sprintf("Still editing %s", m.editorName)
		m.isError = false
		m.editor.Focus()
		return
	}

	if m.isEditing() {
		m.activeDialog = ""
		m.dialogLinks = nil
		m.dialogIndex = -1
		m.dialogOffset = 0
		m.editor.Focus()
		return
	}

	m.activeDialog = ""
	m.dialogNotes = nil
	m.dialogLinks = nil
	m.dialogIndex = -1
	m.dialogOffset = 0
	m.input.SetValue("")
	m.input.Focus()
}

func (m Model) dialogItems() int {
	switch m.activeDialog {
	case "links":
		return len(m.dialogLinks)
	default:
		return len(m.dialogNotes)
	}
}
