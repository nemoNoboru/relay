package serve

import (
	"io/fs"
	"net/http"

	"github.com/nemoNoboru/relay"
	webui "github.com/nemoNoboru/relay/web"
)

// landingFiles maps the node's public single-segment paths to files embedded
// in the binary. Every other path is delegated to next (the site registry).
// These paths are single-segment and can never be a valid node name; site
// requests are always two or more segments (/<node>/<site>/...).
var landingFiles = map[string]string{
	"/host":       "host.html",
	"/why":        "why.html",
	"/styles.css": "styles.css",
}

// Landing serves the node's front door: the public feed at /, the hosting
// pitch at /host, the why page at /why, and the stylesheet companion.
// Every other path is delegated to next (the site registry).
func Landing(next http.Handler) http.Handler {
	dist, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		// The dist tree is fixed at build time; this cannot fail.
		panic("relay: owner UI: " + err.Error())
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFileFS(w, r, dist, "feed.html")
			return
		}
		name, ok := landingFiles[r.URL.Path]
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, relay.SiteFS, name)
	})
}
