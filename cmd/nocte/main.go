package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"

	"github.com/inrd/nocte/internal/app"
	"github.com/inrd/nocte/internal/config"
)

const version = "0.5.0"

func main() {
	cfg, configPath, err := config.LoadOrCreate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nocte: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(app.New(cfg, configPath, version), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "nocte: %v\n", err)
		os.Exit(1)
	}
}
