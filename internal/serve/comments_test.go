package serve_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

type commentsReply struct {
	Comments []serve.CommentView `json:"comments"`
}

type commentReply struct {
	Comment serve.CommentView `json:"comment"`
}

func newCommentsHandler(t *testing.T) (http.Handler, *serve.KeyStore, *social.CommentStore, string, serve.KeyInfo, string, serve.KeyInfo) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
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
	at, ak, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	bt, bk, err := ks.Create("bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := posts.Upsert(social.Post{Site: "blog", Member: ak.ID}); err != nil {
		t.Fatal(err)
	}
	return serve.NewComments(ks, posts, comments), ks, comments, at, ak, bt, bk
}

func TestCommentsAddAndList(t *testing.T) {
	h, _, _, at, _, _, _ := newCommentsHandler(t)
	if rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/blog/comments", "", "", ""); rec.Code != http.StatusOK {
		t.Fatalf("anonymous comments = %d, want 200", rec.Code)
	}
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/posts/blog/comments", "", `{"body":"Hello **friends**"}`, at)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add comment = %d %s", rec.Code, rec.Body)
	}
	var added commentReply
	if err := json.Unmarshal(rec.Body.Bytes(), &added); err != nil {
		t.Fatal(err)
	}
	if !added.Comment.Mine || added.Comment.Author.DisplayName != "alice" {
		t.Fatalf("added = %+v", added.Comment)
	}
	if added.Comment.HTML != "<p>Hello <strong>friends</strong></p>" {
		t.Fatalf("html = %q", added.Comment.HTML)
	}

	rec = doMember(t, h, http.MethodGet, "/_relay/v1/posts/blog/comments", "", "", at)
	if rec.Code != http.StatusOK {
		t.Fatalf("list comments = %d %s", rec.Code, rec.Body)
	}
	var list commentsReply
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Comments) != 1 || list.Comments[0].ID != added.Comment.ID {
		t.Fatalf("list = %+v", list.Comments)
	}
}

func TestCommentsDeleteOwnOnly(t *testing.T) {
	h, _, _, at, _, bt, _ := newCommentsHandler(t)
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/posts/blog/comments", "", `{"body":"Mine"}`, at)
	var added commentReply
	if err := json.Unmarshal(rec.Body.Bytes(), &added); err != nil {
		t.Fatal(err)
	}
	rec = doMember(t, h, http.MethodDelete, "/_relay/v1/comments/"+added.Comment.ID, "", "", bt)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("other delete = %d, want 403", rec.Code)
	}
	rec = doMember(t, h, http.MethodDelete, "/_relay/v1/comments/"+added.Comment.ID, "", "", at)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("own delete = %d, want 204", rec.Code)
	}
	if rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/blog/comments", "", "", at); rec.Code != http.StatusOK {
		t.Fatalf("list after delete = %d", rec.Code)
	}
}

func TestCommentsRateLimit(t *testing.T) {
	h, _, _, at, _, _, _ := newCommentsHandler(t)
	if rec := doMember(t, h, http.MethodPost, "/_relay/v1/posts/blog/comments", "", `{"body":"one"}`, at); rec.Code != http.StatusCreated {
		t.Fatalf("first comment = %d", rec.Code)
	}
	if rec := doMember(t, h, http.MethodPost, "/_relay/v1/posts/blog/comments", "", `{"body":"two"}`, at); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second comment = %d, want 429", rec.Code)
	}
}

func TestCommentsUnknownPost(t *testing.T) {
	h, _, _, at, _, _, _ := newCommentsHandler(t)
	if rec := doMember(t, h, http.MethodGet, "/_relay/v1/posts/nope/comments", "", "", at); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown post comments = %d, want 404", rec.Code)
	}
}

func TestAdminDeletesAnyCommentAndOverviewLists(t *testing.T) {
	h, ks, comments, at, _, _, _ := newCommentsHandler(t)
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/posts/blog/comments", "", `{"body":"Host can remove this"}`, at)
	var added commentReply
	if err := json.Unmarshal(rec.Body.Bytes(), &added); err != nil {
		t.Fatal(err)
	}
	admin := serve.NewAdmin(adminToken, ks, serve.NewVersions(t.TempDir()), nil, comments, nil, nil, nil)
	rec = doJSON(t, admin, http.MethodGet, "/_relay/v1/admin", adminToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("overview = %d %s", rec.Code, rec.Body)
	}
	var ov struct {
		Comments []serve.CommentView `json:"comments"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &ov); err != nil {
		t.Fatal(err)
	}
	if len(ov.Comments) != 1 || ov.Comments[0].ID != added.Comment.ID {
		t.Fatalf("overview comments = %+v", ov.Comments)
	}
	rec = doJSON(t, admin, http.MethodDelete, "/_relay/v1/admin/comments/"+added.Comment.ID, adminToken, "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("admin delete = %d %s", rec.Code, rec.Body)
	}
	if _, ok := comments.Get(added.Comment.ID); ok {
		t.Fatalf("comment still present after admin delete")
	}
}
