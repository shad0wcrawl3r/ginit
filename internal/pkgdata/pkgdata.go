// Package pkgdata provides the awesome-go package list.
//
// Resolution order:
//  1. If binary was built < 24h ago → use embedded packages.json
//  2. Else → try GitHub raw content (3s timeout), cache on success
//  3. Else → try local cache (if < 24h old)
//  4. Fallback → embedded
//
//go:generate go run ../../cmd/genpkgs -o packages.json
package pkgdata

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Package is a single entry from the awesome-go list.
type Package struct {
	Name        string `json:"name"`
	Import      string `json:"import"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Category    string `json:"category"`
	SearchText  string `json:"-"` // pre-computed lowercase for fast filtering
}

//go:embed packages.json
var raw []byte

const (
	githubRawURL = "https://raw.githubusercontent.com/shad0wcrawl3r/ginit/main/internal/pkgdata/packages.json"
	fetchTimeout = 3 * time.Second
	maxAge       = 24 * time.Hour
	cacheFile    = "ginit-packages.json"
)

// All contains every package, loaded via Load().
var All []Package

func init() {
	// Default: parse embedded so the binary always works even if Load() isn't called.
	All = mustParse(raw)
}

// Load resolves packages using the three-source strategy.
// buildDate should be the value injected via ldflags (RFC3339 or "").
func Load(buildDate string) {
	// 1. If binary is fresh (< 24h), embedded is fine.
	if bd, err := time.Parse(time.RFC3339, buildDate); err == nil {
		if time.Since(bd) < maxAge {
			return // All already set from init()
		}
	}

	// 2. Try GitHub raw content.
	if data, err := fetchRemote(); err == nil {
		if pkgs, err := parse(data); err == nil {
			All = pkgs
			_ = writeCache(data) // best-effort cache
			return
		}
	}

	// 3. Try local cache.
	if data, age, err := readCache(); err == nil && age < maxAge {
		if pkgs, err := parse(data); err == nil {
			All = pkgs
			return
		}
	}

	// 4. Fallback: embedded (already set in init).
}

func fetchRemote() ([]byte, error) {
	client := &http.Client{Timeout: fetchTimeout}
	resp, err := client.Get(githubRawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func cacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "ginit")
	return p, os.MkdirAll(p, 0o755)
}

func cachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFile), nil
}

func writeCache(data []byte) error {
	p, err := cachePath()
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func readCache() ([]byte, time.Duration, error) {
	p, err := cachePath()
	if err != nil {
		return nil, 0, err
	}
	info, err := os.Stat(p)
	if err != nil {
		return nil, 0, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, 0, err
	}
	return data, time.Since(info.ModTime()), nil
}

func parse(data []byte) ([]Package, error) {
	var pkgs []Package
	if err := json.Unmarshal(data, &pkgs); err != nil {
		return nil, err
	}
	for i := range pkgs {
		pkgs[i].SearchText = strings.ToLower(
			pkgs[i].Name + " " + pkgs[i].Description + " " + pkgs[i].Category,
		)
	}
	return pkgs, nil
}

func mustParse(data []byte) []Package {
	pkgs, err := parse(data)
	if err != nil {
		panic("pkgdata: " + err.Error())
	}
	return pkgs
}
