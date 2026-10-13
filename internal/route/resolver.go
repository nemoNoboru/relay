// Package route resolves request paths to a serving target: which node,
// which site, and which file path within that site (contract 5).
package route

import (
	"errors"
	"path"
	"strings"

	"relay/internal/domain"
)

// Target is where a request is served from: a file path (Rest) within a
// site on a named node. Rest is always rooted ("/" for the site root).
type Target struct {
	Node string
	Site string
	Rest string
}

var (
	// ErrNoNode is returned when the path has no first segment, or when
	// the first segment is not a valid name. Invalid is unknown, never a
	// different node, so no probing happens (contract 5).
	ErrNoNode = errors.New("route: no node in path")
	// ErrNoSite is returned when the path names a valid node but has no
	// site segment.
	ErrNoSite = errors.New("route: no site in path")
	// ErrReserved is returned when the first segment is the reserved
	// control prefix. It is domain.ErrReservedName.
	ErrReserved = domain.ErrReservedName
)

// Resolve parses an absolute request path (r.URL.Path) into its node,
// site, and the rooted file path within the site. The path is cleaned
// before segmentation, so "//" collapses and route-level ".." cannot
// change the node or site; the file layer is the second guard.
func Resolve(requestPath string) (*Target, error) {
	clean := path.Clean(requestPath)
	segs := strings.Split(clean, "/") // clean is absolute, so segs[0] == ""
	if len(segs) < 2 || segs[1] == "" {
		return nil, ErrNoNode
	}
	node, err := domain.ParseName(segs[1])
	if err != nil {
		if errors.Is(err, domain.ErrReservedName) {
			return nil, ErrReserved
		}
		return nil, ErrNoNode
	}
	if len(segs) < 3 || segs[2] == "" {
		return nil, ErrNoSite
	}
	site, err := domain.ParseName(segs[2])
	if err != nil {
		return nil, ErrNoSite
	}
	rest := "/" + strings.Join(segs[3:], "/")
	if rest != "/" && strings.HasSuffix(requestPath, "/") {
		rest += "/"
	}
	return &Target{Node: string(node), Site: string(site), Rest: rest}, nil
}
