package main

import (
	"fmt"
	"os"

	"github.com/shad0wcrawl3r/ginit/internal/config"
	"github.com/shad0wcrawl3r/ginit/internal/flags"
	"github.com/shad0wcrawl3r/ginit/internal/pkgdata"
	"github.com/shad0wcrawl3r/ginit/internal/selfupdate"
	"github.com/shad0wcrawl3r/ginit/internal/tui"

	tea "charm.land/bubbletea/v2"
)

var (
	version   = "dev"
	buildDate = "" // injected via -ldflags "-X main.buildDate=..."
)

func main() {
	var cfg config.Config
	switch flags.Parse(&cfg) {
	case flags.ShowVersion:
		fmt.Println("ginit", version)
		return
	case flags.SelfUpdate:
		if err := selfupdate.Run(version); err != nil {
			fmt.Fprintf(os.Stderr, "update failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Resolve packages (embedded → remote → cache → fallback).
	pkgdata.Load(buildDate)

	m := tui.NewModel(cfg)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
