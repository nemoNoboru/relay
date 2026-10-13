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

The owner UI (`web/`) is built with Svelte + Vite and embedded into the node binary via
`web/embed.go`; `make build` keeps the order right (frontend first, then
`go build`). `make check` runs every gate that works without a node binary —
the static site in sync, plus the frontend's typecheck, build, tests, and
bundle smoke test. See [`web/README.md`](web/README.md) for the frontend
layout, scripts, and its draft wire contracts.

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

**Stage 0 in progress (Go).** Static file hosting works: `relay node
--name <node> --data <dir> --listen <addr>` serves the site folders under
`<data>/sites/` at `/<node>/<site>/` on a plain HTTP listener. The owner UI
([`web/`](web/README.md)) is implemented and will be embedded in the node
binary. Keys, claim, deploy, TLS, and the tunnel are not built yet. Read
[`REVIEW.md`](REVIEW.md) before writing code.
