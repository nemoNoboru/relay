# Relay — User's Guide

Put a static site online with Relay: run a node, give someone a deploy key,
and they upload a folder. Sites are served at `{origin}/{node}/{site}/`.

Relay uses **two credentials**. Don't mix them up:

| Credential | Who holds it | What it does |
|---|---|---|
| **Admin token** | the node's operator | signs in to the console, issues and revokes deploy keys, rolls sites back |
| **Deploy key** | each person who uploads | uploads sites |

## 1. Run a node (operator)

```sh
# Docker. Flags after the image name go to `relay node`.
docker run -d --name relay -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemonoboru/relay:latest --name mynode

# or the binary
relay node --name mynode --data /var/lib/relay --listen :8080
```

`--name` is the first path segment of every site on this node. On first
start the admin password is **`relay`** (also written to
`<data>/admin-token`). Sign in at the console and change it before issuing
keys. To set one yourself from the start, pass `--admin-token`. For HTTPS,
pass `--tls-cert` and `--tls-key`, or run a reverse proxy in front.
Kubernetes and the other options are in
[`docs/technicals.md`](docs/technicals.md).

## 2. Issue a deploy key (operator)

1. Open `{origin}/_relay/admin.html` and sign in (password `relay` on a fresh node).
2. If asked, choose a new password (at least 8 characters).
3. Type the person's name (letters, digits, `.`, `_`, `-`) and press **Create key**.
4. Copy the **invite link** and send it over a secure channel (Signal, a password manager — never email or a commit).

People with an invite can invite others from the publish page; those keys are
labelled `name/guest`. **Revoke** stops a key and everyone it invited.
**Rotate** replaces a key with a fresh one for the same person (their invitees stay attached).

## 3. Upload a site

Pick a site name: lowercase letters, digits, and hyphens in the middle, up to
63 characters. Keep links and asset paths relative.

```sh
tar -C ./dist -cf site.tar .        # archive what's inside the folder

curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "https://relay.example.org/_relay/v1/sites/mysite/deploy"
```

```json
{ "site": "mysite", "files": 3, "bytes": 5120, "url": "/mynode/mysite/" }
```

The site is live at `{origin}{url}` as soon as the upload finishes. If you
have the `relay` binary, the same upload takes one line (the token comes from
`$RELAY_DEPLOY_TOKEN`):

```sh
relay deploy --origin https://relay.example.org --site mysite ./dist
```

## 4. Roll back (operator)

Every deploy keeps the last 10 versions. Rolling back saves the current live
version first, so you can always go forward again.

```sh
export RELAY_ADMIN_TOKEN=…
relay rollback --origin https://relay.example.org --site mysite            # previous version
relay rollback --origin https://relay.example.org --site mysite --to <id>  # a specific version
```

To list a site's version ids, send
`GET /_relay/v1/admin/sites/mysite/versions` with the admin token.

## Archive rules

The node rejects any upload that breaks these:

- Relative paths only. No `..`, no absolute paths, no symlinks or hard links.
- Regular files and directories only, up to 65,536 files and 256 MiB.
- Every upload **replaces the whole site**, so include every file each time.

## Errors

Failed uploads return `{ "error": { "code", "message" } }`. Check the `code`
before retrying.

| `code` | HTTP | Meaning |
|---|---|---|
| `bad_token` | 401 | The token is missing, wrong, or revoked. |
| `bad_name` | 400 | The site name isn't valid. |
| `bad_archive` | 400 | The body isn't a tar archive, or an entry breaks the archive rules. |
| `too_large` | 413 | The archive is over the size limit. |

## Keep credentials safe

- The **admin token** controls the whole node. Keep it out of shell
  history, logs, and chat. If it leaks, restart the node with a new
  `--admin-token`.
- A **deploy key** is also a secret. In this early version, any deploy key
  can overwrite *any* site on the node, so only give keys to people you
  trust. Per-site ownership is planned.

## Deploy with an agent

Give the agent the origin, a deploy key, and this skill:
`{origin}/_relay/skills/relay-deploy/SKILL.md`.
