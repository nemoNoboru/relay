# Domain-Driven vs Data-Driven Architecture

The two architectures a team actually picks between, what each buys and costs, how to
decide per bounded context, and what each looks like in Go for Relay.

Sources: Eric Evans, *Domain-Driven Design* (2003); Martin Fowler,
["Domain-Driven Design"](https://martinfowler.com/bliki/DomainDrivenDesign.html) and
["Anemic Domain Model"](https://martinfowler.com/bliki/AnemicDomainModel.html);
["Database-centric vs domain-centric architecture"](https://github.com/Dermendzhiev/database-vs-domain-centric-architecture);
["From CRUD to DDD"](https://codemonkeyprojectiles.com/blog/2021/CrudToDdd/).

## The two shapes

**Data-driven (data-centric, CRUD).** The *database schema* is the center of gravity.
For each entity there is a table, a record type that mirrors the columns, and
Create/Read/Update/Delete operations around it. Business logic is procedural: a set of
service functions or handlers that load records, mutate fields, and save. The UI and the
schema map one-to-one.

**Domain-driven.** The *business rules* are the center of gravity. The model is built
around the concepts, states, and transitions the business names — **ubiquitous language**:
the nouns and verbs of the domain (`claim`, `deploy`, `rollback`, `release`) appear as
type and method names. Data is a detail the model persists through interfaces declared by
the feature. The UI maps to
*tasks*, not tables.

## Data-driven: honest strengths

- Fast to first commit; trivial to describe ("one table, four endpoints").
- Matches the way customers and analysts think about records.
- Low skill floor; anyone on the team can change it.
- Perfect when the system's job *is* storing and serving data.

Its real cost shows up later, slowly:

- **Invariants are enforced in many places** — one handler checks the rule, another
  forgets. Correctness depends on every writer remembering.
- **The domain is coupled to storage.** A change in persistence details (a new SQLite
  index, a moved file layout) ripples into the logic that "just uses the fields".
- **Logic in services becomes procedural** — Fowler's "transaction scripts". The services
  carry all the computation and the model objects are "little more than bags of getters
  and setters" — the **anemic domain model** anti-pattern: "all of the costs of a domain
  model, without yielding any of the benefits".
- **Change gets expensive**: adding a rule means finding every path that mutates the
  record.

## Domain-driven: honest costs

- Slower start: the model must be named and the language agreed before the first table.
- Needs the domain to be *known*; Relay's is (it's written in `REVIEW.md`), but a CRUD
  admin panel over it would not earn the model.
- More concepts to keep in a team's head (aggregate, invariant, event).
- Mapping rich models to storage is deliberate work — in Relay this is small, because the
  storage choice (one SQLite file per site) already forces a clear repository boundary.

## Decision rules (use these, in order)

1. **Are there invariants the business will hold you to** — rules that must be true in
   every state, no matter which entry point mutates the data (signed deploys, monotonic
   versions, claimed-prefix uniqueness, session lifetimes)?
   → **Domain-driven.** Put each invariant on exactly one type.
2. **Is the data the product**, with no rules beyond storage semantics (a blob store,
   static files, a byte relay)?
   → **Data-driven.** A content-addressed store with `Put/Get(hash)` is the entire
   architecture. Stop.
3. **Is it a state machine** — objects with a small number of states and legal
   transitions (claim → released → reclaimed; deploy pending → live → rolled back)?
   → **Domain-driven**, even if the "model" is small. The transitions *are* the rules.
4. **Is it glue/orchestration** between other components (routing a request to a tunnel
   frame, forwarding, heartbeats)?
   → **Neither.** It's plumbing: synchronous functions, clear error flow, good logs.
   Do not dress plumbing up as a domain model.
5. **Mixed system?** The two styles coexist **per bounded context**, in the same binary.
   Relay: `prefix` and `site/deploy` are rich; `metrics` and `blobs` are data-driven;
   `tunnel` is plumbing with one protocol state machine.

The smell that flips #2 into #1: the same rule appears in a **second** place (two
handlers both checking "version must increase"). When an invariant is written twice,
it is not a check — it is a model. Extract it then.

## DDD vocabulary, mapped to Go

No classes, no framework — plain Go types:

| DDD concept | Go shape in this repo |
|---|---|
| Ubiquitous language | Package and type names from `REVIEW.md`: `site`, `deploy`, `prefix`, `friend`, `tunnel`, `key` |
| Entity | Struct with an identity field and methods; mutated through methods only (`*Site.Deploy(...)`) |
| Value object | Small immutable struct, no identity, compared by value: `FileHash`, `Address`, `Signature`, `RuleSet` |
| Aggregate | One root type (e.g. `Site`) that owns its invariants and hands out nothing inside the boundary but the root; persisted as a whole (one SQLite file per site is a natural aggregate boundary) |
| Invariant | A property no method on the aggregate allows to be violated; checked in the mutation method and, secondarily, at load time (corrupt data → `ErrCorrupt`) |
| Repository | An interface declared by the feature: `site.Store` with domain-named methods (`Get`, `Put`, `Release`); the SQLite implementation maps rows ↔ aggregate |
| Domain event | A plain struct + callback or a small in-process channel, used when "X happened" must trigger side effects (metrics count, friend re-route) without coupling the model to them |
| Application service | Thin orchestrator in the feature package: load → verify → mutate aggregate → store → log |
| Bounded context | A top-level feature package (`internal/site`, `internal/metrics`); each owns its own language and storage shape |

## Worked example: a deploy

**Data-driven / anemic** (what not to write):

```go
// handler
func (h *Handler) Deploy(w http.ResponseWriter, r *http.Request) {
    row := h.db.QueryRow("SELECT version FROM sites WHERE name=?", name).Scan(...)
    if ver := body.Version; ver != row+1 {                      // rule #1, in a handler
        http.Error(w, "bad version", 400)
    }
    if !h.verifier.Check(body.Manifest, body.Sig) {             // rule #2, in a handler
        http.Error(w, "bad signature", 400)
    }
    h.db.Exec("UPDATE sites SET manifest=?, version=? ...")
}
```

Every future entry point (CLI, tunnel, rollback path) must re-remember both rules.

**Domain-driven** (what to write):

```go
// internal/domain/site — no I/O imports
func (s *Site) Deploy(m Manifest, sig Signature, at time.Time) (Site, error) {
    if m.Version != s.Version+1 {
        return s, fmt.Errorf("%w: have %d, manifest %d", ErrVersionSkew, s.Version, m.Version)
    }
    if err := s.Key.Verify(m.CanonicalBytes(), sig, at); err != nil {
        return s, fmt.Errorf("%w: %v", ErrNotOwner, err)
    }
    s.Manifest = m
    s.Version = m.Version
    s.UpdatedAt = at
    s.history = append(s.history, m.Version)          // rollback candidates
    return s, nil
}
```

The app service becomes a recipe with no rules:

```go
// internal/site (feature)
func (svc *Service) Deploy(ctx context.Context, p DeployParams) error {
    s, err := svc.store.Get(ctx, p.Name)
    if err != nil {
        return fmt.Errorf("load site %q: %w", p.Name, err)
    }
    s, err = s.Deploy(p.Manifest, p.Signature, svc.clock())
    if err != nil {
        return err
    }
    if err := svc.store.Put(ctx, s); err != nil {
        return fmt.Errorf("persist site %q: %w", p.Name, err)
    }
    svc.events.Publish(Deployed{Name: p.Name, Version: p.Manifest.Version})
    return nil
}
```

Now the CLI, the HTTP handler, and the rollback path all call the same two methods and
cannot violate either rule. The invariant lives in exactly one place, and is unit-tested
without a database.

## When data-driven is the *right* call (Relay cases)

- **Blobs:** content-addressed `Put(hash, r io.Reader) / Get(hash)`; the "model" is the
  hash. A domain layer here is ceremony.
- **Static serving:** Caddy/stdlib does it; Relay's contribution is path mapping and
  per-site scoping (cookies, CSRF, `X-Forwarded-Prefix`) — that scoping *is* domain-ish
  and belongs in the edge, not a fake model.
- **Metrics:** aggregate counters with a daily salt. Data in, numbers out. Keep it a
  data store with one naming rule, and export = copy the file.
- **Tunnel byte relay:** a frame in, a frame out. The only state machine worth modeling
  is connection lifecycle (registered → active → draining), and it earns a small
  domain-type treatment; everything around it is plumbing.

## Transition notes

- **Start data-driven, extract on the second invariant.** It is fine to ship the first
  slice as plain CRUD if the model isn't settled; the moment a rule is written twice,
  extract the type. Never the reverse on a whim: don't pre-build rich models for features
  that are still mockups (`node-ui.html`, `metrics.html` are drafts — don't design their
  domain before the spec exists).
- **A "service layer" is not a cure.** Adding an `internal/services` package around a
  data-driven design produces "an overly complicated onion with fundamentally the same
  design" (see *From CRUD to DDD*). The fix is moving rules into types, not adding layers.
- **Keep the ubiquitous language in one place.** If two packages name the same concept
  differently (`Prefix` vs `Path` vs `Route`), the model is not done — pick the
  `REVIEW.md` term and rename.

## Sources

- Martin Fowler — [Domain-Driven Design (bliki)](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- Martin Fowler — [Anemic Domain Model (bliki)](https://martinfowler.com/bliki/AnemicDomainModel.html)
- Eric Evans — *Domain-Driven Design: Tackling Complexity in the Heart of Software* (2003)
- Dermendzhiev — [Database-centric vs domain-centric architecture](https://github.com/Dermendzhiev/database-vs-domain-centric-architecture)
- CodeMonkeyProjectiles — [From CRUD to DDD](https://codemonkeyprojectiles.com/blog/2021/CrudToDdd/)
- Go Code Review Comments — <https://go.dev/wiki/CodeReviewComments>
- Effective Go — <https://go.dev/doc/effective_go>
- Uber Go Style Guide — <https://github.com/uber-go/guide>