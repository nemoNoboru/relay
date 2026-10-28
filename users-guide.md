# Relay — User's Guide

Relay is a micro-social network for you and your real friends: a small social
space for a group of friends, run on a computer one of them keeps. Posts are
ordinary static pages. People join by invitation link, not by creating an
account.

This guide covers the node as it works **today**, and marks the pieces that are
still being built. The plan is in [`docs/pivot-plan.md`](docs/pivot-plan.md).

```
keep a node  →  send an invite  →  publish a page  →  (soon) comment and browse
```

## 1. What you need

**To host a node:** a computer that stays on, a name for the node, and a data
directory. Docker, a binary, or Kubernetes all work. An old laptop is enough for
a few dozen friends.

**To join:** an invitation link from the host or another member. No email, no
password, no sign-up.

**To publish:** the invitation link and something to share — a folder of files,
a single HTML page, or a Markdown file.

## 2. Run a node (host)

```sh
# Docker
docker run -d --name relay -p 8080:8080 -v relay-data:/data \
  ghcr.io/mansus-org/relay:latest --name my-group

# or the binary
relay node --name my-group --data /var/lib/relay --listen :8080
```

- `--name` is the first path segment of every post on this node.
- On first start the admin password is **`relay`**.
- Open `{origin}/_relay/admin.html`, sign in, and choose a new password before
  inviting anyone.
- For HTTPS, pass `--tls-cert` and `--tls-key`, or put a reverse proxy in front.

Keep the data directory backed up. It contains the posts, the versions, the
invitation keys, and the admin password.

## 3. Invite people (host)

1. Open `{origin}/_relay/admin.html` and sign in.
2. Type a name for the person (letters, digits, `.`, `_`, `-`) and press
   **Create invitation**.
3. Copy the **invite link** and send it over a channel you trust (Signal, a
   password manager — not a public post or a commit).

A member can invite others from the publish page. Those people appear under the
inviter as `name/guest`.

- **Revoke** stops a key and everyone it invited.
- **Rotate** replaces a key with a fresh one for the same person; their invitees
  stay attached.

Treat an invite link like a password. Anyone who has it is that member.

## 4. Publish something (member)

Today there are two ways.

### Your home (easiest)

1. Open the invitation link you were sent. It opens your home and signs this
   browser in. You can close the page and come back; the cookie remembers you
   until the key is revoked or you sign out.
2. Press **Publish a page**.
3. Drop a folder, some files, or a single Markdown file onto the page.
4. Choose a name for the page: lowercase letters, numbers, and hyphens.
   For example `sunday-walk`.
5. Press **Publish**. The reply contains the live address.

A single Markdown file is turned into a styled page for you. A folder is served
as it is; keep links and asset paths relative.

On the home you can also set a display name and a short bio.

### Templates

The publish page lists the built-in templates: **Note**, **Link**, **Gallery**,
and **Article**. Open one, fill the form, check the live preview, and press
publish. The node writes the generated page as a normal post; the template code
runs in your browser, not on the server.

To make your own template, create a folder with `relay-template.json`,
`prepublish.html`, and `publish.js`, tar its contents, and use **Install a
template from a .tar** on the publish page. It appears in the picker for
everyone. The host can also install a folder:

```sh
relay template list --data /var/lib/relay
relay template install --data /var/lib/relay ./my-template
relay template remove --data /var/lib/relay my-template
```

Template code runs in the browser on the node origin. Only install templates
from people you trust.

### Relay a url

The publish page leads with **Relay a url**. Paste a link to a video, song,
photo, or any page and press **Relay**. The node downloads what the link
points at and hosts it as a new page — the video, song, or image files
straight, or a card with the page's preview image, title, and the original
url when there is nothing to download.

Downloads use [yt-dlp](https://github.com/yt-dlp/yt-dlp). If it is not
installed, or it can't download anything, the node still makes the share
card. The container image ships yt-dlp; on a self-built node install it
(`apt install yt-dlp`, `brew install yt-dlp`, …) for the download path.

### The home feed, profiles, and post pages

- **Your home** is an infinite scroll of every member's posts, newest first.
  There is no ranking and no “for you”; scroll and the next page loads itself.
- **Each post** carries its own comments and a comment box at the end. You do
  not have to open a separate page to reply.
- **A post's title** opens the post wrapper: the title, author, date, the page
  itself, and the full comment list.
- **An author's name** opens their profile: an infinite scroll of everything
  that member has shared, with the same inline comments.
- **Open page** goes to the static page at its own address. Gallery and
  Article posts can show a cover image in the feed.
- A post's title and summary come from a `relay-post.json` file in the upload,
  or from the page name if there is none. To set them, include:

  ```json
  { "title": "Sunday walk", "summary": "Six photos from the canal.", "template": "gallery" }
  ```

Existing sites can be added to the feed by the host:

```sh
relay index --data /var/lib/relay
```

### Comments

Comments sit under every post in the feed and on the post wrapper. They are
visible to members only. You can delete your own; the host can delete any from
the host console.

Comments allow plain text with paragraphs, `**bold**`, `*italic*`, and links to
`http(s)://` addresses or paths on the node. Everything else is shown as text.
The limit is 2,000 characters.

### The terminal

```sh
tar -C ./my-page -cf page.tar .

curl -X POST \
  -H "Authorization: Bearer $RELAY_KEY" \
  -H "Content-Type: application/x-tar" \
  --data-binary @page.tar \
  "https://relay.example.org/_relay/v1/sites/sunday-walk/deploy"
```

The reply:

```json
{ "site": "sunday-walk", "files": 3, "bytes": 5120, "url": "/my-group/sunday-walk/" }
```

The page is live at `{origin}{url}` immediately.

### What the current version does not do yet

- Any live key can currently replace any page on the node. Per-page ownership is
  not built yet. Invite only people you trust.

## 5. Edit, roll back, and remove

Publishing again with the same name replaces the page. The previous ten
versions are kept.

Roll back with the CLI (or the admin API):

```sh
export RELAY_ADMIN_TOKEN=…
relay rollback --origin https://relay.example.org --site sunday-walk
relay rollback --origin https://relay.example.org --site sunday-walk --to <id>
```

To remove a page, ask the host. There is no delete button in the browser yet.

## 6. What lives where

```
<data>/
  admin-token          host password (also the first-run default "relay")
  deploy-keys.json     invitations, members, page attribution
  sites/
    <site>/            the live page
    .versions/<site>/  the last ten versions
```

Copy the whole directory to back up the node. To move a page, copy its folder.
There is nothing to export from a proprietary store.

## 7. Keep the keys safe

- The **admin password** controls the whole node. Keep it out of shell history,
  logs, and chat. If it leaks, change it in the console.
- An **invite key** is a member's identity. Do not post it publicly. If it
  leaks, rotate it.
- The node stores plaintext invitation tokens so the host can re-copy links.
  Protect the data directory accordingly.

## 8. Troubleshooting

| What you see | What it means |
|---|---|
| `bad_token` | The key is wrong or was revoked. Ask for a new invite. |
| `bad_name` | Page names use lowercase letters, numbers, and hyphens. |
| `bad_archive` | The upload is not a tar, or it contains a link, an absolute path, or another unsupported entry. |
| `too_large` | The page is over 256 MB or has too many files. Make the images smaller. |
| Page looks broken | Use relative links (`images/cat.jpg`), not root-absolute ones (`/images/cat.jpg`). The page lives under a path, not at the top of the host. |
| Console says the password is still `relay` | The node has not been set up. Change it before inviting anyone. |

## 9. Deploy with an assistant

Give an AI assistant the node address, your invite key, and this skill:

```
{origin}/_relay/skills/relay-deploy/SKILL.md
```

Do not paste the key into a public chat. Use a private session.

## 10. Where this is going

Done: the member cookie and welcome page, the home feed, profile pages, post
wrappers, `relay index`, comments, and the seven post templates.

Next: backup/export, invite expiry, limits, and host moderation polish.

Read [`docs/pivot-plan.md`](docs/pivot-plan.md) for the details and the
acceptance tests. If a feature is not in that plan, it is not coming by
accident.
