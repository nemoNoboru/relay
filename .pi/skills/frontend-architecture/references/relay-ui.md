# Relay UI architecture

Relay-specific rules for the owner UI: the pages, what they show, and how they
talk to the node. The product is a micro-social network for a group of friends.
The product direction is [`docs/product.md`](../../../../docs/product.md); the
visual system is the [`relay-design` skill](../../relay-design/SKILL.md) and
[`docs/redesign.md`](../../../../docs/redesign.md); the API is
[`docs/api.md`](../../../../docs/api.md).

## The pages

| Page | Shell | Audience | Contents |
|---|---|---|---|
| Feed | `feed.html` (served at `/`) | Public read, member writes | People strip, every post newest first, peek/cover cards, comments |
| People | `people.html` | Public | `/_relay/people` is the directory; `/_relay/people/<id>` is a profile (posts + edit for the owner) |
| Post | `post.html` | Public | `/_relay/posts/<site>`: metadata, the microwebsite iframe, comments |
| Publish | `publish.html` | Member | Template picker, page name, drop zone, preview, publish, invite |
| Deploy guide | `deploy.html` | Member | Step-by-step terminal upload instructions |
| Host console | `admin.html` | Host | Admin sign-in, invitations, members, pages by member, comments, rollback |

Cross-page links are plain `<a>` tags — no router. One shared header
(`Shell.svelte`) renders on every member surface; `admin.html` keeps its own
header. Reads are public; writing (publish, comment, invite) is by invitation.

## Same-origin node API

The node serves its own UI and its own API over one origin, path-based:

- Relative URLs only (`/_relay/v1/feed`, `/_relay/v1/people`,
  `/_relay/v1/posts/<site>`, `/_relay/v1/publish`, `/_relay/v1/admin`). No
  CORS, no proxy.
- One `api/` client. Every response is decoded with a `Schema`; no `JSON.parse`
  outside `api/`.
- The API's JSON is the serialized Go domain. UI code uses the same vocabulary:
  node, member, invitation, page/post, version. Schemas mirror it.

## Credentials

- The **admin password** is held in `sessionStorage` for the host console and
  sent as a bearer token. Never put it in a URL, localStorage, or a log.
- A **member invitation token** rides in the invite link fragment (`#key=…`).
  Browsers do not send the fragment to the server. The feed exchanges it for
  the member cookie; the publish page can still use it as a bearer token.
- The **member cookie** is HttpOnly; the browser never reads it in JS. It
  resolves to the member on every request, and revocation is immediate.
- No client-side Ed25519 keys, signing, or key files. The credential is the
  token.

## State and data flow

The feed is public and loads `/me` opportunistically (401 = guest):

```
loading ──me+feed+people──► ready (me: Me | undefined)
   └──error──► error
```

- `src/feed/state.ts` defines the state type.
- `src/main/feed.ts` reads the invite fragment, exchanges it, loads `/me`,
  `/feed`, and `/people`, and folds results into a Svelte store.
- `src/ui/Feed.svelte` renders the store and calls callbacks; it never fetches.

The people page is one shell with two shapes: the directory (`members`) or one
profile (`profile`); `src/main/people.ts` decides from the path. The post page
loads `/me` (optional), the post, and its comments; both are public read.

The host console is action-driven:

```
locked ──connect──► loading ──ok──► ready
                         └──fail──► error (back to locked)
ready ──create/revoke/rotate/refresh──► ready (with notice or actionError)
```

- `src/admin/state.ts` defines the state type; `src/main/admin.ts` holds the
  effects; `src/ui/Admin.svelte` renders and calls callbacks.
- `mustChangeToken` gates the console until the default `relay` password is
  replaced. A one-time new token is shown once and dismissed explicitly.

## Copy and product rules

- Brand line: **a micro-social network for you and your real friends**.
- Member-facing words: member, invitation, page/post, host, node. Avoid "user
  account" and "deploy key"; prefer "page" over "site" in member copy.
- **Less is more** (see the relay-design skill §9): no explainer paragraphs, no
  empty `0 comments` counts, no policy footnotes on member surfaces. If a line
  has nothing to say, don't render it.
- The host console is the one place operational language ("admin password",
  "token", "rollback") is appropriate.
- No ideology in UI copy. A node is a place for a group, not a movement.

## Design

The visual system is the [`relay-design` skill](../../relay-design/SKILL.md)
and [`docs/redesign.md`](../../../../docs/redesign.md): Nunito + Figtree,
hairline borders, rounded corners, no noise overlay, no hard offset shadows,
no rotated tags on member surfaces. Colours are unchanged and used one accent
at a time (`--red`; `--ochre` rarely). Tokens live in `web/shared/styles.css`.

The card is a preview of the real thing: a `Peek` panel for text templates, a
cover image for media posts. Comments are a collapsed `<details>` thread,
rendered only when there is something to show. Native elements only for
interactive things: `button`, `a` with real hrefs, `label` + `input`,
`details/summary`, `table` for data grids. Keyboard: tab order follows visual
order; nothing requires a pointer.

## Testing

- `test/feed.test.ts` — SSR render of the feed (guest, member, empty, error).
- `test/people.test.ts` — SSR render of the directory and profile.
- `test/post.test.ts` — SSR render of the post wrapper with comments.
- `test/admin.test.ts` — SSR render of every host console state.
- `test/client.test.ts` — API client decode, errors, bearer/JSON shape.
- `test/publish.test.ts` — invite parsing, tar, Markdown page, copy contracts.
- `tools/smoke.mjs` — built admin and feed bundles in a DOM stub; asserts
  admin sign-in, feed signed-out, and feed signed-in states.

Add schemas and SSR tests in the same style when a page or wire shape changes.
