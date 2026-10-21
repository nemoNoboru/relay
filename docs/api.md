# Relay API

Two parts: what the node serves **today**, and the endpoints the pivot will add.
The current part is accurate against the code in `internal/serve/`. The planned
part is a contract for [`pivot-plan.md`](pivot-plan.md); it is not built yet.

All JSON errors use one envelope:

```json
{ "error": { "code": "bad_token", "message": "invalid or missing deploy token" } }
```

Credentials are bearer tokens in the `Authorization` header unless stated
otherwise. The pivot adds a member cookie; see [`invites.md`](invites.md).

---

## Current API (built)

### Public and static routes

| Method | Path | Auth | Result |
|---|---|---|---|
| `GET` | `/` | none | Embedded landing page |
| `GET` | `/index.html` | none | 301 to `/` |
| `GET` | `/styles.css` | none | Embedded stylesheet |
| `GET` | `/<node>/<site>/…` | none | Static site file |
| `GET` | `/healthz` | none | `200 text/plain` with `ok` |
| `GET` | `/_relay/` | none | 302 to `/_relay/home.html` |
| `GET` | `/_relay/home.html` | none (page); cookie for member data | Member welcome shell |
| `GET` | `/_relay/admin.html` | none (page) | Host console shell |
| `GET` | `/_relay/publish.html` | none (page) | Publish page |
| `GET` | `/_relay/deploy.html` | none (page) | Step-by-step upload page |
| `GET` | `/_relay/skills/relay-deploy/SKILL.md` | none | Deploy instructions for assistants |

A static site request is `/<node>/<site>/<path>`. Unknown nodes, unknown sites,
the reserved `_relay` prefix, and malformed names are 404. A directory without a
trailing slash gets a 301 that preserves the full path.

### Host console

Base: `/_relay/v1/admin`. Auth: `Authorization: Bearer <admin token>`.
Missing or wrong credentials are `401 bad_token`.

| Method | Path | Body | Result |
|---|---|---|---|
| `GET` | `/_relay/v1/admin` | — | `{keys, sites, mustChangeToken}` |
| `POST` | `/_relay/v1/admin/token` | `{ "token": "…" }` | `{ "changed": true }` |
| `POST` | `/_relay/v1/admin/keys` | `{ "user": "alice" }` | `201 {key, token}` |
| `POST` | `/_relay/v1/admin/keys/{id}/revoke` | — | `{key}` (cascades to invited keys) |
| `POST` | `/_relay/v1/admin/keys/{id}/rotate` | — | `201 {key, token}` |
| `GET` | `/_relay/v1/admin/sites/{site}` | — | `{site, versions}` |
| `GET` | `/_relay/v1/admin/sites/{site}/versions` | — | `{site, versions}` |
| `POST` | `/_relay/v1/admin/sites/{site}/rollback` | `{ "version": "…" }` or `{}` | `{site, version}` |
| `DELETE` | `/_relay/v1/admin/comments/{id}` | — | `204` (delete any comment) |

The new admin token must be at least 8 characters and cannot be the default
`relay`. A node that has never changed it reports `mustChangeToken: true`.

The overview also carries `comments`: the newest 50 comments, each with author,
post, body, sanitized HTML, and timestamp. The host console uses this to delete
any comment.

`KeyInfo` in a response:

```json
{
  "id": "dk_…",
  "user": "alice",
  "token": "…",
  "created": "2026-10-19T12:00:00Z",
  "revoked": false,
  "invitedBy": "dk_…",
  "displayName": "Alice",
  "bio": "Runs the reading group."
}
```

`token` is the plaintext credential, present so the host console can reveal and
copy it. The hash is never returned. `displayName` and `bio` are optional.

Version list:

```json
{
  "site": "sunday-walk",
  "versions": [
    { "id": "20261019T160400Z-…", "created": "2026-10-19T16:04:00Z" }
  ]
}
```

The list is newest first. A rollback with no `version` uses the newest snapshot.
The current live site is snapshotted before the rollback, so the operation can
be reversed.

### Member invites

| Method | Path | Auth | Body | Result |
|---|---|---|---|---|
| `POST` | `/_relay/v1/invites` | deploy key or member cookie | `{ "name": "bob" }` | `201 {key, token}` |

The new key's label is `<inviter>/<name>` and its `invitedBy` points at the
inviter's key. Revoking the inviter revokes this key too. The response `token`
is the full invitation credential.

Errors: `401 bad_token` for an unknown/revoked inviter, `400 bad_request` for a
bad name.

### Member session

The member cookie is the browser identity. The invitation token remains valid
as a bearer token for the CLI and assistants. The cookie is accepted as the
same credential by the deploy and invite endpoints, so a signed-in browser does
not need to paste the key again.

| Method | Path | Auth | Body | Result |
|---|---|---|---|---|
| `POST` | `/_relay/v1/session` | invite token in body | `{ "token": "…" }` | `200 {member, node}`, sets cookie |
| `DELETE` | `/_relay/v1/session` | cookie | — | `204`, clears cookie |
| `GET` | `/_relay/v1/me` | cookie or bearer | — | `{member, node}` |
| `PATCH` | `/_relay/v1/me` | cookie or bearer | `{ "displayName": "…", "bio": "…" }` | `{member, node}` |

`member`:

```json
{
  "id": "dk_…",
  "user": "alice/bob",
  "displayName": "Bob",
  "bio": "Runs the reading group.",
  "created": "2026-10-19T12:00:00Z",
  "invitedBy": "dk_…"
}
```

`node`:

```json
{ "name": "mynode", "members": 12 }
```

Cookie: `relay_member`, `HttpOnly`, `SameSite=Lax`, `Secure` on TLS,
`Path=/`, 30-day lifetime. Its value is the invitation token; the node hashes
and checks it on every request, so revoking a key takes effect immediately.
Signing out clears the cookie but does not revoke the key.

`displayName` and `bio` are optional. An empty display name falls back to the
last segment of `user`. Limits: display name 63 characters, bio 280.

Errors: `401 session_required` when there is no live member credential, and
`400 bad_profile` for an invalid display name or bio.

### Deploy

| Method | Path | Auth | Body | Result |
|---|---|---|---|---|
| `POST` | `/_relay/v1/sites/{site}/deploy` | deploy key or member cookie | `application/x-tar` | `{site, files, bytes, url}` |

Example:

```http
POST /_relay/v1/sites/sunday-walk/deploy
Authorization: Bearer <token>
Content-Type: application/x-tar

<tar bytes>
```

Success:

```json
{
  "site": "sunday-walk",
  "files": 4,
  "bytes": 184320,
  "url": "/mynode/sunday-walk/",
  "postUrl": "/_relay/posts/sunday-walk"
}
```

Rules:

- relative paths only; no `..`, no absolute paths, no NUL, no backslashes;
- regular files and directories only; no symlinks, hard links, or special
  entries;
- at most 65,536 files and 256 MiB of archive bytes by default;
- the upload replaces the whole site atomically;
- the previous 10 versions are kept.

Errors:

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | Missing, wrong, or revoked deploy token |
| `bad_name` | 400 | Invalid site name |
| `bad_archive` | 400 | Not a tar, or an entry breaks the archive rules |
| `too_large` | 413 | Archive or file count over the cap |

A deploy key can currently publish any site on the node. The site is attributed
to the key's user, but there is no per-site ownership check yet. The pivot plan
keeps this rule until per-site ownership is designed.

### Post index

Member-only. The content is static; these endpoints serve the social index
around it. Auth is the member cookie or an invitation bearer token.

| Method | Path | Result |
|---|---|---|
| `GET` | `/_relay/v1/feed?cursor=…&limit=…` | `{posts: [...], next?}`, newest first |
| `GET` | `/_relay/v1/posts/<site>` | `{post}` |
| `GET` | `/_relay/v1/people/<member-id>?cursor=…&limit=…` | `{member, posts, total, next?}` |

`limit` defaults to 10, max 50. `next` is an opaque cursor for infinite scroll;
omit it on the first request and send it back for the next page.

Post:

```json
{
  "id": "post_…",
  "site": "sunday-walk",
  "member": "dk_…",
  "title": "Sunday walk",
  "summary": "Six photos from the canal.",
  "template": "gallery",
  "image": "assets/1-canal.svg",
  "imageUrl": "/mynode/sunday-walk/assets/1-canal.svg",
  "publishedAt": "2026-10-19T16:04:00Z",
  "updatedAt": "2026-10-19T16:04:00Z",
  "author": { "id": "dk_…", "displayName": "Bob" },
  "url": "/mynode/sunday-walk/",
  "postUrl": "/_relay/posts/sunday-walk",
  "authorUrl": "/_relay/people/dk_…",
  "comments": []
}
```

Ordering is `publishedAt` descending, then site name. There is no ranking.
`hidden` posts are omitted from the feed and profile; the static page still
serves. Each post carries its comments so the feed and profile can render the
inline comment list and form without an extra request. The post wrapper also
has a dedicated comments endpoint below.

A deploy indexes the post automatically. The index is `<data>/posts.json`, and
each site also gets a portable `relay-post.json` sidecar. Run `relay index
--data <dir>` after upgrading an old node to backfill sites that predate the
index.

### Comments

Member-only. Comments are attached to indexed posts and read oldest first.

| Method | Path | Body | Result |
|---|---|---|---|
| `GET` | `/_relay/v1/posts/<site>/comments` | — | `{comments: [...]}` |
| `POST` | `/_relay/v1/posts/<site>/comments` | `{ "body": "…" }` | `201 {comment}` |
| `DELETE` | `/_relay/v1/comments/<id>` | — | `204` (own comment only) |

Comment:

```json
{
  "id": "cmt_…",
  "post": "sunday-walk",
  "member": "dk_…",
  "author": { "id": "dk_…", "displayName": "Alice" },
  "body": "The third photo is my favourite. **Lovely.**",
  "html": "<p>The third photo is my favourite. <strong>Lovely.</strong></p>",
  "createdAt": "2026-10-19T16:20:00Z",
  "mine": true
}
```

`html` is rendered by the node from an escaped subset: paragraphs, line breaks,
`**bold**`, `*italic*`, `_italic_`, and links whose URL is `http(s)://` or
node-relative. Scripts and other HTML are escaped. `mine` tells the UI whether
to show a delete action.

Limits: 2,000 characters; one comment per member per 2 seconds (in-memory
guard). Host deletion uses the admin endpoint, not the member cookie. The feed
and profile post objects embed the same comment views, so the inline cards use
these endpoints only when a member posts or deletes.

### Templates and publish

Template code runs in the browser. The node serves the form and accepts the
finished files; it never executes `publish.js`.

| Method | Path | Result |
|---|---|---|
| `GET` | `/_relay/v1/templates` | `{templates: [...]}` |
| `POST` | `/_relay/v1/templates` | Install a template from a tar (`application/x-tar`) |
| `GET` | `/_relay/templates/<id>/prepublish.html` | Template form |
| `GET` | `/_relay/templates/<id>/…` | Template assets |
| `POST` | `/_relay/v1/publish` | `{post, files, bytes}` |

`templates` entries are the manifest plus `builtin`:

```json
{
  "schema": 1,
  "id": "note",
  "name": "Note",
  "description": "A short page of writing.",
  "version": "1.0.0",
  "fields": [
    { "name": "title", "type": "text", "required": true, "max": 160 },
    { "name": "body", "type": "richtext", "required": true }
  ],
  "entry": "prepublish.html",
  "script": "publish.js",
  "builtin": true
}
```

Built-in templates: `note`, `link`, `gallery`, `article`, `songs`, `videos`,
`carousel`, `jetset`, `longform`, `reply`, and `event`. The list is discovered from the embedded folder tree and the
installed folder, not hardcoded in the server. Any live member can install a
template tar from the publish page:

- the tar root must contain `relay-template.json`, `prepublish.html`, and
  `publish.js`;
- the id must not collide with a built-in or an already installed template;
- the archive is capped at 16 MiB and 4,096 files.

The host can also manage installed templates from a folder with the CLI:

```sh
relay template list --data <dir>
relay template install --data <dir> <folder>
relay template remove --data <dir> <id>
```

Publish request:

```json
{
  "template": "article",
  "title": "The canal in October",
  "summary": "A walk, six photos, and a map.",
  "site": "canal-in-october",
  "image": "assets/map.png",
  "files": [
    { "path": "index.html", "content": "<!doctype html>…" },
    { "path": "assets/map.png", "base64": "iVBORw0KGgo…" }
  ]
}
```

Rules:

- the member must be live; the template must exist;
- `index.html` is required;
- paths are relative, contained, and cannot be duplicated;
- at most 65,536 files and 64 MiB of decoded bytes;
- an explicit `site` follows the current replacement rule; an omitted `site`
  is derived from the title and made unique with a numeric suffix;
- `image` is optional, relative, and must be one of the generated files; it is
  used as the feed cover;
- the node snapshots the previous version, writes the site atomically, writes
  `relay-post.json`, and indexes the post.

Success:

```json
{
  "post": {
    "id": "post_…",
    "site": "canal-in-october",
    "title": "The canal in October",
    "template": "article",
    "url": "/mynode/canal-in-october/",
    "postUrl": "/_relay/posts/canal-in-october"
  },
  "files": 2,
  "bytes": 184320
}
```

Errors:

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_template` | 400 | Unknown or uninstalled template |
| `bad_name` | 400 | Invalid explicit site name |
| `bad_file` | 400 | Missing index, bad path, duplicate path, or bad base64 |
| `too_large` | 413 | File count or decoded bytes over the cap |
| `session_required` | 401 | No live member credential |
| `bad_origin` | 403 | Cross-origin write |

The existing tar deploy endpoint remains. A plain upload becomes a post with
`template: "custom"`.

### Relay a url

Member-only. The publish page's "Relay a url" section re-hosts what a link
points at. The node runs [`yt-dlp`](https://github.com/yt-dlp/yt-dlp) out of
process to download the content; when it downloads a file or set of files
(an image, a video, a song, a gallery), the files are hosted directly as a
new page. When it downloads nothing (or `yt-dlp` is not installed), the node
falls back to a share card built from the page's social preview metadata and
the url itself. Both paths publish a normal site and index it as a post.

| Method | Path | Auth | Body | Result |
|---|---|---|---|---|
| `POST` | `/_relay/v1/relay` | deploy key or member cookie | `{ "url": "…", "webhook": "…" }` | `202 {job}` |
| `GET` | `/_relay/v1/relay/<id>` | deploy key or member cookie | — | `{job}` |

Starting a job returns immediately; the download runs in the background and
is polled with the second endpoint. The `webhook` field is optional: when
set to a full http(s) url, the node `POST`s the finished job (the same JSON
as the poll endpoint) to that url once the job is `done` or `failed`.
Delivery is best-effort and never changes the job result.

Third parties use the same member credential the members post with: send the
invitation key as a bearer token. A bot (Telegram, Discord, a cron job) can
relay by POSTing here and either poll the job or pass a `webhook` url to get
the result pushed back.

```json
{
  "id": "rj_…",
  "source": "https://…",
  "status": "running",
  "created": "2026-10-19T16:04:00Z"
}
```

`status` is `running`, `done`, or `failed`. When `done`, the job carries the
post and the live address:

```json
{
  "id": "rj_…",
  "source": "https://…",
  "status": "done",
  "title": "Sunday walk",
  "kind": "media",
  "site": "sunday-walk",
  "url": "/mynode/sunday-walk/",
  "postUrl": "/_relay/posts/sunday-walk",
  "created": "2026-10-19T16:04:00Z"
}
```

`kind` is `media` when files were downloaded and `card` for the social
preview fallback. When `failed`, `error` carries a short reason.

Errors:

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_url` | 400 | Missing or non-http(s) url |
| `bad_webhook` | 400 | `webhook` is not a full http(s) url |
| `too_many` | 429 | The node already has the maximum jobs running |
| `session_required` | 401 | No live member credential |
| `bad_origin` | 403 | Cross-origin write |
| `not_found` | 404 | Unknown job id |

### CLI

```
relay node --name <n> --data <d> --listen <addr>
           [--admin-token <t>]
           [--tls-cert <f> --tls-key <f>]
relay deploy --origin <url> --site <n> [--token <t>] [--insecure] <folder>
relay rollback --origin <url> --site <n> [--admin-token <t>] [--to <v>] [--insecure]
relay index --data <dir>
relay key new [-o <file>]
relay version
```

`relay key new` generates an Ed25519 key file. It is not part of the invitation
flow and is not used by the node. It is kept as a separate, tested primitive
until the pivot decides whether ownership keys have a role.

---

## Planned API (pivot)

These endpoints are contracts for future work, not built features. Phases 1–4
(member session, feed, profiles, post wrappers, comments, and templates) are
built; Phase 5 is host operations. The product rules are in
[`pivot-plan.md`](pivot-plan.md), [`invites.md`](invites.md), and
[`publishing.md`](publishing.md). Further proposals are referenced below:
[`navigation-pivot.md`](navigation-pivot.md),
[`indieweb-syndication.md`](indieweb-syndication.md),
[`indieauth.md`](indieauth.md), and [`accounts.md`](accounts.md).

### Admin additions

| Method | Path | Auth | Result |
|---|---|---|---|
| `GET` | `/_relay/v1/admin/posts` | host | Post list, including hidden |
| `POST` | `/_relay/v1/admin/posts/<site>/hide` | host | `{post}` |
| `POST` | `/_relay/v1/admin/posts/<site>/unhide` | host | `{post}` |
| `DELETE` | `/_relay/v1/admin/posts/<site>` | host | Delete the site and its index entry |
| `GET` | `/_relay/v1/admin/members` | host | Members and invite tree |
| `PATCH` | `/_relay/v1/admin/members/<id>` | host | Edit display name/bio |

### Node discovery (optional, later)

A node may expose a small read-only description for its own members:

```
GET /_relay/v1/node
{ "name": "mynode", "members": 12, "version": "0.2.0", "maxMembers": 29 }
```

This is not a public directory and must not list members or posts to
unauthenticated callers. It is useful for a future "is this node alive?" check
and for the home header.

### Navigation & feed-first pivot (proposal)

From [`navigation-pivot.md`](navigation-pivot.md). Not built.

- `/` serves the public feed (no session); `/host` and `/landing` serve the
  hosting pitch; `/index.html` redirects to `/host`; `/_relay/home.html`
  redirects to `/`.
- Read endpoints become public: `GET /_relay/v1/feed`,
  `GET /_relay/v1/posts/<site>`, `GET /_relay/v1/people/<id>/posts`,
  `GET /_relay/v1/posts/<site>/comments`, and `GET /_relay/v1/templates`
  (list only). Writes stay member-only.
- New `GET /_relay/v1/people` (public members directory).
- The feed/post JSON gains `templateName`.
- An optional `--members-only` flag restores the old members-only reads.

### IndieWeb: feeds, Webmention, Micropub, IndieAuth (proposal)

From [`indieweb-syndication.md`](indieweb-syndication.md),
[`indieauth.md`](indieauth.md), and [`accounts.md`](accounts.md). Not built.

| Method | Path | Auth | Result |
|---|---|---|---|
| `GET` | `/_relay/v1/feed.json` | none | JSON Feed 1.1 of posts |
| `POST` | `/_relay/v1/webmention` | none | Accept a source/target mention; verified + sanitized |
| `POST` | `/_relay/v1/micropub` | IndieAuth token | Create a post from MF2 (`h=entry`) |
| `GET` | `/_relay/v1/micropub?q=syndicate-to` | IndieAuth token | List syndication targets |
| `GET` | `/_relay/v1/indieauth/metadata` | none | IndieAuth server metadata |
| `GET`/`POST` | `/_relay/v1/indieauth/authorization` | member (invitation or passkey) | Authorization endpoint |
| `POST` | `/_relay/v1/indieauth/token` | none (PKCE) | Token endpoint (`access_token`, `me`) |
| `POST` | `/_relay/v1/indieauth/revocation` | token | Revoke a token |
| `POST` | `/_relay/v1/indieauth/introspection` | internal | Introspect a token (RFC 7662 + `me`) |

Profile pages emit `rel="indieauth-metadata"` and `rel="micropub"` links; the
feed, post, and profile pages carry `h-feed`/`h-entry`/`h-card` markup.

---

## Compatibility and versioning

- The current endpoints stay at their paths. New social endpoints are additive.
- The tar deploy and invite endpoints keep working across the pivot.
- A future breaking change gets a new version prefix (`/_relay/v2/…`) and a
  deprecation note in this file.
- The public site paths (`/`, `/index.html`, `/styles.css`)
  stay stable. `manifesto.html` is removed; a redirect may be kept for one
  release.
