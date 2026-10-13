// Package domain holds the relay's vocabulary: validated node and site
// names.
package domain

import (
	"errors"
	"fmt"
	"regexp"
)

// Name is a validated node or site name.
type Name string

// Reserved is the first path segment reserved for relay control
// (contract 2.1). It can never be a node or site name.
const Reserved = "_relay"

// ErrReservedName is returned by ParseName for the reserved name.
var ErrReservedName = errors.New("name _relay is reserved for relay control")

// nameRE is the name grammar from contract 2.2: 1-63 characters,
// lowercase alphanumerics and hyphens, no leading or trailing hyphen.
var nameRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ParseName validates s against the name grammar and returns it as a
// Name.
func ParseName(s string) (Name, error) {
	if s == Reserved {
		return "", ErrReservedName
	}
	if !nameRE.MatchString(s) {
		return "", fmt.Errorf("invalid name %q: must match [a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?", s)
	}
	return Name(s), nil
}
