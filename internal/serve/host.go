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
// site folders at /<self>/<site>/.... Requests for any
// other node, for unknown sites, and for the reserved _relay prefix all
// get a plain 404.
func NodeHandler(self domain.Name, sites Sites) http.Handler {
	bySite := make(map[domain.Name]siteServer, len(sites))
	for name, root := range sites {
		bySite[name] = siteServer{root: root, handler: http.FileServer(http.FS(root))}
	}
	return siteHandler(self, func(name domain.Name) (siteServer, bool) {
		ss, ok := bySite[name]
		return ss, ok
	})
}

// siteHandler resolves a request to a site and serves it. The lookup
// function supplies the live site map, so callers can back it with a
// static map (NodeHandler) or a reloadable registry (Registry).
func siteHandler(self domain.Name, lookup func(domain.Name) (siteServer, bool)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := route.Resolve(r.URL.Path)
		if err != nil || target.Node != string(self) {
			http.NotFound(w, r)
			return
		}
		// A post is static content, not a tenant with the node's identity.
		// The sandbox gives the document an opaque origin so its scripts
		// cannot call member APIs with the viewer's cookie. Scripts, forms,
		// and popups still work; same-origin storage and XHR do not.
		w.Header().Set("Content-Security-Policy", "sandbox allow-scripts allow-popups allow-forms allow-downloads")
		ss, ok := lookup(domain.Name(target.Site))
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
