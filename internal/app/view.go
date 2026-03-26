package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.isEditing() {
		return m.editorView()
	}

	inputBox := inputStyle.Render(m.input.View())
	help := helpStyle.Render("Type a note name and press Enter. Type :help for commands. Use :quit or Esc to quit.")

	status := ""
	switch {
	case m.status == "":
	case m.isError:
		status = errorStyle.Render(m.status)
	default:
		status = successStyle.Render(m.status)
	}

	parts := []string{inputBox}
	if m.isCommandMode() {
		parts = append(parts, m.commandPaletteView())
	} else if m.shouldShowNotePalette() {
		parts = append(parts, m.notePaletteView())
	}
	parts = append(parts, help, status)

	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

	if m.width == 0 || m.height == 0 {
		if m.activeDialog != "" {
			return docStyle.Render(lipgloss.JoinVertical(lipgloss.Center, content, m.dialogView()))
		}

		return docStyle.Render(content)
	}

	horizontal := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(m.height, lipgloss.Center, horizontal)

	if m.activeDialog != "" {
		return docStyle.Render(strings.TrimRight(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.dialogView()), "\n"))
	}

	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}

func (m Model) editorView() string {
	header := dialogTitleStyle.Render(m.editorName)
	pathLine := helpStyle.Render(m.editorPath)
	editorContent := m.editor.View()
	if m.previewVisible() {
		editorContent = lipgloss.JoinVertical(lipgloss.Left, helpStyle.Render("Editor"), editorContent)
	}
	editorBox := inputStyle.Render(editorContent)
	if m.previewVisible() {
		previewContent := lipgloss.JoinVertical(lipgloss.Left, helpStyle.Render("Preview (read-only)"), m.previewContent())
		previewBox := inputStyle.Render(previewContent)
		editorBox = lipgloss.JoinHorizontal(lipgloss.Top, editorBox, strings.Repeat(" ", editorPaneGap), previewBox)
	}

	helpText := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyHintStyle.Render("Ctrl+P"),
		helpStyle.Render(" preview  "),
		keyHintStyle.Render("Esc"),
		helpStyle.Render(" save & close  "),
		keyHintStyle.Render("Ctrl+C"),
		helpStyle.Render(" save & quit"),
	)
	if !m.previewVisible() && m.previewEnabled {
		helpText = lipgloss.JoinHorizontal(
			lipgloss.Left,
			keyHintStyle.Render("Ctrl+P"),
			helpStyle.Render(" preview  "),
			keyHintStyle.Render("Esc"),
			helpStyle.Render(" save & close  "),
			keyHintStyle.Render("Ctrl+C"),
			helpStyle.Render(" save & quit  "),
			helpStyle.Render("(preview hidden on narrow terminals)"),
		)
	}
	help := helpText
	statusLine := m.editorStatusLine()
	warningLine := m.editorWarningLine()

	lines := []string{header, pathLine, "", editorBox, help}
	if statusLine != "" {
		lines = append(lines, statusLine)
	}
	if warningLine != "" {
		lines = append(lines, warningLine)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	if m.activeDialog == "save-error" {
		if m.width == 0 || m.height == 0 {
			return docStyle.Render(lipgloss.JoinVertical(lipgloss.Center, content, m.dialogView()))
		}

		return docStyle.Render(strings.TrimRight(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.dialogView()), "\n"))
	}

	if m.width == 0 || m.height == 0 {
		return docStyle.Render(content)
	}

	horizontal := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(m.height, lipgloss.Center, horizontal)
	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}

func (m Model) editorStatusLine() string {
	sizeInfo := metaStyle.Render(editorSizeStatus(m.editor.Value()))
	if m.status == "" {
		return sizeInfo
	}

	var statusText string
	if m.isError {
		statusText = errorStyle.Render(m.status)
	} else {
		statusText = successStyle.Render(m.status)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, statusText, helpStyle.Render(" | "), sizeInfo)
}

func (m Model) editorWarningLine() string {
	warning := largeNoteWarning(int64(len([]byte(m.editor.Value()))))
	if warning == "" {
		return ""
	}

	return errorStyle.Render(warning)
}

func (m Model) commandPaletteView() string {
	matches := m.filteredCommands()
	if len(matches) == 0 {
		return commandPaletteStyle.Render(errorStyle.Render("No matching commands"))
	}

	lines := make([]string, 0, len(matches))
	for i, command := range matches {
		line := commandPaletteLine(command)
		if i == m.commandIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	return commandPaletteStyle.Render(strings.Join(lines, "\n"))
}

func commandPaletteLine(command command) string {
	return command.name + strings.Repeat(" ", max(1, 8-len(command.name)+1)) + command.description
}

func (m Model) notePaletteView() string {
	if len(m.noteMatches) == 0 {
		return commandPaletteStyle.Render(helpStyle.Render("No matching notes"))
	}

	lines := make([]string, 0, len(m.noteMatches))
	for i, note := range m.noteMatches {
		line := note.name
		if i == m.noteIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	return commandPaletteStyle.Render(strings.Join(lines, "\n"))
}
