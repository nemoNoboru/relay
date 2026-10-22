package serve_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

type feedReply struct {
	Posts []serve.PostView `json:"posts"`
	Next  string           `json:"next"`
}

type profileReply struct {
	Member serve.Member     `json:"member"`
	Posts  []serve.PostView `json:"posts"`
	Total  int              `json:"total"`
}

type postReply struct {
	Post serve.PostView `json:"post"`
}

func newPostsHandler(t *testing.T) (http.Handler, *serve.KeyStore, *social.Store, *social.CommentStore, string, serve.KeyInfo, string, serve.KeyInfo) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	store, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	comments, err := social.NewCommentStore(data)
	if err != nil {
		t.Fatal(err)
	}
	at, ak, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	bt, bk, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}
	return serve.NewPosts("mynode", ks, store, comments, mustTemplateStore(t, data), serve.NewVersions(data), nil), ks, store, comments, at, ak, bt, bk
}

func TestFeedIsChronologicalAndPublic(t *testing.T) {
	h, _, store, _, at, ak, _, bk := newPostsHandler(t)
	old := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	if _, err := store.Upsert(social.Post{Site: "older", Member: ak.ID, Title: "Older", PublishedAt: old}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert(social.Post{Site: "newer", Member: bk.ID, Title: "Newer", PublishedAt: old.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	if rec := doMember(t, h, http.MethodGet, "/_relay/v1/feed", "", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("unauthenticated feed = %d, want 200", rec.Code)
	}
	rec := doMember(t, h, http.MethodGet, "/_relay/v1/feed", "", "", at)
	if rec.Code != http.StatusOK {
		t.Fatalf("feed = %d %s", rec.Code, rec.Body)
	}
	var got feedReply
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Posts) != 2 || got.Posts[0].Site != "newer" || got.Posts[1].Site != "older" {
		t.Fatalf("feed = %+v, want newer then older", got.Posts)
	}
	if got.Posts[0].Author.DisplayName != "bob" {
		t.Fatalf("author = %+v", got.Posts[0].Author)
	}
	if got.Posts[0].URL != "/mynode/newer/" || got.Posts[0].PostURL != "/_relay/posts/newer" {
		t.Fatalf("urls = %+v", got.Posts[0])
	}

	// A hidden post disappears from the feed but the static page is untouched.
	if _, _, err := store.SetHidden("newer", true); err != nil {
		t.Fatal(err)
	}
	rec = doMember(t, h, http.MethodGet, "/_relay/v1/feed", "", "", at)
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Posts) != 1 || got.Posts[0].Site != "older" {
		t.Fatalf("hidden feed = %+v", got.Posts)
	}
}

func TestProfileListsOnlyThatMember(t *testing.T) {
	h, _, store, _, at, ak, _, bk := newPostsHandler(t)
	if _, err := store.Upsert(social.Post{Site: "alice-post", Member: ak.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Upsert(social.Post{Site: "bob-post", Member: bk.ID}); err != nil {
		t.Fatal(err)
	}
	rec := doMember(t, h, http.MethodGet, "/_relay/v1/people/"+ak.ID, "", "", at)
	if rec.Code != http.StatusOK {
		t.Fatalf("profile = %d %s", rec.Code, rec.Body)
	}
	var got profileReply
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Member.ID != ak.ID {
		t.Fatalf("member = %+v", got.Member)
	}
	if len(got.Posts) != 1 || got.Posts[0].Site != "alice-post" || got.Total != 1 {
		t.Fatalf("posts = %+v total=%d, want only alice-post", got.Posts, got.Total)
	}
}

func TestPostLookup(t *testing.T) {
	h, _, store, _, at, ak, _, _ := newPostsHandler(t)
	if _, err := store.Upsert(social.Post{Site: "sunday-walk", Member: ak.ID, Template: "gallery"}); err != nil {
		t.Fatal(err)
	}
	rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/sunday-walk", "", "", at)
	if rec.Code != http.StatusOK {
		t.Fatalf("post = %d %s", rec.Code, rec.Body)
	}
	var got postReply
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Post.Site != "sunday-walk" || got.Post.Author.DisplayName != "alice" {
		t.Fatalf("post = %+v", got.Post)
	}
	if rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/nope", "", "", at); rec.Code != http.StatusNotFound {
		t.Fatalf("missing post = %d, want 404", rec.Code)
	}
}

// TestIndexSiteReadsSidecar pins the rollback reindex path.
func TestIndexSiteReadsSidecar(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	store, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Upsert(social.Post{Site: "blog", Member: "dk_1", Title: "Before"})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(data, "sites", "blog")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "relay-post.json"), []byte(`{"title":"After","summary":"sidecar","template":"note","member":"dk_2"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := serve.IndexSite(store, ks, data, "blog"); err != nil {
		t.Fatal(err)
	}
	post, ok := store.Get("blog")
	if !ok || post.ID != first.ID || !post.PublishedAt.Equal(first.PublishedAt) {
		t.Fatalf("identity changed: %+v -> %+v", first, post)
	}
	if post.Title != "After" || post.Summary != "sidecar" || post.Template != "note" || post.Member != "dk_2" {
		t.Fatalf("reindexed = %+v", post)
	}
}

func TestFeedPaginatesAndCarriesComments(t *testing.T) {
	h, _, store, comments, at, ak, _, _ := newPostsHandler(t)
	base := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	for i, site := range []string{"one", "two", "three"} {
		if _, err := store.Upsert(social.Post{
			Site: site, Member: ak.ID, Title: site,
			PublishedAt: base.Add(time.Duration(i) * time.Hour),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := comments.Add(social.Comment{Post: "three", Member: ak.ID, Body: "A comment"}); err != nil {
		t.Fatal(err)
	}
	rec := doMember(t, h, http.MethodGet, "/_relay/v1/feed?limit=2", "", "", at)
	if rec.Code != http.StatusOK {
		t.Fatalf("feed = %d %s", rec.Code, rec.Body)
	}
	var page feedReply
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Posts) != 2 || page.Posts[0].Site != "three" || page.Next == "" {
		t.Fatalf("first page = %+v next=%q", page.Posts, page.Next)
	}
	if len(page.Posts[0].Comments) != 1 || page.Posts[0].Comments[0].Body != "A comment" {
		t.Fatalf("comments = %+v", page.Posts[0].Comments)
	}
	rec = doMember(t, h, http.MethodGet, "/_relay/v1/feed?limit=2&cursor="+url.QueryEscape(page.Next), "", "", at)
	page = feedReply{}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Posts) != 1 || page.Posts[0].Site != "one" || page.Next != "" {
		t.Fatalf("second page = %+v next=%q", page.Posts, page.Next)
	}
}

func TestPostCarriesCoverImage(t *testing.T) {
	h, _, store, _, at, ak, _, _ := newPostsHandler(t)
	if _, err := store.Upsert(social.Post{Site: "cover", Member: ak.ID, Image: "assets/cover.svg"}); err != nil {
		t.Fatal(err)
	}
	rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/cover", "", "", at)
	var reply postReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Post.ImageURL != "/mynode/cover/assets/cover.svg" {
		t.Fatalf("imageUrl = %q", reply.Post.ImageURL)
	}
}

func TestMemberDeletesOwnPost(t *testing.T) {
	h, _, store, _, at, ak, bt, _ := newPostsHandler(t)
	if _, err := store.Upsert(social.Post{Site: "mine", Member: ak.ID, Title: "Mine"}); err != nil {
		t.Fatal(err)
	}

	// The owner can delete their own post.
	req := httptest.NewRequest(http.MethodDelete, "/_relay/v1/posts/mine", nil)
	req.Header.Set("Authorization", "Bearer "+at)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("own delete = %d %s", rec.Code, rec.Body.String())
	}
	if _, ok := store.Get("mine"); ok {
		t.Fatalf("post still indexed after own delete")
	}

	// Another member cannot delete it.
	if _, err := store.Upsert(social.Post{Site: "alices", Member: ak.ID, Title: "Alices"}); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodDelete, "/_relay/v1/posts/alices", nil)
	req.Header.Set("Authorization", "Bearer "+bt)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("foreign delete = %d, want 403", rec.Code)
	}

	// A guest cannot delete at all.
	req = httptest.NewRequest(http.MethodDelete, "/_relay/v1/posts/alices", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anon delete = %d, want 401", rec.Code)
	}
}
