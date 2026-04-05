# ginit

Interactive CLI tool that scaffolds new Go projects. Think `create-vite` but for Go.

## Features

- Step-by-step TUI wizard with a stacked, create-vite-style interface
- Split module path input (domain / user / module) with tab navigation
- Template selection — cli, web, or tui
- Searchable package picker with 2600+ packages from [awesome-go](https://github.com/avelino/awesome-go)
- Auto-generates `go.mod`, `main.go`, README, and `.gitignore`
- Package index refreshed daily via CI

## Install

**From release binaries:**

```bash
# Linux
curl -Lo ginit https://github.com/shad0wcrawl3r/ginit/releases/latest/download/ginit-linux-amd64
chmod +x ginit
sudo mv ginit /usr/local/bin/

# macOS
curl -Lo ginit https://github.com/shad0wcrawl3r/ginit/releases/latest/download/ginit-darwin-arm64
chmod +x ginit
sudo mv ginit /usr/local/bin/
```

**From source:**

```bash
go install gnit@latest
```

## Usage

```bash
ginit
```

The wizard will guide you through:

```
✔ Project name  › myapp
✔ Directory     › myapp
? Module path
  github.com / johndoe / myapp
  github.com/johndoe/myapp
  $ go mod init github.com/johndoe/myapp
```

### Flags

```
-n, --name       Pre-fill the project name
-t, --template   Pre-fill the template (cli, web, tui)
-v, --version    Print version and exit
```

Example:

```bash
ginit --name myapp --template cli
```

## What gets generated

Depending on your choices, ginit creates:

```
myapp/
├── go.mod          # always
├── main.go         # always
├── .gitignore      # if selected
└── README.md       # if selected, lists chosen packages
```

Selected packages are installed via `go get`.

## Building from source

```bash
git clone https://github.com/shad0wcrawl3r/ginit.git
cd ginit
go build .
```

### Refreshing the package index

```bash
go run ./cmd/genpkgs -o internal/pkgdata/packages.json
```

This fetches the latest awesome-go README and regenerates the embedded package list. CI does this daily.

## License

MIT
