package serve

import (
	"fmt"
	"path/filepath"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

// IndexSite re-reads a site's relay-post.json and upserts its post record.
// It is used after a rollback, where the live bytes changed without going
// through the uploader. If the sidecar has no member and there is no existing
// record, the site is left unindexed.
func IndexSite(posts *social.Store, dataDir string, name domain.Name) error {
	if posts == nil {
		return nil
	}
	dir := filepath.Join(dataDir, "sites", string(name))
	meta := readPostMeta(dir)
	existing, found := posts.Get(string(name))
	if meta.Member == "" {
		if !found {
			return nil
		}
		meta.Member = existing.Member
	}
	post := social.Post{
		Site: string(name), Member: meta.Member,
		Title: meta.Title, Summary: meta.Summary, Template: meta.Template,
	}
	if meta.Title == "" && found {
		post.Title = existing.Title
	}
	if meta.Summary == "" && found {
		post.Summary = existing.Summary
	}
	if meta.Template == "" && found {
		post.Template = existing.Template
	}
	if _, err := posts.Upsert(post); err != nil {
		return fmt.Errorf("index %s: %w", name, err)
	}
	return nil
}
