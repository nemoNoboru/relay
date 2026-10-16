# Relay

**The last mile for independent websites.** A folder becomes a website. A computer
behind a NAT becomes reachable. An address is owned by the key you hold — not by a
platform's permission.

> **Start here → [`REVIEW.md`](REVIEW.md).** The consolidated design review: what
> Relay is, the decisions that are locked, the open questions that need answering,
> and everything that is not finished before development.

---

## Site pages

| File | What it is |
|---|---|
| [`index.html`](index.html) | The landing page |
| [`manifesto.html`](manifesto.html) | The vision. **Generated** from [`docs/vision.md`](docs/vision.md) |
| [`technicals.html`](technicals.html) | How it works and how to deploy it. **Generated** from [`docs/technicals.md`](docs/technicals.md) |
| [`node-ui.html`](node-ui.html) | Draft: the Relay node admin UI (static mockup) |
| [`metrics.html`](metrics.html) | Draft: the per-site owner console (static mockup) |
| [`web/`](web/README.md) | **The owner UI itself** (Svelte + Vite build, embedded in the node binary) — replaces the two mockups above |

## Design documents

| Doc | What it covers |
|---|---|
| [`REVIEW.md`](REVIEW.md) | **The one to read.** Decisions, open questions, and the pre-build checklist |
| [`docs/vision.md`](docs/vision.md) | The vision. Source of truth for `manifesto.html` |
| [`docs/technicals.md`](docs/technicals.md) | Technical overview and deployment. Source of truth for `technicals.html` |
| [`docs/keys.md`](docs/keys.md) | Ownership keys: the two halves, signatures, fingerprints, the key file format, and how to guard a key |
| [`docs/api-contract.md`](docs/api-contract.md) | **Draft:** the wire contract — tunnel protocol (WebSocket + yamux), control API, runtime API |
| [`docs/architecture.md`](docs/architecture.md) | **Draft:** backend code architecture — embedded Caddy, tunnel transport, package layout |

Everything else was consolidated into `REVIEW.md` so there is a single document to
review. The detailed specs (classifier, addressing, threat model, CLI) can be
re-expanded from there, section by section, as each stage begins; the API
contract and the backend architecture are already drafted in `docs/`.

## Deployment

Install the `relay` binary locally (Linux or macOS, amd64/arm64):

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh
```

The installer downloads a prebuilt binary from GitHub Releases into
`~/.local/bin` (override with `RELAY_INSTALL_DIR`) and falls back to building
from source when no release asset matches. Prebuilt binaries are published on
every `v*` tag (`.github/workflows/release.yml`). Then run a node:

```sh
relay node --name <node> --data <dir> --listen :8080
```

How to run a node — as a binary, in Docker, on Docker Swarm, or on Kubernetes —
is in [`docs/technicals.md`](docs/technicals.md). A container image is also
built and published to GitHub Packages on every push to `main` and on `v*`
tags (`.github/workflows/image.yml`):

```sh
docker pull ghcr.io/nemoNoboru/relay:latest
docker run -d -p 8080:8080 -v relay-data:/data ghcr.io/nemoNoboru/relay:latest
```

The image runs unprivileged and keeps all state under `/data`. The node prints
its admin token on first start; set one explicitly with `--admin-token` via the
image's entrypoint (e.g.
`docker run ... ghcr.io/nemoNoboru/relay:latest --admin-token secret`). The
admin token unlocks the operator console at `/_relay/admin.html`, where the
operator issues per-user deploy keys and revokes or rotates them. To serve
HTTPS directly, pass `--tls-cert` and `--tls-key`; otherwise terminate TLS at
a reverse proxy in front of the node.

### Kubernetes

One command installs a `relay` namespace, a one-replica StatefulSet with a
persistent volume, and a Service — see
[`deploy/kubernetes.yaml`](deploy/kubernetes.yaml):

```sh
kubectl apply -f https://raw.githubusercontent.com/nemoNoboru/relay/main/deploy/kubernetes.yaml
```

The node serves HTTP on `:8080` with liveness and readiness probes on
`/healthz`; all state lives in the `/data` volume. Expose it with TLS at the
cluster edge by filling in your hostname in
[`deploy/ingress.yaml`](deploy/ingress.yaml) and applying it. On first start
the node writes its admin token to `/data/admin-token` (and prints it once to
the logs); read it with either of:

```sh
kubectl logs -n relay relay-0
kubectl exec -n relay relay-0 -- cat /data/admin-token
```

## Building

The two generated pages are rebuilt from their Markdown sources:

```sh
./tools/sync.sh           # rebuild manifesto.html and technicals.html
./tools/sync.sh --check   # exit 1 if either is out of date (CI / pre-commit)
```

The owner UI (`web/`) is built with Svelte + Vite and embedded into the node binary via
`web/embed.go`; `make build` keeps the order right (frontend first, then
`go build`). `make check` runs every gate that works without a node binary —
the static site in sync, the frontend's typecheck, build, tests, and bundle
smoke test, and the Go gates (gofmt, vet, staticcheck, test -race). See
[`web/README.md`](web/README.md) for the frontend layout, scripts, and its
draft wire contracts.

## Go development

The Go side lives under `cmd/relay` and `internal/` (module
`github.com/nemoNoboru/relay`), built to the standard in
[`.pi/skills/go-architecture`](.pi/skills/go-architecture/SKILL.md).

```sh
make check        # all gates that work without a node binary: site, web, Go
make test         # go test -race ./...
make build        # build the relay binary (builds the owner UI first)
```

The Go gates also run in CI on every push to `main` and on pull requests
(`.github/workflows/test.yml`), and a fast subset runs locally before each
commit via [pre-commit](https://pre-commit.com):

```sh
pip install pre-commit
pre-commit install   # once, per clone; needs a Go toolchain on PATH
```

`relay key new` generates the Ed25519 ownership key — the deed to a Relay
address. Guard key files like passwords; with `-o` they are written `0600`.
The [ownership keys guide](docs/keys.md) explains how keys work.

## Agent skills

- The Go engineering standard lives in
  [`.pi/skills/go-architecture`](.pi/skills/go-architecture/SKILL.md) — idiomatic Go,
  clean code, SOLID, package layout, and the domain-driven vs data-driven rule for each
  feature.
- The frontend standard lives in
  [`.pi/skills/frontend-architecture`](.pi/skills/frontend-architecture/SKILL.md) —
  functional programming as a first-class paradigm, Effect-TS as the functional library
  of choice, TypeScript, vanilla DOM, and the live-data patterns for the node UI and
  site console.
- The deploy skill lives in
  [`.pi/skills/relay-deploy`](.pi/skills/relay-deploy/SKILL.md) — instructions for an
  agent to upload a static site into a Relay node. The node also serves it at
  `/_relay/skills/relay-deploy/SKILL.md`, so a hosted node can hand the URL to an
  agent directly.

The [Pi harness](https://pi.dev) discovers both automatically from `.pi/skills/` and
applies the matching one to each kind of work; force one with
`/skill:go-architecture` or `/skill:frontend-architecture`.

## Status

**Stage 0 in progress (Go).** The owner UI ([`web/`](web/README.md)) and the
Go skeleton — `go.mod`, `cmd/relay` (`relay version`, `relay key new`), and
the pure domain packages `internal/domain/{key,prefix}` with tests, CI, and
pre-commit — are in place. The node runtime serves static files and accepts
uploads: `relay node --name <node> --data <dir> --listen <addr>` serves the
site folders under `<data>/sites/` at `/<node>/<site>/`, and
`POST /_relay/v1/sites/<site>/deploy` (a `tar` archive with a bearer deploy
key) atomically replaces a site and attributes it to the key's user, keeping
the last 10 versions for rollback. The
operator console at `/_relay/admin.html` (admin-token protected) issues,
revokes, and rotates per-user deploy keys, lists sites by owner, and rolls a
site back to a previous version. The CLI drives both sides: `relay deploy
--origin … --site … <folder>` and `relay rollback --origin … --site …`. The
node also hosts its owner UI and the upload documentation at `/_relay/` (see
[`web/deploy.html`](web/deploy.html)) and the deploy skill at
`/_relay/skills/relay-deploy/SKILL.md`. Keys, claim, signed deploys, TLS, and
the tunnel are not built yet. Read
[`REVIEW.md`](REVIEW.md) before writing code.
