# Relay owner UI

The node's owner UI: `node.html` (node admin — tunnels, health) and
`site.html` (site console — deploys, metrics). Svelte 5 components + Effect-TS,
built with Vite, embedded into the Go node binary via `embed.go`, served by the
node same-origin with its API.

Locked by `.pi/skills/frontend-architecture/SKILL.md`:

- TypeScript `strict` (plus `noUncheckedIndexedAccess`), no `any`.
- Effect-TS: services/layers, `Stream` for live data, finalizers for teardown.
- Functional paradigm; Svelte 5 (runes) components — see Deviations; shared
  hand-written CSS — no CSS framework.
- Every API response decoded with a `Schema`; `src/api/` is the only
  wire-aware code.
- No router, relative URLs only, no client-side key material.
- Live-data state machine: `loading → ready ⇄ stale → error`
  (`src/fold.ts`, pure, table-driven tested in `test/`).

## Layout

```
node.html, site.html     page shells (thin; the app is the script module)
admin.html               admin mock page shell
deploy.html              static upload docs page (plain HTML, no script)
shared/styles.css        design tokens (mirrored from the root site) + owner UI
public/skills/           generated: relay-deploy/SKILL.md, copied from .pi/skills
                         at build time and served at /_relay/skills/…/SKILL.md
src/live.ts              the live-page driver both pages share: store, scoped
                         poll Stream, fold, Svelte mount, pagehide teardown
src/main/node.ts         node page entry — the options handed to startPage
src/main/site.ts         site page entry — same shape
src/ui/Phase.svelte      shared presentation of the live phases (loading,
                         error, stale); the page's data view is a snippet
src/ui/Panel.svelte      shared panel wrapper (title + body snippet)
src/ui/Node.svelte       node admin view: pure renderer of the live-data store
src/ui/Site.svelte       site console view: same shape
src/shims-svelte.d.ts    .svelte module shim for plain `tsc` (svelte-check does the real check)
src/api/client.ts        fetch + Schema decode, typed ApiError — the only wire code
src/fold.ts              pure (state, event) → state live-data machine (+ `ago`)
src/tunnels/schema.ts    GET /api/tunnels — DRAFT wire contract
src/deploys/schema.ts    GET /api/site/deploy — DRAFT wire contract
tools/copy-skill.mjs     copies .pi/skills/relay-deploy/SKILL.md → public/skills/
test/                    @effect/vitest suites (fold table, client, SSR component render)
tools/smoke.mjs          DOM-stub smoke test of the built production bundles
embed.go                 Go: //go:embed all:dist → webui.FS for the node binary
dist/                    Vite output (generated; .gitkeep keeps embed valid pre-build)
```

## Scripts

| script              | what it does                                        |
| ------------------- | --------------------------------------------------- |
| `npm run dev`       | Vite dev server (no API proxy; a missing node renders the designed stale/error state) |
| `npm run build`     | typecheck (`tsc --noEmit` + `svelte-check`) + Vite build → `dist/` |
| `npm run preview`   | serve `dist/` locally                               |
| `npm run typecheck` | `tsc --noEmit` + `svelte-check` (strict Svelte + prop typing) |
| `npm test`          | `vitest run` — fold table, client, SSR component render |
| `npm run test:watch`| `vitest`                                            |
| `npm run smoke`     | DOM-stub smoke test of the built bundles (all four page/state scenarios) |

Node >= 22.12 is required (engines). `smoke` needs a fresh `dist/`
(`npm run build` first).

## Building for the Go binary

The root `make build` handles the order (frontend first, then `go build`):

1. `web/` → `npm ci` + `npm run build` → `web/dist/`
2. root → `go build ./...` — `web/embed.go`'s `//go:embed all:dist` picks it up.

`dist/` is committed only as `.gitkeep`; real build output is git-ignored.
The relative asset base (`base: "./"` in `vite.config.ts`) is what makes the
embed work at arbitrary path depth — do not change it.

## Wire contracts — drafts

`/api/tunnels` and `/api/site/deploy` have no finalized spec yet (REVIEW.md
§7.5). `src/tunnels/schema.ts` and `src/deploys/schema.ts` are the UI's draft
of the shapes, written in Relay's domain vocabulary. When the control/tunnel
API spec lands, update these two files — they are the only place the wire
shape is known, and every request is decoded through them.

## Testing

Three layers, all browser-free:

1. **`test/fold.test.ts`** — table-driven tests of the pure live-data state
   machine (`loading → ready ⇄ stale → error`) and `ago`.
2. **`test/client.test.ts`** — the API client against a stub `fetch`: schema
   decode, typed errors, the only wire-aware code.
3. **`test/ui.test.ts`** — SSR-rendering of the Svelte components
   (`svelte/server` `render`) for every visible state: loading, ready (with
   data and empty), stale (last data retained), and error.

`tools/smoke.mjs` goes one step further: it imports the **built production
bundles** into Node with a minimal DOM stub (just enough for Svelte's client
runtime — templates, anchors, text nodes) and asserts the rendered HTML for
all four page/state scenarios. It is the closest this setup gets to a browser
— a real headless browser would still be a stronger check.

## Deviation from the skill (flagged)

The skill pins tooling as "esbuild bundle only, no other build machinery" and
the rendering approach as "vanilla DOM, no UI framework". **Vite and Svelte
were both explicitly requested for this setup** and override those two lines:

- Vite uses esbuild for TS and Rollup for bundling.
- Svelte 5 (runes) components replace hand-rolled DOM wiring. The Effect
  layer is unchanged: `src/live.ts` (the shared live-page driver) still owns
  the poll `Stream` and closes it on `pagehide`; Svelte receives the state
  through a `svelte/store` `writable` (the only bridge — Svelte has no public
  state factory usable from `.ts` files), so the state machine, schemas, and
  client stay exactly as before.

Two tooling consequences, also flagged: `svelte-check` (the Svelte 5 checker
in this registry, unscoped) requires TypeScript `^5 || ^6`, so `typescript`
is pinned to 6.0.3; and `.svelte` imports resolve for `tsc` via
`src/shims-svelte.d.ts` (a module shim) — `svelte-check` performs the real
strict checking of the components themselves.

If the skill's tooling/rendering lines should be amended, that should be done
as a skill change, not silently here.