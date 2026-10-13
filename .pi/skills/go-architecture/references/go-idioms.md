# Idiomatic Go

Distilled from the official [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments),
[Effective Go](https://go.dev/doc/effective_go), and the [Uber Go Style Guide](https://github.com/uber-go/guide).
These are the idioms Relay code must follow.

## Errors

Errors are values. Choose the shape by two questions: does the caller need to *match* the
error, and is the message *static* or *dynamic*?

| Match? | Message | Tool |
|---|---|---|
| no | static | `errors.New("...")` returned inline |
| no | dynamic | `fmt.Errorf("...", args...)` |
| yes | static | top-level sentinel: `var ErrNotFound = errors.New("not found")` |
| yes | dynamic | custom error type (implement `Error()`, optionally `Unwrap()`) |

Rules:

- **Never discard.** `_ = f()` is a bug report waiting to happen. Handle, wrap, or return.
- **Wrap with `%w`** so callers can use `errors.Is` / `errors.As`. Add context at every
  layer that knows something the layer below does not:

  ```go
  // good: each layer adds only what it uniquely knows
  if err := s.store.Save(ctx, site); err != nil {
      return fmt.Errorf("save site %q: %w", site.Name, err)
  }
  ```

- **String style:** lowercase start, no trailing punctuation (they get embedded in larger
  messages). `errors.New("no route for prefix /raspberry/")`, not `"No route found!"`.
- **No in-band errors.** Prefer `(value T, ok bool)` or `(T, error)` over returning `""`,
  `nil`, `-1` to signal failure.
- **Don't leak internals.** A wrapped error is fine to log, but user-facing messages come
  from matched sentinels, not from string parsing.
- **`panic` is for unrecoverable state** (corrupt key file in `main`, programming error),
  never for expected failure. In library code, return the error.

Sentinel example (Relay):

```go
var (
    ErrPrefixTaken = errors.New("prefix already claimed")
    ErrNotOwner    = errors.New("request not signed by owner key")
)
```

## Context

- `context.Context` is the **first parameter**, named `ctx`:
  `func (n *Node) DialFriend(ctx context.Context, addr string) (*Tunnel, error)`.
- **Never store ctx in a struct.** Add a parameter to each method that needs it.
- **Never invent your own context type** or use an interface instead of `context.Context`
  in signatures.
- Application data belongs in parameters, the receiver, or (rarely) a `context.WithValue`
  — never use context as a grab bag for dependencies.
- Propagate cancellation. Anything that loops or blocks long takes `ctx.Done()` into
  account. Use `context.WithTimeout` for outbound work (tunnel dials, Laya calls) so a
  stuck friend cannot pin a goroutine.
- `context.Background()` only at true roots: `main`, a test, or a top-level worker.

## Concurrency

**Ownership rule:** whoever starts a goroutine owns its shutdown. A goroutine's exit must
be visible from its spawn site — via `ctx` cancellation, a `sync.WaitGroup`, or a channel
the spawner closes. If you cannot point to where a goroutine ends, it is a leak.

```go
// good: the goroutine's lifetime is bounded by the caller's ctx
func (t *Tunnel) pump(ctx context.Context) {
    go func() {
        if err := t.readLoop(ctx); err != nil && !errors.Is(err, context.Canceled) {
            t.log.Error("read loop", "err", err)
        }
    }()
}
```

- **Prefer synchronous functions.** A function that returns its result is easier to reason
  about and to test than one that "runs and notifies". Callers who want concurrency can add
  a `go` themselves; removing concurrency you baked in is hard.
- **No fire-and-forget.** `go sendAlert(...)` with no way to observe completion is banned.
- **Channels:** size `0` (synchronous hand-off) or `1` (buffer one item as a latch). Never
  design an API around channels; pass them as arguments. Only the **sender** closes, and
  only once. If the receiver should be told the stream ended, close by the sender as part
  of a documented protocol — otherwise keep the channel open and cancel via context.
- **Fan-in/fan-out:** use `errgroup` for "run N things, stop on first error":

  ```go
  g, ctx := errgroup.WithContext(ctx)
  for _, friend := range n.friends {
      friend := friend
      g.Go(func() error { return friend.Heartbeat(ctx) })
  }
  if err := g.Wait(); err != nil { ... }
  ```
- **`sync`:** zero-value `Mutex`/`RWMutex` is ready to use; never copy a lock
  (`go vet` catches it); guard with `defer mu.Unlock()`.
- **Race detector:** `go test -race ./...` is part of CI. A data race is a blocking bug,
  not a warning.

## Interfaces

- **Small.** 1–3 methods. `io.Reader`, `io.Writer` set the bar.
- **Declared by the consumer**, in the package that *uses* it:

  ```go
  // internal/site (consumer)
  type Store interface {
      Get(ctx context.Context, name string) (site.Site, error)
      Put(ctx context.Context, s site.Site) error
  }
  ```
- **Producers return concrete types.** `internal/infra/sqlite.SiteStore` is a struct; the
  site feature's `New(Store SiteStore, ...)` accepts the interface.
- **No premature interfaces.** If nothing implements it yet and no test needs it, write
  the concrete type. Extract the interface when a second implementation (or a fake)
  appears. Do not define interfaces "for mocking" on the producer side.
- **No empty interfaces in new code** — use `any`.
- **Accept interfaces, return structs.**
- **Pointer arguments to interface methods are an anti-pattern.** If the interface value
  is the thing you pass around, the method receiver is the interface type, not `*I`.

## Structs and receivers

- **Zero values should work.** A freshly declared struct should be usable (or clearly not)
  without a constructor. If `New()` is the only way to get a working value, document why.
- **Named fields always** at initialization: `Site{Name: name, Key: key}`, not positional.
- **Receivers:**
  - Pointer receiver if the method mutates state, the struct contains a mutex/sync
    primitive, or the struct is large or may grow.
  - Value receiver for small immutable value types (`time.Time`-like).
  - **Don't mix** value and pointer receivers on the same type. When in doubt: pointer.
- **Receiver names:** short and consistent per type (`n` for `Node`, `s` for `Site`).
  Never `me`, `this`, `self`.
- **Don't embed types in public structs** (composition-by-embedding leaks the embedded
  type's methods into your API). Name the field.
- **Don't copy structs whose methods live on the pointer type** (aliasing bugs, e.g.
  `bytes.Buffer`).
- **Functional options** for optional construction parameters:

  ```go
  func NewNode(cfg Config, opts ...NodeOption) (*Node, error)
  type NodeOption func(*nodeConfig)
  func WithHeartbeat(d time.Duration) NodeOption { ... }
  ```

## Naming

- **Initialisms are all-caps or all-lower:** `URL`, `ID`, `HTTP`, `TLS`, `API`.
  `ServeHTTP`, not `ServeHttp`; `appID`, not `appId`.
- **Packages:** short, lowercase, no underscores or `go` prefix, no plural (`site`,
  `tunnel`). Never `util`, `common`, `misc`, `helpers`, `types`. The package name is part
  of every identifier, so don't repeat it: `site.Site`, not `site.SiteInfo`.
- **Variables:** the further a name is used from its declaration, the more descriptive it
  must be. Loop indices `i` are fine; a variable used 50 lines away needs meaning.
- **Unexported is mixedCaps**, not `snake_case` or `MAX_LEN`: `maxPrefixLen`.
- **Package names should read well in use:** `tunnel.Dial`, `site.Claim`. If the usage
  reads badly, rename the type or the package.

## Slices, maps, and nil

- Prefer `var t []string` (nil) over `t := []string{}`. Exception: JSON marshaling, where
  nil → `null` and `[]` differs. Don't design APIs where nil vs empty matters.
- **Copy slices and maps at trust boundaries** (function entry/exit touching shared state)
  so callers cannot alias your internals.
- **Preallocate** when the size is known: `make([]Frame, 0, n)`.
- Never return a slice of a local array without a copy if the array is reused (pooling).
- Range over maps is unordered; sort keys if output order matters (tests, manifests).

## Time, crypto, and I/O

- **`crypto/rand` for anything secret** — Relay ownership keys, nonces, session tokens,
  salts. `math/rand` is for shuffling and simulation only. This is a security invariant,
  not a style choice.
- **`time` package:** always UTC at the boundary (wire, storage, logs); convert for
  display. Don't do arithmetic on `time.Time` — use `time.Duration`.
- **I/O:** use `io.Reader`/`io.Writer` boundaries, `io.Copy`, `io.ReadAll` only with
  bounded input (or with a size limit for untrusted input — e.g. tunnel frames).

## Defer, range, and version foot-guns

- `defer` runs at function return, **not** scope exit. A `defer` inside a loop defers
  `n` times — hoist it or use a helper function.
- Go 1.22+: `for i := range n` iterates integers; loop variables are per-iteration, so
  `go func() { use(i) }()` without rebinding is safe. Pin the toolchain in `go.mod` and
  don't write code that depends on pre-1.22 semantics.
- Build tags and `//go:linkname` are banned outside `internal/infra` escape hatches.

## Quality gates

Local and CI, in order:

```sh
gofmt -l .                # must be empty (or use goimports)
go vet ./...
staticcheck ./...
go test -race ./...
```

`staticcheck` catches the things `go vet` misses (error handling, shadowing, unused
result variants). Keep the gate list short; a gate nobody reads is a gate that's off.
