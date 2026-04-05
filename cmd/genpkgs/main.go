// genpkgs fetches the awesome-go README and emits packages.json.
//
// Usage:
//
//	go run ./cmd/genpkgs -o internal/pkgdata/packages.json
//	go run ./cmd/genpkgs  # writes to stdout
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Package struct {
	Name        string `json:"name"`
	Import      string `json:"import"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

var (
	categoryRe = regexp.MustCompile(`^## (.+)`)
	linkRe     = regexp.MustCompile(`^\s*-\s+\[([^\]]+)\]\(([^)]+)\)\s*[-–—]\s*(.+)`)
)

var codeHosts = map[string]bool{
	"github.com":    true,
	"gitlab.com":    true,
	"gopkg.in":      true,
	"pkg.go.dev":    true,
	"sr.ht":         true,
	"codeberg.org":  true,
	"bitbucket.org": true,
}

var skipCategories = map[string]bool{
	"Contents":     true,
	"Contributing": true,
	"License":      true,
	"Websites":     true,
	"Tutorials":    true,
	"Style Guides": true,
	"Social Media": true,
	"Meetups":      true,
	"Conferences":  true,
}

func main() {
	outputPath := flag.String("o", "", "output file path (default: stdout)")
	flag.Parse()

	resp, err := http.Get("https://raw.githubusercontent.com/avelino/awesome-go/main/README.md")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "fetch: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}

	var packages []Package
	var currentCategory string
	seen := make(map[string]bool)

	for _, line := range strings.Split(string(body), "\n") {
		if m := categoryRe.FindStringSubmatch(line); m != nil {
			cat := strings.TrimSpace(m[1])
			if skipCategories[cat] {
				currentCategory = ""
			} else {
				currentCategory = cat
			}
			continue
		}

		if currentCategory == "" {
			continue
		}

		m := linkRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		name := strings.TrimSpace(m[1])
		rawURL := strings.TrimSpace(m[2])
		desc := strings.TrimSpace(m[3])

		u, err := url.Parse(rawURL)
		if err != nil || !codeHosts[u.Host] {
			continue
		}

		imp := u.Host + strings.TrimSuffix(u.Path, ".git")
		imp = strings.TrimSuffix(imp, "/")

		// Deduplicate by import path.
		if seen[imp] {
			continue
		}
		seen[imp] = true

		packages = append(packages, Package{
			Name:        name,
			Import:      imp,
			URL:         rawURL,
			Description: desc,
			Category:    currentCategory,
		})
	}

	var w io.Writer = os.Stdout
	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "create:", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(packages); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "wrote %d packages\n", len(packages))
}
