// Package relay is the module root. It exists to embed the public Relay
// website — the landing page and its technical companion — into the node
// binary, so every node hosts the project's front door at its root path.
package relay

import "embed"

// SiteFS embeds the public site files that live in the repository root:
// the hand-written landing page (index.html), the hand-written technicals
// page (technicals.html), and their shared stylesheet. There is no generator;
// the binary embeds whatever is present at build time.
//
//go:embed index.html technicals.html styles.css
var SiteFS embed.FS
