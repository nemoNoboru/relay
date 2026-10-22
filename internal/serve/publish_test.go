package serve_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

type publishReply struct {
	Post  serve.PostView `json:"post"`
	Files int            `json:"files"`
	Bytes int64          `json:"bytes"`
}

func newPublisher(t *testing.T) (http.Handler, *serve.KeyStore, *social.Store, string, serve.KeyInfo) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	posts, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	versions := serve.NewVersions(data)
	up := serve.NewUploader("mynode", data, ks, versions, posts, serve.MaxDeployBytes, nil)
	store := mustTemplateStore(t, data)
	return serve.NewPublisher("mynode", up, store, ks), ks, posts, token, key
}

func publishBody(fields map[string]any) string {
	b, _ := json.Marshal(fields)
	return string(b)
}

func TestPublishWritesPost(t *testing.T) {
	h, _, posts, token, key := newPublisher(t)
	body := publishBody(map[string]any{
		"template": "note",
		"title":    "Sunday Walk",
		"summary":  "A note",
		"files": []map[string]string{
			{"path": "index.html", "content": "<!doctype html><title>hi</title>"},
		},
	})
	rec := doMember(t, h, http.MethodPost, "/_relay/v1/publish", "", body, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish = %d %s", rec.Code, rec.Body)
	}
	var reply publishReply
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.Post.Site != "sunday-walk" || reply.Post.Template != "note" || reply.Post.Member != key.ID {
		t.Fatalf("post = %+v", reply.Post.Post)
	}
	if reply.Post.PostURL != "/_relay/posts/sunday-walk" {
		t.Fatalf("postUrl = %q", reply.Post.PostURL)
	}
	if reply.Files != 1 || reply.Bytes == 0 {
		t.Fatalf("files = %d bytes = %d", reply.Files, reply.Bytes)
	}
	indexed, ok := posts.Get("sunday-walk")
	if !ok || indexed.Title != "Sunday Walk" || indexed.Template != "note" {
		t.Fatalf("indexed = %+v %v", indexed, ok)
	}
	if _, err := os.Stat(filepath.Join(reply.Post.Post.Site, "index.html")); err == nil {
		t.Fatalf("publish wrote into the process working directory")
	}
}

func TestPublishDerivesUniqueSite(t *testing.T) {
	h, _, _, token, _ := newPublisher(t)
	makeBody := func() string {
		return publishBody(map[string]any{
			"template": "note",
			"title":    "Same Title",
			"files":    []map[string]string{{"path": "index.html", "content": "hi"}},
		})
	}
	for i, want := range []string{"same-title", "same-title-2"} {
		rec := doMember(t, h, http.MethodPost, "/_relay/v1/publish", "", makeBody(), token)
		if rec.Code != http.StatusOK {
			t.Fatalf("publish %d = %d %s", i, rec.Code, rec.Body)
		}
		var reply publishReply
		if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Post.Site != want {
			t.Fatalf("publish %d site = %q, want %q", i, reply.Post.Site, want)
		}
	}
}

func TestPublishValidation(t *testing.T) {
	h, _, _, token, _ := newPublisher(t)
	cases := []struct {
		name string
		body string
		code int
	}{
		{"unknown template", publishBody(map[string]any{"template": "nope", "files": []map[string]string{{"path": "index.html", "content": "x"}}}), http.StatusBadRequest},
		{"traversal", publishBody(map[string]any{"template": "note", "files": []map[string]string{{"path": "../evil", "content": "x"}}}), http.StatusBadRequest},
		{"no index", publishBody(map[string]any{"template": "note", "files": []map[string]string{{"path": "a.txt", "content": "x"}}}), http.StatusBadRequest},
		{"bad base64", publishBody(map[string]any{"template": "note", "files": []map[string]string{{"path": "index.html", "base64": "!!!"}}}), http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doMember(t, h, http.MethodPost, "/_relay/v1/publish", "", tc.body, token)
			if rec.Code != tc.code {
				t.Fatalf("status = %d %s, want %d", rec.Code, rec.Body, tc.code)
			}
		})
	}
}

func TestPublishRejectsRevokedMember(t *testing.T) {
	h, ks, _, token, key := newPublisher(t)
	if _, _, err := ks.Revoke(key.ID); err != nil {
		t.Fatal(err)
	}
	body := publishBody(map[string]any{
		"template": "note",
		"files":    []map[string]string{{"path": "index.html", "content": "x"}},
	})
	if rec := doMember(t, h, http.MethodPost, "/_relay/v1/publish", "", body, token); rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked publish = %d, want 401", rec.Code)
	}
}
