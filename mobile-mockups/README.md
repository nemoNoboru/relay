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
