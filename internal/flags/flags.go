package flags

import (
	"gnit/internal/config"

	"github.com/spf13/pflag"
)

// Parse populates cfg from CLI flags. Returns true if --version was requested.
func Parse(cfg *config.Config) bool {
	pflag.StringVarP(&cfg.PackageName, "name", "n", "", "Project name")
	pflag.StringVarP(&cfg.Template, "template", "t", "", "Template (cli, web, tui)")
	showVersion := pflag.BoolP("version", "v", false, "Print version and exit")
	pflag.Parse()
	return *showVersion
}
