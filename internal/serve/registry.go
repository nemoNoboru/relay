package serve

import (
	"net/http"
	"sync"

	"github.com/nemoNoboru/relay/internal/domain"
)

// Registry serves the sites under a data directory and can be reloaded
// after a deploy, so a newly uploaded site (or changed files) is live
// without restarting the node.
type Registry struct {
	self    domain.Name
	dataDir string

	mu    sync.RWMutex
	sites map[domain.Name]siteServer
}

// NewRegistry loads the sites from <dataDir>/sites and returns a registry
// that serves them for node self, plus the names LoadSites skipped.
func NewRegistry(self domain.Name, dataDir string) (*Registry, []string, error) {
	r := &Registry{self: self, dataDir: dataDir}
	skipped, err := r.Reload()
	if err != nil {
		return nil, nil, err
	}
	return r, skipped, nil
}

// Reload re-reads the sites from disk. It returns the names that were
// skipped (invalid names or non-directories), as LoadSites does.
func (r *Registry) Reload() ([]string, error) {
	sites, skipped, err := LoadSites(r.dataDir)
	if err != nil {
		return nil, err
	}
	bySite := make(map[domain.Name]siteServer, len(sites))
	for name, root := range sites {
		bySite[name] = siteServer{root: root, handler: http.FileServer(http.FS(root))}
	}
	r.mu.Lock()
	r.sites = bySite
	r.mu.Unlock()
	return skipped, nil
}

// Handler serves the current sites at /<self>/<site>/...
func (r *Registry) Handler() http.Handler {
	return siteHandler(r.self, func(name domain.Name) (siteServer, bool) {
		r.mu.RLock()
		ss, ok := r.sites[name]
		r.mu.RUnlock()
		return ss, ok
	})
}
