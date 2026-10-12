# Relay Backend Architecture

_Draft for review. The code architecture for the backend described in
`docs/api-contract.md`: one Go binary, embedded Caddy as the edge, and a
multiplexed-WebSocket tunnel (yamux) that lets a node behind NAT serve the web
by dialing out. Follows the Go standard in `.pi/skills/go-architecture`
(`gofmt` law, errors-as-values, `context` first, visible goroutine lifetimes,
consumer-side interfaces, no mutable global state — with one deliberate,
documented exception in §4.4)._

---

## 1. Overview

```
 visitor ──HTTPS──► ┌────────────────────────── public node ──────────────────────────┐
                    │  Caddy (embedded)                                               │
                    │   ├─ TLS/ACME for the single hostname                           │
                    │   ├─ /_relay/tunnel/v1  →  relay_tunnel  (accept module)        │
                    │   ├─ /_relay/v1/...     →  control handler  (in-process)        │
                    │   └─ /<node>/...        →  reverse_proxy                        │
                    │         ├─ own node  → local site service (loopback)            │
                    │         └─ friend    → tunnel transport ── yamux stream ────────┼──┐
                    │                         (one stream per HTTP request)           │  │
                    │  SessionRegistry (in-process, name → *tunnel.Session)           │  │
                    │  node DB (SQLite) · blobs (BLAKE3) · per-site SQLite            │  │
                    └─────────────────────────────────────────────────────────────────┘  │
                                                                                          │ outbound WS
                    ┌────────────────────────── private node (behind NAT) ──────────────┐ │
                    │  Caddy (embedded, loopback only): serves the same route tree      │ │
                    │  Agent: dial friend → WS upgrade → signed hello → yamux server    │◄┘
                    │          accept loop: stream → forward to 127.0.0.1:<local>       │
                    │  per-site SQLite · blobs                                          │
                    └───────────────────────────────────────────────────────────────────┘
```

**The load-bearing idea:** Caddy's `reverse_proxy` is pointed at a custom
`Transport` whose `DialContext` returns a yamux stream instead of a TCP socket.
Everything downstream of that seam — `http.Transport`, request/response
streaming, idle handling, error mapping — is stock Caddy/Go. Everything upstream
is stock Caddy config. Relay's own code is the seam, the handshake, the registry,
and the agent.

**Roles.** One binary, two roles, one flag (`--role public|private`, locked).
Public: listens on `:443`, serves its hosted sites, accepts friend tunnels,
relays requests. Private: serves its sites on loopback only, dials out to its
friend(s). Same route tree, same site code on both sides — the path is
forwarded unmodified, so a site does not know or care which node answered.

## 2. Startup and lifecycle

### 2.1 Public node

1. `cmd/relay` parses flags, loads operator config (YAML; path, data dir,
   limits, rule set, friend policy), opens the node DB (`infra/sqlite`).
2. Builds the `SessionRegistry`, the route resolver, and the control/runtime
   services (constructor injection; §3).
3. Registers the two Caddy modules and calls `caddy.Load` with the config
   assembled in code (§4.1). Caddy owns TLS/ACME, the listen socket, and the
   admin lifecycle (admin API bound to a unix socket for `relay ctl`, off by
   default).
4. Runs until signal: on `SIGTERM`, stop accepting tunnels, drain in-flight
   requests (30 s), close sessions with `1001`, close Caddy, close DBs.

### 2.2 Private node

1. Same 1–2, plus: starts its local Caddy on `127.0.0.1:<local-port>`
   (default 8080) with the same route tree.
2. Starts the **agent dial loop** (`internal/agent`): for the configured
   friend(s), dial → upgrade → `hello` → on `welcome`, run the yamux accept
   loop. On any failure: exponential backoff (1 s base, 60 s cap, full jitter),
   then the next friend in RTT order (v1: one friend; §4.6 of the contract).
3. The accept loop owns one goroutine per stream: parse the HTTP/1.1 request,
   forward to the local Caddy listener, stream the response back, close the
   stream. Goroutines exit when the stream or the session ends (§6).

### 2.3 Session replacement (both roles, public side active)

On a new `hello` for a name that already has a session: the new session is
fully handshaked, then the registry `Put` swaps it in and the old session is
closed (`1000`). In-flight requests on the old session fail as `502
tunnel:session-replaced`. A swap never unroutes a node mid-flap because the
old entry stays until the new handshake is done.

## 3. Components and package layout

Per the Go skill: `cmd → feature → domain`; `feature → infra` only through
interfaces the feature declares; `domain` imports nothing in the repo and no
I/O. Style per bounded context (ddd-vs-data-driven rule): `domain` is rich,
the tunnel is **plumbing with one protocol state machine**, blobs are
data-driven plumbing.

```
cmd/relay/            main: flags, config load, wiring, signal handling. No logic.
internal/
  domain/             key, Name, Site, Deploy, Version, Prefix — pure model.
                      Invariants: claim uniqueness, monotonic versions,
                      signed-canonical-string builder. No I/O imports.
  route/              Resolver: path → (node, site, rest). Pure string logic +
                      the reserved-path rules. Data-driven (plumbing).
  control/            Control API: signed-request verification, claim/deploy/
                      rollback/export handlers + service. Domain-driven.
  runtime/            Runtime API: db (kv + collections), auth, metrics.
                      Per-site access rules. Domain-driven.
  tunnel/             The wire protocol: hello/welcome codec, handshake state
                      machine, session wrapper, stream policy, close codes.
                      Plumbing + one state machine (contract §4 is its spec).
  agent/              Private node: friend dialer with backoff/failover, yamux
                      server accept loop, per-stream forwarder to loopback.
  registry/           *Registry: name → *tunnel.Session, thread-safe, with
                      subscribe/notify for status + metrics. See §4.4.
  infra/
    caddy/            Embedded Caddy: config assembly, module registration,
                      the two custom module types (§4.2–4.3), admin socket.
    ws/               coder/websocket + yamux wiring: upgrade, NetConn,
                      session config (windows, keepalive, MaxStreamCount).
    sqlite/           Node DB + per-site DB repositories (implement the
                      interfaces declared by control/runtime/domain users).
    blobs/            BLAKE3 content-addressed store: Put/Get(hash), GC hook.
    classify/         Classifier providers (none/heuristic/laya-*) — later.
    quickjs/          Wazero sandbox — later stage.
```

**Dependency rule (enforced by import lint in CI):** `domain` → stdlib only.
`route`, `tunnel`, `agent`, `control`, `runtime`, `registry` → `domain` +
stdlib + each other only where noted below. `infra/*` → `domain` and the
features' interfaces. Nothing imports `infra` except `cmd/relay` and other
`infra` packages that need a concrete type (e.g. `infra/caddy` → `infra/ws`).

Allowed cross-feature imports (kept minimal, documented):
- `agent → tunnel` (uses the handshake state machine and `Session`).
- `tunnel → domain` (the shared canonical-string builder; nothing else).
- `infra/caddy → tunnel, registry` (the accept module drives the handshake).
- `control → domain` (claims, deploys). `runtime → domain` (site config).
- `route` imports nothing internal (pure).

## 4. Caddy integration

### 4.1 Embedding

```go
// cmd/relay (sketch)
import (
    "github.com/caddyserver/caddy/v2"
    _ "github.com/caddyserver/caddy/v2/modules/standard"
    _ "relay/internal/infra/caddy" // registers relay modules
)

func main() {
    cfg, err := buildConfig(loadConfig(*configFlag), *roleFlag) // JSON string
    if err != nil { log.Fatal(err) }
    if err := caddy.Load(cfg, true); err != nil { log.Fatal(err) }
    signalHandler(caddy) // stop on SIGTERM, caddy.Stop on exit
}
```

- Config is **assembled in code, not read from disk** — the operator config
  (YAML) is the source of truth; the Caddy JSON is a projection. This keeps
  "one binary, one config file" and makes the Caddy surface auditable.
- Admin API: `{"admin": {"listen": "unix//run/relay/admin.sock"}}` for
  `relay ctl status` and live reconfiguration of limits; disabled entirely in
  the private role.
- Dynamic agent/host addition (adding a friend's prefixes without a restart)
  goes through the control API → registry + route resolver, not by hot-loading
  Caddy. The Caddy config is static per node identity; the *routing table* is
  dynamic state. This is the line between "Caddy owns routing shape" and
  "Relay owns routing data."

### 4.2 Module 1 — `http.handlers.relay_tunnel` (accept)

Caddy handler; matches `/_relay/tunnel/v1` on the public host.

```go
type TunnelAccept struct {
    registry *registry.Registry   // injected via a module-level hook, §4.4
    claims   ClaimsLookup         // consumer-side interface: name+key → claimed?
    limits   tunnel.Limits
    next     caddyhttp.Handler    // always a terminus; kept for interface fit
}

func (h *TunnelAccept) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        InsecureSkipOriginCheck: true, // non-browser clients; no Origin to check
    })
    if err != nil { return err }

    sess, err := tunnel.Handshake(r.Context(), infraWS.NetConn(r.Context(), ws),
        tunnel.ServerParams{ Verify: h.verifyHello, Limits: h.limits })
    // verifyHello: decode, check v/ts/nonce/sig, claims lookup for every prefix
    if err != nil {
        // tunnel writes the JSON error frame + WS close code, then err
        return err
    }
    h.registry.Put(sess.Peer().Name, sess)
    defer h.registry.Delete(sess.Peer().Name, sess)
    <-sess.CloseChan() // block for the tunnel's lifetime; Caddy hands us one
    return nil         // goroutine/goroutine-per-connection, owned by Caddy
}
```

Responsibilities, in order: upgrade → handshake → register → park. It holds no
routing logic and no site logic.

### 4.3 Module 2 — `http.reverse_proxy.transport.relay_tunnel`

The seam. Caddy calls `DialContext` per upstream connection; we return a
yamux stream.

```go
type TunnelTransport struct {
    registry *registry.Registry
    rt       *http.Transport // built in Provision()
}

func (t *TunnelTransport) Provision(ctx caddy.Context) error {
    t.rt = &http.Transport{
        DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
            name, _, err := net.SplitHostPort(addr) // "raspberry:80"
            if err != nil { return nil, err }
            sess := t.registry.Get(name)
            if sess == nil {
                return nil, tunnel.ErrNodeOffline // → 502 tunnel:node-offline
            }
            return sess.OpenStream(ctx) // yamux stream IS a net.Conn
        },
        DisableKeepAlives: true, // one request per stream (contract §4.3)
        TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // no TLS here;
        ForceAttemptHTTP2:   false,                                  // streams are http/1.1
    }
    return nil
}

func (t *TunnelTransport) RoundTrip(r *http.Request) (*http.Response, error) {
    return t.rt.RoundTrip(r)
}
func (t *TunnelTransport) CloseIdleConnections() { t.rt.CloseIdleConnections() }
```

Caddy config projection (public node, per connected friend — one route per
claimed prefix, path untouched):

```json
{
  "handler": "reverse_proxy",
  "headers": { "request": { "set": { "X-Relay-Node": "{http.request.vars.node}" } } },
  "transport": { "protocol": "relay_tunnel" },
  "upstreams": [{ "dial": "raspberry:80" }]
}
```

Route matching is a single Caddy route: `path {path}/_relay/tunnel/v1` →
accept module; `path /_relay/v1*` → control handler; `path /{*node}/*` →
reverse_proxy with a Caddy Vars step that sets `node` from segment 1, and a
small `handle_error` mapping `ErrNodeOffline → 502`, stream-timeout → `504`.
The resolver (`route.Resolver`) is consulted by a Caddy `map`-style middleware
only for the *own node* case (serve locally) vs friend (proxy) — see §5.

### 4.4 The registry and the one global

The skill says no mutable global state. Two Caddy modules, instantiated by
Caddy's module system from JSON, must share one in-process object. The
pragmatic, documented exception:

```go
// internal/registry
var shared = New() // the *Registry value itself is immutable;
                   // its contents are mutex-guarded. Created once, at
                   // package init, before Caddy loads. cmd/relay calls
                   // shared.Reset() in tests.
func Shared() *Registry { return shared }
```

Rationale: Caddy module constructors (`New() caddy.Module`) take no
parameters, so injection through Caddy's normal path is impossible without
reimplementing Caddy's module loader. `registry.Shared()` is the single
package-level variable in the codebase; it is not configuration (nothing is
*configured* — it is a container), it is created once, and every other
dependency is constructor-injected per the skill. CI lint: exactly one
package-level `var` of type `*Registry`, in `internal/registry`.

```go
type Registry struct {
    mu       sync.RWMutex
    sessions map[string]*tunnel.Session
    subs     []chan Event // {Name, Online/Offline, Replaced} for status UI + metrics
}
func (r *Registry) Put(name string, s *tunnel.Session)  // closes the old one after swap
func (r *Registry) Get(name string) *tunnel.Session
func (r *Registry) Delete(name string, s *tunnel.Session) // no-op if replaced
func (r *Registry) Events() <-chan Event
```

## 5. Routing: who decides "local vs tunnel"

`route.Resolver` (pure):

```go
type Resolver struct{ self string; reserved map[string]bool }
type Target struct { Node, Site, Rest string; Local bool }
func (r *Resolver) Resolve(path string) (Target, error) // strict segment parsing
```

Caddy does not know whether `raspberry` is local or a friend — that is dynamic
state. The split is implemented as a **single reverse_proxy per public node**
with a custom `LoadBalancing`/dial layer, not N routes:

- The Caddy route matches `/{node}/{site}/*` and proxies to a dial address of
  `{node}:80`.
- For the **own** node, `TunnelTransport.DialContext` short-circuits:
  `name == self` → `net.Dial("tcp", "127.0.0.1:<local-port>")`. One
  transport, two backends (local loopback or friend tunnel), chosen at dial
  time against live registry state. This keeps Caddy's config static and the
  fail-over decision (friend offline → 502) in exactly one place.
- Own-node traffic is marked so the agent side can distinguish (it never sees
  `X-Relay-Tunnel` for local serving; §4.4 of the contract).

Reserved segments (`_relay`) are handled before this route by the earlier
routes; `Resolve` returning an error for them is belt-and-braces.

## 6. The `tunnel` package (protocol state machine)

Per the ddd-vs-data-driven rule: plumbing with one state machine. The
state machine is the handshake; everything after is byte pumping.

```go
type State int // S_Dialing → S_HelloSent → S_Welcomed → S_Session
type Session struct {
    yamux  *yamux.Session
    peer   Peer // {Name, PublicKey, NodeVersion, Limits}
    conn   net.Conn // the NetConn over the websocket
    closeCh chan struct{}
}
func (s *Session) OpenStream(ctx context.Context) (net.Conn, error) // public side
func (s *Session) Accept(ctx context.Context) (net.Conn, error)     // private side
func (s *Session) Close() error
func (s *Session) CloseChan() <-chan struct{}
func (s *Session) Peer() Peer

// Handshake drives the state machine on either role. ServerParams.Verify is
// the only policy hook (claims lookup); it is declared in tunnel,
// implemented in infra/caddy (consumer-side interface).
func Handshake(ctx context.Context, conn net.Conn, p Params) (*Session, error)
```

Inside: `codec.go` (hello/welcome JSON + canonical-string + Ed25519 verify —
the canonical string builder itself lives in `domain` so the control API
shares one implementation), `state.go` (the four states, timeouts: 10 s for
the handshake, 5 missed keepalives → dead), `limits.go` (effective limits =
min of both sides), `errors.go` (`ErrNodeOffline`, `ErrSessionReplaced`,
`ErrHandshake` with code + close-code mapping per contract §4.8).

**Agent stream forwarder** (`internal/agent/forward.go`, data-driven plumbing):

```go
func forward(ctx context.Context, stream net.Conn, local string, policy ForwardPolicy) {
    defer stream.Close()
    req, err := http.ReadRequest(bufio.NewReader(stream)) // exactly one request
    // apply policy: enforce body cap, strip X-Relay-Tunnel, no visitor XFF
    localConn, err := net.Dial("tcp", local)               // loopback only
    // write request (preserve chunked), then stream response back:
    // io.Copy(stream, resp.Body) with the header block written first
}
```

One goroutine per stream; the accept loop (one goroutine per session) spawns
it and is its owner via the session context — session death cancels the ctx,
every stream's `io.Copy` unblocks, all goroutines exit. No fire-and-forget.

## 7. Goroutine map (skill rule 4: every goroutine has an owner and an exit)

| Goroutine | Owner | Spawns | Exits when |
|---|---|---|---|
| Caddy server goroutines | Caddy | Caddy | `caddy.Stop` |
| Tunnel accept (per connection) | Caddy handler goroutine | — | WS closes / handshake fails |
| Agent dial loop (per friend) | `agent.Dialer` ctx | accept loop | ctx cancel (node stop) |
| Yamux accept loop (per session) | dial loop ctx | one per stream | session closed |
| Stream forwarder (per stream) | accept loop ctx | — | stream done or ctx cancel |
| Keepalive pinger (per session) | `tunnel.Session` | — | session closed |
| Registry event fan-out | registry | — | node stop (drained) |
| Metrics collector | `cmd/relay` | — | node stop |

## 8. Backpressure and resource bounds

- The chain visitor→Caddy→`http.Transport`→yamux stream→yamux conn→WebSocket→
  private node→local Caddy is **all synchronous `io.Copy` over bounded
  windows**. Nothing buffers more than a yamux window (1 MiB) per stream.
- Worst-case public-node memory per session: `max_streams (64) × 1 MiB` ≈ 64
  MiB; sessions are capped (64) — the `overloaded` close code is the valve.
- Request bodies over `max_body_bytes` abort the stream and return `413`
  (enforced by the public side's request-rewriter *and* the agent's forwarder).
- If aggregate throughput through one WebSocket ever becomes the ceiling
  (it can be — one TCP connection), the documented fix is friend
  multiplication (a site's traffic split across two friends is **not** in v1;
  note for the §7.5 decision).

## 9. Security architecture

- **TLS** at the edge (Caddy/ACME, single cert). No plaintext tunnel path.
- **Auth:** Ed25519 signature on `hello` (tunnel) and on every control
  request; `ts` window ±60 s + 5-minute nonce replay cache in the node DB.
  One canonical-string builder in `domain`, shared by both surfaces.
- **Claim binding:** a session may only serve prefixes claimed by its key;
  a new claim for a live prefix does not displace the live session
  (contract §4.7).
- **Header hygiene:** the strip/set lists are contract §4.4, implemented in
  one `policy.go` in `tunnel` used by both the Caddy request rewriter (public
  side) and the forwarder (private side). Table-driven tests pin the list.
- **Agent confinement:** the forwarder's dial target is `127.0.0.1:<local>`
  — a constant from node config, not from the request. No SSRF surface.
- **Origin isolation** (cookies/CSRF/referer) is the runtime API's job —
  **OPEN** (`REVIEW.md` §7.11); the architecture reserves the seams
  (path-scoped cookie names, per-site CSRF in `runtime`).
- **Abuse:** v1 = contract §4.9 limits. Quotas/bandwidth/ToS **OPEN** (§7.8).

## 10. Observability

- **Logs:** `log/slog`, JSON, one line per request on the public side with
  `node`, `site`, `local|tunnel`, `stream_id`, `status`, `duration_ms`,
  `bytes_in/out`. The tunnel adds `session_id` and reconnect counters.
- **Health:** `GET /_relay/v1/node` (contract §6) doubles as the health
  endpoint: 200 with `sessions`, `storage`, `friends` state. Swarm/K8s probes
  hit it.
- **Metrics:** in-process counters (tunnel up/down, sessions, streams open,
  bytes relaying, 502-by-node) exposed on the admin socket; node-side
  **stats** (the opt-in per-site visitor metrics) are a different thing —
  `runtime`'s job, counted where the site is served.
- **Session events** (`Registry.Events()`) drive the node UI's "friends
  online" view — the same channel feeds health, logs, and UI.

## 11. Testing strategy (skill rule 8)

| Layer | What | How |
|---|---|---|
| Unit (table-driven) | `route.Resolver` (all reserved/malformed paths); `tunnel` codec (hello encode/decode, version/field tolerance); `domain` canonical string (vectors); claims & version invariants | plain Go, no I/O |
| Integration | **Two in-process nodes over a real WebSocket on 127.0.0.1**: public Caddy (test port, self-signed) + private node with a stub site server; assert a visitor request end-to-end, headers (§4.4), body streaming > window size (backpressure), 502 when friend down, session replacement on reconnect, 413 on oversize body | `httptest` + the real modules; this is the CI gate for Stage 1 |
| Contract | handshake frames and error codes pinned as golden JSON in `tunnel/testdata`, generated from `docs/api-contract.md` examples | diff on drift |
| Property | yamux window tuning: assert no session exceeds `max_streams × window` memory under a slow-visitor fuzz | `testing/quick` |

Two-node integration tests are mandatory for any change touching `tunnel`,
`agent`, `registry`, or the Caddy modules (pre-existing requirement from
`REVIEW.md` §7.2 — this layout makes them cheap: both roles construct in
process).

## 12. Stages

| Stage | Scope | Unblocks |
|---|---|---|
| 0 (already the plan in `REVIEW.md` §6) | Hosted sites only: control + runtime + static/blob serving on one public node, no tunnel. `TunnelTransport` short-circuit (local dial) is already the full path, so Stage 0 ships without the tunnel code existing | the acceptance test |
| **1 (this draft)** | `tunnel`, `agent`, `registry`, the two Caddy modules, one friend | a machine behind NAT serves the web |
| 2 | Multi-friend failover (RTT-ordered, gap-free **OPEN**), blob dedup/GC, classifier | `REVIEW.md` §7.7, §7.15 |
| 3 | Compute (QuickJS/Wazero), stats polish, key-login | later |

Stage 1 adds **zero** behavior change for Stage 0 users: the tunnel is a new
Caddy route and a new role; hosted serving is untouched.

## 13. Deliberate deviations from the reference sketch

The task's sketch (embedded Caddy + yamux + registry) is adopted; the deltas,
all forced by locked Relay decisions:

1. **Identity:** `?agent=` + Bearer → signed `hello` (Ed25519 deed is locked;
   identity out of URLs). Contract §4.2.
2. **Registry:** not a bare package var in a sketch, but an injected
   `*registry.Registry` with the single documented global (§4.4) plus an
   event channel for status/observability.
3. **Routing data vs shape:** claims/prefixes are dynamic state in the node
   DB, consulted at dial time; the Caddy config stays static per node
   (§4.1, §5). The sketch's "push updated config with `caddy.Load`" is
   rejected as a steady-state path (admin API / control route only).
4. **One request per stream** (`DisableKeepAlives`) in v1: removes
   connection-reuse edge cases on yamux; contract §4.3 names the upgrade.
5. **Own-node short-circuit in the transport** (§5): one reverse_proxy, two
   backends, no per-node Caddy routes, no config reloads on friend changes.

## 14. Open items (carried, not solved here)

HOL/HTTP2-over-yamux/QUIC (§7.5) · prefix release/reclaim anti-impersonation
(§7.6) · gap-free friend failover (§7.7) · abuse/quota model (§7.8) · rule
schema (§7.9) · full threat model (§7.10) · same-origin isolation mechanics
(§7.11) · DB query grammar (§7.12) · auth lockout details (§7.13) · key
lifecycle UX (§7.14) · blob GC (§7.15). Each has its place in the contract
where its code will land.