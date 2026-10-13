// Package serve loads site folders from the data directory and serves
// them over HTTP.
package serve

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"relay/internal/domain"
)

// Sites maps site names to their file systems. Once loaded, a Sites is
// immutable and safe for concurrent use.
type Sites map[domain.Name]fs.FS

// LoadSites reads the site folders from <dataDir>/sites. A missing sites
// directory is not an error: it yields an empty Sites. Entries that are
// not directories or whose names fail validation are skipped and
// reported in the second return value.
func LoadSites(dataDir string) (Sites, []string, error) {
	sites := Sites{}
	dir := filepath.Join(dataDir, "sites")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return sites, nil, nil
		}
		return nil, nil, fmt.Errorf("serve: reading %s: %w", dir, err)
	}
	var skipped []string
	for _, e := range entries {
		if !e.IsDir() {
			skipped = append(skipped, e.Name())
			continue
		}
		if _, err := domain.ParseName(e.Name()); err != nil {
			skipped = append(skipped, e.Name())
			continue
		}
		sites[domain.Name(e.Name())] = os.DirFS(filepath.Join(dir, e.Name()))
	}
	return sites, skipped, nil
}
