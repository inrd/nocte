package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) dialogView() string {
	switch m.activeDialog {
	case "help":
		return helpDialog()
	case "info":
		return infoDialog(m.version, m.configPath, m.config.NotesPath)
	case "list":
		return m.listDialog()
	case "save-error":
		return m.saveErrorDialog()
	default:
		return ""
	}
}

func helpDialog() string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		dialogTitleStyle.Render("Commands"),
		"",
		":help  Show this dialog",
		":files Open the notes folder",
		":info  Show app and path info",
		":list  List existing notes",
		":quit  Exit the app",
		"",
		helpStyle.Render("Press Esc or Enter to close."),
	)

	return dialogStyle.Render(body)
}

func infoDialog(version string, configPath string, notesPath string) string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		dialogTitleStyle.Render("Info"),
		"",
		fmt.Sprintf("Version: %s", version),
		fmt.Sprintf("Config: %s", configPath),
		fmt.Sprintf("Notes: %s", notesPath),
		"",
		helpStyle.Render("Press Esc or Enter to close."),
	)

	return dialogStyle.Render(body)
}

func (m Model) listDialog() string {
	dialogStyle := m.listDialogStyle()
	lines := []string{
		dialogTitleStyle.Render("Notes"),
		"",
	}

	if len(m.dialogNotes) == 0 {
		lines = append(lines, helpStyle.Render("No notes yet."))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Press Esc or Enter to close."))
		return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	start, end := m.dialogRange()
	if start > 0 {
		lines = append(lines, helpStyle.Render("..."))
	}

	for i := start; i < end; i++ {
		note := m.dialogNotes[i]
		line := m.listDialogLine(note)
		if i == m.dialogIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	if end < len(m.dialogNotes) {
		lines = append(lines, helpStyle.Render("..."))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Use Up and Down to choose. Press Enter to open or Esc to close."))
	return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) saveErrorDialog() string {
	lines := []string{
		dialogTitleStyle.Render("Save Failed"),
		"",
		errorStyle.Render(m.status),
		"",
		"Press Enter to discard your unsaved changes.",
		"Press Esc to return to the editor.",
	}

	if m.editorAction == "quit" {
		lines[4] = "Press Enter to discard your unsaved changes and quit."
	}

	return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) dialogRange() (int, int) {
	visible := m.dialogVisibleCount()
	if visible <= 0 || len(m.dialogNotes) <= visible {
		return 0, len(m.dialogNotes)
	}

	start := max(0, m.dialogOffset)
	end := min(len(m.dialogNotes), start+visible)
	return start, end
}

func (m Model) dialogVisibleCount() int {
	if m.height <= 0 {
		return 10
	}

	return max(3, m.height-10)
}

func (m Model) listDialogStyle() lipgloss.Style {
	width := defaultListDialogWidth
	if m.width > 0 {
		width = min(maxListDialogWidth, max(defaultDialogWidth, m.width-8))
	}

	return dialogStyle.Copy().Width(width)
}

func (m Model) listDialogLine(note noteMatch) string {
	dialogStyle := m.listDialogStyle()
	contentWidth := dialogStyle.GetWidth() - dialogStyle.GetHorizontalFrameSize()
	metaWidth := listMetaWidth
	updatedWidth := listUpdatedAtWidth
	nameWidth := max(12, contentWidth-metaWidth-updatedWidth-(listColumnGap*2))

	name := listNameStyle.Copy().
		Width(nameWidth).
		Render(truncateText(note.name, nameWidth))
	updated := listUpdatedStyle.Copy().
		Width(updatedWidth).
		Render(successStyle.Render(formatNoteUpdatedAt(note.modTime)))
	meta := lipgloss.NewStyle().
		Width(metaWidth).
		Align(lipgloss.Right).
		Render(metaStyle.Render(noteMeta(note)))

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		name,
		strings.Repeat(" ", listColumnGap),
		updated,
		strings.Repeat(" ", listColumnGap),
		meta,
	)
}
