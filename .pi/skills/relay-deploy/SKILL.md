---
name: relay-deploy
description: Upload a static page, post, or microwebsite into a Relay node. Use when the task is to deploy, publish, or upload a folder of HTML, CSS, JS, images, and other assets to Relay, or when a user hands you a Relay node address and an invitation link or key.
---

# Relay Deploy

Upload a finished static page — any folder of files, or a single Markdown note —
to a Relay node, where it is served live at `/<node>/<page>/`.

Relay is a micro-social network for a group of friends. A node is a small server
someone in the group keeps. People join by invitation link, not by creating an
account. This skill covers the current upload endpoint: one HTTP request,
authenticated with the invitation key as a bearer token.

## What you need before you start

1. **The node's origin** — the base URL you upload to, e.g.
   `https://relay.example.org`.
2. **An invitation key** — the credential from the link the member was sent.
   Use it as a bearer token: never in the URL, never committed, never logged,
   never repeated back to the user.
3. **A page name** — the path segment the page will live under. Lowercase
   letters, digits, and interior hyphens, 1–63 characters, no leading or
   trailing hyphen (e.g. `sunday-walk`, `canal-photos`, `zine-42`).

If you were given an invite link such as
`https://relay.example.org/_relay/publish.html#key=…`, its scheme and host are
the origin, and the `key` value after `#` is the invitation key. The link may
also carry `&site=…`; use that name. Never repeat the key back in your output.

If any of these is missing, stop and ask for it — do not invent a name or
guess a token.

## Steps

1. Build or collect the page into a single folder. Relative links work; keep
   asset paths relative so the page works at any path depth.
2. Create a tar archive of the **contents** of the folder (not the folder
   itself). Every entry must be a relative path such as `index.html` or
   `assets/style.css`. A leading `./` and a root `.` directory entry are fine.
3. Upload it:

   ```http
   POST {origin}/_relay/v1/sites/{page}/deploy
   Authorization: Bearer {key}
   Content-Type: application/x-tar
   ```

   The body is the raw tar bytes.
4. Read the response. On success (200) the JSON body is:

   ```json
   { "site": "sunday-walk", "files": 3, "bytes": 5120, "url": "/my-group/sunday-walk/" }
   ```

   The page is live at `{origin}{url}`. The field is called `site` because the
   node's path segment is a site; to the group it is a page or a post.
5. On failure, the body is an error envelope — read the `code`, do not retry
   blindly:

   ```json
   { "error": { "code": "bad_token", "message": "invalid or missing deploy token" } }
   ```

## Creating the archive

A folder that contains `index.html` and `assets/style.css`:

```sh
tar -C ./page-folder -cf page.tar .
```

The `-C ./page-folder` makes paths relative to the folder, and `.` archives its
contents. When generating the archive in code, write regular files (and
directories) with relative paths and no `..` segments.

## A single Markdown note

The browser publish page turns one Markdown file into a styled page. From the
command line, render Markdown to `index.html` yourself (or upload the HTML you
already have). Do not upload a bare `.md` unless the page is designed to show
it as text.

## Archive rules (the node rejects anything else)

- Relative paths only. No `..`, no absolute paths, no NUL bytes, no
  backslashes.
- Regular files and directories only. No symlinks, hard links, or special
  entries.
- At most 65,536 files per publish, and 256 MiB of archive bytes by default.
- A successful upload atomically **replaces** the whole page. Send everything
  the page needs every time; do not upload only changed files.
- Directory entries are optional: a regular file creates its parent
  directories.

## Error codes

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | Missing, wrong, or revoked invitation key. |
| `bad_name` | 400 | The `{page}` path segment is not a valid name. |
| `bad_archive` | 400 | The body is not a tar archive, or an entry breaks the archive rules. |
| `too_large` | 413 | The archive exceeds the node's size cap. |

## Credentials and safety

- The **invitation key is a member's identity**. Keep it out of shell history,
  logs, commit messages, chat output, and files you commit. Prefer an
  environment variable (`RELAY_KEY`) or a secret store.
- There is no account, password, or email. If the key is lost or leaked, ask
  the host to rotate it; the old link stops working.
- A live key can currently publish any page on the node. Pages are attributed
  to the key's member, but per-page ownership is not built yet. Treat a leaked
  key as compromised and ask the host to revoke or rotate it.
- The **admin password** is separate and host-only. It unlocks the host
  console, not uploads. Never ask for it — an invitation key is all you need.
- The planned member cookie does not change this API: the bearer key remains
  valid for the CLI and assistants.

## Worked example (shell)

```sh
# Build or collect the page into ./dist first, then:
tar -C ./dist -cf page.tar .
curl -X POST \
  -H "Authorization: Bearer $RELAY_KEY" \
  -H "Content-Type: application/x-tar" \
  --data-binary @page.tar \
  "$RELAY_ORIGIN/_relay/v1/sites/$RELAY_PAGE/deploy"
```

Then check that `"$RELAY_ORIGIN$url"` returns the page.

If the `relay` binary is available, the same upload is one command (it reads
the key from `$RELAY_DEPLOY_TOKEN`):

```sh
relay deploy --origin "$RELAY_ORIGIN" --site "$RELAY_PAGE" ./dist
```

Post templates and the `/_relay/v1/publish` endpoint are planned. Until they
ship, this tar upload is the only publishing path. See
[`docs/publishing.md`](https://github.com/mansus-org/relay/blob/main/docs/publishing.md)
for the template contract.
