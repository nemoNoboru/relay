# Relay owner UI

The node's owner-facing pages, embedded into the Go binary and served
same-origin under `/_relay/`:

| Page | Audience | What it does |
|---|---|---|
| `home.html` | member | Exchange an invite for a cookie, welcome the member, read the feed, edit display name/bio, sign out |
| `people.html` | member | One member's profile and posts (`/_relay/people/<id>`) |
| `post.html` | member | Post wrapper with author, date, static page iframe, comments and delete-own (`/_relay/posts/<site>`) |
| `admin.html` | host | Sign in with the admin password; create, reveal, revoke, rotate invitations; see pages by member |
| `publish.html` | member | Template picker, install a template tar, free-form publish, invite someone else |
| `deploy.html` | member | Step-by-step terminal upload instructions |

The pages are short-lived documents, not a SPA. They are built with Svelte 5
(runes), Effect-TS, and Vite, and use the design tokens from the public site.

The old `node.html` and `site.html` mock consoles were removed in the pivot
cleanup. They polled `/api/tunnels` and `/api/site/deploy`, which the node never
served. Do not bring them back without a real API and a product reason.

## Layout

```
home.html                member welcome + feed shell
people.html              profile shell (nested route)
post.html                post wrapper shell (nested route)
admin.html               host console shell
publish.html             publish page shell
deploy.html              static guide (no script)
shared/styles.css        owner UI tokens + components
public/skills/           generated: relay-deploy/SKILL.md, copied at build time
src/main/home.ts         member home entry: invite exchange, /me, feed, profile
src/main/people.ts       profile entry: member id from the path
src/main/post.ts         post wrapper entry: site from the path
src/main/admin.ts        host console entry: token session + API effects
src/main/publish.ts      publish page entry: invite, files, tar, preview
src/home/state.ts        home UI state
src/people/state.ts      profile UI state
src/post/state.ts        post wrapper UI state
src/member/schema.ts     member, post, feed, profile schemas
src/member/ops.ts        session, profile, feed, post, comment effects
src/templates/schema.ts  template manifest schemas
src/templates/ops.ts     template listing effect (install is in main/publish.ts)
src/admin/schema.ts      /_relay/v1/admin wire schemas
src/admin/ops.ts         host console API effects
src/admin/state.ts       host console UI state type + initial
src/ui/Home.svelte       member welcome + feed view (pure renderer)
src/ui/FeedList.svelte   infinite-scroll feed (cursor + IntersectionObserver)
src/ui/PostCard.svelte   one post: cover, metadata, inline comments/form
src/ui/People.svelte     profile view (pure renderer)
src/ui/PostPage.svelte   post wrapper view (pure renderer)
src/ui/Admin.svelte      host console view (pure renderer)
src/publish.ts           invite links, Markdown-to-page, tar, generated page
src/api/client.ts        fetch + Schema decode, typed ApiError (only wire code)
src/shims-svelte.d.ts    .svelte module shim for plain `tsc`
tools/copy-skill.mjs     copies .pi/skills/relay-deploy/SKILL.md → public/skills
web/public/template/     shared prepublish runtime (JS + CSS)
tools/smoke.mjs          DOM-stub smoke test of the built admin + home bundles
test/                    @effect/vitest suites
embed.go                 Go: //go:embed all:dist → webui.FS
dist/                    Vite output (generated; .gitkeep keeps embed valid)
```

## Scripts

| script | what it does |
|---|---|
| `npm run dev` | Vite dev server (no API proxy; the admin page renders its sign-in state) |
| `npm run build` | typecheck (`tsc --noEmit` + `svelte-check`) + Vite build → `dist/` |
| `npm run preview` | serve `dist/` locally |
| `npm run typecheck` | `tsc --noEmit` + `svelte-check` |
| `npm test` | `vitest run` — API client, publish helpers, SSR home/admin renders |
| `npm run smoke` | DOM-stub smoke test of the built admin and home bundles |

Node >= 22.12 is required. `smoke` needs a fresh `dist/`
(`npm run build` first).

## Building for the Go binary

The root `make build` handles the order:

1. `web/` → `npm ci` + `npm run build` → `web/dist/`
2. root → `go build ./...` — `web/embed.go`'s `//go:embed all:dist` picks it up.

`dist/` is committed only as `.gitkeep`; real build output is git-ignored.
The fixed asset base (`base: "/_relay/"` in `vite.config.ts`) is required for
nested shells such as `/_relay/people/<id>` and `/_relay/posts/<site>`. Do not
change it to a relative base.

## Wire contracts

The host console talks to the admin API in [`docs/api.md`](../docs/api.md).
The member pages talk to `/_relay/v1/session`, `/_relay/v1/me`,
`/_relay/v1/feed`, `/_relay/v1/posts/<site>`, `/_relay/v1/people/<id>`, the
comment endpoints under `/_relay/v1/posts/<site>/comments`, and
`/_relay/v1/templates`. Template forms post finished files to
`/_relay/v1/publish`. The host console deletes any comment with
`DELETE /_relay/v1/admin/comments/<id>`.
Every response is decoded with a `Schema` (`src/admin/schema.ts`,
`src/member/schema.ts`); `src/api/client.ts` is the only code that knows about
HTTP. Invitation tokens are bearer credentials and are never put in a URL.

The member session, welcome, post index, feed, profiles, post wrappers,
comments, and template publishing are built (Phases 1–4). Phase 5 host
operations are planned in [`docs/pivot-plan.md`](../docs/pivot-plan.md).

## Testing

Browser-free:

1. `test/home.test.ts` — SSR-renders the member welcome, empty feed, and feed
   with posts.
2. `test/people.test.ts` — SSR-renders a profile with posts and its empty and
   signed-out states.
3. `test/post.test.ts` — SSR-renders the post wrapper, comment list, and form.
4. `test/admin.test.ts` — SSR-renders the host console in its locked, ready,
   first-sign-in, and one-time-invitation states.
5. `test/client.test.ts` — the API client against a stub `fetch`: schema
   decode, typed errors, bearer/JSON request shape, PATCH/DELETE.
6. `test/publish.test.ts` — invite parsing, tar layout, Markdown-to-page,
   generated page copy, assistant prompt.

The feed uses cursor pagination (`next` from `/_relay/v1/feed`) and a sentinel
watched by `IntersectionObserver`. Each post carries its comments, so the card
renders the list and form without an extra request.

`tools/smoke.mjs` imports the built production bundles into Node with a
minimal DOM stub and asserts the admin sign-in state and the home signed-out/
signed-in states. It is the closest this setup gets to a browser; a real
headless browser would still be a stronger check.

## Frontend rules

The engineering standard is
[`.pi/skills/frontend-architecture`](../.pi/skills/frontend-architecture/SKILL.md)
and its [`relay-ui.md`](../.pi/skills/frontend-architecture/references/relay-ui.md)
reference. The short version:

- strict TypeScript, no `any`;
- decode every wire response with a `Schema`;
- pure renders, effects at the edge;
- no client-side key material in storage;
- relative URLs only;
- native elements for controls;
- no new dependency without a written reason.
