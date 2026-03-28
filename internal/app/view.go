package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.isEditing() {
		return m.editorView()
	}

	inputBox := inputStyle.Render(m.input.View())
	help := helpStyle.Render("Type a note name and press Enter. Type / to search note contents. Type :help for commands. Use :quit or Esc to quit.")

	status := ""
	switch {
	case m.status == "":
	case m.isError:
		status = errorStyle.Render(m.status)
	default:
		status = successStyle.Render(m.status)
	}

	parts := []string{inputBox}
	parts = append(parts, m.launcherPaletteView())
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

	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyHintStyle.Render("Esc"),
		helpStyle.Render(" save & close  "),
		keyHintStyle.Render("Ctrl+C"),
		helpStyle.Render(" save & quit  "),
		keyHintStyle.Render("Ctrl+H"),
		helpStyle.Render(" help"),
	)
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

	if m.activeDialog != "" {
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
	style := m.commandPaletteStyle()
	matches := m.filteredCommands()
	if len(matches) == 0 {
		return style.Render(errorStyle.Render("No matching commands"))
	}

	start, end := m.commandVisibleRange(len(matches))
	lines := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		command := matches[i]
		line := commandPaletteLine(command)
		if i == m.commandIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	if len(matches) > 1 {
		lines = append(lines, helpStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(matches))))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) launcherPaletteView() string {
	switch {
	case m.shouldShowSearchPalette():
		return m.searchPaletteView()
	case m.isCommandMode():
		return m.commandPaletteView()
	case m.shouldShowNotePalette():
		return m.notePaletteView()
	default:
		return m.commandPaletteStyle().Render(helpStyle.Render("Start typing to search notes, create one, or run a command"))
	}
}

func commandPaletteLine(command command) string {
	return command.name + strings.Repeat(" ", max(1, 8-len(command.name)+1)) + command.description
}

func (m Model) notePaletteView() string {
	style := m.commandPaletteStyle()
	if len(m.noteMatches) == 0 {
		return style.Render(helpStyle.Render("No matching notes"))
	}

	start, end := m.noteVisibleRange()
	lines := make([]string, 0, end-start+1)
	for i := start; i < end; i++ {
		note := m.noteMatches[i]
		line := note.name
		if i == m.noteIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	if len(m.noteMatches) > 1 {
		lines = append(lines, helpStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.noteMatches))))
	}

	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) searchPaletteView() string {
	style := m.searchPaletteStyle()
	query := m.searchQuery()
	if !m.isTodoMode() && query == "" {
		return style.Render(helpStyle.Render("Type after / to search inside your notes"))
	}
	if len(m.searchMatches) == 0 {
		if m.isTodoMode() {
			return style.Render(helpStyle.Render("No Markdown tasks found"))
		}
		return style.Render(helpStyle.Render("No matching note content"))
	}

	start, end := m.searchVisibleRange()
	rows := make([]string, 0, end-start+1)
	contentWidth := style.GetWidth() - style.GetHorizontalFrameSize()
	for i := start; i < end; i++ {
		match := m.searchMatches[i]
		row := m.searchPaletteRow(match, contentWidth)
		if i == m.searchIndex {
			rows = append(rows, commandSelectedStyle.Render(row))
			continue
		}
		rows = append(rows, row)
	}
	if len(m.searchMatches) > 1 {
		rows = append(rows, helpStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.searchMatches))))
	}

	return style.Render(strings.Join(rows, "\n"))
}

func (m Model) commandPaletteStyle() lipgloss.Style {
	height := m.launcherPaletteContentBudget()
	return commandPaletteStyle.Copy().Height(height)
}

func (m Model) searchPaletteStyle() lipgloss.Style {
	width := defaultSearchPaletteWidth
	if m.width > 0 {
		width = min(maxSearchPaletteWidth, max(56, m.width-8))
	}

	height := m.launcherPaletteContentBudget()
	return commandPaletteStyle.Copy().Width(width).Height(height)
}

func (m Model) launcherListVisibleCount() int {
	budget := m.launcherPaletteContentBudget()
	if budget <= 1 {
		return 1
	}

	return budget - 1
}

func (m Model) commandVisibleRange(total int) (int, int) {
	if total == 0 {
		return 0, 0
	}

	visible := m.launcherListVisibleCount()
	if visible <= 0 || total <= visible {
		return 0, total
	}

	start := clampInt(m.commandOffset, 0, total-visible)
	return start, min(total, start+visible)
}

func (m Model) noteVisibleRange() (int, int) {
	if len(m.noteMatches) == 0 {
		return 0, 0
	}

	visible := m.launcherListVisibleCount()
	if visible <= 0 || len(m.noteMatches) <= visible {
		return 0, len(m.noteMatches)
	}

	start := clampInt(m.noteOffset, 0, len(m.noteMatches)-visible)
	return start, min(len(m.noteMatches), start+visible)
}

func (m Model) searchPaletteRow(match searchMatch, contentWidth int) string {
	title := searchNameStyle.Render(m.searchMatchHeader(match, contentWidth))
	snippetWidth := max(16, contentWidth)
	lines := []string{title}
	for _, snippetLine := range match.snippetLines {
		lines = append(lines, searchSnippetStyle.Render(truncateText(snippetLine, snippetWidth)))
	}

	return strings.Join(lines, "\n")
}

func (m Model) searchMatchHeader(match searchMatch, width int) string {
	if m.isTodoMode() {
		return truncateText(match.name, width)
	}

	return matchHeader(match, width)
}

func matchHeader(match searchMatch, width int) string {
	return truncateText(match.name+":"+strconv.Itoa(match.lineNumber), width)
}
