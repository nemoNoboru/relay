package serve

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/nemoNoboru/relay/internal/domain"
)

// defaultKeep is how many previous versions of a site are retained. Older
// snapshots are pruned after each deploy and rollback.
const defaultKeep = 10

// ErrUnknownVersion is returned by Rollback for a version id that does not
// exist for the site.
var ErrUnknownVersion = errors.New("unknown version")

// Version is one rollback-able snapshot of a site.
type Version struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created"`
}

// Versions manages immutable snapshots of a site's previous contents under
// <dataDir>/sites/.versions/<site>/<id>/, and the swap that rolls a site
// back to one of them. The live site stays at <dataDir>/sites/<site>/.
//
// Snapshot and Rollback move directories by rename, so they are cheap and
// atomic on one filesystem. The mutex serializes those moves; reads (List)
// do not take it.
type Versions struct {
	sitesDir string
	keep     int
	mu       sync.Mutex
}

// NewVersions returns a version store for the sites under <dataDir>/sites.
func NewVersions(dataDir string) *Versions {
	return &Versions{sitesDir: filepath.Join(dataDir, "sites"), keep: defaultKeep}
}

func (v *Versions) siteDir(name domain.Name) string {
	return filepath.Join(v.sitesDir, string(name))
}

func (v *Versions) root(name domain.Name) string {
	return filepath.Join(v.sitesDir, ".versions", string(name))
}

func (v *Versions) versionDir(name domain.Name, id string) string {
	return filepath.Join(v.root(name), id)
}

// Snapshot moves the live site into a new version, leaving no live site. It
// returns the new version id, or "" when there was no live site to snapshot.
func (v *Versions) Snapshot(name domain.Name) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	target := v.siteDir(name)
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	id, err := newVersionID()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(v.root(name), 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(target, v.versionDir(name, id)); err != nil {
		return "", err
	}
	return id, nil
}

// Restore moves version id back to being the live site. It is the inverse of
// Snapshot and is used to undo a failed swap.
func (v *Versions) Restore(name domain.Name, id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	return os.Rename(v.versionDir(name, id), v.siteDir(name))
}

// Rollback makes version id the live site, first snapshotting the current
// live site so nothing is lost. It returns the id of that displaced snapshot
// ("" when there was no live site).
func (v *Versions) Rollback(name domain.Name, id string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	src := v.versionDir(name, id)
	if fi, err := os.Stat(src); err != nil || !fi.IsDir() {
		return "", fmt.Errorf("%w: %s", ErrUnknownVersion, id)
	}
	displaced, err := v.snapshotLocked(name)
	if err != nil {
		return "", err
	}
	if err := os.Rename(src, v.siteDir(name)); err != nil {
		if displaced != "" {
			_ = os.Rename(v.versionDir(name, displaced), v.siteDir(name))
		}
		return "", err
	}
	return displaced, nil
}

// snapshotLocked moves the live site aside; caller holds the lock.
func (v *Versions) snapshotLocked(name domain.Name) (string, error) {
	target := v.siteDir(name)
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	id, err := newVersionID()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(v.root(name), 0o755); err != nil {
		return "", err
	}
	if err := os.Rename(target, v.versionDir(name, id)); err != nil {
		return "", err
	}
	return id, nil
}

// List returns the site's snapshots, newest first. A site with no history
// yields an empty, non-nil slice.
func (v *Versions) List(name domain.Name) ([]Version, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.listLocked(name)
}

// Prune removes snapshots beyond the retention cap, oldest first.
func (v *Versions) Prune(name domain.Name) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	vs, err := v.listLocked(name)
	if err != nil {
		return err
	}
	if len(vs) <= v.keep {
		return nil
	}
	for _, old := range vs[v.keep:] {
		if err := os.RemoveAll(v.versionDir(name, old.ID)); err != nil {
			return err
		}
	}
	return nil
}

// Remove deletes the live site and all of its snapshots. Removing a site
// that does not exist is a no-op.
func (v *Versions) Remove(name domain.Name) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if err := os.RemoveAll(v.siteDir(name)); err != nil {
		return err
	}
	return os.RemoveAll(v.root(name))
}

func (v *Versions) listLocked(name domain.Name) ([]Version, error) {
	entries, err := os.ReadDir(v.root(name))
	if errors.Is(err, os.ErrNotExist) {
		return []Version{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]Version, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Version{ID: e.Name(), Created: info.ModTime().UTC()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out, nil
}

func newVersionID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate version id: %w", err)
	}
	return time.Now().UTC().Format("20060102T150405.000000000") + "-" + hex.EncodeToString(b), nil
}
