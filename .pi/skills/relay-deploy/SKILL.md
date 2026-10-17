---
name: relay-deploy
description: Upload a static website into a Relay node. Use when the task is to deploy, publish, or upload a static site (a folder of HTML, CSS, JS, images, and other assets) to Relay, or when a user hands you a Relay node address and a deploy token.
---

# Relay Deploy

Upload a finished static site — any folder of files — to a Relay node, where it
is served live at `/<node>/<site>/`. This skill covers the current (MVP) upload
endpoint: one HTTP request, authenticated with a per-user deploy key sent as a
bearer token.

## What you need before you start

1. **The node's origin** — the base URL you upload to, e.g.
   `https://relay.example.org`.
2. **A deploy key** — a per-user credential the node operator creates on the
   operator console. Use its token as a bearer token, never in the URL,
   never committed, never logged.
3. **A site name** — the path segment the site will live under. Lowercase
   letters, digits, and interior hyphens, 1–63 characters, no leading or
   trailing hyphen (e.g. `docs`, `my-blog`, `zine-42`).

If any of these is missing, stop and ask for it — do not invent a name or
guess a token.

## Steps

1. Build or collect the site into a single folder. Relative links work; keep
   asset paths relative so the site works at any path depth.
2. Create a tar archive of the **contents** of the folder (not the folder
   itself). Every entry must be a relative path such as `index.html` or
   `assets/style.css`. A leading `./` and a root `.` directory entry are fine.
3. Upload it:

   ```http
   POST {origin}/_relay/v1/sites/{site}/deploy
   Authorization: Bearer {token}
   Content-Type: application/x-tar
   ```

   The body is the raw tar bytes.
4. Read the response. On success (200) the JSON body is:

   ```json
   { "site": "docs", "files": 3, "bytes": 5120, "url": "/mynode/docs/" }
   ```

   The site is live at `{origin}{url}`.
5. On failure, the body is an error envelope — read the `code`, do not retry
   blindly:

   ```json
   { "error": { "code": "bad_token", "message": "invalid or missing deploy token" } }
   ```

## Creating the archive

A folder that contains `index.html` and `assets/style.css`:

```sh
tar -C ./site-folder -cf site.tar .
```

The `-C ./site-folder` makes paths relative to the folder, and `.` archives
its contents. When generating the archive in code, write regular files (and
directories) with relative paths and no `..` segments.

## Archive rules (the node rejects anything else)

- Relative paths only. No `..`, no absolute paths, no NUL bytes.
- Regular files and directories only. No symlinks, hard links, or special
  entries.
- At most 65536 files per deploy, and 256 MiB of archive bytes by default.
- A successful upload atomically **replaces** the whole site. Send everything
  the site needs every time; do not upload only changed files.
- Directory entries are optional: a regular file creates its parent
  directories.

## Error codes

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | Missing, wrong, or malformed deploy token. |
| `bad_name` | 400 | The `{site}` path segment is not a valid name. |
| `bad_archive` | 400 | The body is not a tar archive, or an entry breaks the archive rules. |
| `too_large` | 413 | The archive exceeds the node's size cap. |

## Credentials and safety

- The **deploy key is a secret**. Keep its token out of shell history, logs,
  commit messages, chat output, and files you commit. Prefer an environment
  variable (`RELAY_DEPLOY_TOKEN`) or a secret store.
- Relay's long-term identity is an **Ed25519 key**: the deed to a name. For
  this MVP the per-user deploy key is the credential that authorizes uploads;
  keys and signed deploys arrive with the control API. Never send or sign
  with a private key unless a later skill says so.
- A deploy key can upload to any site on the node (sites are attributed to
  the key's user, but a key may overwrite any site). Treat a leaked token as
  compromised and ask the operator to revoke or rotate it on the operator
  console.
- The **admin token** is separate and operator-only: it unlocks the operator
  console, not uploads. Never ask for it — a deploy key is all you need.

## Worked example (shell)

```sh
# Build the site into ./dist first, then:
tar -C ./dist -cf site.tar .
curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "$RELAY_ORIGIN/_relay/v1/sites/$RELAY_SITE/deploy"
```

Then check that `"$RELAY_ORIGIN$url"` returns the site.

If the `relay` binary is available, the same upload is one command (it reads
the token from `$RELAY_DEPLOY_TOKEN`):

```sh
relay deploy --origin "$RELAY_ORIGIN" --site "$RELAY_SITE" ./dist
```
