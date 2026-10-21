// Package relay is the module root. It exists to embed the public Relay
// website — the landing page and the why page — into the node binary, so
// every node hosts the project's front door at its root path.
package relay

import "embed"

// SiteFS embeds the public site files that live in the repository root:
// the hand-written landing page (host.html), the hand-written why page
// (why.html), and their shared stylesheet. There is no generator; the
// binary embeds whatever is present at build time.
//
//go:embed host.html why.html styles.css
var SiteFS embed.FS

// FontsFS embeds the self-hosted webfonts and their @font-face stylesheet.
// The node serves them at /fonts/, so neither the public site nor the owner UI
// has to load fonts from Google Fonts at runtime.
//
//go:embed web/fonts
var FontsFS embed.FS
