package serve_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nemoNoboru/relay/internal/serve"
)

// TestUIHandlerRedirectsRoot pins that /_relay/ lands on the node admin page.
func TestUIHandlerRedirectsRoot(t *testing.T) {
	ui, err := serve.UIHandler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/_relay/", nil)
	rec := httptest.NewRecorder()
	ui.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/_relay/node.html" {
		t.Fatalf("Location = %q, want /_relay/node.html", loc)
	}
}

// TestUIHandlerUnknownFile pins that unbuilt/missing UI files are a plain 404,
// never a listing or a panic.
func TestUIHandlerUnknownFile(t *testing.T) {
	ui, err := serve.UIHandler()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/_relay/does-not-exist.html", nil)
	rec := httptest.NewRecorder()
	ui.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
