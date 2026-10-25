# Relay mobile mockups (RELY-78)

Static HTML/CSS mock UIs of every screen of the Expo / React Native app, for
product-owner review. Design only: no app code. Tokens and components live in
`shared.css` (plus `auth.css`, `communities.css`, `events.css`, `admin.css`
for area-specific bits); each screen is `screens/<NN>-<name>.html` with its
render in `png/<NN>-<name>.png`.

## How to render

```sh
./render.sh                      # every screen -> png/
./render.sh screens/01-feed.html # only some screens
```

Needs `chromium` and `node` on PATH. Output is 780×1688 (a 390×844 phone at 2x).
Emoji use Noto Color Emoji, which `render.sh` downloads into `fonts/`
(git-ignored) on first run.

## Design language

Constructivist / Bauhaus poster, after the Relay landing page. Every screen
gets character (big type, a bold header, colour blocking, one geometric
accent), but post text, forms and lists stay plain and readable. Screens
01–05 are the reference set; copy their patterns.

- **Colour.** Cream paper `--paper` #f0e6d2, ink `--ink` #141210, signal red
  `--red` #ce1f2b, ochre `--ochre` #dfa32a. Colour-block, don't tint: no
  translucent red washes. Contrast rules (WCAG AA):
  - On red, text is `--on-red` #fff8ea (5.2:1), never plain cream (4.4:1).
  - Small red text on cream uses `--red-text` #b81a25. Plain `--red` is for
    blocks, rules and display type 24px and up.
  - On black, accents are ochre or cream. Red stays a block or stripe there,
    never text.
  - Muted text uses `--ink-2`, not opacity.
- **Shape.** Square corners everywhere: no `border-radius` except circles in
  decoration. Borders are 2px ink (`--line`), 3px for emphasis (buttons,
  popovers, focus). Shadows are hard offsets only: `--shadow` (5px 5px 0 ink),
  `--shadow-sm` (3px); use a red or ochre offset for a highlighted element.
- **Type.** `--display` Anton is uppercase, heavy, with tight leading (~.86),
  for headlines, titles and avatar letters; colour the key word with
  `<span class="hot">`. `--body` Oswald is for post text, rows and paragraphs
  (17–20px). `--mono` Space Mono, uppercase and tracked (.08–.2em), is for
  labels, buttons, nav, tabs, metadata and counters. `.label` and
  `.section-label` get a short red rule in front automatically.
- **Headers.** Every screen opens with a black band and a 6px red stripe:
  `.navbar` (compact; `.navbar.large` for a big Anton title plus an ochre
  `.navbar-sub`), or `.poster-head` for hero screens (feed). Put
  `.statusbar.dark` above a black band so it runs full-bleed.
- **Decoration** goes in headers, hero areas, empty states, sheet heads and
  active states, never over content. Use one or two shapes per screen and
  let them bleed off the edge.
- **Tap targets** stay at least 44px: buttons are 50px (`.sm` 44px); seg and
  tab items are 44px or more.

### Poster classes (see the shared.css header for the full list)

| Class | Use |
|---|---|
| `.logo` (`.sm`, `.lg`) | The mark: red square rotated -6°, heavy Anton "R", ochre corner chip, hard shadow. `<span class="logo">R</span>`. `.mark` is an alias kept for the auth screens. |
| `.wordmark` | "RELAY" in Anton, next to the logo |
| `.poster-head` + `.ph-top` `.ph-title` `.ph-meta` | Black hero header: logo row, giant title, ochre mono meta line |
| `.display`, `.display .hot`, `.display .ochre` | Huge Anton headline; red or ochre key word |
| `.kicker` | Tracked mono line after a 30px red rule (above headlines) |
| `.stamp` (`.red`, `.ochre`, `.ink`) | Tilted rectangular label for counts and states ("6 replies", "Location off") |
| `.vtag` | Vertical "RELAY" tag with red bars |
| `.geo-sun` `.geo-ring` `.geo-wedge` (`.flip`) `.geo-bar` `.geo-square` `.geo-quarter` | Absolutely positioned shapes; set size and position inline. The parent needs `position:relative; overflow:hidden`, and content siblings need `position:relative` to sit above them. |
| `.grid-paper` (`.dark`) | Graph-paper background for hero blocks and frames |
| `.stripe` | 6px red rule |
| `.mono` | Inline Space Mono metadata |

Restyled components kept their class names:

- `.card`: 2px border, hard shadow, ruled `.card-foot`.
- `.react.mine`: solid red.
- `.btn.primary`: red, 3px ink border, hard shadow.
- `.seg`: ink fill for `.on`.
- `.tab.active`: red bar on top, red icon.
- `.banner`: tilted ink tag with a red arrow block.
- `.fab`: red square, hard shadow.
- `.sheet`: ink edge plus a red band on top, Anton `.sheet-title`.
- `.toggle`: square.
- `.empty`: Anton headline, ochre disc with a red square.

## Screens

Signed-in screens share the bottom tab bar: Feed · Communities · Events · Settings.

| # | Screen | Sub-issue | Shows |
|---|---|---|---|
| 01 | feed | RELY-96 | Locality feed, new-posts banner, post cards, "+" button |
| 02 | feed-location-off | RELY-96 | Location permission denied, prompt to enable |
| 03 | compose | RELY-96 | Text input with 500-char counter, location chip, community picker, Post/Event selector |
| 04 | thread | RELY-99 | Post on top, replies indented up to 3 levels, "show replies", fixed reply bar |
| 05 | reaction-picker | RELY-99 | Long-press emoji popover above a post card, one reaction selected |
| 10 | login | RELY-95 | Sign in to the node with username and password |
| 11 | login-error | RELY-95 | Login with wrong-credentials error |
| 12 | register | RELY-95 | Join via invite link: display name, username, password |
| 13 | register-errors | RELY-95 | Register with "username is taken" error |
| 14 | register-invalid-invite | RELY-95 | Register with invalid or used invite link |
| 15 | settings | RELY-100 | Profile, server, admin entry, sign out |
| 16 | settings-server-sheet | RELY-100 | Change-server bottom sheet with test connection |
| 17 | settings-server-error | RELY-100 | Server sheet with unreachable-address error |
| 20 | communities | RELY-97 | Nearby communities list with joined state |
| 21 | community-feed | RELY-97 | Community header and feed, Join button |
| 22 | community-about | RELY-97 | About tab: description, area, creator, members |
| 23 | join-pseudonym-sheet | RELY-97 | Join sheet to pick a pseudonym |
| 24 | create-community | RELY-97 | New community form: name, description, location, radius, pseudonyms |
| 30 | events-list | RELY-98 | Upcoming events grouped by day, List/Map toggle |
| 31 | events-map | RELY-98 | Map view with event pins |
| 32 | events-map-sheet | RELY-98 | Map with a selected event's bottom card |
| 33 | event-detail | RELY-98 | Event time, place, host, description, RSVP |
| 34 | event-detail-going | RELY-98 | Event detail after RSVP |
| 35 | create-event | RELY-98 | Event compose: title, start/end, venue, location, community |
| 40 | admin-queue | RELY-101 | Moderation queue with flagged posts and AI score |
| 41 | admin-queue-item | RELY-101 | Flagged post with AI review and actions |
| 42 | admin-members | RELY-101 | Member list |
| 43 | admin-member-sheet | RELY-101 | Member actions bottom sheet |
| 44 | admin-remove-confirm | RELY-101 | Confirm removing a member |
| 45 | admin-invites | RELY-101 | Create invite, pending and used invites |

## Using this for the Expo app

The mockups are the visual spec, not code to port: rebuild each screen in
React Native and match the render. The rules for doing that (tokens, fonts,
hard shadows as offset Views, no `borderRadius`, decoration limits) are in the
agent skill [`.pi/skills/relay-mobile-design/SKILL.md`](../.pi/skills/relay-mobile-design/SKILL.md).
Follow:

- **Tokens:** `shared.css` `:root` is the source of truth for colours, borders,
  shadows and the three font families.
- **Screens:** the table above maps each screen to its `screens/<NN>-<name>.html`
  and its render `png/<NN>-<name>.png`; 01–05 are the pattern set.
- **Style origin:** [`reference/landing-poster.jpg`](reference/landing-poster.jpg)
  is the Relay landing page on a phone, the look the design language comes from.

This is separate from `relay-design`, which covers the owner-facing web UI.
