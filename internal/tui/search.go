package tui

import (
	"strings"

	"github.com/shad0wcrawl3r/ginit/internal/pkgdata"
)

const pkgVisible = 12

// filterPackages returns packages whose pre-computed SearchText contains query.
func filterPackages(all []pkgdata.Package, query string) []pkgdata.Package {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	out := make([]pkgdata.Package, 0, 64)
	for _, p := range all {
		if strings.Contains(p.SearchText, q) {
			out = append(out, p)
		}
	}
	return out
}

// clampOffset keeps cursor inside the visible window [offset, offset+pkgVisible).
func clampOffset(cursor, offset int) int {
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+pkgVisible {
		return cursor - pkgVisible + 1
	}
	return offset
}
