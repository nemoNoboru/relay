# Relay — Navigation & Feed-First Pivot

> **Note:** written against the old Go node, which has been removed. The implementation is now [`backend/`](../backend/README.md) (FastAPI) and [`mobile/`](../mobile/README.md) (Expo); endpoint, CLI and file names below describe the old node and are kept as design context.

**Status:** proposal, not implemented. This document is a design for the next
navigation/product pivot. It changes nothing on disk except itself. When it is
approved, fold its decisions into [`product.md`](product.md),
[`pivot-plan.md`](pivot-plan.md), [`api.md`](api.md), the
[`frontend-architecture` skill](../.pi/skills/frontend-architecture/SKILL.md)
and its `relay-ui.md` reference, then implement in that order.

The **visual side** of this pivot — keeping the theme but making the feed a
friendly, micro-social timeline — is designed separately in
[`ui-ux-pivot.md`](ui-ux-pivot.md), with three clickable mockups under
[`../ui-experiments/`](../ui-experiments/). The chosen direction (V4) is
specified for the whole product in [`redesign.md`](redesign.md).

---

## 1. The problem, in one paragraph

Relay is a micro-social network, but it does not look like one. The front door
(`/`) is a hosting pitch for the *minority* who run a node, while the thing the
*majority* does — read what their friends shared — is buried at
`/_relay/home.html` behind a member cookie. The feed renders posts as metadata
rows with an optional cover image, so it reads like a file manager, not a social
timeline. Every shell (`home.html`, `people.html`, `post.html`, `publish.html`,
`admin.html`) carries a slightly different header with different links, and the
first thing a new invitee sees is an invitation-paste form instead of their
friends' posts. The product story is right; the navigation is not.

## 2. What we are proposing

1. **The feed is the front door.** `/` serves every post on the node, newest
   first, as a visual timeline. This is the page everyone lands on.
2. **Hosting content moves out of the way.** The current landing page becomes
   `/host` (with `/landing` as an alias). It is for the minority who want to
   run a node; the feed links to it in the footer, not the other way around.
3. **The feed is visual and template-aware.** Each card shows a preview of the
   *real* generated template output — a live, sandboxed peek of the actual
   microwebsite, or its cover media where that is the better thumbnail. Cards
   are clickable through to the post and commentable in place.
4. **Reading is public; writing is by invitation.** Posts, the feed, profiles,
   and comments are readable by anyone with the address. Publishing and
   commenting remain invite-only. This is a deliberate reopening of a locked
   decision (see §6.3).
5. **One chrome, everywhere.** A single shared header across feed, profile,
   post, and publish pages: `Relay / node name · Home · Publish · People`, plus
   a compact sign-in/sign-out affordance. Host-only pages are hidden from this
   chrome.
6. **The `/_relay/` paths stay internal.** Users never type them. We do not
   rename the API or the control surfaces in this pivot; we add two clean
   top-level paths (`/` and `/host`) and make the rest feel like one product.

## 3. Goals and non-goals

### Goals

- A visitor, a member, and a host all have an obvious first page.
- The home page is recognizably a social network: content first, chrome second.
- A post looks like *what it is* in the feed, not like a row in an index.
- Navigation is consistent and self-explanatory on every page.
- The product remains true to its existing principles (chronological, no
  metrics, small group, posts are files).

### Non-goals (unchanged from the pivot)

- No likes, reposts, follower counts, or algorithmic feed.
- No global directory, discovery, or public profiles beyond the node.
- No account system; invitations remain the identity.
- No renaming of the `/_relay/` API surface.
- No federation, no native apps, no push.

## 4. Patterns we looked at

| Network | Pattern worth borrowing | What we reject |
|---|---|---|
| **Cohost** (2022–2025) | Chronological timeline, no algorithm, no trending; interaction counts *hidden* (no visible like counts, no follower numbers); posts accept Markdown **and full HTML/CSS** — a post is a page, not a status; comments "akin to blogs"; no numbers to perform for | Like/repost mechanics, global feed, the one-company model |
| **Are.na** (2012– ) | "A research tool that looks like a social network and behaves like a filing cabinet"; channels of blocks (image, link, text); no likes, no volume-driven comments, no algorithmic amplification; deliberately calm, "almost broken" interface | Public graph/discovery, anything resembling an engagement dashboard |
| **Pixelfed** | Instagram's media-first single column and profile grid, but **chronological with no content-manipulation algorithm**, no ads, no third-party analytics | Follower counts, algorithm, stories |
| **Instagram** | The media-first card anatomy and profile-as-grid that everyone already knows how to read | Likes, counts, ranking, reels, ads |
| **Pinterest** | Visual card = the whole unit; tap-through to detail; mixed card shapes in one flow | Recommendation engine, public search |
| **Micro.blog** | Plain chronological timeline; conversation under a post; indie-hosted | Global timeline, federation UI |

Three of these are the closest kin to Relay and worth quoting:

- **Cohost** — "a social network with a chronological timeline, no algorithms,
  no statistics of any kind (no engagement, number of followers, etc)", where
  "posts could be long or short-form… accepted CSS and HTML besides Markdown",
  and reposts "showed *below* the original post". The posts-are-pages idea is
  exactly Relay's microwebsite model.
- **Are.na** — "no likes, no algorithms, just blocks, channels, and people
  following their curiosity in public"; "the interface looks, to a first-time
  visitor, almost broken — like a website that forgot to add the engagement
  features. That's the point."
- **The quiet feed** — "no follower counts, no engagement dashboards, no
  algorithm deciding who sees what. Just people sharing work they actually
  care about with readers who actually showed up."

Distilled into five rules we adopt:

1. **The front door is content, not a pitch or a sign-up wall.** (Cohost,
   Are.na, Instagram.)
2. **A persistent, minimal chrome; everything else is one tap away.**
   (Instagram, Pixelfed.)
3. **The card is a preview of the real thing, and the whole card opens it.**
   (Pinterest, Are.na.) This is Relay's differentiator: the "real thing" is a
   microwebsite, so the card can be a live peek of it, not a flattened
   thumbnail.
4. **Comments live close to the content but never crowd the card.** (Cohost's
   blog-like comments, Instagram, Micro.blog.)
5. **No counts, no ranking, newest first.** Already Relay's stance; it is the
   through-line of Cohost, Are.na, Pixelfed, and BeReal.

### 4.1 The IndieWeb is our niche

Relay is not competing with Instagram. It sits in the **IndieWeb / "small
web"** neighbourhood, and the IndieWeb's principles read, almost line for
line, like the product Relay already describes in
[`product.md`](product.md).

The IndieWeb is "a community of independent & personal websites connected by
simple standards, based on the principles of: owning your domain & using it as
your primary identity, publishing on your own site (optionally syndicating
elsewhere), and owning your data." Its key principles are: **own your identity
and data**, **use and publish visible data** (humans first, machines second),
**make what you need**, **use what you make** ("eat what you cook"),
**document**, **open source**, **UX before plumbing**, **modularity** ("small
pieces loosely joined"), **longevity** ("the long web"), **pluralism**, and
**have fun**.

| IndieWeb principle | Where Relay already is | Where Relay differs |
|---|---|---|
| Own your content / data | Posts are static files; backup is "copy the folder"; no database | Content lives on a *friend's* node, not on each person's own domain |
| Own your identity | An invitation is the identity; no corporate account | Identity is scoped to the node, not a personal domain |
| Longevity (the long web) | Static files, portable by design, readable without Relay | — |
| Make what you need / eat what you cook | One node, one group, run by someone you know | — |
| UX before plumbing | The entire point of this navigation pivot | — |
| Modularity (small pieces loosely joined) | Templates, the store interface, one binary | — |
| Pluralism / no politics | "a club, a family, a reading group, a band" | — |

Concrete overlaps worth naming:

- **Micro.blog** ships the same product stance Relay already holds, almost
  verbatim: it "does not show follower counts, does not have hashtags, public
  likes or trending topics, does not have equivalents of retweeting or quote
  tweeting, does not algorithmically recommend users"; "Micro.blog doesn't
  make it particularly easy to discover new users, and posts don't spread
  virally… **the snowball starts small and stays small**… by design." That is
  Relay's "small is the feature" and "no engagement mechanics" in someone
  else's words.
- **Cohost** made posts that were *pages* (Markdown plus full HTML/CSS), with
  comments "akin to blogs" — Relay's microwebsite-and-template model.
- **The IndieWeb reply model** (write a reply on your own site, link to the
  original, send a webmention) is Relay's existing rule — "to share, write
  your own post linking to it" — with the reply kept *on the node* because that
  is what a group needs.
- **POSSE** — "Publish (on your) Own Site, Syndicate Elsewhere" — is the
  philosophy behind "posts are files, not platform."

What we deliberately do **not** borrow from the IndieWeb in this pivot:

- **Domain-first identity.** The IndieWeb's minimum bar is your own domain as
  your primary identity. Relay is *group-first*: one node, one group, invitation
  = identity, no per-person domains. That is the product, not a gap.
- **The connection standards** (Webmention, Micropub, IndieAuth,
  microformats2). Relay connects people with a shared node and a cookie, not
  with cross-site pings. Federation and cross-node identity are already listed
  as non-goals in [`pivot-plan.md`](pivot-plan.md).
- **Protocol-first culture.** The IndieWeb is, at its core, a standards
  community; Relay is a product for non-technical friends. We take the
  principles, not the plumbing.

One cheap, high-signal idea we do take from the IndieWeb: **visible data
first**. The feed is already real JSON (`feed` → posts); keep it shaped like
the thing it is (a list of posts with authors and timestamps — the IndieWeb's
`h-feed`/`h-card` shape), so a future microformats2 or RSS export is a mapping
exercise, not a redesign. This costs nothing now and keeps the door open.

The *mechanics* of that future — POSSE, Micropub, Webmention, and feeds — are
worked out separately in
[`indieweb-syndication.md`](indieweb-syndication.md).

Sources: [Cohost — Wikipedia](https://en.wikipedia.org/wiki/Cohost),
["A late eulogy to cohost"](https://www.whateverthewindbrings.com/a-late-eulogy-to-cohost-i-guess/),
[IndieWeb — cohost.org](https://indieweb.org/cohost.org),
["The Quiet Feed" — Scroll](https://scroll.name/quiet-feed-intentional-online-communities-no-algorithms/),
[Pixelfed — Wikipedia](https://en.wikipedia.org/wiki/Pixelfed),
[openalternative.co/pixelfed](https://openalternative.co/pixelfed),
[IndieWeb principles](https://indieweb.org/principles),
[On POSSE and IndieWeb — Joel Auterson](https://www.joelotter.com/posts/2023/03/indieweb/),
[Micro.blog — Wikipedia](https://en.wikipedia.org/wiki/Micro.blog).

## 5. The proposed shape

### 5.1 What a visitor sees

`https://node.example/` is the group's feed. A visitor scrolls posts newest
first. Each card shows who posted it, when, and a preview of the actual page.
Clicking a card opens the post (the real microwebsite plus comments). A slim
header says the node's name and offers `Home`, `Publish`, `People`. A footer
line says *"This is a Relay node for {name}. Run your own →"* linking to
`/host`. There is no sign-up wall; reading is free.

### 5.2 What a member sees

The same feed, plus:

- a **Publish** button in the header,
- a **comment box** under each post,
- their **profile** (display name, bio, their posts as a grid) at `/people/<id>`,
- a **sign-out** in the header menu.

Opening an invite link still exchanges `#key=…` for the cookie and lands on `/`.
A signed-out member pastes their invite in the header's sign-in, not in a
full-screen gate.

### 5.3 What a host sees

The same feed as everyone else, plus a discreet way to reach the host console.
The console stays at `/_relay/admin.html`, but it is no longer in the public
nav. `/host` is where the operational story lives (what a node is, how to run
one, link to `technicals.html`, link to the console).

## 6. Decisions

### 6.1 `/` is the feed

The embedded `index.html` (today's landing) is repurposed. The mux serves the
feed shell at `/`, not the landing page.

- New shell: `web/feed.html` (a renamed/trimmed `home.html`), built by Vite and
  embedded like the other owner-UI pages.
- `/_relay/` and `/_relay/home.html` redirect to `/`.
- `/_relay/home.html` remains valid forever (invite links and old bookmarks).

### 6.2 Hosting content moves to `/host`

- Rename the root `index.html` to `host.html`, keep it in `site.go`'s embed
  list, and register it in `landingFiles` under `/host`.
- `/landing` and `/index.html` redirect to `/host` (302), so old links land on
  the pitch, not the feed.
- `/technicals.html` and `/styles.css` keep their paths. `technicals.html` is
  linked from `/host`.

The `landingFiles` map currently keys exact single-segment dotted paths and is
the right place for `/host`; `host.html` is a single-segment path that cannot
collide with `/<node>/<site>/…` (which needs two segments) or with the reserved
`_relay` prefix.

### 6.3 Public reading, invite-only writing (reopens pivot decision 3.8)

Today posts are "unlisted, not secret" at `/<node>/<site>/`, but the feed,
profiles, and comments are members-only. That mismatch is what makes Relay feel
like a hosting service. We flip the default:

- `GET /_relay/v1/feed`, `GET /_relay/v1/posts/<site>`,
  `GET /_relay/v1/people/<id>/posts`, and
  `GET /_relay/v1/posts/<site>/comments` become **public**.
- `POST /_relay/v1/posts/<site>/comments`, `POST /_relay/v1/publish`, the
  deploy endpoint, and invites remain **member-only**.
- The `mine` field in a comment is always `false` for guests.

*Written reason for reopening 3.8:* the content is already public at its own
URL; hiding the index and comments while serving the content is security
theatre, not privacy. The group's *privacy* is the invitation list and the
right to publish — those stay members-only. This is also the single change that
makes `/` safe to serve to a stranger.

Escape hatch: a host flag `--members-only` (or `private: true` in a node
config) restores the old members-only feed and comments for groups that want
it. The default is public, matching the product line "posts are public,
publishing is by invitation."

### 6.4 One chrome, everywhere

One shared header component replaces the four hand-written `.bar` headers:

```
[ R · relay-name ]   Home   Publish   People        [ sign in / avatar ]
```

- **Home** → `/`.
- **Publish** → `/_relay/publish.html` (kept path, member).
- **People** → `/_relay/people` (new small directory page; §6.7).
- The brand links to `/` on every page (today some shells link to `../`).
- Sign-in is a header action, not a full page. Guests see **Sign in**; members
  see their display name and **Sign out**.

The header is rendered by the Svelte shells, not duplicated HTML. `admin.html`
keeps its own header and is not part of this chrome.

### 6.5 Template-aware feed cards

The feed card is no longer `title + cover image`. It is a **template preview**:

```
┌──────────────────────────────────────────────┐
│  (avatar) Name · 2 days ago        [template]│
│  ┌────────────────────────────────────────┐  │
│  │        live peek of the real page      │  │   ← sandboxed iframe, or
│  │        (pointer-events: none)          │  │     cover media for media
│  │                                        │  │     templates
│  └────────────────────────────────────────┘  │
│  Title — summary…                     [Open] │
│  ▸ 3 comments · Add a comment                │
└──────────────────────────────────────────────┘
```

Two preview modes, both honest to "a post is a microwebsite":

- **Peek (default for text/page templates — `note`, `link`, `article`,
  `longform`, custom):** a fixed-aspect, `pointer-events: none`, sandboxed
  iframe of the post's real URL (`/<node>/<site>/`), showing the top of the
  rendered page. The iframe gets the same `Content-Security-Policy: sandbox`
  header the site already serves (opaque origin), so a post's scripts cannot
  act as the viewer. Lazy-load peeks with the same `IntersectionObserver`
  machinery already used for pagination, and cap concurrent live iframes (≈4)
  for a small node.
- **Cover (default for media templates — `gallery`, `carousel`, `videos`,
  `songs`, `jetset`):** the post's existing `image`/`imageUrl` cover, because a
  photo grid's cover is the better thumbnail than a page peek.

The whole preview area is a link to `/_relay/posts/<site>`. To keep the client
from needing a template lookup, the feed post JSON gains one field:
`templateName` (the human name, e.g. `"Gallery"`), derived from the template
store. Card style keys off `imageUrl` presence plus the `template` id; a later
`previewMode` field can make it explicit per template (see §12).

### 6.6 Comments stay under the post, inline

Comments continue to render inline on each card (as today) and on the post
detail page. The only change is visibility: anyone can read them; only members
see the form. Guests see *"Members comment here — {n} comments."* instead of a
textarea. Own-comment deletion and host deletion are unchanged.

### 6.7 A small People page

To make `People` in the nav real, add a minimal directory:

- `GET /_relay/v1/people` → `{members: [{id, displayName, bio}]}`, public,
  revoked members omitted.
- `/_relay/people` → a simple list/grid of member cards linking to
  `/_relay/people/<id>`.

This is the only genuinely *new* surface in the pivot. It is small and uses the
member data the node already holds. If it does not earn its place, drop it and
make `People` a link to the member's own profile only.

### 6.8 Profiles become grids

`/_relay/people/<id>` keeps the display name and bio, then lists the member's
posts as the same visual cards as the feed (not a metadata list). A member
edits their display name/bio here (the form that today lives on `home.html`
moves to the profile page). Profiles become public (read), matching §6.3.

## 7. Proposed routes

### Top-level (public)

| Path | Today | Proposed |
|---|---|---|
| `/` | landing pitch (`index.html`) | **the feed** (`feed.html` shell) |
| `/index.html` | 301 → `/` | 302 → `/host` |
| `/host` | 404 | **landing/host pitch** (`host.html`) |
| `/landing` | 404 | 302 → `/host` |
| `/technicals.html` | technicals | unchanged, linked from `/host` |
| `/styles.css` | stylesheet | unchanged |

### Control and social (kept under `/_relay/`, hidden from nav)

| Path | Today | Proposed |
|---|---|---|
| `/_relay/` | 302 → `/_relay/home.html` | 302 → `/` |
| `/_relay/home.html` | member home shell | 302 → `/` (legacy, keep forever) |
| `/_relay/people` | 404 | **members directory** (public) |
| `/_relay/people/<id>` | profile (members-only) | profile grid (**public read**) |
| `/_relay/posts/<site>` | post wrapper (members-only) | post detail (**public read**, member comment form) |
| `/_relay/publish.html` | publish (member) | unchanged |
| `/_relay/deploy.html` | deploy guide (member) | unchanged, not in main nav |
| `/_relay/admin.html` | host console | unchanged, hidden from nav |

### API deltas

| Endpoint | Today | Proposed |
|---|---|---|
| `GET /_relay/v1/feed` | member | **public** |
| `GET /_relay/v1/posts/<site>` | member | **public** |
| `GET /_relay/v1/people/<id>/posts` | member | **public** |
| `GET /_relay/v1/people` | — | **new**, public |
| `GET /_relay/v1/posts/<site>/comments` | member | **public** |
| `GET /_relay/v1/templates` | member | **public** (list manifests only; install stays member) |
| `POST /_relay/v1/posts/<site>/comments` | member | member (unchanged) |
| `DELETE /_relay/v1/comments/<id>` | own/host | own/host (unchanged) |
| `POST /_relay/v1/publish`, deploy, invites, session, `/me` | member | member (unchanged) |

The feed/post JSON gains `templateName`. Comments already carry `mine`; guests
get `mine: false`.

## 8. Navigation chrome spec

Shared header, present on `/` (feed), `/_relay/people…`, `/_relay/posts/<site>`,
and `/_relay/publish.html`:

- **Brand:** `R` mark + node name (from `/me` when signed in; otherwise a
  static "Relay"). Links to `/`.
- **Primary nav:** `Home` (→ `/`), `Publish` (→ `/_relay/publish.html`),
  `People` (→ `/_relay/people`). Active page gets the underline.
- **Identity slot, right:**
  - signed out → `Sign in` (opens a small inline prompt: paste invitation
    key/link; same `readToken` logic as today, but inline).
  - signed in → `{displayName}` → profile; `Sign out` beside it.
- **Mobile:** nav collapses to the horizontal-scroll row already styled in
  `shared/styles.css` (`.bar .wrap` media rule); identity slot stays visible.

Header state is derived from the same `GET /_relay/v1/me` call the feed already
makes; guests simply get the signed-out slot.

## 9. Feed card spec

Reuses `PostCard.svelte` and `FeedList.svelte`; changes the body from
"cover-or-nothing" to the two preview modes, and gates the comment form on
membership.

- **Header:** avatar (initial), display name → `authorUrl`, relative date,
  template badge (`templateName`).
- **Preview:** `imageUrl` present → cover `<img>` (current behavior, kept);
  otherwise → sandboxed peek iframe, fixed aspect ratio (≈16:10),
  `pointer-events: none`, `loading="lazy"`, `sandbox="allow-scripts
  allow-popups allow-forms allow-downloads"`, `aria-hidden`, wrapped in a link.
  The whole card is clickable via the title and an explicit `Open` action.
- **Body:** title, summary, `Open page` link to `post.url`.
- **Comments:** unchanged inline list + form, but the form is hidden for
  guests and replaced with a "Members comment here" note.

Performance guardrails: only render peek iframes when near the viewport
(`IntersectionObserver`, already used), cap live peeks, and pause them when the
tab is hidden. At 29 members this is a non-issue; the cap is defensive.

## 10. Post detail and profile

- **Post detail** (`/_relay/posts/<site>`): keep the full-size iframe of the
  real site plus comments. Public read; the comment form is member-only. Drop
  the "You need an invitation" full-screen gate; instead show the post and a
  members-only note on the form. Add a `← Back` link to `/`.
- **Profile** (`/_relay/people/<id>`): display name, bio, then the member's
  posts as feed cards (visual, not metadata rows). The owner sees an **Edit
  profile** affordance that reveals the display-name/bio form moved here from
  `home.html`.

## 11. Server changes

1. `site.go`: embed `host.html` alongside `index.html`/`technicals.html`/
   `styles.css` (rename `index.html` → `host.html`).
2. `internal/serve/landing.go`: change `landingFiles` to serve `host.html` at
   `/host` and `/landing`; drop the `/` and `/index.html` entries (they become
   redirects or the feed).
3. `cmd/relay/node.go`: serve the built `feed.html` at `/` (the feed shell),
   add the `/host` landing to the mux or keep it inside `Landing`, and change
   `/_relay/` to redirect to `/`.
4. `internal/serve/posts.go`: drop the session requirement for the read
   endpoints (`feed`, `post`, `people/<id>/posts`); keep writes member-only.
   Add `templateName` to the post JSON.
5. `internal/serve/comments.go`: make `GET …/comments` public; keep `POST` and
   `DELETE` gated.
6. `internal/serve/templates.go`: make `GET /_relay/v1/templates` public (list
   only); keep `POST` (install) and file serving member-gated.
7. New `GET /_relay/v1/people` (public members list) in the post index or
   session surface, plus a `/_relay/people` shell route in `ui.go`.
8. `internal/serve/ui.go`: `/_relay/` → `/`; `/_relay/home.html` → `/`.

The `Content-Security-Policy: sandbox` header on static sites already gives the
peek iframes an opaque origin; no new security machinery is needed for the
previews.

## 12. Frontend changes

- `web/feed.html` (new shell, from `home.html`) served at `/`; `home.html`
  becomes a redirect stub or is deleted from the build.
- `web/shared/styles.css`: add the shared-header component styles, the
  `.post-peek` iframe box, the profile grid, and the inline sign-in prompt.
- `web/src/ui/Shell.svelte` (new): the shared header + identity slot; used by
  `Home/Feed`, `People`, `PostPage`, and the publish page.
- `web/src/ui/Home.svelte` → `Feed.svelte`: drop the signed-out full-page
  gate and the profile-edit card (moves to profile); render feed cards with the
  two preview modes; render a small "Run your own Relay → /host" footer.
- `web/src/ui/PostCard.svelte`: add the peek/cover modes and the guest comment
  note.
- `web/src/ui/People.svelte`: add the directory view (list of members).
- `web/src/ui/PostPage.svelte`: public-read state; gate only the form.
- `web/src/member/schema.ts`: add `templateName`; add the members-list schema.
- `web/src/main/feed.ts` (new): `/me` (optional) + public feed load; the same
  state machine minus the invite exchange being a full page.

The dependency rule (`main → feature → api → Schema`) and the functional
paradigm references are unchanged.

## 13. Migration and compatibility

- `/`, `/index.html` stop being the landing; `/index.html` and `/landing`
  redirect to `/host` so external links are not broken.
- `/_relay/home.html` redirects to `/` forever; every invite link ever sent
  keeps working (they land on the feed after exchanging the fragment).
- `/_relay/publish.html#key=…` links keep working unchanged.
- Static sites at `/<node>/<site>/…` are untouched; the CSP sandbox is
  untouched.
- No token or data migration. The change is routing, one JSON field, and auth
  widening on read endpoints.
- A node that wants the old members-only behavior sets `--members-only`.

## 14. Risks and open questions

| # | Risk / question | Stance |
|---|---|---|
| 1 | **Public feed = public directory for the node.** Anyone with the address sees everything. | That is the point. The address is the privacy boundary. `--members-only` is the escape hatch. |
| 2 | **Peek iframes are heavier than thumbnails.** | 29 members / dozens of posts; lazy-load + cap. If it still feels heavy, fall back to cover-only and ship a `previewMode` manifest field. |
| 3 | **`GET /_relay/v1/templates` public leaks installed template names.** | Names/descriptions are not sensitive; the list already renders on the publish page to signed-in members. Install stays gated. |
| 4 | **`/host` vs `/landing` naming.** | `/host` primary (matches vocabulary); `/landing` alias. Trivial to flip. |
| 5 | **Should the People page ship in this pivot?** | Nice-to-have; drop it if it adds friction. The nav item can link to the viewer's own profile instead. |
| 6 | **Do guests see a comment form greyed out or nothing?** | A one-line note ("Members comment here") is clearer than a disabled form. |
| 7 | **Feed shell at `/` vs the resolver.** | `/` is the mux root, already intercepted before the site resolver; `feed.html` is served there, and `/<node>/<site>/` is two segments, so no collision. |

## 15. Phased rollout

1. **Phase A — routes and public reads.** Add `/host` + aliases, serve the feed
   at `/`, redirect legacy paths, widen the read endpoints, add `templateName`.
   No visual changes yet; the feed at `/` is the current feed, public.
2. **Phase B — one chrome.** `Shell.svelte`, shared header, inline sign-in,
   footer link to `/host`, profile-edit moves to the profile page.
3. **Phase C — visual cards.** Peek/cover preview modes, People directory,
   profile grid, guest comment notes.
4. **Phase D — docs and polish.** Update `README.md`, `product.md`,
   `pivot-plan.md` (mark 3.8 reopened), `api.md`, the frontend skill and its
   `relay-ui.md` reference, and `users-guide.md`.

Each phase leaves the node working and `make check` green.

## 16. Acceptance criteria

1. Visiting `/` shows every post, newest first, no sign-in required.
2. The feed card for a `note` shows the note's real page as a peek; the card
   for a `gallery` shows its cover; both open the post when clicked.
3. A guest can open a post, read it, and read its comments, but sees no
   publish button, no comment form, and no way to write.
4. Opening an invite link lands on `/` signed in, with `Publish` and comment
   forms visible.
5. `/host` (and `/landing`, `/index.html`) show the hosting pitch; the feed's
   footer links to it.
6. The same header appears on feed, profile, post, and publish pages, and the
   active item is indicated.
7. A member edits their name and bio on their profile page, not on the feed.
8. `/_relay/admin.html` is reachable by the host and is not in the public nav.
9. Old links (`/_relay/home.html`, `/_relay/publish.html#key=…`, static site
   URLs) keep working.
10. `make check` passes, and every doc that described the old navigation is
    updated.

---

*This document changes no code. It is the source note for the navigation and
feed-first pivot; implementation follows it in the order above.*
