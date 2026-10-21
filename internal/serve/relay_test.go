package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nemoNoboru/relay/internal/social"
)

func newRelayerTest(t *testing.T) (*Relayer, *KeyStore, string, Member) {
	t.Helper()
	data := t.TempDir()
	ks, err := NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	posts, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	up := NewUploader("mynode", data, ks, NewVersions(data), posts, MaxDeployBytes, nil)
	member, _ := ks.MemberByToken(token)
	return NewRelayer("mynode", ks, up), ks, token, member
}

func doRelay(t *testing.T, rl *Relayer, method, path, bearer, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	rl.ServeHTTP(rec, req)
	return rec
}

func pollRelay(t *testing.T, rl *Relayer, token, id string) relayJob {
	t.Helper()
	rec := doRelay(t, rl, http.MethodGet, "/_relay/v1/relay/"+id, token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("poll status = %d %s", rec.Code, rec.Body)
	}
	var job relayJob
	if err := json.Unmarshal(rec.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	return job
}

func TestExtractMetaPrefersOpenGraph(t *testing.T) {
	page := `<html><head><title>Plain Title</title>
<meta name="twitter:title" content="Tweet Title">
<meta property="og:title" content="OG Title">
<meta name="description" content="Meta desc">
<meta property="og:image" content="https://x.example/og.png">
</head><body></body></html>`
	got := extractMeta(page)
	if got.title != "OG Title" || got.description != "Meta desc" || got.image != "https://x.example/og.png" {
		t.Fatalf("extractMeta = %+v", got)
	}
}

func TestExtractMetaFallsBackToTitleTag(t *testing.T) {
	got := extractMeta(`<html><head><title>A &amp; B</title></head></html>`)
	if got.title != "A & B" {
		t.Fatalf("title = %q", got.title)
	}
}

func TestExtractMetaEmptyPage(t *testing.T) {
	if got := extractMeta("nothing here"); got.title != "" || got.image != "" {
		t.Fatalf("extractMeta = %+v", got)
	}
}

func TestRelayClassify(t *testing.T) {
	cases := map[string]struct {
		kind relayKind
		ok   bool
	}{
		"photo.JPG":  {relayKindImage, true},
		"clip.webm":  {relayKindVideo, true},
		"track.m4a":  {relayKindAudio, true},
		"notes.json": {0, false},
		"index.html": {0, false},
		"thumb.webp": {relayKindImage, true},
	}
	for name, want := range cases {
		kind, ok := relayClassify(name)
		if ok != want.ok || (ok && kind != want.kind) {
			t.Fatalf("relayClassify(%q) = %v, %v; want %v, %v", name, kind, ok, want.kind, want.ok)
		}
	}
}

func TestValidateRelayURL(t *testing.T) {
	for raw, wantErr := range map[string]bool{
		"https://example.com/video": false,
		"http://example.com":        false,
		"":                          true,
		"not a url":                 true,
		"ftp://example.com":         true,
		"file:///etc/passwd":        true,
	} {
		got, err := validateRelayURL(raw)
		if (err != nil) != wantErr {
			t.Fatalf("validateRelayURL(%q) err = %v, wantErr %v", raw, err, wantErr)
		}
		if err == nil && got == "" {
			t.Fatalf("validateRelayURL(%q) returned empty", raw)
		}
	}
}

func TestRelayPages(t *testing.T) {
	media := relayMediaPage("Sunday Walk", "https://x.example/v", []relayMedia{
		{Path: "media-1.mp4", Kind: relayKindVideo},
		{Path: "media-2.webp", Kind: relayKindImage},
	})
	if !strings.Contains(media, `<video src="media-1.mp4"`) || !strings.Contains(media, `<img src="media-2.webp"`) {
		t.Fatalf("media page missing media tags:\n%s", media)
	}
	if !strings.Contains(media, "Sunday Walk") {
		t.Fatalf("media page missing title:\n%s", media)
	}

	card := relayCardPage("The Bakery", "Bread and cakes.", "preview.jpg", "https://x.example/")
	if !strings.Contains(card, `<img class="cover" src="preview.jpg"`) || !strings.Contains(card, "Bread and cakes.") {
		t.Fatalf("card page missing pieces:\n%s", card)
	}
	if !strings.Contains(card, "https://x.example/") {
		t.Fatalf("card page missing source url:\n%s", card)
	}
}

func TestRelayHostsDownloadedMedia(t *testing.T) {
	rl, _, _, member := newRelayerTest(t)
	rl.download = func(ctx context.Context, source, dir string) (string, error) {
		if err := os.WriteFile(filepath.Join(dir, "VIDEOID.mp4"), []byte("videobytes"), 0o644); err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dir, "VIDEOID.webp"), []byte("thumb"), 0o644); err != nil {
			return "", err
		}
		return "Sunday Walk", nil
	}
	rl.fetch = func(ctx context.Context, source string) (relayCard, error) {
		return relayCard{}, errors.New("unused")
	}

	job := &relayJob{ID: "rj_test", Source: "https://x.example/video", Status: "running", Created: time.Now().UTC()}
	rl.add(job)
	rl.run(job, member)

	if job.Status != "done" || job.Kind != "media" {
		t.Fatalf("job = %+v", job)
	}
	if job.URL != "/mynode/sunday-walk/" || job.PostURL != "/_relay/posts/sunday-walk" {
		t.Fatalf("job urls = %q %q", job.URL, job.PostURL)
	}
	siteDir := filepath.Join(rl.uploader.sitesDir, "sunday-walk")
	index, err := os.ReadFile(filepath.Join(siteDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), `<video src="media-1.mp4"`) || !strings.Contains(string(index), `<img src="media-2.webp"`) {
		t.Fatalf("index.html missing media tags:\n%s", index)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "media-1.mp4")); err != nil {
		t.Fatalf("video not hosted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "media-2.webp")); err != nil {
		t.Fatalf("thumbnail not hosted: %v", err)
	}
	post, ok := rl.uploader.posts.Get("sunday-walk")
	if !ok || post.Title != "Sunday Walk" || post.Image != "media-2.webp" {
		t.Fatalf("indexed post = %+v ok=%v", post, ok)
	}
}

func TestRelayFallsBackToCard(t *testing.T) {
	rl, _, token, _ := newRelayerTest(t)

	var source string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><head><title>Plain</title>
<meta property="og:title" content="The Bakery">
<meta property="og:description" content="Bread and cakes.">
<meta property="og:image" content="/img.jpg"></head><body></body></html>`)
		case "/img.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("jpegbytes"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	source = srv.URL + "/"

	rec := doRelay(t, rl, http.MethodPost, "/_relay/v1/relay", token, `{"url":"`+source+`"}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("relay = %d %s", rec.Code, rec.Body)
	}
	var job relayJob
	if err := json.Unmarshal(rec.Body.Bytes(), &job); err != nil {
		t.Fatal(err)
	}
	if job.Status != "running" || job.ID == "" {
		t.Fatalf("accepted job = %+v", job)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		got := pollRelay(t, rl, token, job.ID)
		if got.Status == "done" {
			if got.Kind != "card" || got.Title != "The Bakery" {
				t.Fatalf("done job = %+v", got)
			}
			if got.URL == "" {
				t.Fatalf("done job missing url: %+v", got)
			}
			siteDir := filepath.Join(rl.uploader.sitesDir, got.Site)
			if _, err := os.Stat(filepath.Join(siteDir, "preview.jpg")); err != nil {
				t.Fatalf("preview image not hosted: %v", err)
			}
			return
		}
		if got.Status == "failed" {
			t.Fatalf("job failed: %s", got.Error)
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for the relay job")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestRelayRequiresMember(t *testing.T) {
	rl, _, _, _ := newRelayerTest(t)
	rec := doRelay(t, rl, http.MethodPost, "/_relay/v1/relay", "", `{"url":"https://x.example"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRelayRejectsBadURL(t *testing.T) {
	rl, _, token, _ := newRelayerTest(t)
	for _, body := range []string{`{"url":"not a url"}`, `{"url":"ftp://x.example"}`} {
		rec := doRelay(t, rl, http.MethodPost, "/_relay/v1/relay", token, body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d for %s, want 400", rec.Code, body)
		}
	}
}

func TestRelayRejectsBadWebhook(t *testing.T) {
	rl, _, token, _ := newRelayerTest(t)
	rec := doRelay(t, rl, http.MethodPost, "/_relay/v1/relay", token, `{"url":"https://x.example","webhook":"ftp://hook.example"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body)
	}
}

func TestOEmbedURL(t *testing.T) {
	cases := []struct{ src, want string }{
		{"https://www.youtube.com/watch?v=abc", "https://www.youtube.com/oembed?url=" + url.QueryEscape("https://www.youtube.com/watch?v=abc") + "&format=json"},
		{"https://youtu.be/abc", "https://www.youtube.com/oembed?url=" + url.QueryEscape("https://youtu.be/abc") + "&format=json"},
		{"https://vimeo.com/123", "https://vimeo.com/api/oembed.json?url=" + url.QueryEscape("https://vimeo.com/123")},
		{"https://example.com/", ""},
	}
	for _, tc := range cases {
		if got := oembedURL(tc.src); got != tc.want {
			t.Fatalf("oembedURL(%q) = %q, want %q", tc.src, got, tc.want)
		}
	}
}

func TestFetchOEmbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"title":"You Should Check Out the Indie Web","author_name":"You've Got Kat","thumbnail_url":"https://i.ytimg.com/vi/abc/hqdefault.jpg"}`)
	}))
	defer srv.Close()
	reply, err := fetchOEmbed(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if reply.Title != "You Should Check Out the Indie Web" || reply.AuthorName != "You've Got Kat" || reply.Thumbnail == "" {
		t.Fatalf("reply = %+v", reply)
	}
}

func TestRelayDeliversWebhook(t *testing.T) {
	rl, _, _, member := newRelayerTest(t)
	var got []byte
	hook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer hook.Close()

	rl.download = func(ctx context.Context, source, dir string) (string, error) {
		return "", errors.New("no yt-dlp")
	}
	rl.fetch = func(ctx context.Context, source string) (relayCard, error) {
		return relayCard{Title: "The Bakery", Description: "Bread and cakes."}, nil
	}

	job := &relayJob{ID: "rj_hook", Source: "https://x.example/", Status: "running", Created: time.Now().UTC(), Webhook: hook.URL}
	rl.add(job)
	rl.run(job, member)

	if job.Status != "done" {
		t.Fatalf("job = %+v", job)
	}
	if len(got) == 0 {
		t.Fatal("webhook was not called")
	}
	var delivered relayJob
	if err := json.Unmarshal(got, &delivered); err != nil {
		t.Fatal(err)
	}
	if delivered.Status != "done" || delivered.Kind != "card" || delivered.Site == "" {
		t.Fatalf("delivered = %+v", delivered)
	}
	if delivered.Webhook != "" {
		t.Fatalf("webhook url leaked into the payload: %+v", delivered)
	}
}
