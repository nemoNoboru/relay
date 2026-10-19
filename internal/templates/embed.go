// Package templates embeds Relay's built-in post templates. Installed
// templates live in the node's data directory and shadow a built-in with the
// same id.
package templates

import "embed"

// Builtin holds every template folder under builtin/. Adding a folder with a
// valid relay-template.json is enough; the node does not keep a list of
// template ids. The store discovers the folders at runtime.
//
//go:embed all:builtin
var Builtin embed.FS
