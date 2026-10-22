package serve

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/nemoNoboru/relay"
)

// Fonts serves the node's self-hosted webfonts and their @font-face
// stylesheet at /fonts/, so neither the public site nor the owner UI has to
// load fonts from Google Fonts at runtime.
func Fonts() (http.Handler, error) {
	fonts, err := fs.Sub(relay.FontsFS, "web/fonts")
	if err != nil {
		return nil, fmt.Errorf("embedded fonts: %w", err)
	}
	return http.StripPrefix("/fonts/", http.FileServer(http.FS(fonts))), nil
}
