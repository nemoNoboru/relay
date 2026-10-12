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

How to run a node — as a binary, in Docker, on Docker Swarm, or on Kubernetes —
is in [`docs/technicals.md`](docs/technicals.md). There are no deploy files in the
repo yet, because there is no binary to deploy.

## Building

The two generated pages are rebuilt from their Markdown sources:

```sh
./tools/sync.sh           # rebuild manifesto.html and technicals.html
./tools/sync.sh --check   # exit 1 if either is out of date (CI / pre-commit)
```

## Go development

The Go side lives under `cmd/relay` and `internal/` (module
`github.com/nemoNoboru/relay`), built to the standard in
[`.pi/skills/go-architecture`](.pi/skills/go-architecture/SKILL.md).

```sh
make check        # gofmt, go vet, staticcheck, go test -race
make test         # go test -race ./...
make build        # build the relay binary
```

The same gates run in CI on every push to `main` and on pull requests
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

The [Pi harness](https://pi.dev) discovers both automatically from `.pi/skills/` and
applies the matching one to each kind of work; force one with
`/skill:go-architecture` or `/skill:frontend-architecture`.

## Status

**Go scaffolding in place.** The repo is design documents, static UI mockups, and
the first Go skeleton: `go.mod`, `cmd/relay` (`relay version`, `relay key new`),
and the pure domain packages `internal/domain/key` and `internal/domain/prefix`
with tests, CI, and pre-commit. No node runtime yet. Read
[`REVIEW.md`](REVIEW.md) before writing code.
