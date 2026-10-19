package social_test

import (
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/social"
)

func TestCommentStoreAddListDelete(t *testing.T) {
	store, err := social.NewCommentStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.Add(social.Comment{Post: "blog", Member: "dk_1", Body: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Add(social.Comment{Post: "blog", Member: "dk_2", Body: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add(social.Comment{Post: "other", Member: "dk_1", Body: "Other"}); err != nil {
		t.Fatal(err)
	}
	list := store.List("blog")
	if len(list) != 2 || list[0].ID != first.ID || list[1].ID != second.ID {
		t.Fatalf("list = %+v, want first then second", list)
	}
	if all := store.ListAll(); len(all) != 3 {
		t.Fatalf("all = %+v, want 3", all)
	}
	removed, ok, err := store.Delete(first.ID)
	if err != nil || !ok || removed.ID != first.ID {
		t.Fatalf("delete = %+v %v %v", removed, ok, err)
	}
	if list := store.List("blog"); len(list) != 1 || list[0].Body != "Second" {
		t.Fatalf("after delete = %+v", list)
	}
}

func TestCommentStoreRejectsBadBody(t *testing.T) {
	store, err := social.NewCommentStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{"", "   ", strings.Repeat("x", social.MaxCommentLength+1)} {
		if _, err := store.Add(social.Comment{Post: "blog", Member: "dk_1", Body: body}); err == nil {
			t.Fatalf("Add(%q) succeeded", body[:min(10, len(body))])
		}
	}
}

func TestCommentStorePersists(t *testing.T) {
	dir := t.TempDir()
	store, err := social.NewCommentStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	added, err := store.Add(social.Comment{Post: "blog", Member: "dk_1", Body: "Hello"})
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := social.NewCommentStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.Get(added.ID)
	if !ok || got.Body != "Hello" || got.Member != "dk_1" {
		t.Fatalf("reloaded = %+v %v", got, ok)
	}
}

func TestRenderCommentIsSafeAndSmall(t *testing.T) {
	got := social.RenderComment("Hello **bold** and *italic*.\n\nA [link](https://example.org/a?b=1&c=2).\n\n<script>alert(1)</script>")
	if !strings.Contains(got, "<p>Hello <strong>bold</strong> and <em>italic</em>.</p>") {
		t.Fatalf("render = %q", got)
	}
	if !strings.Contains(got, `<a href="https://example.org/a?b=1&amp;c=2" rel="nofollow noopener">link</a>`) {
		t.Fatalf("link render = %q", got)
	}
	if strings.Contains(got, "<script>") || !strings.Contains(got, "&lt;script&gt;") {
		t.Fatalf("script was not escaped: %q", got)
	}
	if strings.Contains(social.RenderComment("[x](//evil.example)"), "href=") {
		t.Fatalf("protocol-relative link was rendered")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
