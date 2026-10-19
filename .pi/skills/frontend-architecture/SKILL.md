---
name: frontend-architecture
description: Relay frontend standard for the owner UI (host console, publish page, deploy guide). Use when writing or reviewing code under web/, adding a page or component, or wiring a node API response into the UI.
---

# Frontend Architecture

The engineering standard for Relay's owner-facing frontend in `web/`. It covers
the **member home** (`home.html`), the **host console** (`admin.html`), the
**publish page** (`publish.html`), and the **deploy guide** (`deploy.html`). The public landing and technicals pages
are hand-written/generated HTML at the repository root and are not part of this
bundle.

The product is a micro-social network for a group of friends. The frontend must
not promise features that are not built. The pivot plan is
[`docs/pivot-plan.md`](../../../docs/pivot-plan.md); the API is
[`docs/api.md`](../../../docs/api.md).

## Stack (locked)

| Concern | Decision |
|---|---|
| Language | TypeScript, strict mode. No `any`. |
| UI | Svelte 5 (runes) components |
| Functional core | Effect-TS for effects, services, decoding, and any future live data |
| Build | Vite (esbuild for TS, Rollup for the bundle) |
| Styling | Hand-written shared CSS (`shared/styles.css`). No CSS framework |
| Data | Same-origin node HTTP API; **every response decoded with `Schema`** |
| Tests | `tsc --noEmit`, `svelte-check`, Vitest (`@effect/vitest`), DOM-stub smoke |

## Route to the right reference first

- State, handlers, pure vs effectful → `references/functional-paradigm.md`
- Effect-TS API — services, layers, Schema, HTTP, testing → `references/effect-tsx.md`
- Relay-specific pages, layout, copy, and wire shapes → `references/relay-ui.md`

Read the matching reference before writing or reviewing frontend code.

## Non-negotiables

1. **Strict TypeScript is law.** `tsc --noEmit` clean, `noUncheckedIndexedAccess`
   on. No `any`: outside data is `unknown` until a `Schema` decodes it.
2. **Untrusted data is decoded at the edge.** Fetch bodies, URL fragments,
   `sessionStorage`, and DOM input cross into app code only through
   `Schema.decodeUnknownEffect`. No manual `JSON.parse` outside `api/`.
3. **State is immutable data flowing one way.** Event → pure fold/derivation →
   new state → render. No mutation of shared values. No module-level `let`
   holding app state.
4. **Pure core, effectful edge.** Derivations and rules are pure functions
   (testable without a DOM). DOM writes, fetches, timers, and storage live in
   the edge, wrapped in `Effect`.
5. **Services are `Context.Service` + `Layer`.** No module-level singletons, no
   mutable global config. Page entries compose the layer graph.
6. **Errors are typed and accounted for.** `Schema.TaggedError`. Every effect
   with a failure channel is caught at the edge and turned into rendered state.
   No silent swallows, no unhandled rejections.
7. **Every subscription has a finalizer.** Polls, listeners, and timers are
   started scoped and cancelled on page teardown. No orphaned `setInterval`.
8. **Live data is a `Stream`.** Do not add polling by hand. The current host
   console is action-driven (load, create, revoke, rotate), not polled.
9. **The view is a function of state.** Svelte renders state; do not sprinkle
   `el.textContent = …` across handlers.
10. **Boring code wins.** No router, no state library, no CSS framework. Every
    new dependency needs a one-line justification in the PR.

## Layout

```
web/
  home.html                member welcome + feed shell
  people.html              profile shell (nested route)
  post.html                post wrapper shell (nested route)
  admin.html               host console shell
  publish.html             publish page shell
  deploy.html              static guide (no app bundle)
  shared/styles.css        owner UI tokens and components
  src/
    main/admin.ts          host console entry: session + API effects + mount
    main/publish.ts        publish page entry: invite, files, tar, preview
    api/client.ts          the only wire-aware code: fetch + Schema decode
    admin/schema.ts        admin wire schemas
    admin/ops.ts           admin API effects
    admin/state.ts         host console UI state
    ui/Home.svelte         pure view of the home store
    ui/PostCard.svelte     one post in a feed/profile
    ui/People.svelte       pure view of the profile store
    ui/PostPage.svelte     pure view of the post wrapper
    ui/Admin.svelte        pure view of the admin store
    publish.ts             invite links, Markdown-to-page, tar, generated page
  test/                    client, publish, and SSR admin tests
  tools/smoke.mjs          DOM-stub smoke test of the built admin + home bundles
```

**Dependency rule:** `main → feature (admin/publish) → api → Schema`. Views
import types only; they never call `fetch`.

## Copy and product rules

- Brand line: **a micro-social network for you and your real friends**.
- Use "member", "invitation", "page/post", "host", "node". Avoid "user account",
  "deploy key" in member-facing copy, and "site" when "page" is clearer.
- Say what is built. Planned features (home feed, profiles, comments,
  templates) are marked as planned in docs and must not appear as live
  controls. The member cookie and welcome page are built.
- The host console may use "invitation", "token", and "admin password"; it is
  the one place where operational language is appropriate.
- No ideology in UI copy. A node is a place for a group, not a movement.

## Definition of done (every PR)

- [ ] `tsc --noEmit` clean; `svelte-check` clean; `npm run build` succeeds
- [ ] Every untrusted input decoded with `Schema` at the edge
- [ ] No module-level mutable state anywhere in `web/src`
- [ ] Every subscription or timer has a finalizer
- [ ] Failure states are rendered, not `alert()`ed or swallowed
- [ ] Tests cover the changed pure logic and every visible state
- [ ] No new dependency without a stated reason
- [ ] Keyboard-usable: native `button`, `a`, `input`, `label`
- [ ] Copy says what the product actually does today
