package main

import (
	"archive/tar"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestTarDir pins that a folder is archived as relative entries (regular
// files and directories), and that symlinks are rejected.
func TestTarDir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>hi</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "assets", "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	archive, err := tarDir(root)
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]string{}
	tr := tar.NewReader(bytes.NewReader(archive))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(tr)
		names[hdr.Name] = string(body)
	}
	want := map[string]string{
		"index.html":       "<h1>hi</h1>",
		"assets":           "",
		"assets/style.css": "body{}",
	}
	if len(names) != len(want) {
		t.Fatalf("archive entries = %v, want %v", names, want)
	}
	for name, body := range want {
		if names[name] != body {
			t.Fatalf("entry %q = %q, want %q", name, names[name], body)
		}
	}

	// A symlink is refused, not followed.
	linkRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(linkRoot, "index.html"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("index.html", filepath.Join(linkRoot, "link")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	if _, err := tarDir(linkRoot); err == nil {
		t.Fatalf("tarDir accepted a symlink")
	}
}

// newTestNode builds the node's mux (registry + uploader + admin) over a
// fresh data dir and returns it with a minted deploy key token.
func newTestNode(t *testing.T) (http.Handler, string) {
	t.Helper()
	data := t.TempDir()
	ks, err := serve.NewKeyStore(data)
	if err != nil {
		t.Fatal(err)
	}
	keyToken, _, err := ks.Create("alice")
	if err != nil {
		t.Fatal(err)
	}
	versions := serve.NewVersions(data)
	reg, _, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatal(err)
	}
	reload := func() error {
		_, err := reg.Reload()
		return err
	}
	up := serve.NewUploader("mynode", data, ks, versions, nil, serve.MaxDeployBytes, reload)
	admin := serve.NewAdmin("admin-secret", ks, versions, nil, nil, reload, nil, nil)

	mux := http.NewServeMux()
	mux.Handle("/_relay/v1/admin", admin)
	mux.Handle("/_relay/v1/admin/", admin)
	mux.Handle("/_relay/v1/", up)
	mux.Handle("/", reg.Handler())
	return mux, keyToken
}

// TestRunDeployAndRollback drives the real deploy/rollback commands against a
// real (in-process) node, end to end.
func TestRunDeployAndRollback(t *testing.T) {
	mux, keyToken := newTestNode(t)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	write := func(body string) {
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out bytes.Buffer

	write("<h1>v1</h1>")
	if code := runDeploy([]string{"--origin", srv.URL, "--site", "blog", "--token", keyToken, dir}, &out, io.Discard); code != 0 {
		t.Fatalf("deploy v1 exit = %d", code)
	}
	if got := fetch(t, srv.URL+"/mynode/blog/"); got != "<h1>v1</h1>" {
		t.Fatalf("served v1 = %q", got)
	}

	write("<h1>v2</h1>")
	if code := runDeploy([]string{"--origin", srv.URL, "--site", "blog", "--token", keyToken, dir}, &out, io.Discard); code != 0 {
		t.Fatalf("deploy v2 exit = %d", code)
	}
	if got := fetch(t, srv.URL+"/mynode/blog/"); got != "<h1>v2</h1>" {
		t.Fatalf("served v2 = %q", got)
	}

	if code := runRollback([]string{"--origin", srv.URL, "--site", "blog", "--admin-token", "admin-secret"}, &out, io.Discard); code != 0 {
		t.Fatalf("rollback exit = %d", code)
	}
	if got := fetch(t, srv.URL+"/mynode/blog/"); got != "<h1>v1</h1>" {
		t.Fatalf("served after rollback = %q", got)
	}
}

func fetch(t *testing.T, url string) string {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestRunDeployInsecureTLS pins that `relay deploy --insecure` works against
// an HTTPS node with a self-signed certificate.
func TestRunDeployInsecureTLS(t *testing.T) {
	mux, keyToken := newTestNode(t)
	srv := httptest.NewTLSServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>tls</h1>"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if code := runDeploy([]string{"--origin", srv.URL, "--site", "blog", "--token", keyToken, "--insecure", dir}, &out, io.Discard); code != 0 {
		t.Fatalf("deploy over TLS exit = %d", code)
	}

	// A TLS client that trusts nothing must be told to skip verification; the
	// default client would fail on the self-signed cert.
	res, err := newHTTPClient(true, time.Second).Get(srv.URL + "/mynode/blog/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if string(b) != "<h1>tls</h1>" {
		t.Fatalf("served over TLS = %q", b)
	}
}
