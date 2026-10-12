# Relay — Pre-Development Review

_Read this once before we start building. It is the whole design in one place:
what Relay is, the decisions that are already made, the questions that still need
your call, and everything that is not finished yet._

_Sources consolidated here: `vision`, `architecture`, `api`, `naming`, `tls`,
`sidecar`, `minimal`, `readiness`, `classifier`, `language`, `research`. The
public site is generated from [`docs/vision.md`](docs/vision.md) and
[`docs/technicals.md`](docs/technicals.md), which are kept because they build
`manifesto.html` and `technicals.html`._

---

## How to use this

- **§2 Decisions** are locked unless you reopen them.
- **§7 Open questions** need your decision before or during development.
- **§8 Not finished** is the pre-build checklist.
- Everything is deliberately condensed for review; we can expand any section
  back into a full spec once you sign off.

---

## 1. The product in one page

**The problem.** Making a website is easy now; putting one online is not. A
finished folder is not a website until someone can type an address and see it,
and today that last mile runs through accounts, billing, and platforms you don't
own. The web has landlords now: you make the thing, they keep it, and you pay to
reach it.

**What Relay is.** The last mile, and nothing else. It gives what you built a
home on the open web without asking a platform for permission and without
handing over the keys. It is not a language, a framework, or a runtime.

**What it gives you:**

- **A home** — point it at a folder; it gets a real address. No build step, no
  server to babysit, no account.
- **A memory** — a site can keep what people give it, in a form you own and can
  carry anywhere.
- **A way in** — the computer on your desk can serve the world from behind a home
  router. No port forwarding, no static IP.
- **People and numbers, if you want them** — opt-in logins and stats, with no
  central account and no tracker.
- **Just the wire** — if you already run your own server, Relay can front it and
  make it reachable (an ngrok you own).

**The network.** A constellation of independently run Relay nodes, in the lineage
of the BBS sysop and the tilde community — not three clouds. Any node can relay
for any other; there is no centre, only a default starting point.

**The rules.** Who decides what gets published? Not an unaccountable ToS, but a
**published rule set checked in the open**, with a human review path for anything
borderline. If you dislike a server's rules, read them, argue, or fork the server.

**The promises.** You hold the key to your address. Your data leaves with you.
There are no central accounts — ever. No lock-in.

---

## 2. Architecture in one page

One Go binary that runs as either a **public node** or a **private node** — the
same code, one flag. It reuses proven parts instead of inventing them.

```
visitor ──HTTPS──► Caddy ──► Relay node ──┬─► per-site SQLite + blobs (hosted)
                                          └─► outbound WebSocket tunnel ──► friend node
```

- **Addressing:** one hostname, one certificate, and every site is a **path**:
  `https://relay.cloudpunk.org/<node>/<site>/`. No wildcards, no per-site certs.
- **Reachability:** a node behind NAT opens one **outbound WebSocket** to a public
  friend and registers its path prefixes; requests route by the first path segment.
- **Storage:** **one SQLite file per site** — isolation, trivial export, no
  cluster — plus content-addressed blobs for files.
- **Login / stats:** opt-in, per site, stored in the site's own database.
- **Compute (later):** functions in a QuickJS sandbox.
- **Deployment:** one binary, or Docker / Compose / Docker Swarm / Kubernetes
  (examples in [`docs/technicals.md`](docs/technicals.md)).

---

## 3. Decisions (locked)

| # | Decision | Choice | Why |
|---|---|---|---|
| 1 | Host language | **Go** | single static binary; cheap concurrency; the whole reference ecosystem (Caddy, piko, Portal) is Go |
| 2 | Node shape | **one binary, two roles** | "every node is the same; some are public" |
| 3 | Addressing | **path-based**, one host, one cert | removes per-site TLS entirely; identical code public and private |
| 4 | Storage | **SQLite per site** + content-addressed blobs | isolation, one-file export, no cluster |
| 5 | Login | **opt-in, per site**, in the site DB | no central account, ever |
| 6 | Metrics | **opt-in, cookieless, aggregate**, owner-keyed, exportable | "see your site, don't follow your visitors" |
| 7 | Classifier | **provider design**; consume Laya, do not reimplement its runtime | avoids an ML project inside a hosting project |
| 8 | Scope | **3 primitives + opt-ins**, invite-only until the classifier exists | minimum complexity; done beats perfect |
| 9 | Deployment | binary · Docker · Compose · Swarm · Kubernetes | where self-hosters already are |

**Accepted trade-offs** (real costs we chose): no origin isolation between sites
(shared hostname); the relay sees plaintext and could tamper; addresses change
host when a site moves; no custom domains yet; one public friend is a single
point of failure until failover exists.

---

## 4. The API in brief

**Identity.** An **Ed25519** keypair is the deed to a name. All owner requests are
signed (canonical string + timestamp + nonce), so the secret never travels. The
private half is the downloadable "key file."

**Surfaces**

| Surface | Base | Caller |
|---|---|---|
| Control | `/_relay/v1/...` on the node | deploy tool / UI |
| Runtime | `/<node>/<site>/_relay/{db,auth,metrics}/v1/...` | the site's own pages (same-origin) |
| Tunnel | `wss://<public-host>/_relay/tunnel/v1` | friend nodes |

**Control (owner):** `GET /node`, claim/`GET`/`PATCH`/`DELETE` a name, manifest-first
**deploy** (+ versions, rollback, preview).

**Database (runtime):** KV (`GET/PUT/DELETE .../kv/{key}`) and collections
(CRUD + cursor pagination) with per-collection access rules
(`public`/`members`/`owner`/`private`); full export as one SQLite file.

**Login (runtime, opt-in):** `signup` / `login` / `logout` / `me` with Argon2id
passwords and a path-scoped session cookie; optional Ed25519 **key login**.

**Metrics (owner):** aggregate views/pages/referrers/countries; uniques by a
**daily-salted hash**; no persistent identifier; exportable with the site.

**Tunnel:** authenticate by key, register **path prefixes**, then multiplex
request/response frames over the one WebSocket.

_Detailed contract drafted in [`docs/api-contract.md`](docs/api-contract.md)
(tunnel wire protocol + control + runtime) and in
[`docs/architecture.md`](docs/architecture.md) for the backend. Both await
sign-off (see §8)._

---

## 5. The classifier (Laya) in brief

**Laya** (`convaiinnovations/laya`, Apache-2.0) is a non-autoregressive, calibrated
**typed-decision** model: give it a state (page/email/JSON) plus typed questions
(`choice` / `score` / `noul`) and it returns typed answers with probabilities in
one forward pass. It never generates text.

- **`laya-multilingual`** is the one to start with: backbone **mmBERT-base**
  (MIT), 322M params, ~650 MB, 100+ languages, context up to 8192, ~2.2× faster
  than the English checkpoint.
- It is **not small enough to casually embed**, and it ships **uncalibrated**
  (over-confident; refitting temperatures moved ECE 0.314 → 0.106).

**Decision: use it as a provider, don't reimplement it.**

1. `none` — default; open deploys **disabled** (invite-only).
2. `heuristic` — a tiny embedded fallback.
3. `laya-service` — talk to `laya-serve` (Python) over a local socket.
4. `laya-onnx` — the official ONNX export via ONNX Runtime (sidecar or cgo).
5. `laya-local` — a pure-Go engine — **last resort only**.

**The decisive experiment:** export `laya-multilingual` to ONNX, quantize to INT8,
and measure calibration (ECE) on Relay's own rule questions.

---

## 6. First slice (what we build first)

One binary, one host, one certificate →
`relay key new` → `relay deploy ./folder` → live at `/<node>/<site>/` → rollback →
`relay export`.

**Acceptance test:** a person with no accounts runs the binary on a fresh machine,
deploys a folder, fetches it over HTTPS at a path, redeploys and rolls back without
the name or key changing, and exports the site's data as a file they can read
elsewhere. When that works end to end, Stage 0 is done — and only then is the
tunnel worth starting.

---

## 7. Open questions (need your call)

**Before writing code**

1. **Caddy, or the standard library?** Every doc leaned on Caddy for **on-demand
   TLS**, which the path decision deferred. With one host and one cert, Go's
   stdlib + `autocert` may suffice and drops a large dependency. *Decide.*
2. **Repo scaffolding:** *Decided:* module `github.com/nemoNoboru/relay`; layout
   `cmd/relay` + `internal/domain/…`; CI (`.github/workflows/test.yml`) runs
   `gofmt`, `go vet`, `staticcheck`, and `go test -race` on push to `main` and
   on PRs, with pre-commit as the local gate. *Still open:* **LICENSE** (pick
   before the first release); two-node tunnel integration tests arrive with
   the tunnel.
3. **Domains:** do we own `cloudpunk.org` / `relay.cloudpunk.org`?
4. **Versioning:** policy for the protocol (`v1`) and for node/per-site data
   migrations.

**Before the tunnel (Stage 1)**

5. **Tunnel protocol spec:** framing, stream lifecycle, **flow control /
   backpressure**, multiplexing, the head-of-line-blocking fix (HTTP/2 vs QUIC),
   heartbeats, reconnect/resume, error semantics.
6. **Prefix claim & namespace policy:** who owns `/raspberry/`, reserved words,
   collisions, and especially **release/reclaim** (can someone take a freed prefix
   and impersonate a site?).
7. **Friend lifecycle:** authentication, binding a prefix, reconnecting to the
   same prefix, moving between public nodes, gap-free failover.

**Before any open hosting**

8. **Abuse & quota model:** storage/bandwidth/CPU limits, rate limits, invite
   tokens, takedown path, and a **ToS stance**.
9. **Rule schema & classifier I/O:** how a rule becomes typed questions, how a
   site becomes the state, and the human review queue.
10. **Threat model:** path traversal/symlinks, header injection, request
    smuggling, SSRF from compute, sandbox escape, signature canonicalization,
    cross-site DB access on a shared origin.

**Before real users**

11. **Same-origin isolation mechanics:** cookie `Path`/`SameSite`, CSRF token
    design, `Referer` policy, and `<base>`/absolute-path handling (does it break
    SPAs?).
12. **DB API semantics:** filter/sort/pagination grammar, limits, indexing, and
    how the server derives the site from the path so one site cannot read another.
13. **Auth details:** session TTL, lockout/backoff, key-login flow, and stating
    loudly that **there is no account recovery**.
14. **Key lifecycle:** generation, on-disk protection, rotation, loss, release.
15. **Blob storage & GC:** on-disk layout and garbage collection of unreferenced
    blobs after rollback/delete.
16. **CLI / first-run UX:** the command surface and what a first-time user sees.
17. **Content integrity:** the relay can modify hosted content; do we add
    verifiable manifests or just document the risk?

**Before scaling**

18. **Observability:** node metrics, structured logs, health endpoints.
19. **Backup / restore / upgrade** procedure (node DB, blobs, per-site SQLite).
20. **Laya decision:** which checkpoint, and does the INT8 calibration experiment
    pass? If not, Laya stays an external service.

---

## 8. Not finished (pre-build checklist)

- [x] **Go scaffolding:** `go.mod` (`github.com/nemoNoboru/relay`), `cmd/relay`
      (`relay version`, `relay key new`), `internal/domain/{key,prefix}` with
      tests, CI, and pre-commit. **`LICENSE` is still missing** — pick one
      before the first release (§7.2).
- [x] **Specs to write:** `tunnel.md` — drafted as the tunnel section of
      [`docs/api-contract.md`](docs/api-contract.md); Stage 1 unblocked pending
      sign-off. Remaining: `threat-model.md`, `abuse.md`, `cli.md`, and the
      **rule schema**.
- [ ] **The calibration experiment** (ONNX → INT8 → ECE on Relay's questions).
- [ ] **Caddy decision** (§7.1) — the tunnel brief selects **embedded Caddy**
      with two custom modules (`docs/architecture.md` §4); confirm to lift the
      ingress block. Stdlib/autocert remains the cheaper path for Stage 0.
- [ ] **Deploy manifests are drafts** (no binary to build yet).
- [ ] **UI pages are static mockups** (`node-ui.html`, `metrics.html`), not code.
- [ ] **Tunnel is unproven** — the riskiest primitive, still a sketch.
- [ ] **No abuse/ToS posture** — keep servers invite-only until it exists.

---

## 9. What changed

- Consolidated into this file: `architecture`, `api`, `naming`, `tls`, `sidecar`,
  `minimal`, `readiness`, `classifier`, `language`, `research`.
- Kept: [`docs/vision.md`](docs/vision.md) and
  [`docs/technicals.md`](docs/technicals.md) (they generate the public pages),
  [`README.md`](README.md), the site HTML, and `tools/`.
- The detailed specs can be re-expanded from here, section by section, as we
  start each stage.
