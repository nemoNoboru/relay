package serve

import (
	"io/fs"
	"net/http"

	"github.com/nemoNoboru/relay"
)

// Fonts serves the node's self-hosted webfonts and their @font-face
// stylesheet at /fonts/, so neither the public site nor the owner UI has to
// load fonts from Google Fonts at runtime.
func Fonts() http.Handler {
	fonts, err := fs.Sub(relay.FontsFS, "web/fonts")
	if err != nil {
		// The embedded tree is fixed at build time; this cannot fail.
		panic("relay: embedded fonts: " + err.Error())
	}
	return http.StripPrefix("/fonts/", http.FileServer(http.FS(fonts)))
}
