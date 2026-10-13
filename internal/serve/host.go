package serve

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/nemoNoboru/relay/internal/domain"
	"github.com/nemoNoboru/relay/internal/route"
)

// siteServer pairs one site's file system with the stdlib handler that
// serves it. The handler is stateless per request, so the value is safe
// for concurrent use.
type siteServer struct {
	root    fs.FS
	handler http.Handler
}

// NodeHandler returns the node's HTTP handler. It serves this node's own
// site folders at /<self>/<site>/... (contract 5). Requests for any
// other node, for unknown sites, and for the reserved _relay prefix all
// get a plain 404.
func NodeHandler(self domain.Name, sites Sites) http.Handler {
	bySite := make(map[domain.Name]siteServer, len(sites))
	for name, root := range sites {
		bySite[name] = siteServer{root: root, handler: http.FileServer(http.FS(root))}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := route.Resolve(r.URL.Path)
		if err != nil || target.Node != string(self) {
			http.NotFound(w, r)
			return
		}
		ss, ok := bySite[domain.Name(target.Site)]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if redirectTrailingSlash(w, r, ss.root, target.Rest) {
			return
		}
		u := *r.URL
		u.Path = target.Rest
		r.URL = &u
		ss.handler.ServeHTTP(w, r)
	})
}

// redirectTrailingSlash sends a 301 to the slash form when the request
// names a directory without a trailing slash, and reports whether it
// wrote the response. The redirect is done here (not by
// http.FileServer) so the Location carries the full request path, not
// the site-relative one.
func redirectTrailingSlash(w http.ResponseWriter, r *http.Request, root fs.FS, rest string) bool {
	if strings.HasSuffix(r.URL.Path, "/") {
		return false
	}
	if rest != "/" {
		info, err := fs.Stat(root, strings.TrimPrefix(rest, "/"))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
	return true
}
