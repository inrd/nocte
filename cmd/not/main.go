package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

	"github.com/thomas/not/internal/app"
	"github.com/thomas/not/internal/config"
)

func main() {
	cfg, err := config.LoadOrCreate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "not: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(app.New(cfg), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "not: %v\n", err)
		os.Exit(1)
	}
}
