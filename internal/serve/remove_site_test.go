package serve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/social"
)

func TestRemoveSiteAbortsBeforeFilesWhenIndexDeleteFails(t *testing.T) {
	data := t.TempDir()
	ks, err := NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	posts, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	comments, err := social.NewCommentStore(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := posts.Upsert(social.Post{Site: "blog", Member: "m"}); err != nil {
		t.Fatal(err)
	}
	if err := ks.RecordSite("blog", "alice"); err != nil {
		t.Fatal(err)
	}

	siteDir := filepath.Join(data, "sites", "blog")
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "index.html"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Sabotage the index so posts.Delete fails after the bookkeeping starts.
	if err := os.Remove(filepath.Join(data, "posts.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(data, "posts.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := removeSite(domain.Name("blog"), NewVersions(data), posts, comments, ks, nil); err == nil {
		t.Fatal("removeSite succeeded despite index failure")
	}
	if _, err := os.Stat(siteDir); err != nil {
		t.Fatalf("site files were removed despite index failure: %v", err)
	}
}
