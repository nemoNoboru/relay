package social_test

import (
	"testing"
	"time"

	"github.com/nemoNoboru/relay/internal/social"
)

func TestStoreUpsertAndOrder(t *testing.T) {
	store, err := social.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	old := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	newer := old.Add(24 * time.Hour)
	if _, err := store.Upsert(social.Post{Site: "old", Member: "dk_1", PublishedAt: old}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert(social.Post{Site: "new", Member: "dk_2", PublishedAt: newer}); err != nil {
		t.Fatal(err)
	}
	list := store.List()
	if len(list) != 2 || list[0].Site != "new" || list[1].Site != "old" {
		t.Fatalf("list = %+v, want new then old", list)
	}
	if list[1].Title != "Old" {
		t.Fatalf("title from site = %q, want Old", list[1].Title)
	}
}

func TestStoreUpsertPreservesIdentity(t *testing.T) {
	store, err := social.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Upsert(social.Post{Site: "sunday-walk", Member: "dk_1", Title: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Upsert(social.Post{Site: "sunday-walk", Member: "dk_1", Title: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("id changed: %q -> %q", first.ID, second.ID)
	}
	if !second.PublishedAt.Equal(first.PublishedAt) {
		t.Fatalf("publishedAt changed: %v -> %v", first.PublishedAt, second.PublishedAt)
	}
	if second.Title != "Second" {
		t.Fatalf("title = %q, want Second", second.Title)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) && !second.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("updatedAt went backwards")
	}
}

func TestStoreHiddenAndDelete(t *testing.T) {
	store, err := social.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert(social.Post{Site: "docs", Member: "dk_1"}); err != nil {
		t.Fatal(err)
	}
	post, ok, err := store.SetHidden("docs", true)
	if err != nil || !ok || !post.Hidden {
		t.Fatalf("hide = %+v %v %v", post, ok, err)
	}
	if ok, err := store.Delete("docs"); err != nil || !ok {
		t.Fatalf("delete = %v %v", ok, err)
	}
	if _, ok := store.Get("docs"); ok {
		t.Fatalf("post still present after delete")
	}
}

func TestStorePersists(t *testing.T) {
	dir := t.TempDir()
	store, err := social.NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert(social.Post{Site: "docs", Member: "dk_1", Title: "Docs"}); err != nil {
		t.Fatal(err)
	}
	reloaded, err := social.NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	post, ok := reloaded.Get("docs")
	if !ok || post.Title != "Docs" || post.Member != "dk_1" {
		t.Fatalf("reloaded = %+v %v", post, ok)
	}
}
