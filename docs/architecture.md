# Relay Architecture

This describes the code as it exists today, then the additions planned for the
pivot. It replaces the old Caddy/yamux/tunnel draft, which described a product
that was never built. The pivot plan is in [`pivot-plan.md`](pivot-plan.md).

## Current system (built)

One Go binary, one data directory, the standard-library HTTP server. No
database. The only optional external process is `yt-dlp`, used by the
"Relay a url" convenience to re-host a link; that work runs as a bounded
background job.

```
visitor ──HTTP(S)──► relay node (net/http)
                       ├─ /                        embedded public site
                       ├─ /_relay/*.html           embedded owner UI
                       ├─ /_relay/v1/admin[/…]     host console API
                       ├─ /_relay/v1/invites       member-creates-invite
                       ├─ /_relay/v1/session       invite ↔ member cookie
                       ├─ /_relay/v1/me            member profile
                       ├─ /_relay/v1/feed          chronological post index
                       ├─ /_relay/v1/posts/…       one post + its comments
                       ├─ /_relay/v1/comments/…    delete own comment
                       ├─ /_relay/v1/people/…      one member's posts
                       ├─ /_relay/v1/templates     template manifests
                       ├─ /_relay/templates/…      prepublish forms and assets
                       ├─ /_relay/v1/publish       template publish endpoint
                       ├─ /_relay/v1/relay[/…]     re-host a url (background job)
                       ├─ /_relay/v1/sites/…       deploy (tar upload)
                       └─ /<node>/<site>/…         static sites on disk
```

### Process and wiring

`cmd/relay/main.go` dispatches subcommands:

```
relay node      run a node
relay deploy    upload a folder as a site
relay rollback  restore a previous version
relay key new   generate an Ed25519 key file (not part of the invite flow)
relay version
```

`runNode`:

1. parses flags (`--name`, `--data`, `--listen`, `--admin-token`,
   `--tls-cert`, `--tls-key`);
2. resolves the admin token (flag, `<data>/admin-token`, or the default
   `relay`);
3. loads sites from `<data>/sites` into a reloadable registry;
4. opens the key store (`<data>/deploy-keys.json`);
5. constructs the version store, uploader, admin console, invite handler,
   session handler, and the url relayer (with a shutdown that cancels and
   joins its background jobs);
6. embeds the owner UI from `web/dist`;
7. mounts the routes above;
8. serves HTTP or HTTPS, and drains on `SIGINT`/`SIGTERM`.

There is no router framework. `http.ServeMux` plus small handlers is enough.

### Packages

```
cmd/relay/          main, node, deploy, rollback, key commands
internal/domain/    Name parsing, reserved names
  key/              Ed25519 key file model (CLI only, not the invite credential)
  prefix/           Node path prefix model (used by tests/early code)
internal/route/     request path → node/site/rest
internal/identity/
  keys.go           invite/member keys, cascade revoke, site attribution
internal/serve/
  sites.go          load site folders from <data>/sites
  host.go           serve /<node>/<site>/…
  registry.go       reloadable site map
  upload.go         tar upload, validation, atomic replace
  versions.go       snapshot/list/rollback, keep 10
  keys.go           identity aliases (thin HTTP shim)
  invites.go        member-creates-invite endpoint
  session.go        member cookie, /me, profile update
  posts.go          feed, post, and profile JSON
  comments.go       comment API, rate guard, sanitized render
  templates.go      discovered built-ins + installed templates, install tar
  publish.go        template publish endpoint, path/size validation
  postmeta.go       relay-post.json sidecar
  admin.go          host console API (including host comment deletion)
  landing.go        embedded public site at /
  ui.go             embedded owner UI under /_relay/
  httputil.go       JSON error envelope, same-origin, rate guard
internal/social/
  post.go           Post model, title/limit/cover rules
  store.go          posts.json, atomic writes, ordering
  comment.go        Comment model, limits
  commentstore.go   comments.json, atomic writes, ordering
  template.go       template manifest loading/validation
  render.go         safe comment markup (escape, emphasis, links)
internal/templates/builtin/ embedded note/link/gallery/article/songs/videos/carousel
web/                Svelte 5 + Vite owner UI, embedded into the binary
site.go             //go:embed of the public site
```

The code follows `.pi/skills/go-architecture`: `cmd → feature → domain`,
consumer-side interfaces, errors as values, `context` where I/O blocks,
table-driven tests, no mutable global state.

### Data directory

```
<data>/
  admin-token            host credential (0600)
  deploy-keys.json       invite/member keys + site attribution
  posts.json             post index (author, title, time, template)
  comments.json          comments (post, member, body, time)
  templates/             installed templates (shadow built-ins)
  sites/
    <site>/              live static site
    .versions/<site>/<id>/   rollback snapshots (last 10)
```

`deploy-keys.json` is a single JSON document:

```json
{
  "keys": [
    {
      "id": "dk_…",
      "user": "alice",
      "token": "…",
      "hash": "…",
      "created": "2026-10-19T12:00:00Z",
      "revoked": false,
      "invitedBy": "",
      "displayName": "Alice",
      "bio": "Runs the Thursday reading group."
    }
  ],
  "sites": { "sunday-walk": "alice" }
}
```

`user` is a display label, not a path. `invitedBy` points at the parent key.
Revoking a key walks the children and revokes the subtree.

### Authentication (today)

| Credential | Holder | Surface |
|---|---|---|
| Admin token | host | `/_relay/v1/admin…` |
| Deploy key | member | `POST /_relay/v1/sites/<site>/deploy`, `POST /_relay/v1/invites` |
| Member cookie | member | the same deploy/invite endpoints, plus `/me`, `/feed`, `/posts/…`, `/people/…`, and own-comment deletion |

The admin token is compared by SHA-256 in constant time. Deploy keys are looked
up by SHA-256 hash and checked for revocation. The plaintext token is stored so
the console can reveal it.

`POST /_relay/v1/session` exchanges an invitation token for the
`relay_member` cookie (`HttpOnly`, `SameSite=Lax`, `Secure` on HTTPS, 30 days).
The cookie value is the invitation token; every request hashes and checks it,
so revocation is immediate and there is no session table. A bearer token still
works for the CLI, uploads, and assistants. See [`invites.md`](invites.md).

### Static serving

`route.Resolve` splits `/<node>/<site>/<rest>` and rejects the reserved
`_relay` segment. The registry looks up the site in a mutex-guarded map loaded
from `<data>/sites`. `http.FileServer` serves the folder. Directory requests
without a trailing slash get a 301 that preserves the full path. Unknown nodes,
unknown sites, and the reserved prefix are plain 404s.

A deploy writes a temporary directory, validates the whole tar (relative paths,
regular files only, file count, size), then renames it into place and calls
`registry.Reload()`. The old site is moved into `.versions` first. The live site
is never half-written.

### Embedded public site and owner UI

`site.go` embeds `host.html`, `why.html`, and `styles.css` from the repo
root, plus the self-hosted webfonts in `web/fonts`. `internal/serve/landing.go`
maps `/` to the built feed shell, `/host` to `host.html`, and `/why`
and `/styles.css` to their files; `serve.Fonts` serves the fonts at `/fonts/`.
Everything else falls through to site serving.

`web/embed.go` embeds the built Vite output (`web/dist`) and `ui.go` serves it
under `/_relay/`. `/_relay/` redirects to `/` (the feed). The UI is a set of
short-lived documents, not a SPA, sharing one `Shell.svelte` chrome:

- `feed.html` (served at `/`) + Svelte + Effect-TS: the public feed, people strip, peek/cover cards, collapsed comments.
- `people.html` + Svelte: the member directory and one member's profile.
- `post.html` + Svelte: post wrapper, static iframe, comments.
- `admin.html` + Svelte + Effect-TS: host console (keeps its own header).
- `publish.html` + plain TypeScript: template picker, invite/publish flow, Markdown-to-page.
- `deploy.html`: static step-by-step instructions.
- `/_relay/templates/<id>/…`: trusted prepublish forms and template JS; the
  browser runs them, the node never does. Members can install a template tar
  from the publish page; the id cannot shadow a built-in.
- `/_relay/template/runtime.js`: shared browser runtime for built-in templates.

The old `node.html` and `site.html` mock consoles were removed in the pivot
cleanup. They polled endpoints the node never served.

### Testing

- Go: table-driven unit tests per package; end-to-end handlers with
  `httptest`; `go test -race ./...`.
- Web: `tsc --noEmit`, `svelte-check`, Vite build, Vitest (SSR renders for the
  home feed, profile, post wrapper with comments, and host console; API client;
  publish helpers), and a DOM-stub smoke test of the built home and host
  console. Go tests cover cursor pagination, comments, template manifests,
  install/remove, and publish validation.
- Site: `host.html` and `why.html` are hand-written HTML. There is no
  page generator; edits go directly into the files.
- `make check` runs all of it; CI runs the same gates.

### Deployment

- Binary + systemd (or any process supervisor).
- Docker image (unprivileged, `/data` volume).
- Kubernetes manifest and ingress example.
- HTTPS either with `--tls-cert`/`--tls-key` or a reverse proxy in front.
- `/healthz` for probes.

## Target system (remaining pivot additions)

The hosting core and the member welcome stay. We add the rest of the social
layer and keep it boring.

```
                    ┌──────────── relay node ────────────┐
member browser ───► │  net/http                         │
                    │   ├─ public landing (no cookie)    │
                    │   ├─ /_relay/home.html (member)    │
                    │   ├─ /_relay/people/<member>       │
                    │   ├─ /_relay/posts/<site>          │
                    │   ├─ /_relay/v1/{session,feed,…}   │
                    │   ├─ /_relay/templates/<t>/…       │
                    │   └─ /<node>/<site>/  static post  │
                    │                                   │
                    │  stores: deploy-keys.json, posts.json,    │
                    │          comments.json             │
                    │  files:  sites/, templates/        │
                    └───────────────────────────────────┘
```

### Next packages (Phase 5)

The social layer is complete through templates. Phase 5 is host operations:
backup/export, invite expiry, limits, and moderation polish. There is no new
content package planned.

The embeddable comment panel remains deferred: static posts are CSP-sandboxed,
so a script inside a post cannot safely call the member API. Comments live on
the post wrapper until there is an isolated embed origin or a scoped token.

The HTTP handlers stay thin. The social package owns the rules; the store owns
atomic persistence. No social package imports `net/http`.

### Feed pagination and inline comments

The feed and profile APIs use an opaque cursor built from `(publishedAt, site)`
and return `next`; the browser loads the next page when a sentinel at the end of
the list enters the viewport. Each post view carries its comments, so the feed
card can show the list and a comment form without a second request. The post
wrapper still uses the dedicated comments endpoint.

Posts may carry an optional `image` path; the API exposes it as `imageUrl` and
the feed renders it as a cover. Gallery and Article templates set it.

### Relay-a-url surface

The publish page's "Relay a url" section lets a member paste a link and have
the node re-host what it points at. The node first asks `yt-dlp` to download
media; when that fails or `yt-dlp` is absent, it builds a share card from the
page's social preview metadata. Either way the result is published as a normal
static post, so the social layer treats it like any other page.

This is the one place Relay runs background work:

- A request returns a job id immediately and the download proceeds in a
  goroutine bounded by a per-job timeout and a small in-flight cap.
- Jobs live in memory only. Restarting the node drops running and finished
  jobs; a job that already published leaves a normal post behind.
- The `Relayer` owns its goroutines: node shutdown cancels the jobs' contexts
  and waits for them to exit before the process returns.
- `yt-dlp` and `ffmpeg` are optional. They are spawned as subprocesses, never
  linked into the node, and their absence degrades to the share-card path.

### Persistence boundaries

- **Files** are the content: `sites/<site>/`, `templates/<t>/`. Built-ins are
  discovered from one embedded `builtin/` tree; installed folders are scanned
  at runtime.
- **JSON** is the index: `deploy-keys.json`, `posts.json`, `comments.json`.
- Every write is mutex-guarded, written to a temp file, fsynced, and renamed.
- The store is an interface so a SQLite implementation can replace it without
  changing handlers.
- The data directory remains the backup unit.

### Architecture rules for the pivot

1. Static serving does not depend on the social index. A broken `posts.json`
   must not take down the sites.
2. The social index must be rebuildable from the files and sidecars. Provide
   `relay index` for backfill and repair.
3. Template code never runs in the node process.
4. Comments and feed require a live member cookie; posts are public by URL.
5. No background job is required to serve a page. Optional pruning can be a CLI
   command. The one exception is the "Relay a url" action, which runs a bounded
   background job and then publishes a normal static page; see
   [Relay-a-url surface](#relay-a-url-surface).
6. No new external dependency without a written reason. A database is not
   needed for dozens of members.
7. The host's admin token remains separate from member invitations.

### Security boundaries

| Boundary | Rule |
|---|---|
| Public landing vs member home | cookie check in the handler; home returns 401/redirect when absent |
| Member cookie | HttpOnly, SameSite=Lax, Secure on TLS; same-origin checks on writes |
| Post content | static files, no server-side execution; served under `CSP: sandbox allow-scripts allow-popups allow-forms allow-downloads` so a post script cannot call member APIs with the viewer's cookie |
| Template code | browser only; host-installed and trusted |
| Comments | sanitized server-side; member-only read/write; host can delete any |
| Upload paths | relative, no `..`, no links, size and count caps |
| Admin API | admin token only; never the member cookie |
| Secrets | tokens never in query strings, paths, logs, or commits |

## Explicit non-goals

- A tunnel/NAT traversal system. The old draft imagined one; it is not in this
  product.
- Per-site databases, a runtime API for site code, or a QuickJS sandbox.
- A classifier/moderation model.
- Federation, a global directory, or cross-node identity.
- Per-post origin isolation with separate hostnames. Posts share one hostname,
  but each document is sandboxed; a post is a page, not a tenant.

If any of these become necessary, they start as a new document, not as code
creeping into the hosting core.
