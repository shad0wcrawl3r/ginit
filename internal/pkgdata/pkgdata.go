// Package pkgdata provides the awesome-go package list embedded at compile time.
//
//go:generate go run ../../cmd/genpkgs -o packages.json
package pkgdata

import (
	_ "embed"
	"encoding/json"
	"strings"
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

// All contains every package parsed from the embedded packages.json.
var All []Package

func init() {
	if err := json.Unmarshal(raw, &All); err != nil {
		panic("pkgdata: " + err.Error())
	}
	for i := range All {
		All[i].SearchText = strings.ToLower(
			All[i].Name + " " + All[i].Description + " " + All[i].Category,
		)
	}
}
