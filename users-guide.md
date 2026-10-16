# Relay — User's Guide to Hosting and Uploading Static Sites

This guide walks through the full flow: running a Relay node on your server,
getting the admin token, issuing per-user deploy keys from the operator
console, and handing a friend (or their agent) everything they need to upload
a static website.

---

## Overview

A Relay node serves static sites. Each site is a folder of files (HTML, CSS,
JS, images, …) served live at:

```
{origin}/{node}/{site}/
```

There are **two credentials**:

| Credential | Held by | Unlocks |
|---|---|---|
| **Admin token** | the operator (you) | the operator console — issue, revoke, rotate deploy keys; see sites by user |
| **Deploy key** | each user (friend) | uploading a site with a bearer token |

Uploading is one HTTP call: a `tar` archive of the site's contents,
authenticated with a deploy key's token. A successful upload goes live
immediately — no restart — and the site is attributed to that user.

| Piece | URL / value |
|---|---|
| Operator console | `{origin}/_relay/admin.html` |
| Upload endpoint | `POST {origin}/_relay/v1/sites/{site}/deploy` |
| Site URL | `{origin}/{node}/{site}/` |
| Owner docs (hosted) | `{origin}/_relay/deploy.html` |
| Agent skill (hosted) | `{origin}/_relay/skills/relay-deploy/SKILL.md` |

---

## Part 1 — Operator: run the node

### 1. Run the node

**Docker** (from the published image):

```sh
docker pull ghcr.io/nemoNoboru/relay:latest
docker run -d --name relay -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemoNoboru/relay:latest
```

**Binary:**

```sh
relay node --name mynode --data /var/lib/relay --listen :8080
```

- `--name` is the node's path segment (lowercase letters, digits, hyphens).
- `--data` holds all state: uploaded sites (`<data>/sites/`), the admin token,
  and the deploy-key store (`<data>/deploy-keys.json`).
- `--listen` is the bind address.

### 2. Get the admin token

On first start the node generates an admin token, writes it to
`<data>/admin-token` (mode `0600`), and prints it **once** to the log:

```sh
docker logs relay                       # look for: generated admin token ... token=...
docker exec relay cat /data/admin-token # or read the file directly
```

Or set your own, so you control the value:

```sh
docker run -d --name relay -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemoNoboru/relay:latest --admin-token "$(openssl rand -base64 32)"
```

> The admin token is the master key to the node. It is **not** what you hand
> out to uploaders — keep it to yourself.

### 3. Put the node behind TLS

Two ways to get HTTPS:

1. **Native TLS:** pass a certificate pair to the node —
   `relay node … --tls-cert cert.pem --tls-key key.pem` (or append the flags to
   the image's entrypoint, mounting the files).
2. **Reverse proxy:** Caddy/Nginx terminates TLS and forwards to
   `localhost:8080`.

With a self-signed certificate, tell the CLI to skip verification with
`--insecure` on `relay deploy` / `relay rollback`. Real deployments should use
a trusted certificate (or the reverse proxy's).

---

## Part 2 — Operator: issue deploy keys

Open the operator console in a browser:

```
https://relay.example.org/_relay/admin.html
```

Sign in with the **admin token**. The console shows two panels:

1. **Deploy keys** — every key you have issued, **grouped by user**, each
   with its id, creation time, status (`active` / `revoked`), a **Reveal /
   Copy** button for its token, and **Rotate** / **Revoke** actions.
2. **Sites by user** — every deployed site, grouped under the user who last
   deployed it.

### Create a key for a friend

Enter the friend's name (a label like `alice`, any of letters/digits/`.`/`_`/`-`,
1–63 chars) and press **Create key**. The console shows the new key's token in
a highlighted box; copy it now. The token is **also listed under that user**
in the Deploy keys panel, where **Reveal** unmasks it and **Copy** puts it on
your clipboard — so you can re-send it later if needed.

Send the friend two things, over a secure channel (Signal / password manager /
1Password secret link — never email plaintext or commit):

1. **Origin** — e.g. `https://relay.example.org`
2. **Their deploy key token** — the value shown above

Optionally also send the hosted docs/skill URLs:

- `https://relay.example.org/_relay/deploy.html`
- `https://relay.example.org/_relay/skills/relay-deploy/SKILL.md`

### Manage keys via the API

The console is one front-end for the admin API. The same actions are plain
HTTP, authenticated with the **admin token** (the master key):

```sh
# list keys + sites
curl -H "Authorization: Bearer $RELAY_ADMIN_TOKEN" \
  "https://relay.example.org/_relay/v1/admin"

# create a deploy key for a user (returns the token)
curl -X POST \
  -H "Authorization: Bearer $RELAY_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user":"alice"}' \
  "https://relay.example.org/_relay/v1/admin/keys"

# rotate a key (returns a new token; the old one stops working)
curl -X POST \
  -H "Authorization: Bearer $RELAY_ADMIN_TOKEN" \
  "https://relay.example.org/_relay/v1/admin/keys/dk_…/rotate"

# revoke a key
curl -X POST \
  -H "Authorization: Bearer $RELAY_ADMIN_TOKEN" \
  "https://relay.example.org/_relay/v1/admin/keys/dk_…/revoke"
```

All return JSON. Errors use the same envelope as uploads (`bad_token` 401,
`bad_request` 400, `not_found` 404).

### Revoke or rotate a key

- **Revoke** immediately stops that key from uploading. Idempotent — revoking
  an already-revoked key is a no-op.
- **Rotate** revokes the old key and mints a fresh one for the same user,
  showing the new token. Use this when a key leaks, or just on schedule.

A revoked key cannot upload, but sites already deployed stay live — revocation
only cuts future uploads.

---

## Part 3 — Friend: upload a site

### Prerequisites

- The node's **origin** and a **deploy key token** (from the operator).
- A **site name**: lowercase letters, digits, and interior hyphens, 1–63
  characters (e.g. `docs`, `my-blog`, `zine-42`).

### 1. Put the site in a folder

```
dist/
├── index.html
└── assets/
    └── style.css
```

Keep asset paths relative so the site works at any path depth.

### 2. Create the archive

Archive the **contents** of the folder (not the folder itself):

```sh
tar -C ./dist -cf site.tar .
```

### 3. Upload

```sh
curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "https://relay.example.org/_relay/v1/sites/nick-site/deploy"
```

### 4. Check the result

On success (HTTP 200):

```json
{ "site": "nick-site", "files": 3, "bytes": 5120, "url": "/mynode/nick-site/" }
```

The site is live immediately at `{origin}{url}`:

```
https://relay.example.org/mynode/nick-site/
```

The operator console now lists `nick-site` under this friend's user.

### Or use the CLI

If you have the `relay` binary, the same upload is one command:

```sh
RELAY_DEPLOY_TOKEN=… relay deploy \
  --origin https://relay.example.org --site nick-site ./dist
```

---

## Roll back a site

Deploys keep the last 10 versions, so a bad deploy is reversible.

```sh
# list a site's versions
curl -H "Authorization: Bearer $RELAY_ADMIN_TOKEN" \
  https://relay.example.org/_relay/v1/admin/sites/nick-site/versions

# roll back to the previous version
RELAY_ADMIN_TOKEN=… relay rollback \
  --origin https://relay.example.org --site nick-site

# or to a specific version id
RELAY_ADMIN_TOKEN=… relay rollback \
  --origin https://relay.example.org --site nick-site --to <version-id>
```

Rollback snapshots the current live site first, so you can roll back and
forth without losing anything.

---

## Using an agent instead

Point the agent at the hosted skill URL and hand it the origin and the deploy
key token:

```
https://relay.example.org/_relay/skills/relay-deploy/SKILL.md
```

The skill contains the full upload procedure; the agent should stop and ask if
the origin, token, or site name is missing rather than guessing.

---

## Archive rules (the node rejects anything else)

- **Relative paths only.** No `..`, no absolute paths, no NUL bytes.
- **Regular files and directories only.** No symlinks, hard links, or special
  entries.
- At most **65536 files** per deploy and **256 MiB** of archive bytes.
- A successful upload **atomically replaces the whole site** — send everything
  the site needs every time, not only changed files.
- Directory entries are optional: a regular file creates its parent
  directories.

## Error codes

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | Missing, wrong, or revoked deploy key (or admin token). |
| `bad_name` | 400 | The `{site}` path segment is not a valid name. |
| `bad_archive` | 400 | The body is not a tar archive, or an entry breaks the rules above. |
| `too_large` | 413 | The archive exceeds the node's size cap. |

Failures return a JSON envelope:

```json
{ "error": { "code": "bad_token", "message": "invalid or missing deploy token" } }
```

Read the `code`; don't retry blindly.

---

## Credentials and safety

- The **admin token** is the operator's master credential: it issues, revokes,
  and rotates every deploy key. Keep it out of shell history, logs, and chat.
  If it leaks, rotate it by restarting the node with a new `--admin-token`
  (and then rotate the deploy keys too, out of caution).
- **Deploy keys are per-user**, and their tokens are stored on the node's
  operator-only console (the `<data>/deploy-keys.json` store is `0600`), so
  you can reveal and copy a token again later. The token is still a secret:
  hand it out over a secure channel, and rotate it if it leaks.
- In this MVP a deploy key can upload to **any** site on the node; sites are
  attributed to the uploading user but a key may overwrite someone else's
  site. Per-site ownership arrives with the control API.
- Relay's long-term identity is an **Ed25519 key** — the deed to a name. Keys
  and signed deploys arrive with the control API; never send or sign with a
  private key.

### Rotating the admin token

Restart the node with a new value; it overwrites `<data>/admin-token`:

```sh
docker run -d ... ghcr.io/nemoNoboru/relay:latest --admin-token "$(openssl rand -base64 32)"
```
