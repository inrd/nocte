package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	contentWidth, contentHeight := m.docContentSize()
	horizontal := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(contentHeight, lipgloss.Center, horizontal)

	if m.activeDialog != "" {
		return docStyle.Render(strings.TrimRight(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, m.dialogView()), "\n"))
	}

	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}

func (m Model) editorView() string {
	header := dialogTitleStyle.Render(m.editorName)
	pathLine := helpStyle.Render(m.editorPath)
	editorPaneWidth := m.editorPaneWidth()
	editorContent := m.editor.View()
	if m.previewVisible() {
		editorContent = lipgloss.JoinVertical(lipgloss.Left, helpStyle.Render("Editor"), editorContent)
	}
	editorBox := inputStyle.Copy().Width(editorPaneWidth).Render(editorContent)
	if m.previewVisible() {
		previewPaneWidth := m.previewPaneWidth()
		previewContent := lipgloss.JoinVertical(lipgloss.Left, helpStyle.Render("Preview (read-only)"), m.previewContent())
		previewBox := inputStyle.Copy().Width(previewPaneWidth).Render(previewContent)
		editorBox = lipgloss.JoinHorizontal(lipgloss.Top, editorBox, strings.Repeat(" ", editorPaneGap), previewBox)
	}

	help := lipgloss.JoinHorizontal(
		lipgloss.Left,
		keyHintStyle.Render("Esc"),
		helpStyle.Render(" save & close  "),
		keyHintStyle.Render("Ctrl+A"),
		helpStyle.Render(" discard  "),
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

		contentWidth, contentHeight := m.docContentSize()
		return docStyle.Render(strings.TrimRight(lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, m.dialogView()), "\n"))
	}

	if m.width == 0 || m.height == 0 {
		return docStyle.Render(content)
	}

	contentWidth, contentHeight := m.docContentSize()
	horizontal := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(contentHeight, lipgloss.Center, horizontal)
	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}

func (m Model) editorStatusLine() string {
	sizeInfo := metaStyle.Render(editorSizeStatus(m.editor.Value()))
	taskInfo := renderEditorTaskProgress(m.editor.Value())
	saveInfo := renderEditorSaveStatus(m.editorDirty, m.editorLastSave)
	metaParts := []string{sizeInfo, saveInfo}
	if taskInfo != "" {
		metaParts = append(metaParts, taskInfo)
	}
	metaInfo := lipgloss.JoinHorizontal(lipgloss.Left, joinWithHelpSeparators(metaParts...)...)
	if m.status == "" {
		return metaInfo
	}

	var statusText string
	if m.isError {
		statusText = errorStyle.Render(m.status)
	} else {
		statusText = successStyle.Render(m.status)
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, statusText, helpStyle.Render(" | "), metaInfo)
}

func renderEditorSaveStatus(dirty bool, lastSave time.Time) string {
	if dirty {
		return mutedMetaStyle.Render("Unsaved changes")
	}
	if lastSave.IsZero() {
		return mutedMetaStyle.Render("Not yet autosaved")
	}

	return mutedMetaStyle.Render(fmt.Sprintf("Last save %s", lastSave.Format("15:04:05")))
}

func (m Model) editorWarningLine() string {
	warning := largeNoteWarning(int64(len([]byte(m.editor.Value()))))
	if warning == "" {
		return ""
	}

	return errorStyle.Render(warning)
}

func renderEditorTaskProgress(content string) string {
	completed, total := countTaskProgress(content)
	if total == 0 {
		return ""
	}

	percent := completed * 100 / total
	style := editorTaskProgressStyle(percent)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		style.Render(fmt.Sprintf("%d%%", percent)),
		helpStyle.Render(" "),
		renderEditorTaskProgressBar(percent),
	)
}

func renderEditorTaskProgressBar(percent int) string {
	const width = 8

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

func editorTaskProgressStyle(percent int) lipgloss.Style {
	switch {
	case percent >= 100:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	case percent > 75:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	case percent >= 50:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
}

func countTaskProgress(content string) (completed int, total int) {
	for _, line := range strings.Split(content, "\n") {
		if !isTaskListLine(line) {
			continue
		}

		total++
		_, marker, _ := parseTaskListLine(line)
		if isCheckedTaskMarker(marker) {
			completed++
		}
	}

	return completed, total
}

func joinWithHelpSeparators(parts ...string) []string {
	filtered := make([]string, 0, len(parts)*2)
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(filtered) > 0 {
			filtered = append(filtered, helpStyle.Render(" | "))
		}
		filtered = append(filtered, part)
	}
	return filtered
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
		lines = append(lines, metaStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(matches))))
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
	style := m.searchPaletteStyle()
	if len(m.noteMatches) == 0 {
		return style.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			helpStyle.Render("No matching notes"),
			helpStyle.Render("Press Enter to create a new note"),
		))
	}

	start, end := m.noteVisibleRange()
	lines := make([]string, 0, end-start+1)
	contentWidth := style.GetWidth() - style.GetHorizontalFrameSize()
	for i := start; i < end; i++ {
		note := m.noteMatches[i]
		line := m.notePaletteRow(note, contentWidth)
		if i == m.noteIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	if len(m.noteMatches) > 1 {
		lines = append(lines, metaStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.noteMatches))))
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
		rows = append(rows, metaStyle.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(m.searchMatches))))
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
		availableWidth, _ := m.docContentSize()
		width = min(maxSearchPaletteWidth, max(1, min(width, availableWidth)))
	}

	height := m.launcherPaletteContentBudget()
	return commandPaletteStyle.Copy().Width(width).Height(height)
}

func (m Model) docContentSize() (int, int) {
	width := max(1, m.width-docStyle.GetHorizontalFrameSize())
	height := max(1, m.height-docStyle.GetVerticalFrameSize())
	return width, height
}

func (m Model) commandVisibleRange(total int) (int, int) {
	if total == 0 {
		return 0, 0
	}

	visible := m.launcherCommandListVisibleCount()
	if visible <= 0 || total <= visible {
		return 0, total
	}

	start := clampInt(m.commandOffset, 0, total-visible)
	return start, min(total, start+visible)
}

func (m Model) noteVisibleRange() (int, int) {
	return m.noteVisibleRangeFrom(m.noteOffset)
}

func (m Model) noteVisibleRangeFrom(start int) (int, int) {
	if len(m.noteMatches) == 0 {
		return 0, 0
	}

	start = clampInt(start, 0, len(m.noteMatches)-1)
	budget := m.launcherPaletteContentBudget()
	if len(m.noteMatches) > 1 {
		budget--
	}
	if budget < 1 {
		budget = 1
	}

	used := 0
	end := start
	for end < len(m.noteMatches) {
		rowHeight := noteMatchHeight(m.noteMatches[end])
		if used > 0 && used+rowHeight > budget {
			break
		}
		if used == 0 && rowHeight > budget {
			end++
			break
		}
		used += rowHeight
		end++
	}

	if end == start {
		end = min(len(m.noteMatches), start+1)
	}

	return start, end
}

func (m Model) launcherCommandListVisibleCount() int {
	budget := m.launcherPaletteContentBudget()
	if budget <= 1 {
		return 1
	}

	return budget - 1
}

func (m Model) notePaletteRow(note noteMatch, contentWidth int) string {
	title := searchNameStyle.Render(truncateText(note.name, contentWidth))
	snippetWidth := max(16, contentWidth)
	lines := []string{title}
	for _, previewLine := range note.preview {
		lines = append(lines, searchSnippetStyle.Render(truncateText(previewLine, snippetWidth)))
	}

	return strings.Join(lines, "\n")
}

func noteMatchHeight(note noteMatch) int {
	return 1 + len(note.preview)
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
		return truncateText(todoMatchHeader(match), width)
	}

	return matchHeader(match, width)
}

func todoMatchHeader(match searchMatch) string {
	if match.taskTotal <= 0 {
		return match.name
	}

	percent := match.taskDone * 100 / match.taskTotal
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		match.name,
		helpStyle.Render(" "),
		editorTaskProgressStyle(percent).Render(fmt.Sprintf("%d%%", percent)),
	)
}

func matchHeader(match searchMatch, width int) string {
	return truncateText(match.name+":"+strconv.Itoa(match.lineNumber), width)
}
