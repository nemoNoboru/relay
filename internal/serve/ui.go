package serve

import (
	"io/fs"
	"net/http"
	"strings"

	webui "github.com/nemoNoboru/relay/web"
)

// UIHandler serves the embedded owner UI (web/dist) under /_relay/, with a
// redirect from /_relay/ to the member home. The host console lives at
// /_relay/admin.html. The frontend must have been built (make web) before
// the binary is compiled; until then the UI files are simply absent and
// return 404.
func UIHandler() (http.Handler, error) {
	dist, err := fs.Sub(webui.FS, "dist")
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_relay" || r.URL.Path == "/_relay/" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// The people directory and each profile share one shell; the client
		// reads the id (or its absence) from the path.
		if r.URL.Path == "/_relay/people" || r.URL.Path == "/_relay/people/" || strings.HasPrefix(r.URL.Path, "/_relay/people/") {
			http.ServeFileFS(w, r, dist, "people.html")
			return
		}
		if strings.HasPrefix(r.URL.Path, "/_relay/posts/") {
			http.ServeFileFS(w, r, dist, "post.html")
			return
		}
		http.StripPrefix("/_relay/", fileServer).ServeHTTP(w, r)
	}), nil
}
