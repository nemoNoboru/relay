package serve

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFontsServesEmbedded pins that the self-hosted fonts and their
// @font-face stylesheet are served at /fonts/ with the right content types.
func TestFontsServesEmbedded(t *testing.T) {
	cases := []struct {
		path     string
		wantType string
	}{
		{path: "/fonts/fonts.css", wantType: "text/css"},
		{path: "/fonts/anton-400.woff2", wantType: "font/woff2"},
		{path: "/fonts/oswald-var.woff2", wantType: "font/woff2"},
		{path: "/fonts/space-mono-400.woff2", wantType: "font/woff2"},
		{path: "/fonts/space-mono-700.woff2", wantType: "font/woff2"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			Fonts().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, tc.wantType) {
				t.Fatalf("Content-Type = %q, want prefix %q", ct, tc.wantType)
			}
			if rec.Body.Len() == 0 {
				t.Fatalf("body for %q is empty", tc.path)
			}
		})
	}
}
