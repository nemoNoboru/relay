# Relay

**A micro-social network for you and your real friends.**

Relay is a small social space you and your people run yourselves. It lives on
one computer that someone in the group keeps, and it holds the things you want
to share: photos, links, writing, a poster, a page of notes.

```
one node  ·  4 to 29 friends  ·  pages, posts, pictures, comments
```

There is no sign-up. People join by invitation link, and the link is the
account. The aim is a quiet home: each person has a page listing what they
shared, and the node's home lists what everyone shared, newest first. No feed
algorithm, no follower counts, no ads, no company in the middle.

Relay is that place. The hosting core and the pivot's social layer are built
through Phase 4; Phase 5 is host operations. Read
[`docs/product.md`](docs/product.md) for the product and
[`docs/pivot-plan.md`](docs/pivot-plan.md) for the plan.

## What works today

- static pages at `/<node>/<post>/`
- invite links, delegated invites, and cascade revoke
- `tar` uploads from the browser or the command line
- atomic deploys, the last 10 versions, and rollback from the CLI or admin API
- a host console for invitations, members, and pages
- a member welcome page: open an invite once, stay signed in, edit your profile
- an infinite chronological home feed, profile pages, and post wrappers with cover images
- comments directly under every post in the feed and profile, with own-comment deletion and host moderation
- eleven post templates (Note, Link, Gallery, Article, Songs, Videos, Carousel, Jet Set Radio, Longform, Reply, Event), a template picker, and member-installable template tars
- "Relay a url" on the publish page: re-host a video, song, photo, or page from a link (yt-dlp download, with a social-preview card as the fallback)
- `relay index` to backfill posts from sites that predate the index
- native HTTPS or a reverse proxy

Not built yet: Phase 5 host operations (backup/export, invite expiry, limits,
and moderation polish).

## Keep a node

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh

relay node --name my-group --data /var/lib/relay --listen :8080
```

On first start the admin password is **`relay`**. Open
`/_relay/admin.html`, change it, and create the first invitation.

Docker:

```sh
docker run -d --name relay --restart unless-stopped \
  -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemonoboru/relay:latest --name my-group
```

## Publish a page

The easy way: open the invitation link, press **Publish a page** on your home,
drop a folder or a Markdown file, choose a name, and press Publish.

From a terminal:

```sh
tar -C ./my-page -cf page.tar .
curl -X POST \
  -H "Authorization: Bearer $RELAY_KEY" \
  -H "Content-Type: application/x-tar" \
  --data-binary @page.tar \
  "$RELAY_ORIGIN/_relay/v1/sites/sunday-walk/deploy"
```

The page is live at `$RELAY_ORIGIN/<node>/sunday-walk/`. With the `relay`
binary, the same upload is `relay deploy --origin $RELAY_ORIGIN --site
sunday-walk ./my-page`.

## Docs

| Doc | What's in it |
|---|---|
| [`users-guide.md`](users-guide.md) | **Start here.** Run a node, invite people, publish, roll back |
| [`docs/product.md`](docs/product.md) | What Relay is, what it is not, and the vocabulary |
| [`docs/pivot-plan.md`](docs/pivot-plan.md) | The phased plan for the social layer |
| [`docs/invites.md`](docs/invites.md) | Invitation links, member identity, cascade revoke |
| [`docs/publishing.md`](docs/publishing.md) | Posts, microwebsites, templates, the publish protocol |
| [`docs/architecture.md`](docs/architecture.md) | How the node is built today and what the pivot adds |
| [`docs/api.md`](docs/api.md) | Current endpoints and the planned API |
| [`web/README.md`](web/README.md) | The owner UI (Svelte + Vite, embedded in the binary) |
| [`docs/pivot.md`](docs/pivot.md) | The original pivot brief, kept as the source note |
| [`docs/navigation-pivot.md`](docs/navigation-pivot.md) | **Proposal.** Feed-first navigation: `/` as the public feed, `/host` as the landing, one chrome |
| [`docs/indieweb-syndication.md`](docs/indieweb-syndication.md) | **Proposal.** POSSE: feeds, Webmention, Micropub, syndication targets |
| [`docs/indieauth.md`](docs/indieauth.md) | **Proposal.** IndieAuth as the node's sign-in and token layer |
| [`docs/accounts.md`](docs/accounts.md) | **Proposal.** Member sign-in: invitation + passkeys via IndieAuth |
| [`docs/ui-ux-pivot.md`](docs/ui-ux-pivot.md) | **Proposal.** The visual redesign: keeps the theme, makes the feed a friendly timeline. Three mockups under [`ui-experiments/`](ui-experiments/) |
| [`docs/redesign.md`](docs/redesign.md) | **Spec.** The adopted whole-product redesign: quiet feed, soft sans (Nunito + Figtree), less-is-more copy |

## Develop

### Nix dev shell

The project has a pinned Nix flake with Go 1.26, Node 24/npm, make, staticcheck,
gopls, delve, Chromium, ImageMagick, ffmpeg, `gh`, and the basic CLI tools:

```sh
nix develop
```

Then use `make dev` / `make check` as below. The shell also sets `CHROME_BIN`
for screenshot and browser tooling.

### Local node

Build the web UI, build the binary, and run a local node:

```sh
make dev
# or, when make is not installed:
./scripts/dev.sh
```

It defaults to node `my-group`, data `./data`, and `:8080`. Override any of
them:

```sh
make dev NODE=my-group DATA=./data LISTEN=:8080
# or:
NODE=my-group DATA=./data LISTEN=:8080 ./scripts/dev.sh
```

The first start uses the admin password `relay`; open
`http://localhost:8080/_relay/admin.html` and change it.

Other gates:

```sh
make check   # web build + tests, gofmt/vet/staticcheck/test -race
make test    # go test -race ./...
make build   # owner UI first, then the relay binary
make run     # run the existing binary without rebuilding
```

`host.html` and `why.html` are written by hand. Edit the HTML directly.
Engineering standards for Go and the frontend live in [`.pi/skills/`](.pi/skills),
alongside the deploy skill every node serves at
`/_relay/skills/relay-deploy/SKILL.md`.

## Status

**Stage 0 (hosting MVP) and Phases 1–4 of the pivot are built.** Phases 1–3
added the member cookie and welcome, post indexing/feed/profiles/wrappers,
comments, and host comment deletion. Phase 4 adds eleven templates, the publish
endpoint, the template picker, and member-installable template tars. The next
shippable step is Phase 5: host operations. See [`docs/pivot-plan.md`](docs/pivot-plan.md).

**Proposals (not built):** the next product direction is documented in five
research notes — [`docs/navigation-pivot.md`](docs/navigation-pivot.md)
(feed-first navigation), [`docs/ui-ux-pivot.md`](docs/ui-ux-pivot.md) (the
visual redesign, with three home-feed mockups under [`ui-experiments/`](ui-experiments/)),
[`docs/indieweb-syndication.md`](docs/indieweb-syndication.md) (POSSE),
[`docs/indieauth.md`](docs/indieauth.md) (sign-in and token layer), and
[`docs/accounts.md`](docs/accounts.md) (member sign-in). They are designs, not
shipped behaviour. The chosen visual direction is specified in
[`docs/redesign.md`](docs/redesign.md).

Free software, owned by no one. Keep the server, invite your people, share what
you make.
