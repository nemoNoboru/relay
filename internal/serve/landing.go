package serve

import (
	"net/http"

	"github.com/nemoNoboru/relay"
)

// landingFiles maps the node's public root paths to the files embedded in
// the binary. Every other path is delegated to next (the site registry).
// The root pages are single-segment file paths with dots (index.html,
// styles.css), which can never be a valid node name; site requests are
// always two or more segments (/<node>/<site>/...), so the two never
// overlap.
var landingFiles = map[string]string{
	"/":                "index.html",
	"/index.html":      "index.html",
	"/technicals.html": "technicals.html",
	"/styles.css":      "styles.css",
}

// Landing serves the embedded public site (landing page, technicals,
// stylesheet) at the node root, and delegates every other path to next.
func Landing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := landingFiles[r.URL.Path]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, relay.SiteFS, name)
	})
}
