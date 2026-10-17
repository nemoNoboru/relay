# Relay — User's Guide

Host a static site on a Relay node in three steps: run the node, issue a
deploy key, upload a tar archive. Sites go live at `{origin}/{node}/{site}/`.

| Piece | URL / value |
|---|---|
| Operator console | `{origin}/_relay/admin.html` |
| Upload endpoint | `POST {origin}/_relay/v1/sites/{site}/deploy` |
| Hosted upload docs | `{origin}/_relay/deploy.html` |

There are **two credentials**, and they are not interchangeable:

| Credential | Held by | Unlocks |
|---|---|---|
| **Admin token** | the operator | the console: issue, revoke, rotate deploy keys |
| **Deploy key** | each uploader | uploading a site (bearer token) |

---

## 1. Run the node (operator)

**Docker:**

```sh
docker run -d --name relay -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemonoboru/relay:latest
```

**Binary:** `relay node --name mynode --data /var/lib/relay --listen :8080`

On first start the node writes the **admin token** to `<data>/admin-token`
and prints it once to the log (`docker logs relay`). Set your own with
`--admin-token`. For HTTPS, pass `--tls-cert`/`--tls-key` or put a reverse
proxy in front. More options: [`docs/technicals.md`](docs/technicals.md).

## 2. Issue a deploy key (operator)

1. Open `{origin}/_relay/admin.html` and sign in with the admin token.
2. Enter the user's name, press **Create key**, and **copy the token**
   (the console can reveal and copy it again later).
3. Send the user, over a secure channel (Signal, a password manager — never
   email or a commit):
   - the **origin** — e.g. `https://relay.example.org`
   - their **deploy key token**

**Revoke** cuts future uploads instantly (deployed sites stay live);
**Rotate** revokes and mints a fresh token for the same user.

## 3. Upload a site (user)

You need the origin, a deploy key token, and a site name (lowercase letters,
digits, interior hyphens, 1–63 chars). Keep asset paths relative.

```sh
tar -C ./dist -cf site.tar .          # archive the *contents* of the folder

curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "https://relay.example.org/_relay/v1/sites/mysite/deploy"
```

```json
{ "site": "mysite", "files": 3, "bytes": 5120, "url": "/mynode/mysite/" }
```

The site is live immediately at `{origin}{url}`. With the `relay` binary the
same upload is `relay deploy --origin … --site mysite ./dist`.

## 4. Roll back

Deploys keep the last 10 versions; the live site is snapshotted first, so
nothing is lost. Use the console, or:

```sh
RELAY_ADMIN_TOKEN=… relay rollback --origin … --site mysite [--to <version-id>]
```

## Archive rules (the node rejects anything else)

- Relative paths only — no `..`, no absolute paths, no symlinks or hard links.
- Regular files and directories only; at most 65536 files and 256 MiB.
- Every upload **atomically replaces the whole site** — send everything, every time.

## Errors

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | Missing, wrong, or revoked token. |
| `bad_name` | 400 | Invalid site name. |
| `bad_archive` | 400 | Not a tar, or an entry breaks the rules. |
| `too_large` | 413 | Archive over the size cap. |

Failures return `{ "error": { "code", "message" } }` — read the `code`, don't
retry blindly.

## Safety

- The **admin token** is the master key to the node. Keep it out of history,
  logs, and chat; if it leaks, restart with a new `--admin-token`.
- A **deploy key** is a secret too — hand it out over a secure channel and
  rotate it if it leaks. In this MVP a deploy key can upload to *any* site on
  the node; per-site ownership arrives with the control API.

## Agents

Point the agent at the hosted skill and hand it the origin and token:
`{origin}/_relay/skills/relay-deploy/SKILL.md`.
