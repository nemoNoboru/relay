---
name: relay-design
description: Relay's visual design system — the quiet feed, Nunito/Figtree typography, the one-chrome shell, and the less-is-more copy rule. Use when writing or reviewing web-facing CSS (host.html, why.html, ui-experiments/), adding a surface, or restyling a page.
---

# Relay Design

The adopted design system for Relay's owner-facing UI. The whole-product spec
is [`docs/redesign.md`](../../../docs/redesign.md); the reference code is
[`ui-experiments/v4-soft-sans.html`](../../../ui-experiments/v4-soft-sans.html).
When a styling question is not answered here, read those two files before
writing code.

## Read first

- [`docs/redesign.md`](../../../docs/redesign.md) — the spec (whole product).
- [`ui-experiments/v4-soft-sans.html`](../../../ui-experiments/v4-soft-sans.html) —
  the reference implementation: tokens, type scale, shell, and card anatomy.
- [`docs/ui-ux-pivot.md`](../../../docs/ui-ux-pivot.md) — how V4 was chosen;
  V1–V3 are the rejected alternatives.

## Non-negotiables

1. **Colours are the only unchanged theme asset.** Use them one accent at a
   time (`--red` is the accent; `--ochre` is rare).
2. **Nunito + Figtree, self-hosted.** Never Anton, Oswald, or Space Mono in new
   member-facing UI. Fonts load from `web/fonts/fonts.css`.
3. **Soft grammar.** Hairline borders (`--hair`), rounded corners, no noise
   overlay, no hard 3–4px borders, no offset block shadows, no rotated tags on
   member surfaces.
4. **One chrome.** One shared header on every member surface.
5. **The card is a preview of the real thing.** Peek for text templates, cover
   for media templates; the whole preview is the link, never a metadata row.
6. **Comments render only when there is something to show.** `N comments` only
   when N > 0; the reply form is member-only, guests see the thread.
7. **Less is more.** No explainer paragraphs, no empty comment counts, no
   policy footnotes. If a line has nothing to say, don't render it.
8. **Guest / member / host share the same content; only the write controls
   differ.**

## Tokens

| Token | Value |
|---|---|
| `--red` | `#ce1f2b` |
| `--red-deep` | `#a5141f` |
| `--ink` | `#141210` |
| `--paper` | `#f0e6d2` |
| `--paper-2` | `#e6d8bc` |
| `--ochre` | `#dfa32a` |
| `--hair` | `rgba(20,18,16,.16)` |
| `--surface` | `#fbf6e9` |
| `--r-card` | `12px` |
| `--r-inner` | `9px` |
| `--r-avatar` | `11px` large / `8px` small |

## Type

| Role | Face | Weight |
|---|---|---|
| Display | Nunito | 700 (mark: 800) |
| Body | Figtree | 300 / 400 / 600 |
| Eyebrow / meta | Figtree | 600, uppercase, `.08em` tracking |

## Chrome

```
[ R · node-name ]     Home   Publish   People      ( alice · sign out )
```

Sticky paper bar at ~92% opacity with a hairline bottom border, 58px tall,
max-width 820px. The active nav item is underlined in red. Sign-in is an
inline prompt (paste the invitation key or link), never a full page.

## Definition of done

- [ ] Tokens and type match this skill; no Anton/Oswald/Space Mono in new UI
- [ ] One chrome
- [ ] Cards use peek/cover, not metadata rows
- [ ] No filler copy; empty states omitted (§ "less is more")
- [ ] Native `a`, `button`, `details/summary`, `input`, `label`; keyboard-usable
- [ ] Screenshots updated when the visual changes
