---
name: go-architecture
description: Go engineering standard for this repository, covering idiomatic Go, clean code, SOLID, layered package layout, and choosing domain-driven vs data-driven design per feature. Use when writing, reviewing, or refactoring Go code, deciding where a new package, type, or piece of logic belongs, or adding persistence to a Relay feature.
---

# Go Architecture

The engineering standard for Relay's Go codebase. Relay's product decisions are locked in
`REVIEW.md` (one Go binary, path-based addressing, per-site SQLite, outbound WebSocket
tunnels, Ed25519 ownership keys). This skill governs **how the code is written**: idioms,
structure, and where logic lives.

## Route to the right reference first

- New feature, package, or layer question → `references/solid-architecture.md`
- Writing functions, naming, errors, comments, tests → `references/clean-code-go.md`
- Concurrency, context, interfaces, structs, stdlib usage → `references/go-idioms.md`
- "Where does this logic go?" model vs service vs handler, adding persistence →
  `references/ddd-vs-data-driven.md`

Read the matching reference before writing or reviewing the code. Do not skip this step for
"small" changes — the routing rule that keeps the codebase healthy is cheap to follow.

## Non-negotiables

1. **gofmt is law.** `gofmt -l .` must be empty. `go vet ./...` and `staticcheck ./...`
   must be clean. No exceptions for "temporary" code.
2. **Errors are values.** Never discard an error (`_ = f()`), never `panic` outside `main`
   and tests for control flow. Wrap at every layer that adds context:
   `fmt.Errorf("deploy site %q: %w", name, err)`. Error strings: lowercase, no trailing
   punctuation.
3. **`context.Context` is the first parameter** of anything that blocks, does I/O, or crosses
   a call boundary. Never store a context in a struct.
4. **Goroutines have visible lifetimes.** The spawner owns shutdown: cancellation via
   context, a waitgroup, or a channel the spawner closes. No fire-and-forget.
5. **Interfaces are small, live on the consumer side, and appear only when needed** (a
   second implementation or a test). Producer packages return concrete types.
6. **The domain layer imports no I/O.** `internal/domain` (and the domain half of each
   feature) has no `database/sql`, `net/http`, `os`, or CLI-flag imports. Interfaces are
   declared by the feature that uses them; their implementations live in `internal/infra`.
7. **No mutable global state.** Configuration is built once in `cmd/relay` and passed in.
8. **Table-driven tests** for branchy logic; **integration tests** for anything crossing a
   real boundary (SQLite file, WebSocket tunnel between two in-process nodes).
9. **Names carry the domain vocabulary** from `REVIEW.md`: node, site, prefix, deploy,
   key, tunnel, friend. Initialisms are consistent: `URL`, `ID`, `HTTP`, `TLS`, never `Url`.
10. **Boring code wins.** Prefer the standard library and the proven choices already locked
    in `REVIEW.md` (Caddy/autocert, SQLite, Ed25519, Wazero). Every new external
    dependency needs a one-line justification in the PR.

## Package layout

```
cmd/relay/            main(): flags, wiring, exit codes. No logic.
internal/
  domain/             pure business model: site, deploy, prefix, key. No I/O imports.
  <feature>/          application layer per feature: tunnel, control, runtime, auth,
                      metrics, rules. Declares the interfaces it needs.
  infra/              implementations: sqlite, blobs, caddy/autocert, ws transport,
                      wazero, laya provider.
pkg/                  only if genuinely importable by other modules.
```

**Dependency rule:** `cmd → feature → domain`, and `feature → infra` only through
interfaces the feature declares. `domain` imports nothing else in the repo.

## Domain-driven vs data-driven: the default stance

Relay is not a CRUD app. Its value is in the *rules* — who owns what, how a deploy is
signed, how a prefix is routed, what a rollback means. So:

- **Default: domain-driven (rich model).** Invariants live on domain types
  (`site.Deploy`, `prefix.Claim`). A handler that can violate an invariant is a missing
  model, not an oversight.
- **Data-driven is fine for plumbing** that has no rule to protect: static file serving,
  the blob store, relaying bytes over a tunnel. Do not build a model on top of plumbing.
- The mix is per bounded context, decided feature by feature — see
  `references/ddd-vs-data-driven.md` before adding persistence or a "manager"/"service"
  layer, or before writing any handler that mutates state.

## Definition of done (every PR)

- [ ] `gofmt -l .` empty; `go vet ./...` and `staticcheck ./...` clean
- [ ] New domain logic has table-driven unit tests; new boundary behavior has an
      integration test
- [ ] No new `internal/infra`, `net/http`, or `database/sql` import inside `internal/domain`
- [ ] No new external dependency without a stated reason
- [ ] Every error from a changed code path is wrapped with context at each layer
- [ ] Goroutines spawned by the change have an owner and an exit
