package serve_test

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
	"github.com/nemoNoboru/relay/internal/social"
)

// tarSite builds an in-memory tar archive of the given regular files.
func tarSite(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		body := files[n]
		if err := tw.WriteHeader(&tar.Header{Name: n, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func doDeploy(t *testing.T, up http.Handler, path, auth string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}
	req.Header.Set("Content-Type", "application/x-tar")
	rec := httptest.NewRecorder()
	up.ServeHTTP(rec, req)
	return rec
}

// testUploader returns an uploader backed by a fresh key store with one
// deploy key minted for user "tester", plus that key's token.
func testUploader(t *testing.T, data string, maxBytes int64, reload func() error) (*serve.Uploader, string) {
	t.Helper()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("tester")
	if err != nil {
		t.Fatal(err)
	}
	return serve.NewUploader("mynode", data, ks, serve.NewVersions(data), nil, maxBytes, reload), token
}

// TestUploadRequiresToken pins that uploads are gated by a deploy key.
func TestUploadRequiresToken(t *testing.T) {
	up, _ := testUploader(t, t.TempDir(), serve.MaxDeployBytes, nil)
	body := tarSite(t, map[string]string{"index.html": "hi"})

	for name, auth := range map[string]string{"missing": "", "wrong": "nope"} {
		t.Run(name, func(t *testing.T) {
			rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", auth, body)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

// TestUploadRejectsRevokedKey pins that a revoked deploy key stops working.
func TestUploadRejectsRevokedKey(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, key, err := ks.Create("tester")
	if err != nil {
		t.Fatal(err)
	}
	up := serve.NewUploader("mynode", data, ks, serve.NewVersions(data), nil, serve.MaxDeployBytes, nil)
	body := tarSite(t, map[string]string{"index.html": "hi"})

	if rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, body); rec.Code != http.StatusOK {
		t.Fatalf("deploy before revoke = %d", rec.Code)
	}
	if _, _, err := ks.Revoke(key.ID); err != nil {
		t.Fatal(err)
	}
	if rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, body); rec.Code != http.StatusUnauthorized {
		t.Fatalf("deploy after revoke = %d, want 401", rec.Code)
	}
}

// TestUploadAndServe is the end-to-end acceptance: upload a folder, then
// fetch it live at /<node>/<site>/.
func TestUploadAndServe(t *testing.T) {
	data := t.TempDir()
	reg, _, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatal(err)
	}
	up, token := testUploader(t, data, serve.MaxDeployBytes, func() error {
		_, err := reg.Reload()
		return err
	})

	body := tarSite(t, map[string]string{"index.html": "<h1>hello</h1>"})
	rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("deploy status = %d, body %s", rec.Code, rec.Body.String())
	}

	srv := httptest.NewServer(reg.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + "/mynode/docs/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, _ := io.ReadAll(res.Body)
	if string(got) != "<h1>hello</h1>" {
		t.Fatalf("served body = %q", got)
	}
}

// TestUploadRecordsSiteOwner pins that a deploy attributes the site to the
// user whose key uploaded it.
func TestUploadRecordsSiteOwner(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	up := serve.NewUploader("mynode", data, ks, serve.NewVersions(data), nil, serve.MaxDeployBytes, nil)

	rec := doDeploy(t, up, "/_relay/v1/sites/blog/deploy", token, tarSite(t, map[string]string{"index.html": "hi"}))
	if rec.Code != http.StatusOK {
		t.Fatalf("deploy = %d", rec.Code)
	}
	ov := ks.Overview()
	if len(ov.Sites) != 1 || ov.Sites[0].Site != "blog" || ov.Sites[0].Label != "alice" {
		t.Fatalf("sites = %+v, want blog->alice", ov.Sites)
	}
}

// TestUploadReplacesLiveSite pins that a second deploy replaces the first.
func TestUploadReplacesLiveSite(t *testing.T) {
	data := t.TempDir()
	reg, _, _ := serve.NewRegistry("mynode", data)
	up, token := testUploader(t, data, serve.MaxDeployBytes, func() error {
		_, err := reg.Reload()
		return err
	})

	if rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, tarSite(t, map[string]string{"index.html": "v1"})); rec.Code != http.StatusOK {
		t.Fatalf("first deploy = %d", rec.Code)
	}
	if rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, tarSite(t, map[string]string{"index.html": "v2", "style.css": "body{}"})); rec.Code != http.StatusOK {
		t.Fatalf("second deploy = %d", rec.Code)
	}

	srv := httptest.NewServer(reg.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + "/mynode/docs/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, _ := io.ReadAll(res.Body)
	if string(got) != "v2" {
		t.Fatalf("served body = %q, want v2", got)
	}
}

// TestUploadBadArchive pins the 400 on non-tar bodies.
func TestUploadBadArchive(t *testing.T) {
	up, token := testUploader(t, t.TempDir(), serve.MaxDeployBytes, nil)
	rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, []byte("not a tar"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %s)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bad_archive") {
		t.Fatalf("body = %s, want bad_archive code", rec.Body.String())
	}
}

// TestUploadRejectsEscape pins that traversal and absolute entries are 400
// and leave an existing site untouched.
func TestUploadRejectsEscape(t *testing.T) {
	data := t.TempDir()
	reg, _, _ := serve.NewRegistry("mynode", data)
	up, token := testUploader(t, data, serve.MaxDeployBytes, func() error {
		_, err := reg.Reload()
		return err
	})

	if rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, tarSite(t, map[string]string{"index.html": "safe"})); rec.Code != http.StatusOK {
		t.Fatalf("seed deploy = %d", rec.Code)
	}

	for _, entry := range []tar.Header{
		{Name: "../evil.txt", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg},
		{Name: "/etc/passwd", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg},
		{Name: `..\evil.txt`, Mode: 0o644, Size: 1, Typeflag: tar.TypeReg},
		{Name: "link", Mode: 0o777, Typeflag: tar.TypeSymlink, Linkname: "../../secret"},
	} {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		if err := tw.WriteHeader(&entry); err != nil {
			t.Fatal(err)
		}
		if entry.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte("x")); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, buf.Bytes())
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("entry %q status = %d, want 400", entry.Name, rec.Code)
		}
	}

	// The original site is still the live one.
	srv := httptest.NewServer(reg.Handler())
	defer srv.Close()
	res, err := http.Get(srv.URL + "/mynode/docs/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, _ := io.ReadAll(res.Body)
	if string(got) != "safe" {
		t.Fatalf("served body = %q, want safe (failed deploys must not clobber)", got)
	}
}

// TestUploadTooLarge pins the 413 on archives over the configured cap.
func TestUploadTooLarge(t *testing.T) {
	up, token := testUploader(t, t.TempDir(), 16, nil)
	rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, tarSite(t, map[string]string{"index.html": "this body is longer than sixteen bytes"}))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413 (body %s)", rec.Code, rec.Body.String())
	}
}

// TestUploadBadNameAndRoute pins 400/404 for malformed paths.
func TestUploadBadNameAndRoute(t *testing.T) {
	up, token := testUploader(t, t.TempDir(), serve.MaxDeployBytes, nil)
	body := tarSite(t, map[string]string{"index.html": "hi"})

	for _, path := range []string{
		"/_relay/v1/sites/Bad_Name/deploy",
		"/_relay/v1/sites/docs/other",
		"/_relay/v1/other",
	} {
		t.Run(path, func(t *testing.T) {
			rec := doDeploy(t, up, path, token, body)
			if rec.Code != http.StatusBadRequest && rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 400 or 404", rec.Code)
			}
		})
	}
}

// TestUploadWritesSiteDirOnDisk pins the on-disk layout the serving side reads.
func TestUploadWritesSiteDirOnDisk(t *testing.T) {
	data := t.TempDir()
	up, token := testUploader(t, data, serve.MaxDeployBytes, nil)
	rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, tarSite(t, map[string]string{"assets/logo.svg": "<svg/>"}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	got, err := os.ReadFile(filepath.Join(data, "sites", "docs", "assets", "logo.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "<svg/>" {
		t.Fatalf("on-disk body = %q", got)
	}
}

// TestUploadAcceptsArchiveRootEntry pins that `tar -c .` (which emits a "."
// directory entry first) is accepted — the common way a folder is archived.
func TestUploadAcceptsArchiveRootEntry(t *testing.T) {
	data := t.TempDir()
	up, token := testUploader(t, data, serve.MaxDeployBytes, nil)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range []tar.Header{
		{Name: ".", Mode: 0o755, Typeflag: tar.TypeDir},
		{Name: "index.html", Mode: 0o644, Size: 2, Typeflag: tar.TypeReg},
	} {
		if err := tw.WriteHeader(&e); err != nil {
			t.Fatal(err)
		}
		if e.Typeflag == tar.TypeReg {
			if _, err := tw.Write([]byte("ok")); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	rec := doDeploy(t, up, "/_relay/v1/sites/docs/deploy", token, buf.Bytes())
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(data, "sites", "docs", "index.html")); err != nil {
		t.Fatalf("index.html not written: %v", err)
	}
}

// TestUploadIndexesPost pins that a deploy creates a post, writes the
// portable sidecar, and updates the same post on redeploy.
func TestUploadIndexesPost(t *testing.T) {
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	posts, err := social.NewStore(data)
	if err != nil {
		t.Fatal(err)
	}
	up := serve.NewUploader("mynode", data, ks, serve.NewVersions(data), posts, serve.MaxDeployBytes, nil)

	rec := doDeploy(t, up, "/_relay/v1/sites/sunday-walk/deploy", key.Token, tarSite(t, map[string]string{"index.html": "v1"}))
	if rec.Code != http.StatusOK {
		t.Fatalf("deploy = %d %s", rec.Code, rec.Body)
	}
	var reply struct {
		PostURL string `json:"postUrl"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	if reply.PostURL != "/_relay/posts/sunday-walk" {
		t.Fatalf("postUrl = %q", reply.PostURL)
	}
	first, ok := posts.Get("sunday-walk")
	if !ok || first.Member != key.ID || first.Title != "Sunday Walk" || first.Template != "custom" {
		t.Fatalf("indexed post = %+v %v", first, ok)
	}
	if _, err := os.Stat(filepath.Join(data, "sites", "sunday-walk", "relay-post.json")); err != nil {
		t.Fatalf("sidecar: %v", err)
	}

	// A sidecar in the upload supplies the title/summary/template.
	body := tarSite(t, map[string]string{
		"index.html":      "v2",
		"relay-post.json": `{"title":"A walk","summary":"Six photos","template":"gallery"}`,
	})
	rec = doDeploy(t, up, "/_relay/v1/sites/sunday-walk/deploy", key.Token, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("redeploy = %d %s", rec.Code, rec.Body)
	}
	second, ok := posts.Get("sunday-walk")
	if !ok || second.ID != first.ID || !second.PublishedAt.Equal(first.PublishedAt) {
		t.Fatalf("identity changed: %+v -> %+v", first, second)
	}
	if second.Title != "A walk" || second.Summary != "Six photos" || second.Template != "gallery" {
		t.Fatalf("updated post = %+v", second)
	}
}
