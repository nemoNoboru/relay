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
// (see docs/api.md). It can never be a node or site name.
const Reserved = "_relay"

// ErrReservedName is returned by ParseName for the reserved name.
var ErrReservedName = errors.New("name _relay is reserved for relay control")

// nameRE is the name grammar: 1-63 characters,
// lowercase alphanumerics and hyphens, no leading or trailing hyphen.
var nameRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// labelRE is the wider member-label grammar used for invite-chain labels:
// 1-63 letters, digits, dots, underscores, and hyphens.
var labelRE = regexp.MustCompile(`^[A-Za-z0-9._-]{1,63}$`)

// ValidName reports whether s satisfies the node/site name grammar.
func ValidName(s string) bool {
	return nameRE.MatchString(s)
}

// ValidLabel reports whether s satisfies the member-label grammar.
func ValidLabel(s string) bool {
	return labelRE.MatchString(s)
}

// ParseName validates s against the name grammar and returns it as a
// Name.
func ParseName(s string) (Name, error) {
	if s == Reserved {
		return "", ErrReservedName
	}
	if !ValidName(s) {
		return "", fmt.Errorf("invalid name %q: must match [a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?", s)
	}
	return Name(s), nil
}
