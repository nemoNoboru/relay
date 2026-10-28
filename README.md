# Relay

**A micro-social network for you and your real friends.**

Relay is a small social space you and your people run yourselves. It holds the
things you want to share: photos, links, writing, a poster, a page of notes.
There is no sign-up. People join by invitation link, and the link is the
account. No feed algorithm, no follower counts, no ads, no company in the middle.

Read [`docs/product.md`](docs/product.md) for the product.

## Where things live

| Path | What it is |
|---|---|
| [`backend/`](backend/README.md) | Python (FastAPI) backend. Quickstart, config and checks in its README |
| [`mobile/`](mobile/README.md) | Expo / React Native app |
| `host.html`, `why.html`, `styles.css`, `web/fonts/` | The static site published to GitHub Pages |
| [`mobile-mockups/`](mobile-mockups/README.md), [`ui-experiments/`](ui-experiments/) | Visual specs and design experiments |
| [`docs/`](docs/) | Product, design and backend development notes |
| [`.pi/skills/`](.pi/skills) | Engineering and design standards (`relay-design`, `relay-mobile-design`) |

The earlier Go implementation has been removed. Several design notes in `docs/`
were written against it and say so at the top.

## Docs

| Doc | What's in it |
|---|---|
| [`docs/product.md`](docs/product.md) | What Relay is, what it is not, and the vocabulary |
| [`docs/backend-development.md`](docs/backend-development.md) | Backend architecture, DI, testing plan, ruff/ty/pre-commit, Docker |
| [`docs/pivot-plan.md`](docs/pivot-plan.md) | The phased plan for the social layer |
| [`docs/invites.md`](docs/invites.md) | Invitation links, member identity, cascade revoke |
| [`docs/publishing.md`](docs/publishing.md) | Posts, microwebsites, templates, the publish protocol |
| [`docs/pivot.md`](docs/pivot.md) | The original pivot brief, kept as the source note |
| [`docs/navigation-pivot.md`](docs/navigation-pivot.md) | **Proposal.** Feed-first navigation |
| [`docs/indieweb-syndication.md`](docs/indieweb-syndication.md) | **Proposal.** POSSE: feeds, Webmention, Micropub |
| [`docs/indieauth.md`](docs/indieauth.md) | **Proposal.** IndieAuth as sign-in and token layer |
| [`docs/accounts.md`](docs/accounts.md) | **Proposal.** Member sign-in: invitation + passkeys |
| [`docs/ui-ux-pivot.md`](docs/ui-ux-pivot.md) | **Proposal.** The visual redesign, with mockups under [`ui-experiments/`](ui-experiments/) |
| [`docs/redesign.md`](docs/redesign.md) | **Spec.** The adopted redesign: quiet feed, soft sans, less-is-more copy |

## Develop

Backend (needs only [uv](https://docs.astral.sh/uv/)):

```sh
cd backend
uv sync
uv run pytest
```

Mobile: see [`mobile/README.md`](mobile/README.md). A pinned Nix dev shell with
uv, Node and bun is available via `nix develop`.

`host.html` and `why.html` are written by hand. Edit the HTML directly.

Free software, owned by no one.
