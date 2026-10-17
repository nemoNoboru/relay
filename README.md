# Relay

**The last mile for independent websites.** A folder becomes a website — one
command, no account, no platform. An address is owned by the key you hold, not
by a platform's permission.

## Deploy a site

On any Relay node (try one you know, or run your own):

```sh
tar -C ./dist -cf site.tar .
curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "$RELAY_ORIGIN/_relay/v1/sites/mysite/deploy"
```

Live at `{origin}/{node}/mysite/`. A deploy key comes from the node's
operator — or from the console at `/_relay/admin.html` on your own node. The
full flow is in [`users-guide.md`](users-guide.md).

## Run a node

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh
relay node --data /var/lib/relay --listen :8080
```

Or with Docker / Kubernetes — one command each in
[`docs/technicals.md`](docs/technicals.md). Every node also hosts this
project's landing page at its root. First start prints the **admin token**;
it unlocks the operator console at `/_relay/admin.html`, where deploy keys
are issued, sites are listed, and deploys are rolled back.

## Documentation

| Doc | What it is |
|---|---|
| [`users-guide.md`](users-guide.md) | Run a node, issue keys, upload, roll back — start here |
| [`docs/vision.md`](docs/vision.md) | Why Relay exists (source of `manifesto.html`) |
| [`docs/technicals.md`](docs/technicals.md) | How it works + deployment (source of `technicals.html`) |
| [`REVIEW.md`](REVIEW.md) | The design review: locked decisions, open questions. **Read before writing code.** |
| [`docs/keys.md`](docs/keys.md) · [`docs/api-contract.md`](docs/api-contract.md) · [`docs/architecture.md`](docs/architecture.md) | Ownership keys, the wire contract, the backend design |
| [`web/README.md`](web/README.md) | The owner UI (Svelte + Vite, embedded in the node binary) |

## Develop

```sh
make check   # site in sync + web gates + Go gates (gofmt, vet, staticcheck, test -race)
make test    # go test -race ./...
make build   # owner UI first, then the relay binary
```

`index.html` is hand-written; `manifesto.html` and `technicals.html` are
generated — edit their `docs/*.md` sources and run `python3 tools/build.py`
(`make check` fails when they drift). Engineering standards live in
[`.pi/skills/`](.pi/skills) (Go, frontend, and the deploy skill, also hosted
by every node at `/_relay/skills/relay-deploy/SKILL.md`).

## Status

**Stage 0 (MVP).** Built: static hosting at `/<node>/<site>/`, tar uploads
with per-user deploy keys, atomic deploys with the last 10 versions kept for
rollback, the operator console, native HTTPS, and the `relay` CLI
(`deploy`, `rollback`, `key new`, `node`). Not built yet: NAT tunnels,
per-site databases, login/stats, signed deploys. See [`REVIEW.md`](REVIEW.md).
