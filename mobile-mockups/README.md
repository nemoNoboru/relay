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
