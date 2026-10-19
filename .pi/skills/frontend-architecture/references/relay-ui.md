# Relay UI architecture

Relay-specific rules for the owner UI: the pages, what they show, and how they
talk to the node. The product is a micro-social network for a group of friends;
the pivot plan is [`docs/pivot-plan.md`](../../../../docs/pivot-plan.md).

## The pages

| Page | Shell | Audience | Contents |
|---|---|---|---|
| Host console | `admin.html` | Host | Admin sign-in, invitations, members, pages by member, rollback |
| Publish | `publish.html` | Member | Invite greeting, page name, file/folder/Markdown drop, preview, publish, "invite someone" |
| Deploy guide | `deploy.html` | Member | Step-by-step terminal upload instructions |

There is no fourth app page without a stated reason. Cross-page links are plain
`<a>` tags — no router. The pages are short-lived documents, not a SPA.

Comments are built on the post wrapper. The embeddable comment panel is
deferred because static posts are CSP-sandboxed; do not add a post-script embed
that calls member APIs.

## Same-origin node API

The node serves its own UI and its own API over one origin, path-based:

- Relative URLs only (`/_relay/v1/admin`, `/_relay/v1/invites`,
  `/_relay/v1/sites/<site>/deploy`). No CORS, no proxy.
- One `api/` client. The host console uses the admin endpoints; the publish
  page uses the invite and deploy endpoints.
- The API's JSON is the serialized Go domain. UI code uses the same vocabulary:
  node, member, invitation, page/site, version. Schemas mirror it.

## Credentials

- The **admin password** is held in `sessionStorage` for the host console and
  sent as a bearer token. Never put it in a URL, localStorage, or a log.
- A **member invitation token** rides in the invite link fragment (`#key=…`).
  Browsers do not send the fragment to the server. The home page exchanges it
  for the member cookie; the publish page can still use it as a bearer token
  for the upload.
- The **member cookie** is HttpOnly; the browser never reads it in JS. It
  resolves to the member on every request, and revocation is immediate.
- No client-side Ed25519 keys, signing, or key files. The credential is the
  token.

## State and data flow

The member home is a small state machine:

```
loading ──me──► ready ──sign out──► signedOut
   └──invite──► ready
   └──401/error──► signedOut / error
```

- `src/home/state.ts` defines the state type.
- `src/main/home.ts` reads the invite fragment, exchanges it, loads `/me`, and
  folds the result into a Svelte store.
- `src/ui/Home.svelte` renders the store and calls callbacks; it never fetches.

The host console is action-driven:

```
locked ──connect──► loading ──ok──► ready
                         └──fail──► error (back to locked)
ready ──create/revoke/rotate/refresh──► ready (with notice or actionError)
```

- `src/admin/state.ts` defines the state type.
- `src/main/admin.ts` holds the effects and folds results into a Svelte store.
- `src/ui/Admin.svelte` renders the store and calls callbacks; it never fetches.
- `mustChangeToken` gates the console until the default `relay` password is
  replaced.
- A one-time new token is shown after create/rotate and dismissed explicitly.

The publish page is a form, not a state machine:

1. Read the invite fragment.
2. Collect files (drag/drop, file picker, folder picker, or Markdown).
3. Render a Markdown page client-side (bar, banner, styles) when needed.
4. Build a tar in the browser and POST it.
5. Render the live URL or the error code.

## Copy

- Brand: **a micro-social network for you and your real friends**.
- Member-facing words: member, invitation, page, post, host, node.
- Operational words (host console only): admin password, token, invitation
  key, version, rollback.
- The member cookie and welcome page are live. Do not claim the feed, profiles,
  comments, or templates are live; they are planned.
- No ideology. The product is a place, not a campaign.

## Design

- Tokens live in `web/shared/styles.css` and mirror the public site.
- Constructivist language: paper, ink, red, ochre; Anton display type; Oswald
  body; Space Mono for labels and code; hard borders, offset shadows, rotated
  stamps, wedges.
- The owner UI and the public pages should read as one product.
- Native elements only for interactive things: `button` (not `div onclick`),
  `a` with real hrefs, `label` + `input`, `table` for data grids.
- Status changes get an `aria-live` region where appropriate.
- Keyboard: tab order follows visual order; nothing requires a pointer.

## Testing

- `test/home.test.ts` — SSR render of the welcome, empty feed, and feed.
- `test/people.test.ts` and `test/post.test.ts` — SSR renders of the profile
  and post wrapper with comments.
- `test/admin.test.ts` — SSR render of every host console state.
- `test/client.test.ts` — API client decode, errors, bearer/JSON shape.
- `test/publish.test.ts` — invite parsing, tar, Markdown page, copy contracts.
- `tools/smoke.mjs` — built admin and home bundles in a DOM stub; asserts
  admin sign-in, home signed-out, and home signed-in states.

When the member pages land, add their schemas and SSR tests in the same style.
