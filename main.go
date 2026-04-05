package main

import (
	"fmt"
	"os"

	"gnit/internal/config"
	"gnit/internal/flags"
	"gnit/internal/tui"

	tea "charm.land/bubbletea/v2"
)

var version = "dev"

func main() {
	var cfg config.Config
	if flags.Parse(&cfg) {
		fmt.Println("ginit", version)
		return
	}

	m := tui.NewModel(cfg)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
