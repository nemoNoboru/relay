# Relay API Contract (v1)

_Draft for review. Expands `REVIEW.md` §4 ("The API in brief") into a wire-level
contract. The tunnel section covers the Caddy reverse proxy over a multiplexed
WebSocket — the primitive that blocks Stage 1. Open items carried from
`REVIEW.md` §7 are marked **OPEN** inline._

---

## 1. Scope

This contract defines everything a Relay node speaks on the wire:

| Surface | Base path | Caller | Defined here |
|---|---|---|---|
| **Tunnel** | `wss://<public-host>/_relay/tunnel/v1` | private (friend) nodes | full wire contract (§4) |
| **Control** | `https://<host>/_relay/v1/...` | deploy tool / UI (owner key) | endpoint contract (§6) |
| **Runtime** | `https://<host>/<node>/<site>/_relay/...` | a site's own pages (same-origin) | endpoint contract (§7) |
| **Routing** | `https://<host>/<node>/<site>/...` | visitors | forwarding contract (§5) |

One hostname per public node, one TLS certificate (locked, `REVIEW.md` §3).
No per-site certificates, no wildcards.

## 2. Conventions

- **Versioning.** Major version in the path (`/v1`). Breaking change → new major,
  old major kept running until the node operator says otherwise. Non-breaking
  additions (new optional fields, new endpoints) do not bump.
- **Encoding.** JSON for metadata and control; UTF-8; compact. Bodies are
  content-addressed (BLAKE3 hex, 64 chars) wherever file bytes are involved.
- **Times.** Unix seconds, integer, in `ts` fields.
- **Names.** A *node name* is the first path segment of an address:
  `[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?`. A *site name* follows the same grammar.
  Reserved node names: `_relay`, and anything the operator lists in its rules file.
- **Error envelope** (all HTTP surfaces):

  ```json
  { "error": { "code": "name_unavailable", "message": "…", "details": {} } }
  ```

  `code` is a stable machine-readable token; `message` is human-readable and may
  change; `details` is optional and endpoint-specific. Clients branch on `code`
  only.

- **Status codes.** `400` malformed · `401` missing/invalid signature or session ·
  `403` signed by a key that does not own the resource · `404` unknown ·
  `409` conflict (claim, version) · `413` body too large · `429` rate limited ·
  `502` tunnel peer offline · `504` tunnel peer timed out.

## 3. Identity and signatures

Locked (`REVIEW.md` §3/§4): an **Ed25519 keypair is the deed to a name**. The
secret half never travels.

- The private node and the deploy tool each hold a key. The **public half**
  identifies the node; the **key file** (private half + node metadata, see §6.2)
  is the only thing that proves ownership.
- Every owner request (control API) and every tunnel `hello` carries a
  signature over a **canonical string**:

  ```
  canonical = METHOD + "\n" + PATH + "\n" + BLAKE3(body) + "\n" + ts + "\n" + nonce
  signature = Ed25519_sign(private, canonical)     // 64 raw bytes, hex-encoded
  ```

  - `PATH` is the request target path, query string included, exactly as received.
  - `ts` within ±60 s of server time, else reject.
  - `nonce` is 16 random bytes, hex; the node keeps a 5-minute replay window.
- The public half travels in `X-Relay-Key` (control) or inside `hello` (tunnel).

## 4. Tunnel protocol (wire contract)

### 4.1 Transport

```
private node (behind NAT)                          public node
      │  1. TLS 1.2/1.3, single cert                     │
      │  2. HTTP/1.1 Upgrade: websocket  /_relay/tunnel/v1
      │  3. JSON handshake over WS binary messages       │
      │  4. socket becomes a yamux session               │
      │  5. one HTTP/1.1 request per yamux stream  ◄──── │ (opened by Caddy)
```

- The **public node is the yamux client** (it opens streams); the private node is
  the yamux server (it accepts). This follows from the topology: only the public
  side needs to push work downstream.
- yamux over `github.com/coder/websocket` `NetConn`, `MessageBinary` only.
  No text messages after the handshake. No subprotocols.
- **Identity is not in the URL.** No `?agent=`, no Bearer token on the upgrade.
  URLs are logged by proxies and the operator's access logs; identity rides in the
  signed `hello` instead (§4.2). The upgrade itself is unauthenticated (TLS is
  the trust boundary at that point) — see the note in §4.11 for the rejected
  Bearer alternative.

### 4.2 Handshake

| Step | Direction | Frame |
|---|---|---|
| 1 | private → public | `hello` |
| 2 | public → private | `welcome` **or** `error` (then close) |

After `welcome`, both sides switch the socket to yamux. A peer that sends a yamux
byte before the handshake completes gets close code `4000`.

**`hello`** (first binary WS message, JSON):

```json
{
  "v": 1,
  "name": "raspberry",
  "key": "ed25519-public-hex",
  "ts": 1760000000,
  "nonce": "16-random-bytes-hex",
  "sig": "ed25519-signature-hex",
  "node_version": "relay 0.1.0",
  "prefixes": ["/raspberry/"],
  "limits": {
    "max_streams": 64,
    "max_body_bytes": 1073741824
  }
}
```

- `sig` is over the canonical string
  `"HELLO\n" + name + "\n" + BLAKE3(hello-without-sig) + "\n" + ts + "\n" + nonce`
  with `hello-without-sig` being the compact JSON of the same object minus the
  `sig` field.
- `prefixes` is the full list of path prefixes this node will serve — every
  prefix **must already be claimed** by `key` on this public node (claims are
  made via the control API, §6.3). An unknown or foreign prefix is a hard
  reject: the node either serves its claimed prefixes or not at all.
- `limits` is what the private node *can* honor; the effective session limits are
  `min(hello.limits, welcome.limits, operator config)`.

**`welcome`**:

```json
{
  "v": 1,
  "server": "relay 0.1.0",
  "node": "raspberry",
  "session": "random-16-bytes-hex",
  "limits": {
    "max_streams": 64,
    "max_body_bytes": 1073741824,
    "max_concurrent_requests": 32,
    "idle_timeout_s": 180,
    "keepalive_interval_s": 30
  }
}
```

**`error`** (then WS close):

```json
{ "error": { "code": "prefix_not_claimed", "message": "…", "details": { "prefixes": ["/other/"] } } }
```

Handshake error codes: `bad_version` · `bad_signature` · `name_mismatch`
(public name differs from claimed name) · `prefix_not_claimed` · `prefix_taken`
(two live nodes claim the same prefix — see §4.7) · `overloaded` (public node at
session capacity) · `rules_rejected` **OPEN** (classifier hold; until §7.9 lands,
the code is never sent).

### 4.3 Stream semantics

- Each **yamux stream carries exactly one HTTP/1.1 request and its response**.
  One request per stream; the stream closes when the response body is complete
  (or the request fails). No pipelining on a stream in v1.
- Request line uses the **public host**: `GET /raspberry/docs/index.html HTTP/1.1`
  with `Host: relay.example.org`. The path is forwarded **unmodified** — the
  private node's local server runs the same route tree and resolves
  `<node>/<site>` itself. The public node does zero site-level routing; it maps
  only the first segment to a tunnel (§5).
- **Body transfer:** `Content-Length` preferred; `chunked` allowed and preserved
  end to end. Neither side upgrades `Transfer-Encoding` mid-stream.
- **Request body cap** enforced on both ends from the effective
  `max_body_bytes`; exceeding it aborts the stream, public side returns `413`.
- **Response:** the private node streams status line + headers + body; the public
  side forwards them verbatim to the visitor. No response buffering.
- **Backpressure is the design, not a feature:** slow visitor ⇒ yamux stream
  window stops ⇒ yamux connection window stops ⇒ WebSocket stops ⇒ the private
  node's read from the stream blocks. No application-level queues, no `io.Copy`
  with unbounded buffers. yamux window is sized at 1 MiB (default is 256 KiB —
  raise it for file deploys); per-connection memory is bounded by
  `max_streams × window`. **OPEN** (`REVIEW.md` §7.5): if the per-stream
  HTTP/1.1 HOL or the single-WebSocket aggregate proves insufficient, the
  sanctioned fix is HTTP/2 framing over the yamux session (same streams, better
  framing) — not a new transport.

### 4.4 Header policy

Headers the public node **sets** on the tunnel-bound request:

| Header | Value |
|---|---|
| `X-Relay-Tunnel` | `1` |
| `X-Relay-Node` | the node name (first path segment) |
| `X-Forwarded-For` | visitor IP, appended |
| `X-Forwarded-Proto` | `https` |
| `X-Forwarded-Prefix` | `/<node>` |
| `X-Real-Ip` | visitor IP |

Headers the public node **strips** before the tunnel (hop-by-hop and
proxy-privileged): `Connection`, `Keep-Alive`, `Proxy-Authenticate`,
`Proxy-Authorization`, `TE`, `Trailers`, `Upgrade`, `Transfer-Encoding` (only
when the body is re-chunked), and any second `X-Forwarded-*`/`X-Real-Ip`/
`X-Relay-*` received from the visitor (a visitor must not be able to pre-seed
trust headers).

The private node **does not forward** `X-Relay-Tunnel` onward (it is only
meaningful inside the tunnel) and must ignore visitor-supplied
`X-Forwarded-*` the same way.

### 4.5 Keepalive and liveness

- yamux keepalive pings every `keepalive_interval_s` (default 30), 5 missed →
  session dead.
- The private node also sends WS `ping` every 20 s as a second probe, so a
  middlebox that only speaks WebSocket keepalive cannot be the silent killer.
- **Idle timeout:** a session with no streams and no data for
  `idle_timeout_s` (default 180) is closed by the public node with code `1000`.
  The private node reconnects immediately (idle ≠ error).
- Operators behind load balancers must set WS idle timeouts **above**
  `idle_timeout_s + 2 × keepalive_interval_s`.

### 4.6 Reconnect and session replacement

- The private node reconnects on any failure with **exponential backoff, full
  jitter**, base 1 s, cap 60 s, plus a `welcome`-independent health re-check.
  It dials its friends in RTT order and fails over to the next friend after 3
  consecutive failed dials (failover **OPEN** until `REVIEW.md` §7.7 lands; in
  v1 there is one active friend).
- Reconnect is a **fresh handshake**: new `hello`, new yamux session. There is
  no stream resume in v1; in-flight requests on the replaced session get
  `502 tunnel:session-replaced` and visitors retry.
- The public node **replaces the stale session on reconnect** (registry
  semantics): the old session is closed with `1000` *after* the new one
  completes its handshake, so a flapping connection never unroutes a node.
- A node that is absent has no session; requests to its prefix return `502`
  with `code: "tunnel:node-offline"` — fail fast, no queueing.

### 4.7 Prefix conflict

- Two live sessions claiming the same prefix: the **later** `hello` is rejected
  with `prefix_taken` and the older session keeps serving.
- The older session may re-`hello` (it owns the live claim) and wins.
- Claim lifecycle (release, grace period, reclamation/anti-impersonation) is
  **OPEN** (`REVIEW.md` §7.6) and lives in the control API; the tunnel only
  enforces "claimed and signed by the same key."

### 4.8 Close codes

| Code | Meaning |
|---|---|
| `1000` | normal (idle timeout, graceful shutdown, session replacement) |
| `1001` | going away (operator restart) |
| `4000` | protocol error (bad frame order, non-JSON handshake, yamux before handshake) |
| `4401` | handshake auth failure (`bad_signature`, `name_mismatch`) |
| `4409` | prefix conflict (`prefix_not_claimed`, `prefix_taken`) |
| `4429` | public node overloaded |

Application-level failure (one bad request) never closes the session; it closes
the stream and the public node answers the visitor with the mapped status.

### 4.9 Limits (defaults, operator-overridable)

| Limit | Default | Enforced by |
|---|---|---|
| Concurrent sessions per public node | 64 | public |
| Concurrent streams per session | 64 | public (yamux `MaxStreamCount`) |
| Concurrent in-flight HTTP requests per session | 32 | public |
| Request body | 1 GiB | both |
| Response body | unbounded (backpressured) | — |
| Idle timeout | 180 s | public |
| Keepalive interval | 30 s | both |
| yamux stream window | 1 MiB | both |

### 4.10 Security rules (normative)

1. TLS is mandatory; the tunnel path is not served over plain HTTP (Caddy
   auto-redirects).
2. `hello` must be the first frame; its `key` must match a claim for every
   prefix in `prefixes`; the signature must verify; `ts`/`nonce` must pass the
   replay window.
3. A session may only be dialed by the public node's own proxy: there is no API
   by which a visitor or a third node opens streams to another node.
4. The private node dials only the configured friend hosts and only its own
   loopback as a forward target (no SSRF surface: the agent is the client).
5. Neither side trusts visitor-supplied `X-Relay-*`, `X-Forwarded-*`, or
   `X-Real-Ip` (§4.4).
6. Per-session request rate limiting is **OPEN** (`REVIEW.md` §7.8, abuse model);
   v1 ships the limits in §4.9 and nothing more.

### 4.11 Versioning and compatibility

- `v` in `hello`/`welcome` is the major protocol version. The public node
  supports the version in the path plus at most one older major; a `hello` with
  an unsupported major gets `error { "code": "bad_version" }` and close `4000`.
- Unknown fields in handshake frames are ignored (forward compatibility).
- New capabilities are announced as strings in a future `hello.features` list;
  a feature is used only if **both** sides announced it.

**Rejected alternative (for the record):** the reference sketch used
`?agent=<id>` + a Bearer token on the upgrade. That puts identity in logs and
authenticates a string the server chose the meaning of; Relay's locked identity
model is an Ed25519 deed with a signed challenge-free canonical string, so the
signed `hello` replaces both. The yamux session is otherwise exactly as
sketched: public node = yamux client, registry keyed by node name.

## 5. Routing contract (visitor → node → site)

```
https://relay.example.org/<node>/<site>/<rest>
```

- The public node's router reads **segment 1** (`<node>`):
  - `<node>` == this node's own name → serve locally (hosted site: static +
    per-site SQLite + runtime API). No tunnel involved.
  - `<node>` is a connected friend → `reverse_proxy` through the tunnel
    transport, path **unmodified**, one stream per request (§4.3).
  - `<node>` is claimed but offline → `502 tunnel:node-offline`.
  - `<node>` unknown → `404` (never "is that a node or a site?" probing —
    unknown first segments are 404, full stop).
- `<node>` == `_relay` is reserved: `/relay/v1/...` is control (§6),
  `/_relay/tunnel/v1` is the tunnel (§4). `_relay` may never be claimed as a
  node name.
- Sites inside a node share the origin; per-site isolation is the runtime API's
  job (path-scoped cookies, CSRF, referer check — `REVIEW.md` §7.11, **OPEN**).
- The runtime API base is `/<node>/<site>/_relay/` — i.e. `_relay` as the
  **second** segment is the site's runtime namespace, not the node's.

## 6. Control API (`/_relay/v1`)

Owner surface. Every request is signed per §3. Base: `https://<host>/_relay/v1`.

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/node` | Node info: `name`, `version`, `rules` (version hash), `capacity` (sessions, storage), `friends` (public nodes it can reach). Unsigned, unauthenticated. |
| `GET` | `/rules` | Published rule set (plain-language rules + machine questions). Unsigned. |
| `POST` | `/claim` | Claim `name` with this key. Body: `{ "name", "node_version" }`. Returns the **key file** (see below). `409 name_unavailable` if taken; `403 rules_rejected` when the classifier holds a claim **OPEN**. |
| `GET` | `/names` | Names claimed by this key on this node. |
| `GET` | `/names/{name}` | Claim record: owner key (public), site versions, prefix status, tunnel state (`online`/`offline`, last-seen). |
| `PATCH` | `/names/{name}` | Update site metadata (title, description). Signed by the owning key. |
| `DELETE` | `/names/{name}` | Release the name. Release/reclaim grace is **OPEN** (§7.6). |
| `POST` | `/names/{name}/deploys` | **Manifest-first deploy.** Body: the deploy manifest (§6.1). Returns `201` + `version`. |
| `GET` | `/names/{name}/deploys` | List versions (newest first, cursor-paginated). |
| `POST` | `/names/{name}/deploys/{version}/rollback` | Roll back to a listed version. Returns the new live version. |
| `POST` | `/names/{name}/preview` | Render a folder/manifest to a throwaway preview path, valid `TTL` (default 30 min). Returns `{ "url", "expires" }`. |
| `GET` | `/names/{name}/export` | Full export: one SQLite file + blobs tar, streamed. `application/octet-stream`. |

### 6.1 Deploy manifest

```json
{
  "v": 1,
  "files": [
    { "path": "/index.html", "hash": "blake3-hex", "bytes": 1234, "mime": "text/html" },
    { "path": "/logo.png",   "hash": "blake3-hex", "bytes": 9011, "mime": "image/png" }
  ],
  "blobs": [
    { "hash": "blake3-hex", "bytes": 9011, "url": "https://…/upload-ticket" }
  ],
  "meta": { "title": "…", "description": "…" }
}
```

- Small files (`≤ 64 KiB`) may be inlined as base64 in `files[].data` instead of
  a blob. The server **refuses** any deploy where a hash does not match the
  bytes received. Content-addressed dedup happens on the node (locked, `REVIEW`
  §4).
- A deploy is atomic: all-or-nothing version bump; the previous version stays
  live until the new one is verified, enabling instant rollback.

### 6.2 Key file

What `POST /claim` returns (also what `relay key new` generates locally before
claiming):

```json
{
  "relay_key": "1",
  "name": "raspberry",
  "node": "relay.example.org",
  "public_key": "ed25519-hex",
  "private_key": "ed25519-hex",
  "created": 1760000000
}
```

The private half is the deed. Losing it means losing the name (no recovery —
to be stated loudly in the CLI, `REVIEW.md` §7.14 **OPEN**).

### 6.3 Operator admin + uploads (MVP bearer-token surface)

The Stage-0 MVP runs before §3 signatures land, so it uses two bearer tokens
instead of signatures: an **admin token** and per-user **deploy keys**, both
sent as `Authorization: Bearer <token>` and compared in constant time. Deploy
keys are stored on the node in a `0600` store with their plaintext token so
the operator console can reveal and copy them; uploads authenticate against
the token's SHA-256 in constant time.

Upload (any valid, non-revoked deploy key):

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/sites/{site}/deploy` | Replace `{site}` with a `tar` archive of its contents (`application/x-tar`). Returns `{ "site", "files", "bytes", "url" }`. |

Operator admin (admin token only; web console at `/_relay/admin.html`):

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/admin` | Overview: `{ "keys": [{ "id", "user", "token", "created", "revoked" }], "sites": [{ "site", "user" }] }`. |
| `POST` | `/admin/keys` | Issue a deploy key. Body `{ "user" }`; returns `{ "key", "token" }` (the token is also listed in `GET /admin`). |
| `POST` | `/admin/keys/{id}/revoke` | Revoke a key (idempotent). Returns `{ "key" }`. |
| `POST` | `/admin/keys/{id}/rotate` | Revoke and re-issue for the same user. Returns `{ "key", "token" }`. |

Error codes: `bad_token` (401), `bad_request` (400), `not_found` (404),
`too_large` (413), `bad_name` / `bad_archive` (400, uploads). This surface is
the MVP stand-in for the signed `POST /names/{name}/deploys` flow and is
removed once §3 signing is wired.

## 7. Runtime API (`/<node>/<site>/_relay/...`)

Called by a site's own pages, same-origin, no owner signature. Access is gated
by the site's configuration (collections' access rules, session cookies).

| Method | Path (under `/<node>/<site>/_relay/`) | Purpose |
|---|---|---|
| `GET` / `PUT` / `DELETE` | `db/v1/kv/{key}` | Flat KV. Key: 1–255 chars, any UTF-8 except NUL. Value: JSON, ≤ 1 MiB. `PUT` upserts. |
| `GET` | `db/v1/collections/{col}/items?cursor=&limit=` | List items, cursor-paginated (`limit` ≤ 200, default 50). |
| `POST` | `db/v1/collections/{col}/items` | Create item (JSON object, ≤ 256 keys, ≤ 1 MiB). Returns `201` + `id`. |
| `GET` / `PUT` / `DELETE` | `db/v1/collections/{col}/items/{id}` | Read / update (full replace) / delete one item. |
| `GET` | `db/v1/export` | Owner-only: the site's whole SQLite file. |
| `POST` | `auth/v1/signup` | Opt-in sites: `{ "id", "password" }` → Argon2id hash, session cookie. |
| `POST` | `auth/v1/login` | `{ "id", "password" }`. Lockout/backoff **OPEN** (§7.13). |
| `POST` | `auth/v1/logout` | Kill the session. |
| `GET` | `auth/v1/me` | Current session's identity, or `401`. |
| `GET` | `metrics/v1/summary?from=&to=` | Owner-only aggregate: views, pages, referrers, countries. |

- **Collection access rules** (per collection, set by the owner): `public` (any
  visitor may read; write owner-only) · `members` (authenticated readers) ·
  `owner` (only) · `private` (site code only, via an internal route the pages
  cannot reach — implementation detail, `REVIEW.md` §7.12 **OPEN**).
- **Session cookie:** name `_relay_s`, `Path=/<node>/<site>/`,
  `SameSite=Strict`, `HttpOnly`, `Secure`. Opaque value; stored hashed.
- **CSRF:** every mutating runtime call requires the per-site `X-Relay-Csrf`
  header, issued in the page bootstrap and bound to the session
  (mechanics **OPEN**, §7.11).
- **Filter/sort grammar** for collection queries is **OPEN** (§7.12); v1 is
  list + cursor only.

## 8. Visitor-visible behavior (summary)

| Situation | Response |
|---|---|
| Site on a connected private node | the site, streamed, with §4.4 headers added |
| Friend node offline | `502`, `code: "tunnel:node-offline"` |
| Friend tunnel stalled > 30 s | `504`, `code: "tunnel:timeout"` |
| Body over limit mid-upload | `413` |
| Unknown first path segment | `404` |

## 9. Open items (from `REVIEW.md` §7)

| Item | §7 ref | Status here |
|---|---|---|
| HOL / HTTP/2 over yamux / QUIC | 5 | Deferred; v1 = HTTP/1.1 per stream + 1 MiB window (§4.3) |
| Prefix release/reclaim, anti-impersonation | 6 | Control API has `DELETE /names`; grace period **OPEN** |
| Friend failover, gap-free handoff | 7 | v1 = single active friend; RTT-ordered list reserved (§4.6) |
| Abuse & quotas, rate limits, ToS | 8 | §4.9 limits only; rest **OPEN** |
| Rule schema / classifier I/O | 9 | `rules_rejected` code reserved, never sent |
| Threat model (smuggling, header injection…) | 10 | §4.10 normative rules are the seed; full model **OPEN** |