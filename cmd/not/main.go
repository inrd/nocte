package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

	"github.com/thomas/not/internal/app"
)

func main() {
	program := tea.NewProgram(app.New(), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "not: %v\n", err)
		os.Exit(1)
	}
}
