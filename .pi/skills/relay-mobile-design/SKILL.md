---
name: relay-mobile-design
description: Relay's mobile visual language, the "Constructivist poster" (cream paper, black header band with red stripe, hard offset shadows, Anton / Oswald / Space Mono). Use when building or reviewing UI in the Expo / React Native app, adding a screen or component, or checking a screen against the mockups.
---

# Relay Mobile Design

The approved design for the Expo / React Native app. The mockups in
[`mobile-mockups/`](../../../mobile-mockups/README.md) are the visual spec, not
code to port: rebuild each screen in React Native and match the PNG.

**Two design systems, do not mix them.**

| Surface | Skill | Look |
|---|---|---|
| Web-facing UI (static pages, design experiments) | [`relay-design`](../relay-design/SKILL.md) | Quiet feed, Nunito/Figtree, soft rounded corners |
| The Expo mobile app | this skill | Poster: Anton/Oswald/Space Mono, square corners, hard shadows |

Anton, Oswald and Space Mono belong to the app (and the landing page), never to
web member UI. Nunito/Figtree and soft corners never go in the app.

## Reference

- Tokens, source of truth: [`mobile-mockups/shared.css`](../../../mobile-mockups/shared.css) (`:root`, and the header comment listing every component).
- Per-screen images: [`mobile-mockups/png/`](../../../mobile-mockups/png) (780x1688, a 390x844 phone at 2x). Screen table and area CSS: [`mobile-mockups/README.md`](../../../mobile-mockups/README.md).
- Origin of the style, the Relay landing page on a phone: [`mobile-mockups/reference/landing-poster.jpg`](../../../mobile-mockups/reference/landing-poster.jpg).
- Reference set to copy patterns from: `png/01-feed.png` to `05-reaction-picker.png`.

If this file and `shared.css` disagree on a value, `shared.css` wins; fix this file.

## Tokens

| Token | Value | Use |
|---|---|---|
| `paper` | `#f0e6d2` | Screen background |
| `paper2` | `#e4d5b7` | Secondary blocks, grid paper |
| `surface` | `#f8f1e3` | Cards, inputs, sheets |
| `ink` | `#141210` | Text, borders, shadows, header band |
| `ink2` | `#5c544a` | Muted text (never opacity) |
| `red` | `#ce1f2b` | Blocks, stripes, rules, display type 24px+ |
| `redText` | `#b81a25` | Small red text on cream |
| `redDeep` | `#a5141f` | Pressed / deep red |
| `ochre` | `#dfa32a` | Accent on black, highlights |
| `onRed` | `#fff8ea` | Text on red |
| `hair` | `rgba(20,18,16,.22)` | Hairline dividers only |

Contrast rules (WCAG AA):

- Text on red is `onRed` (5.2:1), never plain paper (4.4:1).
- Small red text on cream is `redText`. Plain `red` only for blocks, rules and display type 24px and up.
- On black, accents are `ochre` or paper. Red stays a block or stripe there, never text.
- Colour-block, don't tint: no translucent red washes.

## Shape, borders, shadows

- Square corners. No `borderRadius` except circles in decoration.
- Borders 2px ink; 3px for emphasis (buttons, popovers, focus).
- Shadows are hard offsets: 5px (cards, buttons) or 3px (small), ink. A red or ochre offset marks a highlighted element.

## Type

| Role | Font | Rules |
|---|---|---|
| Display | Anton | Uppercase, line height about .86, headlines, titles, avatar letters. Key word in `red` (`ochre` on black). |
| Body | Oswald | Post text, rows, paragraphs, 17-20px |
| Mono | Space Mono | Uppercase, `letterSpacing` .08-.2em: labels, buttons, tabs, nav, metadata, counters |

Section labels get a short red rule in front.

## Page anatomy

- **Header.** Every screen opens with a black band and a 6px red bottom stripe. Compact: Anton 22px title, back/action buttons 44px. Large: Anton 38px title plus an ochre mono subtitle. Hero screens (feed) use the poster header: logo row, giant title, ochre mono meta line. The status bar area is black too, with light content.
- **Logo.** Red square rotated -6deg, heavy Anton "R", ochre corner chip, hard shadow. Next to it the "RELAY" wordmark in Anton.
- **Tab bar.** Paper, 3px ink top border, mono 9.5px labels. Active tab: 6px red bar on top, red icon.
- **Decoration** (geometric shapes: sun, ring, wedge, bar, square, quarter; grid paper; tilted stamps; vertical "RELAY" tag): headers, hero areas, empty states, error states, sheet heads. One or two shapes per screen, bleeding off the edge.
- **Content stays plain.** Post text, forms and lists have no decoration behind or over them. Flash lives only in headers, hero, empty/error states and sheets.
- **Tap targets** at least 44px. Buttons are 50px (small 44px); segmented items and tabs 44px or more.

## Components

Class names are from `shared.css`; build RN components with the same names.

- Navigation: `Navbar` (compact / large), `PosterHead`, `TabBar`, `Fab` (red square, hard shadow), `Banner` ("3 new posts" tilted ink tag).
- Posts: `Card` (2px border, hard shadow, ruled footer), `Avatar` (Anton letter on a colour block; sizes sm 30, 42, lg 76), `ReactionRow` (`mine` = solid red, `add`), repliesCount `Stamp`.
- Actions: `Button` (`primary` red, `dark`, `ghost`, `danger`, `block`, disabled, `sm`), `Seg` (ink fill for the selected item), `Toggle`, `Chip`.
- Forms: `Field`, `Input`, `Textarea`, `Label`, `Hint`, `Counter`. Errors in `redText` with a red border.
- Lists: `Row` (title, sub, end, chevron), `SectionLabel`.
- Overlays: `Scrim`, `BottomSheet` (grab bar, Anton title), `Popover`/`ReactionPicker` (3px border, hard shadow).
- Empty state: `Empty` with a badge and decoration, short copy.
- Poster pieces: `Logo`, `Wordmark`, `Display` (with `hot` / `ochre` words), `Kicker` (mono line after a 30px red rule), `Stamp` (red/ochre/ink), `Vtag`, `Geo*` shapes, `GridPaper`, `Stripe`, `Mono`.

## Translating to React Native

- **Tokens object.** One `theme.ts` exporting colours, `line` (2/3), `shadow` offsets, `fonts`, `type` presets. Components read it; no hex literals in screens. Mirror `shared.css` names.
- **Fonts.** Load Anton, Oswald and Space Mono with `expo-font` (`useFonts`, e.g. from `@expo-google-fonts/*`) and hold the splash screen until ready. RN has no synthetic weights per family: use the exact family names (`Anton_400Regular`, `Oswald_400Regular` / `_700Bold`, `SpaceMono_400Regular` / `_700Bold`) and never set `fontWeight` on a custom font.
- **Hard shadows.** RN `shadow*` and `elevation` are soft and platform-specific: do not use them. Render the shadow as a sibling `View` behind the element, same size, `ink` background, offset with `top`/`left` (5 or 3), inside a wrapper that reserves the extra space. Build one `HardShadow` wrapper and reuse it for cards, buttons, FAB, popovers.
- **No rounding.** Never set `borderRadius`, except `width/2` circles in decoration.
- **Mono labels.** RN has no CSS `text-transform`: write `.toUpperCase()` in a `MonoLabel` component and set `letterSpacing` in px (em x font size). Same `Display` helper for Anton.
- **Geometric shapes.** `position: 'absolute'` Views inside a parent with `overflow: 'hidden'`; rotation via `transform: [{ rotate: '-6deg' }]`. Circles are `borderRadius: size/2`; wedges and triangles use a rotated square, or `react-native-svg` when needed. Keep content siblings later in the tree (or `zIndex`) so they sit above. Decoration gets `pointerEvents="none"` and is hidden from accessibility.
- **Grid paper.** A repeated `react-native-svg` pattern, or skip it; never a large bitmap.
- **Icons.** Square-capped, 2px-stroke line icons (see the `.ic-*` set in `shared.css`); `react-native-svg` with `color` from the theme.
- **Status bar.** `expo-status-bar` style `light` over a black header band.
- **Touch.** `Pressable` with `hitSlop` to reach 44px; pressed state shifts the element onto its shadow (translate by the shadow offset) instead of fading.
- **Safe area.** `react-native-safe-area-context`; the black band extends under the status bar, the tab bar under the home indicator.

## Review checklist

1. Matches the PNG for that screen at 390 wide.
2. Colours come from the theme; contrast rules above hold.
3. No `borderRadius`, no soft shadows or elevation.
4. Display text is Anton uppercase, labels are mono uppercase tracked, body is Oswald.
5. Header has the black band and red stripe.
6. Decoration only in headers, hero, empty/error states and sheets; content is plain.
7. Every tap target is 44px or more.
