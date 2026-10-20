# Relay — Redesign Spec (the quiet, friendly feed)

**Status:** adopted direction. This is the whole-product UI/UX spec for the
redesign. It supersedes the exploration in
[`ui-ux-pivot.md`](ui-ux-pivot.md), whose **V4** mockup
([`../ui-experiments/v4-soft-sans.html`](../ui-experiments/v4-soft-sans.html))
is the reference for the feed. Routes, data, and public-read decisions live in
[`navigation-pivot.md`](navigation-pivot.md); the code standards in the
[frontend-architecture skill](../.pi/skills/frontend-architecture/SKILL.md)
still govern implementation. Nothing here is built yet.

> **Reference code.** [`../ui-experiments/v4-soft-sans.html`](../ui-experiments/v4-soft-sans.html)
> is the source-of-truth implementation of §3–§6. Its inline `<style>` is the
> literal CSS to port into `web/shared/styles.css` — the token values, the
> Nunito/Figtree type scale, the shell, and the card/peek/comments anatomy all
> come from that file. The screenshot
> ([`../ui-experiments/screenshots/v4-soft-sans.png`](../ui-experiments/screenshots/v4-soft-sans.png))
> is for review only; when the two disagree, the code file wins.

---

## 1. The decision, in one paragraph

Relay keeps its colours and its "a place that belongs to the people in it"
stance, and softens everything else. The feed is the front door, one calm
column of posts that look like what they are; the chrome is a single slim
header; the type is warm and rounded (Nunito + Figtree) instead of brutalist;
and the copy says only what it needs to. If a line has nothing to say, it
isn't rendered.

## 2. Principles

1. **Content first, chrome second.** `/` is the feed. No pitch, no sign-up
   wall in front of it.
2. **The card is a preview of the real thing.** A Relay post *is* a
   microwebsite, so the card shows a peek of that page (or its cover when the
   cover is the better thumbnail). Never a metadata row.
3. **A group, not a feed.** The People strip, real names, and profiles make
   "your people" legible in one glance. No counts-as-status, no ranking.
4. **Calm.** Hairline borders, rounded corners, generous whitespace, one
   accent colour used sparingly. No noise texture, no hard offset shadows, no
   rotated tags on member surfaces.
5. **Less is more.** No explainer paragraphs, no empty "0 comments" states, no
   policy footnotes. Copy is the last resort, not the filler.
6. **Writing is by invitation; reading is public.** The comment form and
   publish affordances appear only for members; the content never hides.

## 3. Design tokens

Colours are unchanged from today. What changes is *how they are used* (one
accent at a time, hairline borders instead of hard ink).

| Token | Value | Use |
|---|---|---|
| `--red` | `#ce1f2b` | The single accent: active nav, links, primary action, the mark |
| `--red-deep` | `#a5141f` | Hover states |
| `--ink` | `#141210` | Text, hover borders, the dark bar on `/host` and the host console |
| `--paper` | `#f0e6d2` | Page background |
| `--paper-2` | `#e6d8bc` | Muted fills, placeholders |
| `--ochre` | `#dfa32a` | Rare warm accent (member avatars, one-off highlights) |

New tokens for the softened grammar:

| Token | Value | Use |
|---|---|---|
| `--hair` | `rgba(20,18,16,.16)` | All default borders and dividers |
| `--surface` | `#fbf6e9` | Card background (a step lighter than paper) |
| `--r-card` | `12px` | Card corner radius |
| `--r-inner` | `9px` | Peek/cover corner radius |
| `--r-avatar` | `11px` / `8px` | Large / small avatar radius |
| `--shadow-soft` | `0 1px 2px rgba(20,18,16,.08)` | The only elevation, if any |

What is kept from the old grammar: the red wedge as *accent*, the ink-on-paper
contrast, the small uppercase eyebrows (now in Figtree, not Space Mono), and
the `R` mark. What is dropped on member surfaces: the noise overlay, hard
3–4px borders, offset block shadows, rotated tags, and Anton slab headlines.
`/host` and the host console may keep *one* bold moment (a single wedge or a
dark bar), but in the same type and tokens.

## 4. Typography

Self-hosted in `web/fonts/fonts.css` (already extended). Two pairings exist;
**the product uses pairing B**.

| Role | Face | Weight | Size / notes |
|---|---|---|---|
| Display | Nunito | 700 (800 for the mark) | Page/card titles, all-caps, tight leading |
| Body | Figtree | 300 / 400 / 600 | Reading text, form fields |
| Eyebrow / meta | Figtree | 600 | Uppercase, `letter-spacing:.08em`, ~`.6rem` |

Type scale (feed):

| Element | Face/weight | Size |
|---|---|---|
| Card title | Nunito 700 | `1.15rem` |
| Page title (where one exists) | Nunito 700 | `1.7rem` |
| Body | Figtree 400 | `1rem` |
| Meta (date, chip) | Figtree 600 | `.6rem` uppercase |
| Comment body | Figtree 400 | `.94rem` |

The original Anton / Oswald / Space Mono files stay bundled for the public
landing and any un-migrated surface, but no new member-facing UI references
them.

## 5. Chrome — the shell

One header, rendered by a new `web/src/ui/Shell.svelte`, present on `/`,
`/_relay/people`, `/_relay/people/<id>`, `/_relay/posts/<site>`, and
`/_relay/publish.html`. It replaces the four hand-written `.bar` headers.

```
[ R · my-group ]          Home   Publish   People        ( alice · sign out )
```

- **Brand:** red rounded-square `R` mark + the node name (lowercase). Links to
  `/`. Guests see `Relay`; members see the node name from `/me`.
- **Primary nav:** `Home` → `/`, `Publish` → `/_relay/publish.html`, `People`
  → `/_relay/people`. The active item is underlined in red.
- **Identity slot (right):**
  - signed out → `Sign in` (opens a small inline prompt: paste the invitation
    key or link; same `readToken` logic as today, but inline, not a full page).
  - signed in → avatar + display name → profile, and `Sign out`.
- **Styling:** sticky, paper background at ~92% with `backdrop-filter: blur`,
  hairline bottom border, 58px tall, max-width 820px. No dark bar, no stamp.

`admin.html` keeps its own header (with the `Relay host` stamp), because the
host console is operational and deliberately out of the member nav. It uses the
same tokens and type.

## 6. Components

| Component | File | Notes |
|---|---|---|
| Shell | `web/src/ui/Shell.svelte` (new) | The shared header + inline sign-in |
| People strip | `web/src/ui/PeopleStrip.svelte` (new) | Row of avatar + name; `+ invite` for members |
| Feed list | `web/src/ui/FeedList.svelte` | Unchanged mechanics (scroll sentinel, pagination) |
| Post card | `web/src/ui/PostCard.svelte` | Reworked to peek/cover modes + collapsed comments |
| Peek | `web/src/ui/Peek.svelte` (new) | The little page-preview panel |
| Comments | `web/src/ui/Comments.svelte` (new) | Collapsed `<details>` thread + member form |
| Profile view | `web/src/ui/People.svelte` | Directory + profile grid; edit form moves here |
| Post page | `web/src/ui/PostPage.svelte` | Public-read; gates only the form |

Card anatomy (reference: V4):

```
┌───────────────────────────────────────────────┐
│ (A) alice · 2 days ago                [note]  │
│  ┌───────────────────────────────────────────┐│
│  │  SUNDAY WALK, THE LONG WAY                ││
│  │  We took the canal path and it was…       ││
│  │  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ (skeleton lines)     ││
│  └───────────────────────────────────────────┘│
│  ▸ 2 comments                                  │
│    bob — the long way is the whole point       │
│    dana — next time I'm coming                 │
│    [ Add a comment…            Reply ]         │
└───────────────────────────────────────────────┘
```

- **Header:** avatar (coloured rounded square, initial), display name →
  profile, relative date, template chip (pill, hairline).
- **Preview:** the whole preview is the link to the post.
- **Comments:** a collapsed `<details>` labelled `N comments` **only when
  `N > 0`**; the reply form is inside, for members. Guests see the thread, not
  the form. No empty states.

## 7. Pages

### 7.1 Feed — `/`

The V4 layout, exactly. Top to bottom:

1. Shell header.
2. People strip (avatars + names; `+ invite` for members only).
3. The feed, one column, newest first, infinite scroll.
4. A single footer link: `Run your own Relay →` → `/host`.

States:

| State | Renders |
|---|---|
| Loading | a quiet "…" or nothing but the shell |
| Empty (no posts) | one line: *"Nothing here yet."* with a single `Publish` action for members |
| Error | one short line, not a wall |
| Guest | same feed, no `Publish`/comment forms, `Sign in` in the header |
| Member | `Publish` nav, comment forms, `+ invite` in the strip |

### 7.2 People — `/_relay/people`

A simple grid of member cards (avatar, display name), linking to profiles.
Public read. Revoked members omitted. No bios here — the card is just who they
are.

### 7.3 Profile — `/_relay/people/<id>`

- Avatar, display name, bio (when set).
- The member's posts as the **same cards** as the feed, newest first.
- The owner sees an `Edit profile` action that reveals the display-name/bio
  form (moved here from `home.html`). Public read.

### 7.4 Post — `/_relay/posts/<site>`

- `← Back` to `/`.
- Title, author → profile, date, template chip.
- The full-size iframe of the real microwebsite (unchanged, CSP-sandboxed).
- Comments below, public read; the form is member-only. No full-screen
  "you need an invitation" gate — the post is readable by anyone.

### 7.5 Publish — `/_relay/publish.html`

Same shell header (Publish active). Trimmed to the action:

- Template picker (cards, calm styling).
- Page name field.
- Drop zone (hairline dashed, rounded; the current `.drop` restyled, not the
  heavy pubcard).
- Preview iframe.
- `Publish`.

Cut: the "What is Relay?" prose column moves to `/host` (or collapses behind a
`What is this?` details). The invitation and assistant blocks stay but their
copy is trimmed per §9.

### 7.6 Deploy guide — `/_relay/deploy.html`

Keep as a static step-by-step, restyled to the same type/tokens. Copy trimmed:
each step is a heading and the command, no reassurances.

### 7.7 Host console — `/_relay/admin.html`

Operational language is allowed here and only here. Same tokens/type; the
panels keep their structure but soften to hairlines and rounded corners. Keeps
its own header. Not in the member nav.

### 7.8 Landing — `/host` (+ `/landing`, `/index.html` redirect)

The hosting pitch, for the minority who run a node. Same fonts and tokens;
calmer than today: one headline, one sentence, the install command, three
steps, a link to `technicals.html` and the console. The red wedge appears
once, as an accent — not as a full-bleed hero. `technicals.html` is
regenerated with the same type and spacing.

## 8. Template preview modes

One field on the feed JSON drives the card: `templateName` (added per
[`navigation-pivot.md`](navigation-pivot.md) §7). Mapping:

| Template | Preview |
|---|---|
| `note`, `article`, `longform`, `link`, `reply`, `event`, custom | **Peek** — a small rendered excerpt of the page (static excerpt now; sandboxed iframe peek later) |
| `gallery`, `carousel`, `videos`, `songs`, `jetset` | **Cover** — the post's `imageUrl`, the whole cover is the link |
| `link` | Peek + a one-line link target (the domain) |

If a text template has an `imageUrl`, it may still use the cover; the rule is
"cover when the cover is the better thumbnail".

## 9. Voice and copy — less is more

Rules:

- A line must tell the user something they need, or do something they asked
  for. Otherwise it is deleted.
- One accent sentence per surface at most. No "here's what this is" paragraphs.
- Counts appear only when they are a real affordance (e.g. `2 comments` that
  expands). No `0 comments`, no `14 frames`, no follower numbers.
- Policy notes ("posts are public…") do not live in the feed; the one required
  escape hatch is the footer link to `/host`.

Before → after:

| Surface | Today | Redesign |
|---|---|---|
| Feed intro | "Your people · newest first / What the group made / Photos, links, notes and little websites from the people you know…" | *(nothing)* |
| Feed footer | "This is a Relay node for my-group · Run your own → / Posts are public · publishing and replying are by invitation" | `Run your own Relay →` |
| Empty comments | "0 comments" | *(omit the block)* |
| Comment placeholder | "Say something to your friend…" | `Add a comment…` |
| Publish sub | "Drop a page, some files or a Markdown note. It goes live the moment you press Publish. No account, nothing to install." | `Drop a page, some files or a Markdown note.` |
| Publish explainer | the whole "What is Relay?" column | moved to `/host` |

## 10. States, motion, accessibility

- **Guest / member / host:** one feed, three renderings. Guests read everything
  but see no write controls; members see write controls; hosts see the console
  only through `/_relay/admin.html` (never in the nav).
- **Motion:** hover = border darkens / link turns red; the People-strip avatar
  lifts 2px. No other animation. Respect `prefers-reduced-motion`.
- **Keyboard:** native `a`, `button`, `details/summary`, `input`, `label`
  everywhere. The inline sign-in is a real form; focus is visible.
- **Type safety:** every new wire field (`templateName`, the people list) is
  decoded with `Schema` at the edge, per the skill.

## 11. Implementation

Ordered by [`navigation-pivot.md`](navigation-pivot.md) §15 phases:

1. **Phase A — routes + public reads.** Serve the feed shell at `/`, `/host`
   for the landing, redirect legacy paths, widen read endpoints, add
   `templateName`. No visual change yet.
2. **Phase B — one chrome.** `Shell.svelte`, inline sign-in, footer link to
   `/host`, profile-edit moves to the profile page.
3. **Phase C — visual cards.** Rework `PostCard.svelte` to peek/cover modes;
   add `PeopleStrip`, the People directory, and profile grid; guest comment
   notes.
4. **Phase D — tokens and copy.** Adopt the §3 tokens and §4 type scale in
   `web/shared/styles.css`, swap the shells to `Shell.svelte`, restyle the
   publish/deploy/admin pages, and run the §9 copy pass across every surface.

Files touched: `web/shared/styles.css`, `web/*.html` shells,
`web/src/ui/*.svelte`, `web/src/member/schema.ts` (add `templateName` +
people list), `web/src/main/*.ts`, the root `styles.css` + `index.html` →
`host.html` (+ `technicals.html`), and `web/fonts/fonts.css` (already done).

The dependency rule (`main → feature → api → Schema`) and the functional
paradigm references are unchanged.

## 12. Acceptance criteria

1. `/` shows every post newest first, no sign-in, no intro paragraph.
2. A `note` card shows a peek of its page; a `gallery` card shows its cover;
   both open the post when clicked.
3. A guest reads a post and its comments but sees no publish button, no
   comment form, and no `+ invite`.
4. Opening an invite link lands on `/` signed in, with `Publish` and comment
   forms visible.
5. The same header appears on feed, profile, post, and publish pages, with the
   active item underlined.
6. A member edits their name and bio on their profile page, not the feed.
7. `/_relay/admin.html` is reachable by the host and absent from the member
   nav.
8. `/host`, `/landing`, and `/index.html` show the hosting pitch; the feed's
   footer links to it.
9. No surface renders an empty comment count, an explainer paragraph, or a
   policy footnote (§9).
10. `make check` passes; `docs/api.md`, `README.md`, and the frontend skill's
    `relay-ui.md` reference are updated to match.

## 13. Open questions

| # | Question | Stance |
|---|---|---|
| 1 | Peek now or later? | Ship the static excerpt first; the sandboxed iframe peek is a later, cheap swap. |
| 2 | Does `/host` keep a dark bar? | Optional. The member surfaces never use it; `/host` may keep one dark moment. |
| 3 | People strip on mobile? | Horizontal scroll, avatars shrink; no labels beyond the name. |
| 4 | Comments open or collapsed by default? | Collapsed; expand on click. |
| 5 | Do the old fonts stay bundled? | Yes, until every surface is migrated off them. |

---

*This is the source of truth for the redesign. The visual reference for the
feed is [`../ui-experiments/v4-soft-sans.html`](../ui-experiments/v4-soft-sans.html)
and its screenshot; the exploration that led here is
[`ui-ux-pivot.md`](ui-ux-pivot.md).*
