package serve

import (
	"io/fs"
	"net/http"
	"strings"

	"relay/internal/domain"
	"relay/internal/route"
)

// NodeHandler returns the node's HTTP handler. It serves this node's own
// site folders at /<self>/<site>/... (contract 5). Requests for any
// other node, for unknown sites, and for the reserved _relay prefix all
// get a plain 404.
func NodeHandler(self domain.Name, sites Sites) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := route.Resolve(r.URL.Path)
		if err != nil || target.Node != string(self) {
			http.NotFound(w, r)
			return
		}
		site, ok := sites[domain.Name(target.Site)]
		if !ok {
			http.NotFound(w, r)
			return
		}
		origPath := r.URL.Path
		if !strings.HasSuffix(origPath, "/") {
			// Directory requested without a trailing slash: redirect to
			// the slash form. Done here (not by http.FileServer) so the
			// Location carries the full request path, not the
			// site-relative one.
			isDir := target.Rest == "/"
			if !isDir {
				if info, err := fs.Stat(site, strings.TrimPrefix(target.Rest, "/")); err == nil {
					isDir = info.IsDir()
				}
			}
			if isDir {
				http.Redirect(w, r, origPath+"/", http.StatusMovedPermanently)
				return
			}
		}
		u := *r.URL
		u.Path = target.Rest
		r.URL = &u
		http.FileServer(http.FS(site)).ServeHTTP(w, r)
	})
}
