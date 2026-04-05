package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	checkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	questionStyle = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// renderHistory returns all completed step answers stacked above the current prompt,
// matching the create-vite-app style.
func renderHistory(m Model) string {
	check := checkStyle.Render("✔")
	sep := dimStyle.Render(" › ")
	var lines []string

	if m.step > StepProjectName {
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("Project name"), sep, m.projectName))
	}
	if m.step > StepProjectDir {
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("Directory   "), sep, m.projectDir))
	}
	if m.step > StepModulePath {
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("Module path "), sep, m.modulePath))
	}
	if m.step > StepTemplate {
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("Template    "), sep, string(m.template)))
	}
	if m.step > StepGitInit {
		v := yesNo(m.gitInit)
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("Git init    "), sep, v))
	}
	if m.step > StepReadme {
		v := yesNo(m.readme)
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render("README.md   "), sep, v))
	}
	if m.step > StepGitIgnore {
		v := yesNo(m.gitignore)
		lines = append(lines, fmt.Sprintf("%s %s%s%s", check, dimStyle.Render(".gitignore  "), sep, v))
	}

	return strings.Join(lines, "\n")
}

func withHistory(m Model, current string) string {
	h := renderHistory(m)
	if h == "" {
		return current
	}
	return h + "\n" + current
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func (m Model) LoadingView() tea.View {
	return tea.NewView(m.spinner.View() + " Loading...")
}

func (m Model) ProjectNameView() tea.View {
	content := questionStyle.Render("? Project name") + "\n" + m.textInput.View()
	return tea.NewView(withHistory(m, content))
}

var (
	domainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	userStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	moduleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	cmdStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true)
)

func (m Model) ProjectDirView() tea.View {
	val := m.textInput.Value()

	var where string
	if val == "" || val == "." {
		where = dimStyle.Render("current directory")
	} else {
		where = moduleStyle.Render("./"+val+"/")
	}

	var action string
	if val == "" || val == "." {
		action = dimStyle.Render("  files will be written into the ") + where
	} else {
		action = dimStyle.Render("  a new directory ") + where + dimStyle.Render(" will be created")
	}

	hint := dimStyle.Render("  . to use the current directory")

	content := questionStyle.Render("? Project directory") + "\n" +
		m.textInput.View() + "\n\n" +
		action + "\n" +
		hint
	return tea.NewView(withHistory(m, content))
}

func (m Model) ModulePathView() tea.View {
	sep := dimStyle.Render(" / ")
	row := m.domainInput.View() + sep + m.userInput.View() + sep + m.moduleInput.View()

	domain := m.domainInput.Value()
	user := m.userInput.Value()
	module := m.moduleInput.Value()

	// Colour each segment of the preview path to match its input
	pathPreview := domainStyle.Render(domain) +
		dimStyle.Render("/") +
		userStyle.Render(user) +
		dimStyle.Render("/") +
		moduleStyle.Render(module)

	// Show the go mod init command that will run
	fullPath := domain + "/" + user + "/" + module
	initCmd := cmdStyle.Render("$ go mod init " + fullPath)

	hint := dimStyle.Render("tab · / next   shift+tab prev")

	content := questionStyle.Render("? Module path") + "\n" +
		row + "\n\n" +
		"  " + pathPreview + "\n" +
		"  " + initCmd + "\n\n" +
		"  " + hint
	return tea.NewView(withHistory(m, content))
}

func (m Model) TemplateView() tea.View {
	lines := []string{questionStyle.Render("? Select a template")}
	for i, t := range templates {
		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("> "+string(t)))
		} else {
			lines = append(lines, dimStyle.Render("  "+string(t)))
		}
	}
	return tea.NewView(withHistory(m, strings.Join(lines, "\n")))
}

func (m Model) GitInitView() tea.View {
	return tea.NewView(withHistory(m, yesNoView("? Initialize git repository?", m.cursor)))
}

func (m Model) ReadmeView() tea.View {
	hint := dimStyle.Render("  will list selected packages")
	content := yesNoView("? Create README.md?", m.cursor) + "\n" + hint
	return tea.NewView(withHistory(m, content))
}

func (m Model) GitIgnoreView() tea.View {
	hint := dimStyle.Render("  prefilled with Go template")
	content := yesNoView("? Create .gitignore?", m.cursor) + "\n" + hint
	return tea.NewView(withHistory(m, content))
}

func (m Model) PackagesView() tea.View {
	header := questionStyle.Render("? Select packages") +
		dimStyle.Render("  ↑↓ navigate · space toggle · enter confirm")

	// search input row
	searchRow := m.searchInput.View()

	// result count
	total := len(m.filteredPkgs)
	sel := len(m.selectedImports)
	stats := dimStyle.Render(fmt.Sprintf("  %d results · %d selected", total, sel))

	lines := []string{header, searchRow, ""}

	start := m.listOffset
	end := min(start+pkgVisible, total)

	for i := start; i < end; i++ {
		p := m.filteredPkgs[i]
		selected := m.selectedImports[p.Import]

		box := "[ ]"
		if selected {
			box = "[x]"
		}

		name := truncateRunes(p.Name, 20)
		desc := truncateRunes(p.Description, 48)
		cat := dimStyle.Render(p.Category)

		nameCol := fmt.Sprintf("%-22s", name)
		descCol := fmt.Sprintf("%-50s", desc)

		var line string
		switch {
		case i == m.cursor:
			line = selectedStyle.Render("> "+box+" "+nameCol+"  "+descCol) + "  " + cat
		case selected:
			line = checkStyle.Render("  "+box+" "+nameCol) + dimStyle.Render("  "+descCol) + "  " + cat
		default:
			line = dimStyle.Render("  "+box+" "+nameCol+"  "+descCol+"  "+p.Category)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", stats)
	return tea.NewView(withHistory(m, strings.Join(lines, "\n")))
}

func (m Model) DoneView() tea.View {
	if m.executing {
		return tea.NewView(withHistory(m, m.spinner.View()+" Setting up project..."))
	}
	if m.err != nil {
		return tea.NewView(withHistory(m, "✖ "+m.err.Error()+"\n"))
	}
	msg := checkStyle.Render("✔") + " Done! Project " + questionStyle.Render(m.projectName) + " created.\n"
	return tea.NewView(withHistory(m, msg))
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}

func yesNoView(question string, cursor int) string {
	yes := dimStyle.Render("  Yes")
	no := dimStyle.Render("  No")
	if cursor == 0 {
		yes = selectedStyle.Render("> Yes")
	} else {
		no = selectedStyle.Render("> No")
	}
	return strings.Join([]string{questionStyle.Render(question), yes, no}, "\n")
}
