package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
	case ":list":
		m.openListDialog()
		return m, nil
	case ":files":
		if err := os.MkdirAll(m.config.NotesPath, 0o755); err != nil {
			m.status = fmt.Sprintf("Could not prepare notes dir: %v", err)
			m.isError = true
			return m, nil
		}
		if err := openPathWithSystemApp(m.config.NotesPath); err != nil {
			m.status = fmt.Sprintf("Could not open notes dir: %v", err)
			m.isError = true
			return m, nil
		}
		m.input.SetValue("")
		m.input.Focus()
		m.syncLauncherState()
		m.status = "Opened notes folder"
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
