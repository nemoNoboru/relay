package serve

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

// buildSitePost reads a site's relay-post.json sidecar and resolves its
// member from the sidecar first, then the ownership map. ok=false means the
// site has no resolvable member and should be reported rather than indexed.
// The returned post is not normalized; callers decide how to fill gaps.
func buildSitePost(ks *KeyStore, sitesDir string, name domain.Name) (social.Post, bool) {
	meta := readPostMeta(filepath.Join(sitesDir, string(name)))
	memberID := meta.Member
	if memberID == "" {
		if owner := ks.SiteUser(string(name)); owner != "" {
			if m, ok := ks.MemberByUser(owner); ok {
				memberID = m.ID
			}
		}
	}
	if memberID == "" {
		return social.Post{}, false
	}
	return social.Post{
		Site:     string(name),
		Member:   memberID,
		Title:    meta.Title,
		Summary:  meta.Summary,
		Template: meta.Template,
		Image:    meta.Image,
	}, true
}

// BackfillIndex rebuilds the post index from the site folders on disk. It is
// the shared backfill path used by `relay index`; each site is read from its
// relay-post.json sidecar, falling back to the site name and the ownership
// map. Unattributed sites are reported in the returned skip list.
func BackfillIndex(posts *social.Store, ks *KeyStore, dataDir string) (int, []string, error) {
	if posts == nil {
		return 0, nil, nil
	}
	sitesDir := filepath.Join(dataDir, "sites")
	entries, err := os.ReadDir(sitesDir)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil, nil
	}
	if err != nil {
		return 0, nil, fmt.Errorf("read sites: %w", err)
	}

	indexed := 0
	var skipped []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || !entry.IsDir() {
			continue
		}
		if _, err := domain.ParseName(name); err != nil {
			skipped = append(skipped, name+": invalid name")
			continue
		}
		if _, ok := posts.Get(name); ok {
			continue
		}
		post, ok := buildSitePost(ks, sitesDir, domain.Name(name))
		if !ok {
			skipped = append(skipped, name+": no recorded member")
			continue
		}
		info, err := entry.Info()
		if err != nil {
			skipped = append(skipped, name+": "+err.Error())
			continue
		}
		post.PublishedAt = info.ModTime()
		post.Normalize()
		if _, err := posts.Upsert(post); err != nil {
			skipped = append(skipped, name+": "+err.Error())
			continue
		}
		indexed++
	}
	return indexed, skipped, nil
}

// IndexSite re-reads one site's sidecar and upserts its post record. It is
// used after a rollback, where the live bytes changed without going through
// the uploader. Existing posts keep their member and any fields the sidecar
// leaves empty.
func IndexSite(posts *social.Store, ks *KeyStore, dataDir string, name domain.Name) error {
	if posts == nil {
		return nil
	}
	sitesDir := filepath.Join(dataDir, "sites")
	post, ok := buildSitePost(ks, sitesDir, name)
	existing, found := posts.Get(string(name))
	if !ok && !found {
		return nil
	}
	if !ok {
		post = existing
		post.Site = string(name)
	}
	if found {
		if post.Member == "" {
			post.Member = existing.Member
		}
		if post.Title == "" {
			post.Title = existing.Title
		}
		if post.Summary == "" {
			post.Summary = existing.Summary
		}
		if post.Template == "" {
			post.Template = existing.Template
		}
		if post.Image == "" {
			post.Image = existing.Image
		}
	}
	post.Normalize()
	if _, err := posts.Upsert(post); err != nil {
		return fmt.Errorf("index %s: %w", name, err)
	}
	return nil
}
