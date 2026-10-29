# Relay — IndieWeb Syndication (POSSE) Research

> **Note:** written against the old Go node, which has been removed. The implementation is now [`backend/`](../backend/README.md) (FastAPI) and [`mobile/`](../mobile/README.md) (Expo); endpoint, CLI and file names below describe the old node and are kept as design context.

**Status:** research proposal, not implemented. This document studies how Relay
could syndicate with the IndieWeb using POSSE and its neighbouring standards.
It changes nothing except itself. It is the companion to
[`navigation-pivot.md`](navigation-pivot.md), which positions Relay in the
IndieWeb niche; this document works out the *mechanics* of that niche.

---

## 1. TL;DR

Relay is unusually well positioned for POSSE because **a Relay post already has
a permanent, canonical URL** (`/<node>/<site>/`). POSSE only requires that you
publish to that URL first and then copy a *link back* to it elsewhere. Most of
the work is therefore not inventing new publishing, but exposing Relay posts in
the IndieWeb's standard, machine-readable shapes and then adding two small
endpoints.

The cheapest, highest-signal order is:

1. **Visible data** — a JSON Feed / RSS / `h-feed` of the group's posts, and
   `h-entry` + `h-card` markup on the post and profile pages. Near-zero cost,
   makes Relay readable by the whole indieweb ecosystem.
2. **Webmention receiving** — external replies, likes, and reposts land under a
   Relay post as comments (optionally bridged from Mastodon etc. via Bridgy).
3. **Micropub endpoint** — indieweb clients (Quill, IndieKit, Micro.blog apps)
   can publish *to* a Relay node, reusing the invitation token as the credential.
4. **POSSE out** — a member opts to syndicate a teaser + canonical link to
   external silos (Mastodon, Bluesky, …), and Relay records the copies with
   `u-syndication`.

Step 4 is the only part with real operational weight (per-silo credentials and
rate limits); steps 1–3 are small and additive.

## 2. Definitions

- **POSSE** — *Publish (on your) Own Site, Syndicate Elsewhere.* Post to your
  own canonical URL first, then publish copies (or link-only teasers) on
  third-party services, each linking back to the original. The original is the
  source of truth; the copies are disposable pointers.
- **PESOS** — *Publish Elsewhere, Syndicate to your Own Site.* The reverse:
  post to a silo first, then import into your own site. Relay's existing
  `relay index` backfill is a mild cousin of this.
- **Syndication copy** — one of those third-party copies. The original marks
  them with the microformats2 `u-syndication` property (`rel="syndication"`),
  so parsers can see "this post also lives at these URLs".
- **rel-canonical / link back** — the copy links to the original (ideally
  `rel="canonical"`), which is what lets search engines and POSSE-discovery
  tools know the copy is a copy.

## 3. The building blocks

| Standard | What it is | Why Relay might want it |
|---|---|---|
| **Micropub** (W3C Rec) | API for creating/updating/deleting posts on *your own* site; clients discover it and post MF2 (`h=entry`) form/JSON; a `q=syndicate-to` query lists syndication targets | Lets third-party indieweb clients publish **to** a Relay node without learning Relay's API |
| **Webmention** (W3C Rec) | A source URL notifies a target URL that it links to it; the target fetches the source, verifies the link, and renders the reply/like/repost | Lets external people reply to or like a Relay post from *their own* site, and lets Relay receive those as comments |
| **microformats2** (`h-entry`, `h-card`, `h-feed`, `u-syndication`) | HTML class names that mark posts, people, and feeds as visible, parseable data | Makes Relay post/profile pages machine-readable by indieweb parsers and Bridgy |
| **IndieAuth** (OAuth2-based) | Sign in with *your domain*; issues a token for Micropub | The one building block that does **not** fit Relay's group-first model (see §5) |
| **Feeds** (Atom, RSS, JSON Feed, `h-feed`) | Subscribe-and-poll syndication of a list of posts | The floor of "syndication": a reader or aggregator can follow a group's feed |
| **Bridgy** (brid.gy) | Hosted bridge: backfeeds silo likes/replies as webmentions, and can POSSE *out* to silos | Lets Relay receive Mastodon/Twitter/etc. interactions without speaking ActivityPub |

## 4. Where Relay already stands

Relay has most of the *hard* parts of POSSE already, and is missing mostly the
*standardized labels and endpoints*:

| POSSE requirement | Relay today |
|---|---|
| Canonical permalink per post | Yes — `/<node>/<site>/` (public, static, CSP-sandboxed) |
| A list of posts (the feed) | Yes — `GET /_relay/v1/feed` (JSON; the navigation pivot makes it public) |
| A post page | Yes — `/_relay/posts/<site>` (wrapper: metadata + iframe + comments) |
| An author page | Yes — `/_relay/people/<id>` |
| Comments on a post | Yes — node-local, member-written, sanitized by `social.RenderComment` |
| A publish credential | Yes — the invitation token as a bearer token (CLI + assistants) |
| Machine-readable feed (RSS/JSON Feed/`h-feed`) | No |
| `h-entry` / `h-card` markup | No |
| A Webmention endpoint | No |
| A Micropub endpoint | No |
| POSSE out to silos | No |

In other words: the *content model* is ready; the *interchange layer* is absent.

## 5. The core tension: group-first vs domain-first

The IndieWeb's floor is "own your domain and use it as your primary identity."
Relay is **group-first**: one node is one group, and the identity is an
invitation, not a domain. This tension resolves cleanly for everything except
IndieAuth:

- **Content and permalinks** — a Relay post *is* on a domain (the node's). `h-entry`,
  `h-feed`, feeds, and Webmention all work at the node level with no change to
  the identity model.
- **Authorship** — the indieweb `h-card` author is the member's profile URL
  (`/_relay/people/<id>`), not a personal domain. That is fine: `h-card` is
  just a URL with a name and photo.
- **IndieAuth is the one thing that does not fit.** IndieAuth signs in with a
  domain and scopes tokens to that domain's identity. Relay has no per-member
  domain, so the initial plan was to skip it and reuse the invitation token as
  the Micropub bearer. On reflection this is wrong: IndieAuth's identity is
  just a URL, and a Relay member has one (`/_relay/people/<id>`), with the
  invitation as the "authenticate the user" step. IndieAuth is therefore the
  enabling piece for standard-client Micropub, and it is worked out fully in
  [`indieauth.md`](indieauth.md). The invite-token bearer remains a valid
  shortcut for Relay's own tools only.

## 6. The two directions of syndication

POSSE is bidirectional in practice. Relay can do both, at very different costs.

### 6.1 Syndicate *out* (classic POSSE)

On publish, the node optionally posts a **teaser** to an external silo:

```
Relay post (canonical)  ──teaser + link──►  Mastodon / Bluesky / …
        ▲                                          │
        └──────── u-syndication link ─────────────┘
```

- The teaser is `title + summary + canonical URL (+ cover image)`, not the
  whole microwebsite — a gallery post syndicates its first image, a note
  syndicates its text.
- The node stores the returned silo URL on the post and renders it as
  "also on Mastodon" (`u-syndication`).
- Requires the **host** to configure each target's credentials (a Mastodon app
  token, a Bluesky app password, …). Members opt in per post; they never hold
  the silo keys.

### 6.2 Syndicate *in* (be a target)

Two independent ways in:

```
IndieWeb client ──Micropub──► Relay node ──► post
Other people's sites ──Webmention──► Relay node ──► comment under a post
Readers/aggregators ──feed──► Relay node ──► subscribed
```

- **Feeds** let a group's posts be followed from any reader. This is the
  cheapest "syndication" and the prerequisite for several POSSE discovery
  tools.
- **Webmention** lets someone reply to a Relay post *from their own site*;
  Relay fetches and renders that reply beside its local comments. Via Bridgy,
  the same mechanism carries Mastodon replies and likes without Relay speaking
  ActivityPub.
- **Micropub** lets a member use standard indieweb apps to post to the group.

## 7. Proposed plan (ordered by value ÷ cost)

### Phase 0 — Visible data (feeds + microformats2)

Zero new endpoints; markup and one serialiser.

- Add **`GET /_relay/v1/feed.json`** returning [JSON Feed 1.1](https://www.jsonfeed.org/)
  (`items` = posts, each with `id`, `url`, `title`, `summary`, `date_published`,
  `authors`, and `image` where present). Consider an Atom feed too if RSS
  readers matter to the group.
- Emit **`h-feed`** markup on the public feed page (`/`, per the navigation
  pivot), with each post as `h-entry` (`p-name`, `p-summary`, `dt-published`,
  `p-author` as `h-card`, `u-url`).
- Emit **`h-entry`** on the post wrapper (`/_relay/posts/<site>`) and
  **`h-card`** on the profile page (`/_relay/people/<id>`).

**Outcome:** a reader can subscribe to the group; Bridgy and indieweb parsers
can see Relay posts as posts. Cost is a template change and a serialiser. This
is the "visible data first" principle from
[`navigation-pivot.md`](navigation-pivot.md) §4.1 made concrete.

### Phase 1 — Webmention receiving

Add `POST /_relay/v1/webmention` (form-encoded `source` + `target`).

1. Validate `target` is a real post URL on this node.
2. Fetch `source` (http/https only, timeout, size cap, **no private/loopback
   addresses** — SSRF guard).
3. Verify the fetched HTML links to `target`.
4. Parse the source's `h-entry` (name, content, author `h-card`); fall back to
   the link's surrounding text.
5. Sanitize exactly like a comment (`social.RenderComment`), store as a
   **mention** (a comment with a `via` source URL), and render it under the post.

Bridgy can be pointed at the node's webmention endpoint to backfeed silo
interactions. Webmentions are public read (posts are public); only the
endpoint's spam controls are new.

**Outcome:** a Relay post can be replied to from the wider indieweb, and (via
Bridgy) from Mastodon, without Relay implementing ActivityPub.

### Phase 2 — Micropub endpoint

Add `POST /_relay/v1/micropub` (and `GET …?q=…` for `config`/`syndicate-to`).

- **Auth:** `Authorization: Bearer <token>` — an IndieAuth access token
  obtained by signing in with the member's profile URL (see
  [`indieauth.md`](indieauth.md)). The raw invitation token remains an accepted
  shortcut for Relay's own CLI/assistants.
- **Create:** accept MF2 `h=entry` in form-encoded, multipart, or JSON. Map the
  common shapes to Relay templates: `content` → Note/Article, `photo` → Gallery,
  `name` → title, `summary` → summary, `bookmark-of` → Link.
- **Respond:** `201` with `Location: /_relay/posts/<site>` and, when
  syndication exists later, `Link: <…>; rel="syndication"`.
- `q=syndicate-to` returns `[]` until Phase 3.

This reuses the existing publish pipeline (`/_relay/v1/publish`) and the
uploader's validation; it is a thin adapter, not a second publishing system.

**Outcome:** standard indieweb clients can post to the group; a member gets
"post from my phone" without Relay building a mobile app.

### Phase 3 — POSSE out (syndication targets)

The full classic POSSE, and the only phase with operational weight.

- The host configures syndication targets (a list of `{id, kind, endpoint,
  secret}` — Mastodon, Bluesky, and a generic webhook first).
- `GET /_relay/v1/micropub?q=syndicate-to` (and the publish page) lists them.
- On publish, a member may pick targets; the node POSTs a teaser + canonical
  link and stores the returned URL as `u-syndication` on the post.
- The post wrapper shows "also on …" links from `u-syndication`.

**Outcome:** a post exists on the node *and* reaches friends who live on silos,
with the node as the canonical source.

## 8. Security and moderation notes

- **Webmention SSRF.** Fetching `source` to verify the link is a server-side
  request. Restrict to http/https, resolve and reject private/loopback/link-local
  addresses, cap size and redirects, and time out. Verify the source actually
  links to the target before rendering anything.
- **Webmention spam.** A webmention is an unsolicited external render. Run it
  through the same sanitizer as comments, mark mentions distinctly ("from the
  web"), and let the host delete them with the existing comment-deletion path.
- **Micropub auth.** The invitation token is already a bearer credential with
  the same trust as the deploy endpoint; a Micropub client holding it can do
  exactly what the member can. No new surface is exposed beyond a second
  wire format.
- **POSSE out secrets.** Silo credentials live with the host, never with
  members, and are stored beside the admin token (a new `syndication.json`).
- **Revocation.** A revoked invitation stops Micropub and comment writes
  immediately, because both resolve the same key store.

## 9. What we explicitly do *not* do

- **Federation between Relay nodes.** POSSE is one-way copy/syndication, not a
  Relay-to-Relay protocol. Two nodes syncing is a different product and stays a
  non-goal (see [`pivot-plan.md`](pivot-plan.md)).
- **ActivityPub.** Bridgy is the bridge; Relay stays out of the fediverse
  protocol itself.
- **IndieAuth / domain identity.** Skipped for Micropub in favour of the
  invitation token (§5).
- **Per-member silo accounts in v1.** POSSE out uses host-configured targets;
  per-member "post to *my* Mastodon" is a later product if it earns its place.

## 10. Open questions

| # | Question | Stance |
|---|---|---|
| 1 | Should `feed.json` be public or members-only? | Public, matching the navigation pivot's "reading is public". |
| 2 | Does Webmention receiving need per-post opt-out? | Host-level "accept webmentions" toggle; per-post opt-out later. |
| 3 | Should mentions be a new store or a field on comments? | A `source` field on comments, plus a `type` (comment/mention); smallest change. |
| 4 | Which silo first for POSSE out? | Mastodon (open API, token per app) then Bluesky; both link-based teasers. |
| 5 | Does the teaser strip the microwebsite's richness? | Yes, by design — the teaser is a pointer, the canonical URL is the content. |
| 6 | Are there rate/abuse concerns for `feed.json`? | It is the same data as the feed page; no new data is exposed. |

## 11. Acceptance criteria (per phase)

- **Phase 0:** `/_relay/v1/feed.json` validates against JSON Feed 1.1; the feed,
  post, and profile pages carry `h-feed`/`h-entry`/`h-card` a parser can read.
- **Phase 1:** a webmention from an external page appears under the target post,
  sanitized; a forged/non-linking source is rejected; a loopback source is
  rejected.
- **Phase 2:** a Micropub `h=entry` create with an invite token produces a post
  and returns its `Location`; `q=syndicate-to` returns a valid (possibly empty)
  list.
- **Phase 3:** publishing with a target posts a teaser with the canonical link
  and the post shows the `u-syndication` link; a revoked member cannot POSSE out.

---

## Sources

- [POSSE — IndieWeb](https://indieweb.org/POSSE)
- [Micropub — IndieWeb](https://indieweb.org/Micropub) and
  [Micropub — W3C](https://www.w3.org/TR/micropub/)
- [Webmention — IndieWeb](https://indieweb.org/Webmention)
- [syndication / u-syndication — IndieWeb](https://indieweb.org/syndication)
- [IndieWeb principles](https://indieweb.org/principles)
- [Micro.blog — Wikipedia](https://en.wikipedia.org/wiki/Micro.blog)

*This document changes no code. It is the source note for a future
IndieWeb/POSSE integration; implementation follows it in the phase order above.*
