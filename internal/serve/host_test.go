package serve_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

const indexHTML = "<!doctype html><title>hello</title>"

// newTestNode builds a handler for node "mynode" with two sites under a
// temp data dir:
//
//	<data>/sites/docs/index.html
//	<data>/sites/docs/assets/style.css
//	<data>/sites/docs/assets/logo.png
//	<data>/sites/docs/sub/notes.txt
//	<data>/sites/blog/index.html
//	<data>/topsecret.txt   (directly in the data dir, outside sites/)
func newTestNode(t *testing.T) http.Handler {
	t.Helper()
	data := t.TempDir()
	writeFile(t, filepath.Join(data, "sites", "docs", "index.html"), indexHTML)
	writeFile(t, filepath.Join(data, "sites", "docs", "assets", "style.css"), "body{}")
	writeFile(t, filepath.Join(data, "sites", "docs", "assets", "logo.png"), "PNGDATA")
	writeFile(t, filepath.Join(data, "sites", "docs", "sub", "notes.txt"), "notes")
	writeFile(t, filepath.Join(data, "sites", "blog", "index.html"), "blog home")
	writeFile(t, filepath.Join(data, "topsecret.txt"), "secret")

	sites, _, err := serve.LoadSites(data)
	if err != nil {
		t.Fatal(err)
	}
	return serve.NodeHandler("mynode", sites)
}

// AC-1: GET /mynode/docs/index.html returns 200, body byte-equal to the
// fixture, Content-Type "text/html; charset=utf-8".
func TestServesSiteFiles(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/mynode/docs/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != indexHTML {
		t.Fatalf("body = %q, want %q", body, indexHTML)
	}
	if ct := res.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", ct, "text/html; charset=utf-8")
	}
}

// AC-3: nested file, correct MIME types from extension.
func TestServesNestedFiles(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	for path, wantCT := range map[string]string{
		"/mynode/docs/assets/style.css": "text/css; charset=utf-8",
		"/mynode/docs/assets/logo.png":  "image/png",
	} {
		t.Run(path, func(t *testing.T) {
			res, err := http.Get(srv.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", res.StatusCode)
			}
			if ct := res.Header.Get("Content-Type"); ct != wantCT {
				t.Fatalf("Content-Type = %q, want %q", ct, wantCT)
			}
		})
	}
}

// AC-2: directory index. /mynode/docs/ serves docs/index.html with 200;
// /mynode/docs (no trailing slash) 301-redirects to /mynode/docs/.
func TestDirectoryIndexAndRedirect(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	res, err := client.Get(srv.URL + "/mynode/docs")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("status for /mynode/docs = %d, want 301", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "/mynode/docs/" {
		t.Fatalf("Location = %q, want /mynode/docs/", loc)
	}

	res, err = http.Get(srv.URL + "/mynode/docs/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status for /mynode/docs/ = %d, want 200", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != indexHTML {
		t.Fatalf("body = %q, want %q", body, indexHTML)
	}
}

// AC-4: missing file inside a known site is 404.
func TestMissingFile(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/mynode/docs/nope.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

// AC-5: unknown first segment and bare root are 404 (no probing, contract 5).
func TestUnknownNodeAndRoot(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	for _, path := range []string{"/", "/ghost/", "/ghost/docs/"} {
		t.Run(path, func(t *testing.T) {
			res, err := http.Get(srv.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", res.StatusCode)
			}
		})
	}
}

// AC-6: known node, unknown site is 404.
func TestUnknownSite(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/mynode/ghost/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

// AC-7: the reserved prefix is 404 now and never serves anything.
func TestReservedRelayPrefix(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	for _, path := range []string{"/_relay/", "/_relay/v1/node"} {
		t.Run(path, func(t *testing.T) {
			res, err := http.Get(srv.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", res.StatusCode)
			}
		})
	}
}

// AC-8: path traversal cannot escape the site root.
func TestNoPathTraversal(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	for _, path := range []string{
		"/mynode/docs/../../../etc/passwd",
		"/mynode/docs/%2e%2e/%2e%2e/%2e%2e/etc/passwd",
		"/mynode/docs/assets/../../../../topsecret.txt",
		"/mynode/docs/sub/../../topsecret.txt",
	} {
		t.Run(path, func(t *testing.T) {
			res, err := http.Get(srv.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", res.StatusCode)
			}
		})
	}
}

// AC-9: sites are isolated; both sites of one node are served.
func TestSiteIsolation(t *testing.T) {
	srv := httptest.NewServer(newTestNode(t))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/mynode/docs/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("docs status = %d, want 200", res.StatusCode)
	}

	res, err = http.Get(srv.URL + "/mynode/blog/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("blog status = %d, want 200", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "blog home" {
		t.Fatalf("blog body = %q, want %q", body, "blog home")
	}
}
