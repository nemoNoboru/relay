// Package prefix holds Relay's node addressing model: the path a node
// registers under a public host and is routed to. It is pure domain
// logic: no I/O.
package prefix

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalid is returned when a prefix does not satisfy the node
// addressing rules.
var ErrInvalid = errors.New("invalid node prefix")

// maxNodeLen is the longest node name, mirroring a DNS label.
const maxNodeLen = 63

// Prefix is a node's canonical path under a public host: "/<node>/".
// The node name is a single label of lowercase letters, digits, and
// interior hyphens; the trailing slash is part of the canonical form.
//
// The reserved control path "/_relay/" can never be a node prefix:
// node names do not contain underscores.
type Prefix string

// Parse normalizes s to a canonical node prefix. Both "/raspberry" and
// "/raspberry/" parse to "/raspberry/".
func Parse(s string) (Prefix, error) {
	node, err := validNode(s)
	if err != nil {
		return "", err
	}
	return Prefix("/" + node + "/"), nil
}

// String returns the canonical form, e.g. "/raspberry/".
func (p Prefix) String() string {
	return string(p)
}

// Node returns the node name without the surrounding slashes.
func (p Prefix) Node() string {
	return strings.TrimSuffix(string(p), "/")[1:]
}

// Contains reports whether path lies under p. p must be canonical, as
// produced by Parse.
func (p Prefix) Contains(path string) bool {
	prefix := string(p)
	if strings.HasPrefix(path, prefix) {
		return true
	}
	// path may omit the trailing slash: it must then be the bare node
	// root or continue with a path segment.
	bare := prefix[:len(prefix)-1]
	if strings.HasPrefix(path, bare) {
		return len(path) == len(bare) || path[len(bare)] == '/'
	}
	return false
}

// validNode extracts and validates the node label of s. Every failure
// wraps ErrInvalid with the specific reason.
func validNode(s string) (string, error) {
	parts := strings.Split(s, "/")
	var node string
	switch {
	case len(parts) == 3 && parts[0] == "" && parts[2] == "": // "/raspberry/"
		node = parts[1]
	case len(parts) == 2 && parts[0] == "": // "/raspberry"
		node = parts[1]
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalid, s)
	}
	if node == "" {
		return "", fmt.Errorf("%w: %q: empty node name", ErrInvalid, s)
	}
	if len(node) > maxNodeLen {
		return "", fmt.Errorf("%w: %q: node name longer than %d characters", ErrInvalid, s, maxNodeLen)
	}
	if node[0] == '-' || node[len(node)-1] == '-' {
		return "", fmt.Errorf("%w: %q: node name must not start or end with a hyphen", ErrInvalid, s)
	}
	for i, c := range node {
		switch {
		case c >= 'a' && c <= 'z' || c >= '0' && c <= '9':
		case c == '-' && i > 0 && i < len(node)-1:
		default:
			return "", fmt.Errorf("%w: %q: node names use lowercase letters, digits, and interior hyphens", ErrInvalid, s)
		}
	}
	return node, nil
}
