# Clean Code in Go

Language-agnostic clean-code discipline, translated to Go and Relay. The rules that
matter most here: names, small functions, honest comments, and tests that fail clearly.

## Naming

- **The distance rule:** the further a name is used from its declaration, the more
  descriptive it must be. `i` for a loop, `n` for a receiver, but `outboundFriend`
  for something that lives across functions.
- **Use the domain vocabulary** from `REVIEW.md`, consistently: `node`, `site`, `prefix`,
  `deploy`, `friend`, `tunnel`, `key`. If the design doc calls it a "friend node", the
  code says `Friend`, not `Peer`/`Upstream`/`Remote` in different packages.
- **Functions are verbs or verb phrases:** `Claim`, `DeploySite`, `Rollback`.
  **Types are nouns:** `Deploy`, `SiteStore`. **Booleans read as predicates:**
  `isPublic`, `hasRoute`, `canSign`.
- **No clever abbreviations** that require decoding. `prefix` not `pfx`. Acronyms allowed
  only when universal (`URL`, `ID`, `HTTP`).
- If a name needs a comment to explain it, rename the thing.
- **Prefixes that duplicate the package are banned:** `site.Site`, `tunnel.Tunnel`
  are the exception (the type is the domain noun); `site.SiteManager` is not.

## Functions

- **One function, one job, one level of abstraction.** If the function's name needs
  "and", split it: `readFrameAndDispatch` → `readFrame`, `dispatch`.
- **Early return on error; keep the happy path at the left margin** ("indent error flow"):

  ```go
  func (n *Node) routePrefix(r *http.Request) (handler.Handler, error) {
      prefix := firstSegment(r.URL.Path)
      if prefix == "" {
          return nil, ErrMissingPrefix
      }
      fr, ok := n.friends[prefix]
      if !ok {
          return nil, fmt.Errorf("%w: %q", ErrNoRoute, prefix)
      }
      return fr, nil
  }
  ```

- **No `else` after `return`** (or after `continue`/`break`).
- **Arguments:** 0–2 is comfortable; 3 is the ceiling; more means a missing struct
  (`DeployParams{Name, Version, Manifest}`).
- **No output parameters.** Multiple return values are for `(result, error)` and
  `(value, ok)`, not for smuggling state out.
- **Line/function length:** there is no N-line rule. A function is too long when it has
  two jobs or two abstractions mixed; split on the job, not the line count.
- **Private helpers are fine and cheap**; don't artificially inflate small functions into
  one-liner wrappers either.

## Comments

- **Code says what; comments say why.** `// retry because the tunnel may be mid-reconnect`
  is worth writing; `// loop over friends` is noise.
- **Doc comments** on every exported identifier (and non-trivial unexported ones):
  full sentence, starts with the name being described:

  ```go
  // Claim records that key owns prefix until the key is released.
  // It returns ErrPrefixTaken if the prefix is already claimed by another key.
  func (s *Site) Claim(key ed25519.PublicKey, prefix string) error { ... }
  ```

- **Package comment** on every package, adjacent to the `package` clause, no blank line:

  ```go
  // Package site holds the pure domain model for hosted sites: deploys,
  // versions, rollback, and ownership. It performs no I/O.
  package site
  ```

- **TODO:** `// TODO(username): short reason + Kandev task ref`. A TODO without an owner
  and a reason is rot.
- Never comment out code; delete it (it's in git).

## Constants, magic values, and enums

- No bare numbers or strings with meaning. `time.Second * 15` is a heartbeat; name it
  `heartbeatInterval`.
- **Enums with `iota` start at 1** so `0` is a meaningful "unset":

  ```go
  type DeployStatus int

  const (
      DeployPending DeployStatus = iota + 1
      DeployLive
      DeployRolledBack
  )
  ```

- Give enums a `String()` (and `Parse*` where they cross the wire).
- Magic strings that appear in two places (reserved paths, header names) become named
  constants in the package that owns the concept — `/_relay/` lives in exactly one place.

## Structs

- Named fields at every literal. Field order: required first, then optional, grouping
  related fields.
- **Immutable value types** for concepts with identity + invariants that are checked
  once at construction:

  ```go
  type Manifest struct {
      Version  int
      Files    []FileHash
      SignedAt time.Time
  }

  func NewManifest(v int, files []FileHash, now time.Time) (Manifest, error) {
      if v <= 0 {
          return Manifest{}, fmt.Errorf("version must be positive, got %d", v)
      }
      return Manifest{Version: v, Files: append([]FileHash(nil), files...), SignedAt: now}, nil
  }
  ```

- A struct whose every method is a getter/setter is a bag of fields, not a type. Either
  it earns behavior (see `ddd-vs-data-driven.md`) or it is admitted to be a data record.

## Testing

**Table-driven is the default** for anything with branches:

```go
func TestFirstSegment(t *testing.T) {
    tests := []struct {
        name  string
        path  string
        want  string
        wantOK bool
    }{
        {"plain", "/raspberry/docs/", "raspberry", true},
        {"root", "/", "", false},
        {"no slash after", "/raspberry", "raspberry", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, ok := firstSegment(tt.path)
            if ok != tt.wantOK || (ok && got != tt.want) {
                t.Errorf("firstSegment(%q) = %q, %v; want %q, %v",
                    tt.path, got, ok, tt.want, tt.wantOK)
            }
        })
    }
}
```

- **Fail with useful messages.** Assume the person debugging your failing test is not you.
  Order is *got, want*: `t.Errorf("Route(%q) = %v; want %v", in, got, want)`.
- **No sleeping.** Use contexts with timeouts, channels, or test seams. A `time.Sleep`
  in a test is a flake factory.
- `t.Parallel()` at the top of independent subtests; `t.TempDir()` for per-test files
  (SQLite test databases belong here); `t.Cleanup` for anything started.
- **Boundary tests are integration tests.** For Relay: deploy → serve → rollback → export
  against a real temp-dir SQLite file; tunnel tests run **two in-process nodes** and move
  a real HTTP request through the WebSocket. These are not "slow tests to skip" — they are
  the point.
- **Examples:** every new package ships one runnable `Example` or a small test showing the
  intended usage sequence.

## Logging

- Use `log/slog`, structured fields, no `fmt.Println`.
- Levels: `Error` for "a page 404'd where it should have been 200", `Warn` for
  "reconnecting, attempt 3", `Info` for lifecycle (node started, deploy live), `Debug`
  for per-request trace. A healthy node at `Info` should be quiet.
- Every log line that can carry an identifier does: `site`, `prefix`, `node`.
- **Never log secrets:** Ed25519 private halves, session tokens, key-file contents. Log
  fingerprints (`sha256` prefix), never material.

## Code-smell checklist (for review)

- A function with more than one abstraction level mixed
- A handler/service method longer than ~40 lines doing domain work (move it to the model)
- `else` after `return`
- Error returned but not wrapped where context exists; error compared by string
- A struct field written only in `New` and never read outside
- A channel whose closer (or absence of one) you cannot point to
- A new `interface` with one implementor and no test
- A new external dependency without a stated reason
- A copy-pasted block in a third location (extract it then, not at the second)