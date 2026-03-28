package app

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) isEditing() bool {
	return m.editorPath != ""
}

func (m *Model) openEditor(path string, name string) {
	content, err := os.ReadFile(path)
	if err != nil {
		m.status = fmt.Sprintf("Could not open %s: %v", name, err)
		m.isError = true
		return
	}

	m.editorPath = path
	m.editorName = name
	m.lastSaved = string(content)
	m.editorCreated = false
	m.editor.SetValue(m.lastSaved)
	m.editor.Focus()
	m.resizeEditor()
	m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyCtrlHome})
	m.input.SetValue("")
	m.input.Blur()
	m.activeDialog = ""
	m.commandIndex = 0
	m.noteIndex = -1
	m.searchIndex = -1
	m.noteMatches = nil
	m.searchMatches = nil
	m.todoMode = false
	m.dialogNotes = nil
	m.dialogIndex = -1
	m.dialogOffset = 0
	m.editorAction = ""
	m.status = fmt.Sprintf("Editing %s", name)
	m.isError = false
}

func (m *Model) closeEditor(saved bool) {
	name := m.editorName
	m.editorPath = ""
	m.editorName = ""
	m.lastSaved = ""
	m.editorCreated = false
	m.editorAction = ""
	m.editor.SetValue("")
	m.editor.Blur()
	m.input.SetValue("")
	m.input.Focus()
	m.syncLauncherState()
	if saved {
		m.status = fmt.Sprintf("Saved and closed %s", name)
	} else {
		m.status = fmt.Sprintf("Closed %s without changes", name)
	}
	m.isError = false
}

func (m *Model) discardEditor() {
	name := m.editorName
	m.editorPath = ""
	m.editorName = ""
	m.lastSaved = ""
	m.editorCreated = false
	m.editorAction = ""
	m.editor.SetValue("")
	m.editor.Blur()
	m.activeDialog = ""
	m.input.SetValue("")
	m.input.Focus()
	m.syncLauncherState()
	m.status = fmt.Sprintf("Discarded changes in %s", name)
	m.isError = false
}

func (m *Model) deleteEditorNote() error {
	name := m.editorName
	if err := os.Remove(m.editorPath); err != nil {
		return fmt.Errorf("could not delete %s: %w", name, err)
	}

	m.editorPath = ""
	m.editorName = ""
	m.lastSaved = ""
	m.editorCreated = false
	m.editorAction = ""
	m.activeDialog = ""
	m.dialogNotes = nil
	m.dialogLinks = nil
	m.dialogIndex = -1
	m.dialogOffset = 0
	m.editor.SetValue("")
	m.editor.Blur()
	m.input.SetValue("")
	m.input.Focus()
	m.syncLauncherState()
	m.status = fmt.Sprintf("Deleted %s", name)
	m.isError = false
	return nil
}

func (m *Model) resizeEditor() {
	width := m.editorPaneWidth()
	height := 12

	if m.height > 0 {
		height = max(8, m.height-10)
	}

	m.editor.SetWidth(width)
	m.editor.SetHeight(height)
}

func (m *Model) saveEditor() (bool, error) {
	content := m.editor.Value()
	if content == m.lastSaved {
		return false, nil
	}

	if err := os.WriteFile(m.editorPath, []byte(content), 0o644); err != nil {
		return false, fmt.Errorf("could not save %s: %w", m.editorName, err)
	}

	m.lastSaved = content
	m.status = fmt.Sprintf("Saved %s", m.editorName)
	m.isError = false
	return true, nil
}

func (m *Model) discardEmptyCreatedNote() (bool, error) {
	if !m.editorCreated || m.editor.Value() != "" || m.lastSaved != "" {
		return false, nil
	}

	content, err := os.ReadFile(m.editorPath)
	if err != nil {
		return false, fmt.Errorf("could not check %s before closing: %w", m.editorName, err)
	}
	if len(content) != 0 {
		return false, nil
	}

	if err := os.Remove(m.editorPath); err != nil {
		return false, fmt.Errorf("could not remove empty note %s: %w", m.editorName, err)
	}

	m.status = fmt.Sprintf("Discarded empty note %s", m.editorName)
	m.isError = false
	return true, nil
}

func (m *Model) finishEditing(action string) bool {
	discarded, err := m.discardEmptyCreatedNote()
	if err != nil {
		m.activeDialog = "save-error"
		m.editorAction = action
		m.status = err.Error()
		m.isError = true
		return false
	}
	if discarded {
		name := m.editorName
		m.closeEditor(false)
		m.status = fmt.Sprintf("Discarded empty note %s", name)
		return false
	}

	saved, err := m.saveEditor()
	if err != nil {
		m.activeDialog = "save-error"
		m.editorAction = action
		m.status = err.Error()
		m.isError = true
		return false
	}

	m.closeEditor(saved)
	return false
}

func (m *Model) openExistingNote(note noteMatch) error {
	m.openEditor(note.path, note.name)
	if m.isError {
		return errors.New(m.status)
	}

	return nil
}

func (m *Model) openSearchMatch(match searchMatch) error {
	m.openEditor(match.path, match.name)
	if m.isError {
		return errors.New(m.status)
	}

	m.jumpEditorTo(match.lineNumber-1, match.column)
	m.status = fmt.Sprintf("Editing %s at line %d", match.name, match.lineNumber)
	return nil
}

func (m *Model) jumpEditorTo(line int, column int) {
	if line < 0 {
		line = 0
	}

	m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyCtrlHome})
	for m.editor.Line() < line {
		m.editor.CursorEnd()
		previousLine := m.editor.Line()
		m.editor, _ = m.editor.Update(tea.KeyMsg{Type: tea.KeyDown})
		if m.editor.Line() == previousLine {
			break
		}
	}

	m.editor.SetCursor(column)
}

func (m *Model) toggleEditorTask() {
	lines := strings.Split(m.editor.Value(), "\n")
	line := m.editor.Line()
	if line < 0 || line >= len(lines) {
		return
	}

	column := m.editor.LineInfo().CharOffset
	lines[line], column = toggleTaskLine(lines[line], column)
	m.editor.SetValue(strings.Join(lines, "\n"))
	m.jumpEditorTo(line, column)
}

func (m *Model) copyCodeAtCursor() error {
	content, ok := m.codeContentAtCursor()
	if !ok {
		return errors.New("cursor is not inside Markdown code")
	}

	if err := writeClipboardText(content); err != nil {
		return err
	}

	return nil
}

func (m Model) codeContentAtCursor() (string, bool) {
	lines := strings.Split(m.editor.Value(), "\n")
	lineIndex := m.editor.Line()
	if lineIndex < 0 || lineIndex >= len(lines) {
		return "", false
	}

	if content, ok := codeBlockContentAtLine(lines, lineIndex); ok {
		return content, true
	}

	column := m.editor.LineInfo().CharOffset
	return inlineCodeContentAtColumn(lines[lineIndex], column)
}

func codeBlockContentAtLine(lines []string, target int) (string, bool) {
	inCodeBlock := false
	blockStart := -1

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inCodeBlock {
				if target > blockStart && target < i {
					return strings.Join(lines[blockStart+1:i], "\n"), true
				}
				inCodeBlock = false
				blockStart = -1
				continue
			}

			inCodeBlock = true
			blockStart = i
		}
	}

	if inCodeBlock && target > blockStart {
		return strings.Join(lines[blockStart+1:], "\n"), true
	}

	return "", false
}

func inlineCodeContentAtColumn(line string, column int) (string, bool) {
	runes := []rune(line)
	if len(runes) == 0 {
		return "", false
	}

	column = max(0, min(column, len(runes)))
	for i := 0; i < len(runes); i++ {
		if runes[i] != '`' {
			continue
		}

		delimiterWidth := 1
		for i+delimiterWidth < len(runes) && runes[i+delimiterWidth] == '`' {
			delimiterWidth++
		}

		closeIndex := -1
		for j := i + delimiterWidth; j < len(runes); j++ {
			matched := true
			for k := 0; k < delimiterWidth; k++ {
				if j+k >= len(runes) || runes[j+k] != '`' {
					matched = false
					break
				}
			}
			if matched {
				closeIndex = j
				break
			}
		}
		if closeIndex < 0 {
			break
		}

		if column > i && column < closeIndex+delimiterWidth {
			return string(runes[i+delimiterWidth : closeIndex]), true
		}

		i = closeIndex + delimiterWidth - 1
	}

	return "", false
}

func toggleTaskLine(line string, column int) (string, int) {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	indent := line[:len(line)-len(trimmed)]

	oldPrefix := indent
	newPrefix := indent + "- [ ] "
	updated := newPrefix + trimmed

	switch {
	case isTaskListLine(line):
		oldPrefix = indent + trimmed[:6]
		marker := "[x]"
		if trimmed[3] == 'x' || trimmed[3] == 'X' {
			marker = "[ ]"
		}
		newPrefix = indent + string(trimmed[0]) + " " + marker + " "
		updated = newPrefix + trimmed[6:]
	case isBulletLine(line):
		oldPrefix = indent + trimmed[:2]
		newPrefix = indent + string(trimmed[0]) + " [ ] "
		updated = newPrefix + trimmed[2:]
	}

	return updated, adjustTaskToggleColumn(column, len([]rune(oldPrefix)), len([]rune(newPrefix)))
}

func adjustTaskToggleColumn(column int, oldPrefixLen int, newPrefixLen int) int {
	if column < 0 {
		return 0
	}
	if oldPrefixLen == newPrefixLen {
		return column
	}
	if column <= oldPrefixLen {
		return newPrefixLen
	}
	return column + (newPrefixLen - oldPrefixLen)
}
