# Relay UI architecture

Relay-specific rules for the frontend: the two pages, what data they show, how they talk
to the node, and how the static mockups become real pages.

## The pages (locked)

| Page | Shell | Audience | Contents |
|---|---|---|---|
| Node admin | `node.html` (mockup: `node-ui.html`) | Node operator | tunnels + friends, sites list, node health |
| Site console | `site.html` (mockup: `metrics.html`) | Site owner | deploys, versions, rollback, daily-salted metrics, prefix claims |

There is no third page without a stated reason. Cross-page links are plain `<a>` tags —
no router. The pages are short-lived documents, not a SPA.

## Same-origin node API

The node serves its own UI and its own API over one origin, path-based (the same
addressing model as Relay traffic). Consequences:

- Relative URLs only (`/api/tunnels`, `/api/site/deploy`). No CORS, no cross-origin
  config, no proxy.
- One `api/` service for both pages (different feature services on top).
- The API's JSON is the serialized Go domain. UI code uses the same vocabulary —
  `site`, `prefix`, `deploy`, `tunnel`, `friend`, `key` — and Schemas mirror it.

## Ownership and secrets stay server-side

- Site identity is Ed25519 on the Go side; console actions (deploy, rollback, claim)
  are authorized by the node from the browser's session on this origin.
- The client never models private keys, never stores key material in
  `localStorage`/`IndexedDB`, and never signs requests itself. If a design requires
  client-side key material, stop and re-discuss — it contradicts the locked backend.

## Live data

| Data | Source | Cadence (default, `Context.Reference`-overridable) |
|---|---|---|
| Tunnels, friends, node health | `/api/tunnels` poll | 5s |
| Deploys, versions | `/api/site/deploy` poll | 10s |
| Metrics (daily-salted counters) | `/api/site/metrics` poll | 30s |

- Polling is a `Stream.fromEffectSchedule` in a scoped subscription (see
  `effect-tsx.md`), not `setInterval`.
- A failed poll is a `stale` event: the view keeps last-known data and shows
  "last updated …s ago". Retries use the HTTP client's `retryTransient`; the poll loop
  itself never dies.
- Do not poll faster than the data changes. Daily-salted counters change at salt
  granularity; 30s is plenty and it is what the Go side was designed around.

## Page state machine

Every page's fold tracks, at minimum:

```
loading  →  ready
ready    ⇄  stale        (poll failed; last data still shown)
stale    →  error        (never got data, or data expired past threshold)
```

These are rendered (banner/section state), never `alert()`ed. Actions (deploy, rollback,
claim) are optimistic about UX, pessimistic about truth: the page shows "working…" from
the click, and the next poll is what confirms it.

## Mockup migration

`node-ui.html` and `metrics.html` (repo root, shared `styles.css`) are static mockups and
the visual reference. When a page becomes real:

1. Create `web/<page>.html` as a thin shell: `shared/styles.css` + one module entry.
2. Copy the mockup's structure and classes as the initial view output — visual parity is
   the acceptance bar.
3. Build the feature services and fold; the view renders the shell's regions from state.
4. Leave the mockups at the root untouched — they remain the design reference. Do not
   edit them in place.

## Do not confuse the two JavaScripts

- **This frontend** = the owner's browser, TypeScript + Effect, talks to the node API.
- **Site runtime** = per-site QuickJS (Wazero) on the Go side, sandboxed, CPU-bounded —
  the thing sites deploy code into. It is not the frontend, shares no code with it, and
  must not be referenced from `web/` code or tests. If a PR blurs the two, it is wrong.

## Accessibility baseline

- Native elements only for interactive things: `button` (not `div onclick`), `a` with
  real hrefs, `label` + `input`, `table` for metrics grids.
- Status changes (tunnel open/closed, deploy progress) get an `aria-live` region so
  screen readers hear the live data.
- Keyboard: tab order follows visual order; nothing requires a pointer.