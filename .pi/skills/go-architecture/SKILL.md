---
name: go-architecture
description: Go engineering standard for this repository, covering idiomatic Go, clean code, SOLID, layered package layout, and choosing domain-driven vs data-driven design per feature. Use when writing, reviewing, or refactoring Go code, deciding where a new package, type, or piece of logic belongs, or adding persistence to a Relay feature.
---

# Go Architecture

The engineering standard for Relay's Go codebase. Relay is a micro-social
network for a group of friends: one Go binary serves a node's static posts,
invitation-based membership, and (in progress) a small social layer. The
product definition is [`docs/product.md`](../../../docs/product.md); the current
architecture is [`docs/architecture.md`](../../../docs/architecture.md); the
plan is [`docs/pivot-plan.md`](../../../docs/pivot-plan.md).

This skill governs **how the code is written**: idioms, structure, and where
logic lives. It does not override the product or the plan.

## Route to the right reference first

- New feature, package, or layer question → `references/solid-architecture.md`
- Writing functions, naming, errors, comments, tests → `references/clean-code-go.md`
- Concurrency, context, interfaces, structs, stdlib usage → `references/go-idioms.md`
- "Where does this logic go?" model vs service vs handler, adding persistence →
  `references/ddd-vs-data-driven.md`

Read the matching reference before writing or reviewing the code. Do not skip this
step for "small" changes — the routing rule that keeps the codebase healthy is
cheap to follow.

## Non-negotiables

1. **gofmt is law.** `gofmt -l .` must be empty. `go vet ./...` and `staticcheck ./...`
   must be clean. No exceptions for "temporary" code.
2. **Errors are values.** Never discard an error (`_ = f()`), never `panic` outside
   `main` and tests for control flow. Wrap at every layer that adds context:
   `fmt.Errorf("deploy site %q: %w", name, err)`. Error strings: lowercase, no
   trailing punctuation.
3. **`context.Context` is the first parameter** of anything that blocks, does I/O, or
   crosses a call boundary. Never store a context in a struct.
4. **Goroutines have visible lifetimes.** The spawner owns shutdown: cancellation via
   context, a waitgroup, or a channel the spawner closes. No fire-and-forget.
5. **Interfaces are small, live on the consumer side, and appear only when needed** (a
   second implementation or a test). Producer packages return concrete types.
6. **The domain layer imports no I/O.** `internal/domain` has no `database/sql`,
   `net/http`, `os`, or CLI-flag imports. Interfaces are declared by the feature that
   uses them; implementations live beside the feature or in `internal/infra`.
7. **No mutable global state.** Configuration is built once in `cmd/relay` and passed
   in. A shared store is a dependency, not a package variable.
8. **Table-driven tests** for branchy logic; **integration tests** for anything crossing
   a real boundary (a temp data directory, an HTTP handler, a file store).
9. **Names carry the domain vocabulary** from `docs/product.md`: node, member,
   invitation, page/post, version, host. Initialisms are consistent: `URL`, `ID`,
   `HTTP`, `TLS`, never `Url`. Keep the Go code's existing `site` vocabulary for a
   post's path segment; "post" is the social record around it.
10. **Boring code wins.** Prefer the standard library. The current stack is
    `net/http`, the filesystem, and JSON. A new dependency needs a one-line
    justification in the PR; a database needs a measured reason.

## Package layout

```
cmd/relay/            main(): flags, wiring, exit codes. No logic.
internal/
  domain/             pure business model: names, reserved paths. No I/O imports.
    key/              Ed25519 key file (CLI primitive, not the invite credential)
    prefix/           node path prefix model
  route/              request path → node/site/rest. Pure.
  serve/              HTTP: sites, upload, versions, keys, invites, admin, UI.
  social/             (planned) members, posts, comments, templates, JSON store.
```

**Dependency rule:** `cmd → serve/social → domain`. `serve` may use `social`;
`social` must not import `net/http`. `domain` imports nothing else in the repo.
The filesystem is the persistence boundary; the social store interface belongs to
`internal/social` and its file implementation lives beside it (or in
`internal/infra` when a second implementation appears).

## Domain-driven vs data-driven: the default stance

Relay is not a CRUD app. Its value is in the *rules* — who is a member, what an
invitation can do, what a post is, what a rollback means. So:

- **Default: domain-driven (rich model) for identity and content rules.**
  Invariants live on types (`KeyStore`, `Invitation`, `Post`, `Version`). A handler
  that can violate an invariant is a missing model, not an oversight.
- **Data-driven is fine for plumbing** that has no rule to protect: static file
  serving, tar extraction, JSON persistence, the registry map. Do not build a model
  on top of plumbing.
- The mix is per bounded context, decided feature by feature — see
  `references/ddd-vs-data-driven.md` before adding persistence or a "manager"/"service"
  layer, or before writing any handler that mutates state.

## The pivot boundary

The social layer is additive. Do not rewrite the hosting core:

- Static serving must not depend on `posts.json` or `comments.json`.
- The social index must be rebuildable from the site files and `relay-post.json`.
- Template JavaScript never runs in the node process.
- The member cookie and the admin token are separate credentials.
- No tunnel, NAT traversal, per-site SQLite runtime, classifier, or WASM
  sandbox. Those were drafts, not this product.

## Definition of done (every PR)

- [ ] `gofmt -l .` empty; `go vet ./...` and `staticcheck ./...` clean
- [ ] New domain logic has table-driven unit tests; new boundary behavior has an
      integration test
- [ ] No new `net/http`, `os`, or filesystem import inside `internal/domain`
- [ ] No new external dependency without a stated reason
- [ ] Every error from a changed code path is wrapped with context at each layer
- [ ] Goroutines spawned by the change have an owner and an exit
- [ ] Docs updated: `docs/api.md` for a wire change, `docs/architecture.md` for a
      structural change
