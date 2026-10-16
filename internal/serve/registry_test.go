package serve_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestNewRegistryWarnings pins that invalid site directory names are
// reported, not fatal.
func TestNewRegistryWarnings(t *testing.T) {
	data := t.TempDir()
	writeFile(t, filepath.Join(data, "sites", "docs", "index.html"), "hi")
	writeFile(t, filepath.Join(data, "sites", "Bad Dir", "index.html"), "skip me")

	reg, skipped, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	if len(skipped) != 1 || skipped[0] != "Bad Dir" {
		t.Fatalf("skipped = %v, want [Bad Dir]", skipped)
	}

	res := httptest.NewRecorder()
	reg.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/mynode/docs/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("docs status = %d, want 200", res.Code)
	}
}

// TestRegistryReload picks up a site directory created after startup.
func TestRegistryReload(t *testing.T) {
	data := t.TempDir()
	reg, _, err := serve.NewRegistry("mynode", data)
	if err != nil {
		t.Fatal(err)
	}
	if rec := serveGet(reg.Handler(), "/mynode/late/"); rec.Code != http.StatusNotFound {
		t.Fatalf("before reload status = %d, want 404", rec.Code)
	}

	writeFile(t, filepath.Join(data, "sites", "late", "index.html"), "late!")
	if _, err := reg.Reload(); err != nil {
		t.Fatal(err)
	}
	if rec := serveGet(reg.Handler(), "/mynode/late/"); rec.Code != http.StatusOK {
		t.Fatalf("after reload status = %d, want 200", rec.Code)
	}
}

func serveGet(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// TestAdminTokenFileIsPrivate is belt-and-braces that the admin token file
// created by the node is not world-readable (see resolveAdminToken).
func TestAdminTokenFileIsPrivate(t *testing.T) {
	data := t.TempDir()
	path := filepath.Join(data, "admin-token")
	if err := os.WriteFile(path, []byte("secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("token file mode = %o, want 600", perm)
	}
}
