package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) isSearchMode() bool {
	return strings.HasPrefix(strings.TrimSpace(m.input.Value()), "/")
}

func (m Model) isTodoMode() bool {
	return m.todoMode && strings.TrimSpace(m.input.Value()) == ":todo"
}

func (m Model) searchQuery() string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(m.input.Value()), "/"))
}

func (m Model) shouldShowSearchPalette() bool {
	return m.isSearchMode() || m.isTodoMode()
}

func (m Model) hasSearchPalette() bool {
	if m.isTodoMode() {
		return len(m.searchMatches) > 0
	}

	return m.searchQuery() != "" && len(m.searchMatches) > 0
}

func (m *Model) moveSearchSelection(delta int) {
	if len(m.searchMatches) == 0 {
		m.searchIndex = -1
		m.searchOffset = 0
		return
	}

	if m.searchIndex == -1 {
		if delta > 0 {
			m.searchIndex = 0
			return
		}

		m.searchIndex = len(m.searchMatches) - 1
		return
	}

	m.searchIndex = (m.searchIndex + delta + len(m.searchMatches)) % len(m.searchMatches)
	m.syncSearchOffset()
}

func (m Model) findSearchMatches(query string) []searchMatch {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	entries, err := os.ReadDir(m.config.NotesPath)
	if err != nil {
		return nil
	}

	queryLower := strings.ToLower(query)
	matches := make([]searchMatch, 0, min(searchResultLimit, len(entries)))
	for _, entry := range entries {
		if len(matches) >= searchResultLimit || entry.IsDir() {
			if len(matches) >= searchResultLimit {
				break
			}
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(m.config.NotesPath, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for lineIndex, line := range lines {
			matchColumns := allMatchColumns(line, queryLower)
			for _, column := range matchColumns {
				matches = append(matches, searchMatch{
					name:         name,
					path:         path,
					lineNumber:   lineIndex + 1,
					column:       column,
					snippetLines: snippetForMatch(lines, lineIndex),
				})
				if len(matches) >= searchResultLimit {
					break
				}
			}
			if len(matches) >= searchResultLimit {
				break
			}
		}
	}

	sort.Slice(matches, func(i int, j int) bool {
		if matches[i].name != matches[j].name {
			return matches[i].name < matches[j].name
		}
		if matches[i].lineNumber != matches[j].lineNumber {
			return matches[i].lineNumber < matches[j].lineNumber
		}
		return matches[i].column < matches[j].column
	})

	return matches
}

func (m Model) findTodoMatches() []searchMatch {
	entries, err := os.ReadDir(m.config.NotesPath)
	if err != nil {
		return nil
	}

	matches := make([]searchMatch, 0, min(searchResultLimit, len(entries)))
	for _, entry := range entries {
		if len(matches) >= searchResultLimit || entry.IsDir() {
			if len(matches) >= searchResultLimit {
				break
			}
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}

		path := filepath.Join(m.config.NotesPath, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for lineIndex, line := range lines {
			if !isOpenTaskListLine(line) {
				continue
			}

			matches = append(matches, searchMatch{
				name:         name,
				path:         path,
				lineNumber:   lineIndex + 1,
				column:       0,
				snippetLines: []string{strings.TrimSpace(line)},
			})
			if len(matches) >= searchResultLimit {
				break
			}
		}

		if len(matches) >= searchResultLimit {
			break
		}
	}

	sort.Slice(matches, func(i int, j int) bool {
		if matches[i].name != matches[j].name {
			return matches[i].name < matches[j].name
		}
		return matches[i].lineNumber < matches[j].lineNumber
	})

	return matches
}

func allMatchColumns(line string, queryLower string) []int {
	if queryLower == "" {
		return nil
	}

	lineLower := strings.ToLower(line)
	columns := make([]int, 0, 1)
	offset := 0
	for offset <= len(lineLower) {
		index := strings.Index(lineLower[offset:], queryLower)
		if index < 0 {
			break
		}

		byteIndex := offset + index
		columns = append(columns, utf8.RuneCountInString(line[:byteIndex]))
		offset = byteIndex + len(queryLower)
	}

	return columns
}

func snippetForMatch(lines []string, lineIndex int) []string {
	if len(lines) == 0 || lineIndex < 0 || lineIndex >= len(lines) {
		return nil
	}

	start := max(0, lineIndex-searchSnippetContextLines)
	end := min(len(lines), lineIndex+searchSnippetContextLines+1)
	snippet := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		snippet = append(snippet, strings.TrimSpace(lines[i]))
	}

	return snippet
}

func (m Model) handleSearchResult() (tea.Model, tea.Cmd) {
	if m.searchIndex >= 0 && m.searchIndex < len(m.searchMatches) {
		if err := m.openSearchMatch(m.searchMatches[m.searchIndex]); err != nil {
			m.status = err.Error()
			m.isError = true
		}
		return m, nil
	}

	if m.isTodoMode() {
		if len(m.searchMatches) == 0 {
			m.status = "No Markdown tasks found"
			m.isError = true
			return m, nil
		}

		m.status = "Select a task with Up or Down"
		m.isError = false
		return m, nil
	}

	if len(m.searchMatches) == 0 {
		m.status = fmt.Sprintf("No note content matches %q", m.searchQuery())
		m.isError = true
		return m, nil
	}

	m.status = "Select a search result with Up or Down"
	m.isError = false
	return m, nil
}

func (m *Model) syncSearchOffset() {
	if len(m.searchMatches) == 0 {
		m.searchOffset = 0
		return
	}

	if m.searchIndex < 0 {
		m.searchOffset = min(m.searchOffset, len(m.searchMatches)-1)
		if m.searchOffset < 0 {
			m.searchOffset = 0
		}
		return
	}

	m.searchOffset = clampInt(m.searchOffset, 0, len(m.searchMatches)-1)
	for m.searchIndex < m.searchOffset {
		m.searchOffset--
	}
	for {
		start, end := m.searchVisibleRangeFrom(m.searchOffset)
		if m.searchIndex >= start && m.searchIndex < end {
			return
		}
		m.searchOffset++
		if m.searchOffset >= len(m.searchMatches) {
			m.searchOffset = len(m.searchMatches) - 1
			return
		}
	}
}

func (m Model) searchVisibleRange() (int, int) {
	return m.searchVisibleRangeFrom(m.searchOffset)
}

func (m Model) searchVisibleRangeFrom(start int) (int, int) {
	if len(m.searchMatches) == 0 {
		return 0, 0
	}

	start = clampInt(start, 0, len(m.searchMatches)-1)
	budget := m.launcherPaletteContentBudget()
	if len(m.searchMatches) > 1 {
		budget--
	}
	if budget < 1 {
		budget = 1
	}

	used := 0
	end := start
	for end < len(m.searchMatches) {
		rowHeight := searchMatchHeight(m.searchMatches[end])
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
		end = min(len(m.searchMatches), start+1)
	}

	return start, end
}

func (m Model) launcherPaletteContentBudget() int {
	if m.height <= 0 {
		return 12
	}

	return max(4, (m.height/2)-4)
}

func searchMatchHeight(match searchMatch) int {
	return 1 + len(match.snippetLines)
}

func clampInt(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
