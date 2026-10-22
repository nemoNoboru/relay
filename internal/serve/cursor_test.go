package serve

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nemoNoboru/relay/internal/social"
)

func TestCursorRoundTrip(t *testing.T) {
	c := pageCursor{At: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC), Site: "blog"}
	enc := encodeCursor(c)
	dec, ok := decodeCursor(enc)
	if !ok || !dec.At.Equal(c.At) || dec.Site != c.Site {
		t.Fatalf("round trip = %+v %v, want %+v", dec, ok, c)
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	for _, raw := range []string{"", "not-base64!", "bm8tbnVsbA"} {
		if _, ok := decodeCursor(raw); ok {
			t.Fatalf("decodeCursor(%q) succeeded", raw)
		}
	}
}

func TestAfterCursor(t *testing.T) {
	c := pageCursor{At: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), Site: "blog"}
	before := social.Post{PublishedAt: c.At.Add(-time.Hour), Site: "a"}
	equalEarlierSite := social.Post{PublishedAt: c.At, Site: "a"}
	equalLaterSite := social.Post{PublishedAt: c.At, Site: "z"}
	after := social.Post{PublishedAt: c.At.Add(time.Hour), Site: "a"}
	if !afterCursor(before, c) || afterCursor(equalEarlierSite, c) || !afterCursor(equalLaterSite, c) || afterCursor(after, c) {
		t.Fatalf("afterCursor misordered: %v %v %v %v", afterCursor(before, c), afterCursor(equalEarlierSite, c), afterCursor(equalLaterSite, c), afterCursor(after, c))
	}
}

func TestPagePaginatesAndReturnsCursor(t *testing.T) {
	p := &Posts{}
	all := []social.Post{
		{Site: "one", PublishedAt: time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)},
		{Site: "two", PublishedAt: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)},
		{Site: "three", PublishedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)},
	}
	req := httptest.NewRequest("GET", "/feed?limit=2", nil)
	page, next := p.page(req, all, func(social.Post) bool { return true })
	if len(page) != 2 || next == "" {
		t.Fatalf("page = %d posts, next = %q", len(page), next)
	}
	req = httptest.NewRequest("GET", "/feed?limit=2&cursor="+next, nil)
	page, next = p.page(req, all, func(social.Post) bool { return true })
	if len(page) != 1 || next != "" {
		t.Fatalf("second page = %d posts, next = %q", len(page), next)
	}
}
