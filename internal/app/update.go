package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeEditor()
		if m.activeDialog == "list" || m.activeDialog == "links" {
			m.syncDialogOffset()
		}
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(msg)
	}

	var cmd tea.Cmd
	previousValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != previousValue {
		m.syncLauncherState()
	}
	return m, cmd
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isEditing() {
		return m.updateEditorKey(msg)
	}

	if m.activeDialog != "" {
		return m.updateDialogKey(msg)
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up":
		if m.isCommandMode() {
			m.moveCommandSelection(-1)
			return m, nil
		}
		if m.hasNotePalette() {
			m.moveNoteSelection(-1)
			return m, nil
		}
	case "down":
		if m.isCommandMode() {
			m.moveCommandSelection(1)
			return m, nil
		}
		if m.hasNotePalette() {
			m.moveNoteSelection(1)
			return m, nil
		}
	case "enter":
		if m.isCommandMode() {
			return m.handleCommand()
		}
		if m.noteIndex >= 0 && m.noteIndex < len(m.noteMatches) {
			if err := m.openExistingNote(m.noteMatches[m.noteIndex]); err != nil {
				m.status = err.Error()
				m.isError = true
			}
			return m, nil
		}
		return m.createNoteFromInput()
	}

	var cmd tea.Cmd
	previousValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != previousValue {
		m.syncLauncherState()
	}
	return m, cmd
}

func (m Model) updateEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.activeDialog == "links" {
		switch msg.String() {
		case "ctrl+c":
			if shouldQuit := m.finishEditing("quit"); shouldQuit {
				return m, tea.Quit
			}
			return m, nil
		case "ctrl+l":
			m.closeDialog()
			return m, nil
		case "esc":
			m.closeDialog()
			return m, nil
		case "up":
			m.moveDialogSelection(-1)
			return m, nil
		case "down":
			m.moveDialogSelection(1)
			return m, nil
		case "enter":
			_ = m.openSelectedDialogLink()
			return m, nil
		}

		return m, nil
	}

	if m.activeDialog == "save-error" {
		switch msg.String() {
		case "ctrl+c", "enter":
			action := m.editorAction
			m.discardEditor()
			if action == "quit" {
				return m, tea.Quit
			}
			return m, nil
		case "esc":
			m.closeDialog()
			return m, nil
		}

		return m, nil
	}

	if m.activeDialog == "delete-confirm" {
		switch msg.String() {
		case "ctrl+c":
			if err := m.deleteEditorNote(); err != nil {
				m.activeDialog = ""
				m.status = err.Error()
				m.isError = true
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			if err := m.deleteEditorNote(); err != nil {
				m.activeDialog = ""
				m.status = err.Error()
				m.isError = true
				return m, nil
			}
			return m, nil
		case "esc", "ctrl+d":
			m.closeDialog()
			return m, nil
		}

		return m, nil
	}

	switch msg.String() {
	case "ctrl+l":
		m.openLinksDialog()
		return m, nil
	case "ctrl+d":
		m.activeDialog = "delete-confirm"
		m.status = ""
		m.isError = false
		return m, nil
	case "ctrl+p":
		m.togglePreview()
		return m, nil
	case "ctrl+c":
		if shouldQuit := m.finishEditing("quit"); shouldQuit {
			return m, tea.Quit
		}
		return m, nil
	case "esc":
		m.finishEditing("close")
		return m, nil
	}

	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	return m, cmd
}

func (m Model) updateDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.closeDialog()
		return m, nil
	case "up":
		if m.activeDialog == "list" || m.activeDialog == "links" {
			m.moveDialogSelection(-1)
		}
		return m, nil
	case "down":
		if m.activeDialog == "list" || m.activeDialog == "links" {
			m.moveDialogSelection(1)
		}
		return m, nil
	case "enter":
		if m.activeDialog == "list" {
			if err := m.openSelectedDialogNote(); err != nil {
				m.status = err.Error()
				m.isError = true
			}
			return m, nil
		}
		if m.activeDialog == "links" {
			_ = m.openSelectedDialogLink()
			return m, nil
		}
		m.closeDialog()
		return m, nil
	}

	return m, nil
}

func (m Model) createNoteFromInput() (tea.Model, tea.Cmd) {
	filename, err := sanitizeFilename(m.input.Value())
	if err != nil {
		m.status = err.Error()
		m.isError = true
		return m, nil
	}

	if err := os.MkdirAll(m.config.NotesPath, 0o755); err != nil {
		m.status = fmt.Sprintf("Could not prepare notes dir: %v", err)
		m.isError = true
		return m, nil
	}

	path := filepath.Join(m.config.NotesPath, filename+".md")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			m.status = fmt.Sprintf("%s already exists", filename+".md")
			m.isError = true
			return m, nil
		}

		m.status = fmt.Sprintf("Could not create note: %v", err)
		m.isError = true
		return m, nil
	}
	_ = file.Close()

	m.openEditor(path, filename+".md")
	m.editorCreated = true
	m.status = fmt.Sprintf("Created %s", filename+".md")
	m.isError = false
	return m, nil
}
