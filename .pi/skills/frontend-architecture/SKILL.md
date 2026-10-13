---
name: frontend-architecture
description: Relay frontend standard for the node UI and the per-site owner console. Functional programming is a first-class, valid paradigm for this codebase, and Effect-TS (effect.website) is the functional library of choice. Use when writing or reviewing frontend code, adding a UI page, or wiring live data from the Relay node API into a page.
---

# Frontend Architecture

The engineering standard for Relay's frontend. The frontend is the owner-facing side of
the node: `node-ui.html` (node admin — tunnels, sites, health) and `metrics.html`
(per-site owner console — deploys, rollback, daily-salted metrics). Today these are static
mockups; this skill is the standard they grow into. The Go side's decisions are locked in
`REVIEW.md` — the frontend must not contradict them (same-origin node API, domain
vocabulary, Ed25519 ownership stays server-side).

## Stack (locked)

| Concern | Decision |
|---|---|
| Language | TypeScript, strict mode. No `any`. |
| Paradigm | **Functional — a valid, first-class paradigm for this codebase** (`references/functional-paradigm.md`) |
| Functional library | **Effect-TS** — <https://effect.website> (`references/effect-tsx.md`) |
| UI rendering | Vanilla DOM. No UI framework until a page demonstrably outgrows it, with a stated reason |
| Styling | Hand-written shared CSS (`styles.css`). No CSS framework |
| Tooling | `tsc --noEmit` (types), `esbuild` (bundle only), `@effect/vitest` (tests). No other build machinery |
| Data | Same-origin node HTTP API; **every response decoded with `Schema`** |

## Route to the right reference first

- "Is this the right shape?" state, handlers, data flow, pure vs effectful →
  `references/functional-paradigm.md`
- Effect-TS API — services, layers, Schema, streams, HTTP client, testing →
  `references/effect-tsx.md`
- Relay-specific pages, layout, polling cadences, mockup migration → `references/relay-ui.md`

Read the matching reference before writing or reviewing frontend code. Do not skip this
step for "small" changes — it is cheap.

## Non-negotiables

1. **strict TypeScript is law.** `tsc --noEmit` clean, `noUncheckedIndexedAccess` on.
   No `any`: values from the outside world are `unknown` until a `Schema` decodes them.
2. **Untrusted data is decoded at the edge.** Fetch bodies, URL parameters, `localStorage`,
   and DOM input cross into app code only through `Schema.decodeUnknownEffect`. Decoded
   values are domain values. No manual `JSON.parse` outside `api/`.
3. **State is immutable data flowing one way.** Event → pure fold → new state → render.
   No mutation of shared values. No module-level `let` holding app state.
4. **Pure core, effectful edge.** Derivations and rules are pure functions (testable
   without a DOM). DOM writes, fetches, timers, and storage live in the edge, wrapped in
   `Effect`.
5. **Services are `Context.Service` + `Layer`.** No module-level singletons, no mutable
   global config. Page entries compose the layer graph; that is the only wiring point.
6. **Errors are typed and accounted for.** `Schema.TaggedError`. Every effect with a
   failure channel is caught at the edge and turned into rendered state or a log. No
   silent swallows, no unhandled rejections.
7. **Every subscription has a finalizer.** Polls, listeners, and timers are started
   scoped (`forkScoped`, `acquireRelease`) and cancelled on page teardown. No orphaned
   `setInterval`.
8. **Live data is a `Stream`.** Polling = `Stream.fromEffectSchedule(eff, Schedule.spaced(...))`;
   user events = `Stream.fromEventListener`. Views subscribe; nobody hand-loops.
9. **The view is a function of state.** DOM is updated by rendering state, not by
   sprinkling `el.textContent = …` across handlers. State is never read back out of the DOM.
10. **Boring code wins.** No UI framework, no state library, no CSS framework, no
    bundler tricks. Every new dependency needs a one-line justification in the PR.

## Layout (once the UI starts moving)

Mockups stay at the repo root as static pages. Real pages are built under `web/`:

```
web/
  node.html                   # node admin page (shell: styles + module entry)
  site.html                   # per-site owner console (shell)
  shared/styles.css
  src/
    main/node.ts              # entry: compose layers, mount, register teardown
    main/site.ts
    api/                      # node HTTP client service + Schemas. The only wire-aware code.
    tunnels/                  # feature service: tunnels, friends, node health
    deploys/                  # feature service: deploys, versions, rollback
    metrics/                  # feature service: counters, daily-salt aware
    dom/                      # pure views: (state) -> DOM. Imports types only, no services.
  test/
```

**Dependency rule:** `main → feature services → api`; `dom` imports types only; features
never touch the DOM; `api` is the only place that knows the wire format. Domain
vocabulary from `REVIEW.md` — site, prefix, deploy, tunnel, friend, key — is the UI's
vocabulary too.

## Definition of done (every PR)

- [ ] `tsc --noEmit` clean; esbuild build succeeds
- [ ] Every untrusted input decoded with `Schema` at the edge
- [ ] No module-level mutable state anywhere in `web/src`
- [ ] Every poll, listener, and timer has a finalizer; page teardown cancels all of them
- [ ] Failure and stale-data states are rendered, not `alert()`ed or swallowed
- [ ] Tests: pure functions table-driven; services tested with `layer(...)`; time via
      `TestClock`
- [ ] No new dependency without a stated reason
- [ ] Keyboard-usable: every control is a native element (`button`, `a`, `input`, …)
