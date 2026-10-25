# Relay mobile mockups (RELY-78)

Static HTML/CSS mock UIs of every screen of the Expo / React Native app, for
product-owner review. Design only: no app code. Tokens and components live in
`shared.css`; each screen is `screens/<NN>-<name>.html`.

Render everything (or only some screens) to `png/`:

```sh
./render.sh                      # all
./render.sh screens/01-feed.html # one
```

| # | Screen | Sub-issue | Shows |
|---|---|---|---|
| 01 | feed | RELY-96 | Locality feed, new-posts banner, post cards, "+" button |
| 02 | feed-location-off | RELY-96 | Location permission denied, prompt to enable |
| 03 | compose | RELY-96 | Text input with 500-char counter, location chip, community picker, Post/Event selector |
| 04 | thread | RELY-99 | Post on top, replies indented up to 3 levels, "show replies", fixed reply bar |
| 05 | reaction-picker | RELY-99 | Long-press emoji popover over a post card, one reaction selected |

Emoji render through Noto Color Emoji, which `render.sh` downloads into
`fonts/` (git-ignored) on first run.
