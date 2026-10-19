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
			http.Redirect(w, r, "/_relay/home.html", http.StatusFound)
			return
		}
		// Profile and post wrappers are single shells whose clients read the
		// id/site from the path. Any path under the prefix gets the shell.
		if strings.HasPrefix(r.URL.Path, "/_relay/people/") {
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
