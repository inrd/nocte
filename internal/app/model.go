package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/thomas/not/internal/config"
)

var (
	docStyle = lipgloss.NewStyle().
			Padding(1, 2)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 2).
			Width(64)

	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true)

	commandPaletteStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1).
				Width(48)

	commandSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("12"))
)

type Model struct {
	input        textinput.Model
	editor       textarea.Model
	width        int
	height       int
	status       string
	isError      bool
	config       config.Config
	configPath   string
	version      string
	activeDialog string
	commandIndex int
	noteIndex    int
	noteMatches  []noteMatch
	editorPath   string
	editorName   string
	lastSaved    string
}

type command struct {
	name        string
	description string
}

type noteMatch struct {
	name  string
	path  string
	score int
}

func New(cfg config.Config, configPath string, version string) Model {
	input := textinput.New()
	input.Placeholder = "Search or create a note..."
	input.Prompt = ""
	input.Focus()
	input.Width = 48

	editor := textarea.New()
	editor.Placeholder = "Start writing..."
	editor.Prompt = ""
	editor.ShowLineNumbers = false
	editor.SetHeight(12)
	editor.SetWidth(64)

	return Model{
		input:      input,
		editor:     editor,
		noteIndex:  -1,
		config:     cfg,
		configPath: configPath,
		version:    version,
	}
}

var invalidFileChars = regexp.MustCompile(`[^a-z0-9._-]+`)
var commands = []command{
	{name: ":help", description: "Show available commands"},
	{name: ":info", description: "Show app and path info"},
	{name: ":quit", description: "Exit the app"},
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeEditor()
		return m, nil
	case tea.KeyMsg:
		if m.isEditing() {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.closeEditor()
				return m, nil
			}

			before := m.editor.Value()
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			if m.editor.Value() != before {
				m.saveEditor()
			}
			return m, cmd
		}

		if m.activeDialog != "" {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "enter":
				m.activeDialog = ""
				m.input.SetValue("")
				m.input.Focus()
				return m, nil
			}

			return m, nil
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
			m.status = fmt.Sprintf("Created %s", filename+".md")
			m.isError = false
			return m, nil
		}
	}

	var cmd tea.Cmd
	previousValue := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != previousValue {
		m.syncLauncherState()
	}
	return m, cmd
}

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
	editorBox := inputStyle.Render(m.editor.View())
	help := helpStyle.Render("Plain text editor. Autosaves as you type. Press Esc to return, Ctrl+C to quit.")

	status := ""
	switch {
	case m.status == "":
	case m.isError:
		status = errorStyle.Render(m.status)
	default:
		status = successStyle.Render(m.status)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, pathLine, "", editorBox, help, status)

	if m.width == 0 || m.height == 0 {
		return docStyle.Render(content)
	}

	horizontal := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(m.height, lipgloss.Center, horizontal)
	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}

func (m Model) handleCommand() (tea.Model, tea.Cmd) {
	matches := m.filteredCommands()
	if len(matches) == 0 {
		m.status = fmt.Sprintf("Unknown command: %s", strings.TrimSpace(m.input.Value()))
		m.isError = true
		return m, nil
	}

	command := matches[m.commandIndex].name

	switch command {
	case ":help":
		m.activeDialog = "help"
		m.input.SetValue(command)
		m.status = ""
		m.isError = false
		return m, nil
	case ":info":
		m.activeDialog = "info"
		m.input.SetValue(command)
		m.status = ""
		m.isError = false
		return m, nil
	case ":quit":
		return m, tea.Quit
	default:
		m.status = fmt.Sprintf("Unknown command: %s", command)
		m.isError = true
		return m, nil
	}
}

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

func (m Model) dialogView() string {
	switch m.activeDialog {
	case "help":
		return helpDialog()
	case "info":
		return infoDialog(m.version, m.configPath, m.config.NotesPath)
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
		":info  Show app and path info",
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

func (m Model) isCommandMode() bool {
	return strings.HasPrefix(strings.TrimSpace(m.input.Value()), ":")
}

func (m Model) filteredCommands() []command {
	query := strings.ToLower(strings.TrimSpace(m.input.Value()))
	if !strings.HasPrefix(query, ":") {
		return nil
	}

	if query == ":" {
		return commands
	}

	filtered := make([]command, 0, len(commands))
	for _, command := range commands {
		name := strings.ToLower(command.name)
		description := strings.ToLower(command.description)
		if strings.Contains(name, query) || strings.Contains(description, strings.TrimPrefix(query, ":")) {
			filtered = append(filtered, command)
		}
	}

	return filtered
}

func (m *Model) syncCommandSelection() {
	matches := m.filteredCommands()
	if len(matches) == 0 {
		m.commandIndex = 0
		return
	}

	typed := strings.TrimSpace(m.input.Value())
	for i, command := range matches {
		if command.name == typed {
			m.commandIndex = i
			return
		}
	}

	if m.commandIndex >= len(matches) {
		m.commandIndex = len(matches) - 1
	}
}

func (m *Model) moveCommandSelection(delta int) {
	matches := m.filteredCommands()
	if len(matches) == 0 {
		m.commandIndex = 0
		return
	}

	m.commandIndex = (m.commandIndex + delta + len(matches)) % len(matches)
}

func (m Model) commandPaletteView() string {
	matches := m.filteredCommands()
	if len(matches) == 0 {
		return commandPaletteStyle.Render(errorStyle.Render("No matching commands"))
	}

	lines := make([]string, 0, len(matches))
	for i, command := range matches {
		line := fmt.Sprintf("%-8s %s", command.name, command.description)
		if i == m.commandIndex {
			lines = append(lines, commandSelectedStyle.Render(line))
			continue
		}

		lines = append(lines, line)
	}

	return commandPaletteStyle.Render(strings.Join(lines, "\n"))
}

func (m Model) shouldShowNotePalette() bool {
	return !m.isCommandMode() && strings.TrimSpace(m.input.Value()) != ""
}

func (m Model) hasNotePalette() bool {
	return m.shouldShowNotePalette() && len(m.noteMatches) > 0
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
	m.editor.SetValue(m.lastSaved)
	m.editor.Focus()
	m.resizeEditor()
	m.input.SetValue("")
	m.input.Blur()
	m.activeDialog = ""
	m.commandIndex = 0
	m.noteIndex = -1
	m.noteMatches = nil
	m.status = fmt.Sprintf("Editing %s", name)
	m.isError = false
}

func (m *Model) closeEditor() {
	name := m.editorName
	m.editorPath = ""
	m.editorName = ""
	m.lastSaved = ""
	m.editor.SetValue("")
	m.editor.Blur()
	m.input.SetValue("")
	m.input.Focus()
	m.syncLauncherState()
	m.status = fmt.Sprintf("Closed %s", name)
	m.isError = false
}

func (m *Model) resizeEditor() {
	width := 64
	height := 12

	if m.width > 0 {
		width = max(24, m.width-12)
	}

	if m.height > 0 {
		height = max(8, m.height-10)
	}

	m.editor.SetWidth(width)
	m.editor.SetHeight(height)
}

func (m *Model) saveEditor() {
	content := m.editor.Value()
	if content == m.lastSaved {
		return
	}

	if err := os.WriteFile(m.editorPath, []byte(content), 0o644); err != nil {
		m.status = fmt.Sprintf("Could not save %s: %v", m.editorName, err)
		m.isError = true
		return
	}

	m.lastSaved = content
	m.status = fmt.Sprintf("Autosaved %s", m.editorName)
	m.isError = false
}

func (m *Model) syncLauncherState() {
	if m.isCommandMode() {
		m.syncCommandSelection()
		m.noteMatches = nil
		m.noteIndex = -1
		return
	}

	m.commandIndex = 0
	m.noteMatches = m.findNoteMatches(strings.TrimSpace(m.input.Value()))
	m.noteIndex = -1
}

func (m *Model) moveNoteSelection(delta int) {
	if len(m.noteMatches) == 0 {
		m.noteIndex = -1
		return
	}

	if m.noteIndex == -1 {
		if delta > 0 {
			m.noteIndex = 0
			return
		}

		m.noteIndex = len(m.noteMatches) - 1
		return
	}

	m.noteIndex = (m.noteIndex + delta + len(m.noteMatches)) % len(m.noteMatches)
}

func (m *Model) openExistingNote(note noteMatch) error {
	m.openEditor(note.path, note.name)
	if m.isError {
		return fmt.Errorf(m.status)
	}

	return nil
}

func (m Model) findNoteMatches(query string) []noteMatch {
	if query == "" {
		return nil
	}

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

		score, ok := fuzzyScore(strings.TrimSuffix(name, filepath.Ext(name)), query)
		if !ok {
			continue
		}

		matches = append(matches, noteMatch{
			name:  name,
			path:  filepath.Join(m.config.NotesPath, name),
			score: score,
		})
	}

	sort.Slice(matches, func(i int, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].name < matches[j].name
		}
		return matches[i].score < matches[j].score
	})

	return matches
}

func fuzzyScore(candidate string, query string) (int, bool) {
	candidate = strings.ToLower(candidate)
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return 0, false
	}

	idx := 0
	lastMatch := -1
	score := 0

	for _, r := range query {
		found := false
		for idx < len(candidate) {
			if rune(candidate[idx]) == r {
				score += idx
				if lastMatch != -1 {
					score += idx - lastMatch - 1
				}
				lastMatch = idx
				idx++
				found = true
				break
			}
			idx++
		}

		if !found {
			return 0, false
		}
	}

	score += len(candidate) - len(query)
	if strings.Contains(candidate, query) {
		score -= len(query)
	}

	return score, true
}
