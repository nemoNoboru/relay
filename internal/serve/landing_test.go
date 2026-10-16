package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestLandingServesPublicSite pins that the node root and its companion
// paths serve the embedded public site, with the right content types.
func TestLandingServesPublicSite(t *testing.T) {
	cases := []struct {
		path        string
		wantType    string // Content-Type prefix
		wantContent string // substring of the served body
	}{
		{path: "/", wantType: "text/html", wantContent: "Relay"},
		{path: "/manifesto.html", wantType: "text/html", wantContent: "Manifesto"},
		{path: "/technicals.html", wantType: "text/html", wantContent: "Technicals"},
		{path: "/styles.css", wantType: "text/css", wantContent: "Relay"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatalf("Landing served nothing for %q and fell through to next", r.URL.Path)
			})
			rec := httptest.NewRecorder()
			Landing(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tc.wantType) {
				t.Fatalf("Content-Type = %q, want prefix %q", ct, tc.wantType)
			}
			if !strings.Contains(rec.Body.String(), tc.wantContent) {
				t.Fatalf("body for %q does not contain %q", tc.path, tc.wantContent)
			}
		})
	}
}

// TestLandingIndexRedirect pins the stdlib canonicalisation: /index.html is
// a 301 to /, so the landing page has one canonical URL.
func TestLandingIndexRedirect(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Landing served nothing for %q and fell through to next", r.URL.Path)
	})
	rec := httptest.NewRecorder()
	Landing(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/index.html", nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if loc := rec.Header().Get("Location"); loc != "./" {
		t.Fatalf("Location = %q, want %q", loc, "./")
	}
}

// TestLandingDelegatesOtherPaths pins that every non-landing path — site
// routes and anything else — is handed to the registry, not swallowed.
func TestLandingDelegatesOtherPaths(t *testing.T) {
	for _, p := range []string{"/demo/hello/", "/demo", "/demo/hello/assets/style.css", "/_relay/node.html", "/nope"} {
		t.Run(p, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusTeapot)
			})
			rec := httptest.NewRecorder()
			Landing(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
			if !called {
				t.Fatalf("Landing(%q) did not delegate to next", p)
			}
		})
	}
}
