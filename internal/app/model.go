package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	width        int
	height       int
	status       string
	isError      bool
	config       config.Config
	configPath   string
	version      string
	activeDialog string
	commandIndex int
}

type command struct {
	name        string
	description string
}

func New(cfg config.Config, configPath string, version string) Model {
	input := textinput.New()
	input.Placeholder = "Search or create a note..."
	input.Prompt = ""
	input.Focus()
	input.Width = 48

	return Model{
		input:      input,
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
		return m, nil
	case tea.KeyMsg:
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
		case "down":
			if m.isCommandMode() {
				m.moveCommandSelection(1)
				return m, nil
			}
		case "enter":
			if m.isCommandMode() {
				return m.handleCommand()
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

			m.status = fmt.Sprintf("Created %s", filename+".md")
			m.isError = false
			m.input.SetValue("")
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.syncCommandSelection()
	return m, cmd
}

func (m Model) View() string {
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
