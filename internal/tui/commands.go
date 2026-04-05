package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) == 0
}

func initProject(m Model) tea.Cmd {
	return func() tea.Msg {
		// Validate prerequisites.
		if _, err := exec.LookPath("go"); err != nil {
			return doneMsg{err: fmt.Errorf("'go' is not installed or not in PATH")}
		}
		if m.gitInit {
			if _, err := exec.LookPath("git"); err != nil {
				return doneMsg{err: fmt.Errorf("'git' is not installed or not in PATH")}
			}
		}

		dir := m.projectDir

		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return doneMsg{err: fmt.Errorf("create directory: %w", err)}
			}
		}

		if m.gitInit {
			out, err := exec.Command("git", "init", dir).CombinedOutput()
			if err != nil {
				return doneMsg{err: fmt.Errorf("git init: %s: %w", strings.TrimSpace(string(out)), err)}
			}
		}

		if m.gitignore {
			if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(goGitignore), 0644); err != nil {
				return doneMsg{err: fmt.Errorf("create .gitignore: %w", err)}
			}
		}

		modInit := exec.Command("go", "mod", "init", m.modulePath)
		modInit.Dir = dir
		if out, err := modInit.CombinedOutput(); err != nil {
			return doneMsg{err: fmt.Errorf("go mod init: %s: %w", strings.TrimSpace(string(out)), err)}
		}

		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGoContent()), 0644); err != nil {
			return doneMsg{err: fmt.Errorf("create main.go: %w", err)}
		}

		for _, pkg := range m.packages {
			cmd := exec.Command("go", "get", pkg.Import+"@latest")
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				return doneMsg{err: fmt.Errorf("go get %s: %s: %w", pkg.Name, strings.TrimSpace(string(out)), err)}
			}
		}

		if m.readme {
			if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readmeContent(m)), 0644); err != nil {
				return doneMsg{err: fmt.Errorf("create README.md: %w", err)}
			}
		}

		return doneMsg{}
	}
}

func readmeContent(m Model) string {
	var sb strings.Builder
	sb.WriteString("# " + m.projectName + "\n\n")
	sb.WriteString("A " + string(m.template) + " project.\n\n")

	if len(m.packages) > 0 {
		sb.WriteString("## Dependencies\n\n")
		for _, p := range m.packages {
			sb.WriteString("- [" + p.Name + "](https://pkg.go.dev/" + p.Import + ")\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Getting Started\n\n")
	sb.WriteString("```bash\ngo run .\n```\n")
	return sb.String()
}

func mainGoContent() string {
	return "package main\n\nfunc main() {\n\tprint(\"Package created\")\n}\n"
}

const goGitignore = `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Go workspace
go.work
go.work.sum

# Environment
.env
`
