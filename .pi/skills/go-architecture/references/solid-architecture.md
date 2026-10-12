# SOLID and Layering in Go

SOLID is not a class-hierarchy technique. In Go it is achieved with small types, small
consumer-side interfaces, and constructor injection — no inheritance, no DI container.
This reference states each principle in Go terms and then fixes Relay's layering.

## The five principles, in Go

### S — Single Responsibility

A type (and each function) has **one reason to change**.

- A `Router` that also persists sites and also formats HTTP responses is three types.
- Package-level check: if you can't state in one sentence why a package changes, split it.
- Relay example: `internal/tunnel` changes when the tunnel *protocol* changes;
  `internal/infra/ws` changes when the *transport* changes. Those are different reasons,
  so they are different packages.

### O — Open for Extension, Closed for Modification

Design extension points where the design doc already anticipates variants, and only there.

- Relay's clearest case is the **classifier**: the locked design has providers
  (`none`, `heuristic`, `laya-service`, `laya-onnx`). The feature declares one interface:

  ```go
  // internal/rules (consumer)
  type Classifier interface {
      Score(ctx context.Context, state RuleState, rules RuleSet) (Decision, error)
  }
  ```

  New providers land as new implementations in `internal/infra`; `internal/rules` never
  changes.
- Extension via **composition and configuration** (choose the implementation in `cmd/relay`),
  not via plugin machinery. No `plugin` package.
- Don't invent extension points "just in case". If there is one provider and no planned
  second, concrete type.

### L — Liskov Substitution

Every value that satisfies an interface must be substitutable without the caller noticing.

- Keep interfaces narrow so this is true: a `TunnelTransport` that says "it can be used
  anywhere a tunnel is expected" must not surprise with `nil` returns, blocking without
  ctx, or mutating its arguments.
- If a "variant" needs an `if type X` special case at a call site, the interface is too
  wide — split it.
- Concrete test: every implementation has the same integration test as every other
  implementation (same test file, parameterized). If one implementation needs a
  special-cased test, the substitution contract is broken or the interface is mis-shaped.

### I — Interface Segregation

Callers should depend only on what they use.

- The Go idiom: **many small interfaces, one big concrete type**. The site feature may declare
  `SiteStore` (Get/Put) and `SiteIndex` (List/PrefixSearch) separately even if
  `infra/sqlite.SiteStore` implements both.
- A five-method interface where a handler uses one is four methods of coupling it doesn't
  pay attention to.
- Corollary from the Go review wiki: don't define interfaces before they are used.

### D — Dependency Inversion

High-level policy depends on interfaces; low-level code implements them.

- **Interfaces are declared by the feature** (the high-level side); **implementations live
  in `internal/infra`**:

  ```
  internal/site (policy)          internal/infra (implementations)
  ┌───────────────────────┐       ┌────────────────────────────────┐
  │ type Store interface  │ ◄──── │ sqlite.SiteStore               │
  │ type Signer interface │ ◄──── │ ed25519 signer (stdlib)        │
  │ New(store, signer)    │       │ ws transport, caddy, ...       │
  └───────────────────────┘       └────────────────────────────────┘
          imports nothing from infra     implements the interfaces
  ```

- **Constructor injection, explicit and boring:** `site.NewService(store Store, signer Signer, clock func() time.Time)`.
  No global DI container, no service locator, no `init()` wiring. `cmd/relay/main.go` is
  the only place concrete types are chosen — it is the composition root.
- Inject **time** (`clock func() time.Time`) in code where "now" is an input to policy
  (signatures, sessions, salted daily metrics). Tests then stop being time-dependent.

## Relay layering (interfaces and implementations)

```
        HTTP / WebSocket / CLI                 (edge — transport only)
        ┌──────────────────────────────────────┐
        │  cmd/relay + internal/edge handlers  │  parse, sign-check, encode
        └───────────────┬──────────────────────┘
                        │ calls
        ┌───────────────▼──────────────────────┐
        │  feature application services        │  orchestrate one use case:
        │  internal/tunnel, control, runtime,  │  claim, deploy, serve, route,
        │  auth, metrics, rules                │  export. Thin. No business rules.
        └───────┬──────────────────────┬───────┘
                │ declares interfaces  │ calls
        ┌───────▼─────────────┐  ┌─────▼──────────────────────────┐
        │  internal/domain    │  │  internal/infra implementations│
        │  site, deploy,      │  │  sqlite, blobs, ws, caddy,      │
        │  prefix, key        │  │  wazero, laya provider          │
        │  (pure, no I/O)     │  │  (implement the interfaces,     │
        │                     │  │   own all the I/O)              │
        └─────────────────────┘  └────────────────────────────────┘

  Dependency arrows point inward and down: cmd → feature → domain;
  feature → infra only through the interfaces feature declared.
  domain imports nothing in the repo.
```

**What lives where** (the questions to ask when placing code):

| Question | Answer → location |
|---|---|
| Does it parse/serialize transport bytes? | edge handler |
| Does it orchestrate steps of one use case (load, check, save, notify)? | feature app service |
| Is it a rule the business will hold you to ("a deploy must be signed", "versions increase", "a released prefix cannot be re-claimed within the grace period")? | **domain type/method** |
| Does it touch SQLite, disk, network, TLS, WASM? | infra implementation |
| Does it choose which implementation, or build config? | `cmd/relay` |

**App service vs domain — the one-line test.** An app service method may be read as a
recipe: "load site, verify signature, apply change, save, log". If a sentence in that
recipe is a *business rule* ("a rolled-back version is not counted"), the rule has leaked
out of the domain — move it to a domain method and keep the recipe.

## Repository pattern: when, when not

- **Use** a repository-shaped interface when the domain has an aggregate with identity and
  invariants (sites with their deploys): `site.Store` with `Get/Put/Release`, methods named
  after *domain* operations, not queries.
- **Don't** build repositories for plumbing: static file serving needs a
  `FileServer`; the blob store is a content-addressed `Put/Get(hash)` — that's an API, not
  a repository. The tunnel is not a repository for anything.
- Repository methods return domain types, not rows. Row↔domain mapping is an infra
  detail.

## What not to do

- **Anemic "Manager" services.** A `SiteManager` with `UpdateStatus(id, status)` that
  contains the rules is the anemic model — see `ddd-vs-data-driven.md`.
- **Premature abstraction.** One implementation, zero planned second → concrete type.
- **Over-layering plumbing.** Static file serving does not get a domain layer.
- **God packages.** `internal/relay` containing everything is how this codebase dies.
  Feature packages are allowed to be small.
- **`init()` for wiring.** It runs in import order, hides dependencies, and makes tests
  fight the package. All wiring in `cmd/relay`.
- **Interface sprawl at the edge.** Handlers may take narrow interfaces, but the edge is
  transport: it may use `net/http` freely. The rule is about domain and feature purity,
  not about banning `net/http` everywhere.

## Verification (review checklist)

- [ ] Each new type has a one-sentence reason-to-change
- [ ] Each new interface has ≥1 real implementor and a concrete consumer
- [ ] No import of `internal/infra` (or I/O stdlib) in `internal/domain` or feature
      domain halves
- [ ] Composition root is `cmd/relay`; implementations are never constructed elsewhere
- [ ] Implementations are testable without the rest of the node (temp dir, in-process peers)
- [ ] An app-service method's sentences are all "recipe", not "rule"
