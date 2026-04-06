package flags

import (
	"github.com/shad0wcrawl3r/ginit/internal/config"

	"github.com/spf13/pflag"
)

// Result indicates which action main should take after parsing flags.
type Result int

const (
	RunTUI    Result = iota // normal TUI mode
	ShowVersion             // --version
	SelfUpdate              // --update
)

// Parse populates cfg from CLI flags and returns the requested action.
func Parse(cfg *config.Config) Result {
	pflag.StringVarP(&cfg.PackageName, "name", "n", "", "Project name")
	pflag.StringVarP(&cfg.Template, "template", "t", "", "Template (cli, web, tui)")
	showVersion := pflag.BoolP("version", "v", false, "Print version and exit")
	update := pflag.BoolP("update", "u", false, "Self-update to latest release")
	pflag.Parse()

	switch {
	case *update:
		return SelfUpdate
	case *showVersion:
		return ShowVersion
	default:
		return RunTUI
	}
}
