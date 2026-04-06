// Package selfupdate checks GitHub releases and replaces the running binary.
package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	releasesURL = "https://api.github.com/repos/shad0wcrawl3r/ginit/releases/latest"
	timeout     = 10 * time.Second
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Run checks for the latest release and updates the binary in-place.
// currentVersion is the value of main.version (e.g. "v0.1.2" or "dev").
func Run(currentVersion string) error {
	fmt.Println("Checking for updates...")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(releasesURL)
	if err != nil {
		return fmt.Errorf("failed to check releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return fmt.Errorf("failed to parse release: %w", err)
	}

	if rel.TagName == currentVersion {
		fmt.Printf("Already up to date (%s)\n", currentVersion)
		return nil
	}

	// Find the asset for this OS/arch.
	wantName := assetName(runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range rel.Assets {
		if a.Name == wantName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
	}

	fmt.Printf("Updating %s → %s ...\n", currentVersion, rel.TagName)

	// Download new binary.
	dlResp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", dlResp.StatusCode)
	}

	// Write to a temp file next to the current executable.
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate current binary: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "ginit-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("download write failed: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod failed: %w", err)
	}

	// Atomic swap: rename temp over current binary.
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Cross-device? Fall back to copy.
		if err2 := copyFile(tmpPath, execPath); err2 != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("cannot replace binary: %w (rename: %v)", err2, err)
		}
		os.Remove(tmpPath)
	}

	fmt.Printf("Updated to %s\n", rel.TagName)
	return nil
}

func assetName(goos, goarch string) string {
	ext := ""
	if goos == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("ginit-%s-%s%s", goos, goarch, ext)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// StripV removes a leading "v" prefix for display if present.
func StripV(version string) string {
	return strings.TrimPrefix(version, "v")
}
