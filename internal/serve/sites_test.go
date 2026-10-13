package serve_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// AC-12: a <data>/sites directory with docs/ and blog/ subdirectories
// produces a Sites map with exactly keys "docs" and "blog". A malformed
// subdirectory name ("Bad Dir") is skipped without error.
func TestLoadSites(t *testing.T) {
	data := t.TempDir()
	writeFile(t, filepath.Join(data, "sites", "docs", "index.html"), "<html></html>")
	writeFile(t, filepath.Join(data, "sites", "docs", "assets", "style.css"), "body{}")
	writeFile(t, filepath.Join(data, "sites", "blog", "index.html"), "<html></html>")
	writeFile(t, filepath.Join(data, "sites", "Bad Dir", "index.html"), "should be skipped")

	sites, skipped, err := serve.LoadSites(data)
	if err != nil {
		t.Fatalf("LoadSites: %v", err)
	}
	if len(sites) != 2 {
		t.Fatalf("len(sites) = %d, want 2 (keys: %v)", len(sites), siteKeys(t, sites))
	}
	if _, ok := sites["docs"]; !ok {
		t.Error(`sites missing key "docs"`)
	}
	if _, ok := sites["blog"]; !ok {
		t.Error(`sites missing key "blog"`)
	}
	if len(skipped) != 1 || skipped[0] != "Bad Dir" {
		t.Fatalf("skipped = %v, want [Bad Dir]", skipped)
	}
}

// AC-13 (data side): a data directory without a sites/ subdirectory returns
// an empty Sites and a nil error.
func TestLoadSitesMissingDir(t *testing.T) {
	data := t.TempDir()

	sites, skipped, err := serve.LoadSites(data)
	if err != nil {
		t.Fatalf("LoadSites on empty data dir: %v", err)
	}
	if len(sites) != 0 {
		t.Fatalf("len(sites) = %d, want 0", len(sites))
	}
	if len(skipped) != 0 {
		t.Fatalf("skipped = %v, want empty", skipped)
	}
}

func siteKeys(t *testing.T, sites serve.Sites) []string {
	t.Helper()
	keys := make([]string, 0, len(sites))
	for k := range sites {
		keys = append(keys, string(k))
	}
	return keys
}
