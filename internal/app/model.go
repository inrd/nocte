package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
)

type Model struct {
	input  textinput.Model
	width  int
	height int
}

func New() Model {
	input := textinput.New()
	input.Placeholder = "Search or create a note..."
	input.Prompt = ""
	input.Focus()
	input.Width = 48

	return Model{
		input: input,
	}
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
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	inputBox := inputStyle.Render(m.input.View())
	help := helpStyle.Render("Dummy input bar for now. Esc quits.")
	content := lipgloss.JoinVertical(lipgloss.Center, inputBox, help)

	if m.width == 0 || m.height == 0 {
		return docStyle.Render(content)
	}

	horizontal := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	vertical := lipgloss.PlaceVertical(m.height, lipgloss.Center, horizontal)

	return docStyle.Render(strings.TrimRight(vertical, "\n"))
}
