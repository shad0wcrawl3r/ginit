# AGENTS.md — ginit

## What this project is

ginit is an interactive CLI tool that scaffolds new Go projects. It presents a step-by-step TUI wizard (similar to create-vite) that collects project configuration and then generates the project directory, `go.mod`, `main.go`, selected dependencies, README, and `.gitignore`.

The binary ships with an embedded JSON index of ~2600 Go packages sourced from the [awesome-go](https://github.com/avelino/awesome-go) list. A GitHub Actions workflow refreshes this index daily and auto-releases new binaries when packages change.

## Architecture

```
ginit/
├── main.go                          # entrypoint — parses flags, launches TUI
├── go.mod                           # module: gnit (Go 1.26)
├── internal/
│   ├── config/config.go             # Config struct + Validate()
│   ├── flags/flags.go               # CLI flag parsing (--name, --template, --version)
│   ├── pkgdata/
│   │   ├── pkgdata.go               # //go:embed packages.json, Package type, init()
│   │   └── packages.json            # ~2600 awesome-go entries (auto-generated, do not hand-edit)
│   └── tui/
│       ├── model.go                 # Bubble Tea model, steps, Update loop
│       ├── views.go                 # View functions for each step (stacked history style)
│       ├── commands.go              # initProject() tea.Cmd — runs git/go/file operations
│       └── search.go               # filterPackages(), clampOffset()
├── cmd/genpkgs/main.go              # generator: fetches awesome-go README → packages.json
└── .github/workflows/
    ├── update-packages.yml          # daily cron: regenerate JSON, auto-tag + release if changed
    └── release.yml                  # on v* tag push: cross-compile 6 binaries, GitHub release
```

## TUI flow

The wizard progresses through these steps (defined as `Step` iota in `model.go`):

| Step | Type | What it collects |
|------|------|------------------|
| `StepStart` | automatic | Spinner shown for one tick, then auto-advances |
| `StepProjectName` | text input | Project name (pre-filled from `--name` flag if given) |
| `StepProjectDir` | text input | Directory to create (defaults to `.` if cwd is empty, else project name) |
| `StepModulePath` | 3 split text inputs | Domain / user / module — tab or `/` advances between them |
| `StepTemplate` | single select | cli, web, or tui |
| `StepGitInit` | yes/no | Whether to run `git init` |
| `StepReadme` | yes/no | Whether to generate README.md |
| `StepGitIgnore` | yes/no | Whether to generate .gitignore |
| `StepPackages` | searchable multi-select | Pick from embedded awesome-go index |
| `StepDone` | execution + result | Runs `initProject()`, shows spinner then result |

All completed steps remain visible above the current prompt (create-vite stacked style via `renderHistory()` + `withHistory()`).

## Key design decisions

- **Value receiver model** — Bubble Tea's `Update()` returns `(tea.Model, tea.Cmd)`. All mutation happens on the local copy within Update and is returned. Pointer-receiver methods like `textinput.Focus()` and `textinput.Blur()` work because the struct fields are addressable within the function body.

- **Three separate text inputs for module path** — `domainInput`, `userInput`, `moduleInput` in the Model. Tab/`/` advances focus, shift+tab retreats. Only arrow keys control list cursors (not j/k, which must be typeable in text fields).

- **Package search** — `pkgdata.Package.SearchText` is pre-computed (lowercased name+description+category) at `init()` time. `filterPackages()` does a single `strings.Contains` per package per keystroke. Re-filtering only runs when the query actually changes (compared via `searchQuery` field).

- **`selectedImports map[string]bool`** — toggling off deletes the key (not sets to false) so `len()` always reflects the true count.

- **Execution in a tea.Cmd** — `initProject()` returns a closure that runs all file/shell operations off the main thread. It validates `go`/`git` on PATH before starting. Returns `doneMsg{err}` which triggers quit.

## packages.json lifecycle

1. `cmd/genpkgs` fetches `https://raw.githubusercontent.com/avelino/awesome-go/main/README.md`
2. Parses `## Category` headings and `- [Name](URL) - Description` links via regex
3. `moduleRoot()` extracts clean import paths from URLs (strips `/tree/main/...` deep links, keeps `/vN` suffixes)
4. Deduplicates by import path, filters to known code hosts (github, gitlab, etc.)
5. Outputs JSON array of `{name, import, url, description, category}`

Regenerate locally: `go run ./cmd/genpkgs -o internal/pkgdata/packages.json`

The `update-packages.yml` workflow does this daily at 06:00 UTC. If the JSON changes, it commits, computes the next patch version from existing git tags, cross-compiles, and publishes a GitHub release.

## Building

```bash
go build .                                           # dev binary (version: "dev")
go build -ldflags "-s -w -X main.version=v1.0.0" .  # release binary with version
```

The release workflow builds for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64.

## Dependencies

| Package | Purpose |
|---------|---------|
| `charm.land/bubbletea/v2` | TUI framework (Elm architecture) |
| `charm.land/bubbles/v2` | Text input and spinner components |
| `charm.land/lipgloss/v2` | Terminal styling/colours |
| `github.com/spf13/pflag` | POSIX-compliant CLI flag parsing |

## Things to watch out for

- **Do not hand-edit `packages.json`** — it is auto-generated and will be overwritten by CI. Modify `cmd/genpkgs/main.go` instead.
- **bubbletea v2 key names** — space is `"space"` (not `" "`), shift+tab is `"shift+tab"`, arrows are `"up"`/`"down"`. Check the ultraviolet key table if adding new keybindings.
- **Module name is `gnit`** not `ginit` — this is the Go module path in `go.mod`. The binary is named `ginit` in release workflows via `-o` flag.
- **`initProject` runs shell commands** — it calls `git init`, `go mod init`, `go get`. These require network access and the respective tools on PATH. The function validates prerequisites before starting.
