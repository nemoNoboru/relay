// Package relay is the module root. It exists to embed the public Relay
// website — the landing page and its companions — into the node binary, so
// every node hosts the project's front door at its root path.
package relay

import "embed"

// SiteFS embeds the public site files that live in the repository root:
// the hand-written landing page (index.html), the two generated pages
// (manifesto.html, technicals.html), and their shared stylesheet. The
// generated pages are rebuilt from their Markdown sources by
// tools/build.py; `make check` fails if they are stale, and the binary
// embeds whatever is present at build time.
//
//go:embed index.html manifesto.html technicals.html styles.css
var SiteFS embed.FS
