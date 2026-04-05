package tui

import (
	"gnit/internal/config"
	"gnit/internal/pkgdata"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Step int

const (
	StepStart Step = iota
	StepProjectName
	StepProjectDir
	StepModulePath
	StepTemplate
	StepGitInit
	StepReadme
	StepGitIgnore
	StepPackages
	StepDone
)

type Template string

const (
	TemplateCLI Template = "cli"
	TemplateWeb Template = "web"
	TemplateTUI Template = "tui"
)

var templates = []Template{TemplateCLI, TemplateWeb, TemplateTUI}

// Package is the trimmed representation used by initProject.
type Package struct {
	Name   string
	Import string
}

type doneMsg struct{ err error }

type Model struct {
	step        Step
	projectName string
	projectDir  string
	modulePath  string
	template    Template
	gitInit     bool
	readme      bool
	gitignore   bool
	packages    []Package // final selection, populated on StepPackages confirm

	// StepProjectName / StepProjectDir
	textInput textinput.Model

	// StepModulePath — three split inputs
	domainInput      textinput.Model
	userInput        textinput.Model
	moduleInput      textinput.Model
	moduleActivePart int // 0=domain 1=user 2=module

	// StepTemplate / yes-no steps
	cursor int

	// StepPackages
	searchInput     textinput.Model
	allPkgs         []pkgdata.Package
	filteredPkgs    []pkgdata.Package
	searchQuery     string
	selectedImports map[string]bool
	listOffset      int

	spinner   spinner.Model
	cfg       config.Config
	executing bool
	err       error
}

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

func newTextInput(placeholder string, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 255
	ti.SetWidth(width)
	return ti
}

func NewModel(cfg config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Moon
	s.Style = spinnerStyle

	si := textinput.New()
	si.Placeholder = "Search packages..."
	si.CharLimit = 100
	si.SetWidth(40)

	all := pkgdata.All
	return Model{
		step:            StepStart,
		spinner:         s,
		cfg:             cfg,
		textInput:       textinput.New(),
		searchInput:     si,
		allPkgs:         all,
		filteredPkgs:    all,
		selectedImports: make(map[string]bool),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case doneMsg:
		m.executing = false
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.executing {
			break
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.step != StepProjectName && m.step != StepProjectDir && m.step != StepModulePath {
				return m, tea.Quit
			}

		case "q":
			if m.step != StepProjectName && m.step != StepProjectDir &&
				m.step != StepModulePath && m.step != StepPackages {
				return m, tea.Quit
			}
			if m.step == StepPackages {
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case "tab", "/":
			if m.step == StepModulePath {
				switch m.moduleActivePart {
				case 0:
					m.domainInput.Blur()
					m.moduleActivePart = 1
					cmds = append(cmds, m.userInput.Focus())
				case 1:
					m.userInput.Blur()
					m.moduleActivePart = 2
					cmds = append(cmds, m.moduleInput.Focus())
				}
			}

		case "shift+tab":
			if m.step == StepModulePath {
				switch m.moduleActivePart {
				case 1:
					m.userInput.Blur()
					m.moduleActivePart = 0
					cmds = append(cmds, m.domainInput.Focus())
				case 2:
					m.moduleInput.Blur()
					m.moduleActivePart = 1
					cmds = append(cmds, m.userInput.Focus())
				}
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down":
			switch m.step {
			case StepTemplate:
				if m.cursor < len(templates)-1 {
					m.cursor++
				}
			case StepPackages:
				if m.cursor < len(m.filteredPkgs)-1 {
					m.cursor++
				}
			default:
				if m.cursor < 1 {
					m.cursor++
				}
			}

		case "space":
			if m.step == StepPackages && len(m.filteredPkgs) > 0 {
				imp := m.filteredPkgs[m.cursor].Import
				if m.selectedImports[imp] {
					delete(m.selectedImports, imp)
				} else {
					m.selectedImports[imp] = true
				}
			}

		case "enter":
			switch m.step {
			case StepProjectName:
				if m.textInput.Value() == "" {
					return m, nil
				}
				m.projectName = m.textInput.Value()
				m.textInput.Reset()
				m.textInput.Placeholder = "directory"
				m.textInput.CharLimit = 255
				m.textInput.SetWidth(30)
				if isDirEmpty(".") {
					m.textInput.SetValue(".")
				} else {
					m.textInput.SetValue(m.projectName)
				}
				m.step++
				cmds = append(cmds, m.textInput.Focus())

			case StepProjectDir:
				if m.textInput.Value() == "" {
					return m, nil
				}
				m.projectDir = m.textInput.Value()
				m.textInput.Reset()
				m.domainInput = newTextInput("domain", 14)
				m.domainInput.SetValue("github.com")
				m.userInput = newTextInput("username", 16)
				m.moduleInput = newTextInput("module", 20)
				m.moduleInput.SetValue(m.projectName)
				m.moduleActivePart = 1
				m.step++
				cmds = append(cmds, m.userInput.Focus())

			case StepModulePath:
				if m.domainInput.Value() == "" || m.userInput.Value() == "" || m.moduleInput.Value() == "" {
					return m, nil
				}
				m.modulePath = m.domainInput.Value() + "/" + m.userInput.Value() + "/" + m.moduleInput.Value()
				m.domainInput.Blur()
				m.userInput.Blur()
				m.moduleInput.Blur()
				m.cursor = 0
				m.step++

			case StepTemplate:
				m.template = templates[m.cursor]
				m.cursor = 0
				m.step++

			case StepGitInit:
				m.gitInit = m.cursor == 0
				m.cursor = 0
				m.step++

			case StepReadme:
				m.readme = m.cursor == 0
				m.cursor = 0
				m.step++

			case StepGitIgnore:
				m.gitignore = m.cursor == 0
				m.cursor = 0
				// reset package search state
				m.searchQuery = ""
				m.searchInput.Reset()
				m.filteredPkgs = m.allPkgs
				m.listOffset = 0
				m.step++
				cmds = append(cmds, m.searchInput.Focus())

			case StepPackages:
				var selected []Package
				seen := make(map[string]bool)
				for _, p := range m.allPkgs {
					if m.selectedImports[p.Import] && !seen[p.Import] {
						seen[p.Import] = true
						selected = append(selected, Package{Name: p.Name, Import: p.Import})
					}
				}
				m.packages = selected
				m.step = StepDone
				m.executing = true
				return m, initProject(m)
			}

		default:
			switch m.step {
			case StepProjectName, StepProjectDir:
				m.textInput, cmd = m.textInput.Update(msg)
				cmds = append(cmds, cmd)
			case StepModulePath:
				switch m.moduleActivePart {
				case 0:
					m.domainInput, cmd = m.domainInput.Update(msg)
				case 1:
					m.userInput, cmd = m.userInput.Update(msg)
				case 2:
					m.moduleInput, cmd = m.moduleInput.Update(msg)
				}
				cmds = append(cmds, cmd)
			case StepPackages:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	if m.step == StepStart {
		m.step = StepProjectName
		m.textInput.Placeholder = "Project name"
		m.textInput.CharLimit = 255
		m.textInput.SetWidth(25)
		if m.cfg.PackageName != "" {
			m.textInput.SetValue(m.cfg.PackageName)
		}
		cmds = append(cmds, m.textInput.Focus())
	}

	// Re-filter packages whenever search query changes.
	if m.step == StepPackages {
		if q := m.searchInput.Value(); q != m.searchQuery {
			m.searchQuery = q
			m.filteredPkgs = filterPackages(m.allPkgs, q)
			m.cursor = 0
			m.listOffset = 0
		}
		m.listOffset = clampOffset(m.cursor, m.listOffset)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	switch m.step {
	case StepStart:
		return m.LoadingView()
	case StepProjectName:
		return m.ProjectNameView()
	case StepProjectDir:
		return m.ProjectDirView()
	case StepModulePath:
		return m.ModulePathView()
	case StepTemplate:
		return m.TemplateView()
	case StepGitInit:
		return m.GitInitView()
	case StepReadme:
		return m.ReadmeView()
	case StepGitIgnore:
		return m.GitIgnoreView()
	case StepPackages:
		return m.PackagesView()
	case StepDone:
		return m.DoneView()
	default:
		return m.LoadingView()
	}
}
