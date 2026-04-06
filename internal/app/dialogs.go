package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) dialogView() string {
	switch m.activeDialog {
	case "help":
		return helpDialog()
	case "editor-help":
		return editorHelpDialog()
	case "info":
		return infoDialog(m.version, m.configPath, m.config.NotesPath, m.config.BackupPath)
	case "list":
		return m.listDialog()
	case "links":
		return m.linksDialog()
	case "delete-confirm":
		return m.deleteConfirmDialog()
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
		"/search Search inside note contents from the launcher",
		"",
		":export-all Render all notes to HTML",
		":help  Show this dialog",
		":files Open the notes folder",
		":info  Show app and path info",
		":list  List existing notes",
		":todo  List Markdown tasks",
		":quit  Exit the app",
		"",
		helpStyle.Render("Press Esc or Enter to close."),
	)

	return dialogStyle.Render(body)
}

func editorHelpDialog() string {
	rows := []string{
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Esc"), helpStyle.Render("     Save and close the editor")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+A"), helpStyle.Render("  Discard changes and return to the launcher")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+H"), helpStyle.Render("  Show this help dialog")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+P"), helpStyle.Render("  Toggle the Markdown preview")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+T"), helpStyle.Render("  Toggle the current line as a task")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+K"), helpStyle.Render("  Copy the current inline or block code")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+E"), helpStyle.Render("  Export the current note to HTML")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+L"), helpStyle.Render("  List links in the current note")),
		lipgloss.JoinHorizontal(lipgloss.Left, keyHintStyle.Render("Ctrl+D"), helpStyle.Render("  Delete the current note")),
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		dialogTitleStyle.Render("Editor Shortcuts"),
		"",
		lipgloss.JoinVertical(lipgloss.Left, rows...),
		"",
		helpStyle.Render("Press Esc, Enter, or Ctrl+H to close."),
	)

	return dialogStyle.Render(body)
}

func infoDialog(version string, configPath string, notesPath string, backupPath string) string {
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		dialogTitleStyle.Render("Info"),
		"",
		fmt.Sprintf("Version: %s", version),
		fmt.Sprintf("Config:  %s", configPath),
		fmt.Sprintf("Notes:   %s", notesPath),
		fmt.Sprintf("Backups: %s", backupPath),
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
		lines = append(lines, m.listDialogLine(note, i == m.dialogIndex))
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

	return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) deleteConfirmDialog() string {
	lines := []string{
		dialogTitleStyle.Render("Delete Note"),
		"",
		errorStyle.Render(fmt.Sprintf("Delete %s?", m.editorName)),
		"",
		"Press Enter to delete this note.",
		"Press Esc to keep editing.",
	}

	return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) linksDialog() string {
	dialogStyle := m.listDialogStyle()
	lines := []string{
		dialogTitleStyle.Render("Links"),
		"",
	}

	if len(m.dialogLinks) == 0 {
		lines = append(lines, helpStyle.Render("No links found in this note."))
		lines = append(lines, "")
		lines = append(lines, helpStyle.Render("Press Esc or Enter to close."))
		return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
	}

	start, end := m.dialogRange()
	if start > 0 {
		lines = append(lines, helpStyle.Render("..."))
	}

	for i := start; i < end; i++ {
		line := m.linkDialogLine(m.dialogLinks[i], i == m.dialogIndex)
		if i == m.dialogIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}
		lines = append(lines, line)
	}

	if end < len(m.dialogLinks) {
		lines = append(lines, helpStyle.Render("..."))
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Use Up and Down to choose. Press Enter to open or Esc to close."))
	return dialogStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) dialogRange() (int, int) {
	visible := m.dialogVisibleCount()
	total := m.dialogItems()
	if visible <= 0 || total <= visible {
		return 0, total
	}

	start := max(0, m.dialogOffset)
	end := min(total, start+visible)
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

func (m Model) listDialogLine(note noteMatch, selected bool) string {
	dialogStyle := m.listDialogStyle()
	contentWidth := dialogStyle.GetWidth() - dialogStyle.GetHorizontalFrameSize()
	metaWidth := listMetaWidth
	updatedWidth := listUpdatedAtWidth
	progressWidth := listTaskProgressWidth
	nameWidth := max(12, contentWidth-metaWidth-updatedWidth-progressWidth-(listColumnGap*2))
	nameWidth = max(12, nameWidth-listColumnGap)

	nameStyle := listNameStyle.Copy().Width(nameWidth)
	progress := strings.Repeat(" ", progressWidth)

	if selected {
		nameStyle = commandSelectedStyle.Copy().Width(nameWidth)
	}
	if note.taskTotal > 0 {
		progress = renderListTaskProgress(note.taskDone, note.taskTotal)
	}

	name := nameStyle.Render(truncateText(note.name, nameWidth))
	updated := listUpdatedStyle.Copy().
		Width(updatedWidth).
		Render(renderListUpdatedAt(note.modTime))
	meta := lipgloss.NewStyle().
		Width(metaWidth).
		Align(lipgloss.Right).
		Render(renderListMeta(note))
	progress = lipgloss.NewStyle().
		Width(progressWidth).
		Align(lipgloss.Left).
		Render(progress)

	parts := []string{
		name,
		strings.Repeat(" ", listColumnGap),
		updated,
		strings.Repeat(" ", listColumnGap),
		progress,
		strings.Repeat(" ", listColumnGap),
		meta,
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func renderListTaskProgress(done int, total int) string {
	if total <= 0 {
		return ""
	}

	percent := done * 100 / total
	style := editorTaskProgressStyle(percent)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.Render(fmt.Sprintf("%d%%", percent)),
		helpStyle.Render(" "),
		renderListTaskProgressBar(percent),
	)
}

func renderListTaskProgressBar(percent int) string {
	const width = 4

	filled := percent * width / 100
	if percent > 0 && filled == 0 {
		filled = 1
	}
	if percent >= 100 {
		filled = width
	}

	style := editorTaskProgressStyle(percent)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.Render(strings.Repeat("█", filled)),
		metaStyle.Render(strings.Repeat("░", max(0, width-filled))),
	)
}

func renderListUpdatedAt(modTime time.Time) string {
	formatted := formatNoteUpdatedAt(modTime)
	return successStyle.Render(formatted)
}

func renderListMeta(note noteMatch) string {
	return metaStyle.Render(noteMeta(note))
}

func (m Model) linkDialogLine(link noteLink, selected bool) string {
	dialogStyle := m.listDialogStyle()
	contentWidth := dialogStyle.GetWidth() - dialogStyle.GetHorizontalFrameSize()
	labelWidth := max(16, (contentWidth-listColumnGap)/2)
	urlWidth := max(16, contentWidth-labelWidth-listColumnGap)

	label := link.label
	if label == "" {
		label = "(raw URL)"
	}

	labelStyle := linkLabelStyle.Copy()
	urlStyle := linkURLStyle.Copy()
	if selected {
		labelStyle = labelStyle.Foreground(lipgloss.Color("0"))
		urlStyle = urlStyle.Foreground(lipgloss.Color("0"))
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Width(labelWidth).Render(truncateText(label, labelWidth)),
		strings.Repeat(" ", listColumnGap),
		urlStyle.Width(urlWidth).Render(truncateText(link.url, urlWidth)),
	)
}
