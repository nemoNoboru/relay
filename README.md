# Relay

**Put your site on the web.** One upload and your site is live at a real
address: no account, no build step, and no platform in between. Relay is
free hosting software that any collective, club or neighbour can run, so the
web stops being rented from a handful of companies.

## Deploy a site

The easy way: the operator copies an **invite link** from the keys page
(`/_relay/admin.html`) and sends it to you. Open it, drop your pages in, and
press Publish. You can also hand the link to an AI
assistant and ask it to publish your site.

From a terminal:

You need a Relay node's address (`$RELAY_ORIGIN`) and a deploy key from the
person who runs it (`$RELAY_DEPLOY_TOKEN`):

```sh
tar -C ./dist -cf site.tar .
curl -X POST \
  -H "Authorization: Bearer $RELAY_DEPLOY_TOKEN" \
  -H "Content-Type: application/x-tar" \
  --data-binary @site.tar \
  "$RELAY_ORIGIN/_relay/v1/sites/mysite/deploy"
```

Your site is now live at `$RELAY_ORIGIN/<node>/mysite/`. With the `relay`
binary, the same upload is
`relay deploy --origin $RELAY_ORIGIN --site mysite ./dist`.

## Run a node

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh
relay node --name mynode --data /var/lib/relay --listen :8080
```

Docker and Kubernetes each take one command; see
[`docs/technicals.md`](docs/technicals.md). On first start the node prints
an **admin token**. Use it to sign in to the operator console at
`/_relay/admin.html`, where you issue deploy keys and see who last
deployed each site. Every node also serves this project's landing page at `/`.

## Docs

| Doc | What's in it |
|---|---|
| [`users-guide.md`](users-guide.md) | **Start here.** Run a node, issue keys, upload, and roll back |
| [`docs/technicals.md`](docs/technicals.md) | How it works, what's built, and every way to deploy a node |
| [`docs/vision.md`](docs/vision.md) | Why Relay exists |
| [`REVIEW.md`](REVIEW.md) | The design review, with locked decisions and open questions. **Read it before writing code.** |
| [`docs/keys.md`](docs/keys.md) · [`docs/api-contract.md`](docs/api-contract.md) · [`docs/architecture.md`](docs/architecture.md) | Ownership keys, the wire contract, and the backend design |
| [`web/README.md`](web/README.md) | The owner UI (Svelte + Vite, embedded in the binary) |

## Develop

```sh
make check   # every gate: site in sync, web build + tests, gofmt/vet/staticcheck/test -race
make test    # go test -race ./...
make build   # owner UI first, then the relay binary
```

`index.html` is written by hand. `manifesto.html` and `technicals.html` are
generated from `docs/vision.md` and `docs/technicals.md`. Edit the Markdown,
then run `python3 tools/build.py`; `make check` fails if they drift.
Engineering standards for Go and the frontend live in [`.pi/skills/`](.pi/skills),
alongside the deploy skill that every node serves at
`/_relay/skills/relay-deploy/SKILL.md`.

## Status

**Stage 0 (MVP).** Built:

- static hosting at `/<node>/<site>/`
- `tar` uploads authenticated with per-user deploy keys
- atomic deploys that keep the last 10 versions for rollback
- the operator console and native HTTPS
- the `relay` CLI: `node`, `deploy`, `rollback`, `key new`

Not built yet: NAT tunnels, per-site ownership and signed deploys, per-site
databases, and login and stats. See [`REVIEW.md`](REVIEW.md).
