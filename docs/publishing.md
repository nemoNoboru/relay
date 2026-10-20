# Posts, Microwebsites, and Templates

A **post** on Relay is a small static website at
`/<node>/<site>/`, plus an entry in the node's post index. The site is the
content. The index is the social layer: who published it, when, with which
template, and under what title.

This document defines the post model, the template package, and the publish
protocol. The post index and templates are built (Phases 2 and 4 of
[`pivot-plan.md`](pivot-plan.md)); the wire details are in [`api.md`](api.md).

## Why posts are sites

- The hosting core already serves static files safely and atomically.
- A post can be read without Relay: open the folder, or fetch the URL.
- Rollback already works per site.
- A template is just a small site that generates another site.
- Export and backup stay simple: copy the data directory.

The social layer never owns the bytes. It indexes them.

## The post model

```json
{
  "id": "post-2026-10-19-a1b2",
  "site": "sunday-walk",
  "member": "dk_…",
  "title": "Sunday walk",
  "summary": "Six photos from the canal, and a map.",
  "template": "gallery",
  "publishedAt": "2026-10-19T16:04:00Z",
  "updatedAt": "2026-10-19T16:04:00Z",
  "hidden": false
}
```

| Field | Meaning |
|---|---|
| `id` | Stable post id. Can be derived from the site and a short random suffix. |
| `site` | The path segment under `/<node>/`. Must satisfy the site-name grammar. |
| `member` | The member key id that published the current version. |
| `title` | Shown in the feed, profile, and wrapper. |
| `summary` | One or two lines. Optional. |
| `template` | Template id, or `custom` for a plain upload. |
| `image` | Optional site-relative cover image for feeds and profiles. |
| `publishedAt` | First publication time. |
| `updatedAt` | Last publish time. |
| `hidden` | Host moderation flag. Hidden posts are omitted from feed/profile but the URL still serves. |

A post is created on the first successful publish of a site and updated on
later publishes. Rolling back a site updates `updatedAt` but does not create a
second post.

### Where the metadata lives

Two places, on purpose:

1. **`<data>/posts.json`** — the authoritative index for ordering and profiles.
2. **`<data>/sites/<site>/relay-post.json`** — a portable sidecar written by
   Relay, so an exported site carries its own title, author label, template, and
   date.

If the index and the sidecar disagree, the index wins and the sidecar is
rewritten on the next publish. The sidecar is never required to serve the site.

## Site layout

A post is an ordinary site folder:

```
<data>/sites/sunday-walk/
  index.html
  relay-post.json
  assets/
    photo-1.jpg
    photo-2.jpg
  styles.css
```

Rules (the same as today's uploader):

- relative paths only; no `..`, no absolute paths, no symlinks or hard links;
- regular files and directories only;
- at most 65,536 files and 64 MiB of decoded files for a template publish;
  the plain tar endpoint keeps its 256 MiB cap;
- paths are relative to the site root;
- a successful publish replaces the whole site atomically;
- the previous 10 versions are kept for rollback.

## Templates

A template is a small site that knows how to make a post. It has a form for the
author and code that turns the form into files.

### Package layout

```
<data>/templates/article/
  relay-template.json      manifest
  prepublish.html          the form the member fills in
  publish.js               runs in the browser, produces the files
  preview.css              optional, used by the form's preview
  assets/                  optional, copied into the post if declared
```

Built-in templates ship inside the binary (Note, Link, Gallery, Article,
Songs, Videos, Images carousel, Jet Set Radio, Longform, Reply, Event) and are
served from the embedded folders. Installed templates live in
`<data>/templates/<id>/`; the host can shadow a built-in with the CLI, while a
member tar install cannot take a built-in id. The host manages installed
templates with `relay template list|install|remove`.

### Manifest

```json
{
  "schema": 1,
  "id": "article",
  "name": "Article",
  "description": "A longer piece of writing with a title and a hero image.",
  "version": "1.0.0",
  "fields": [
    { "name": "title", "type": "text", "required": true, "max": 120 },
    { "name": "standfirst", "type": "textarea", "max": 280 },
    { "name": "body", "type": "richtext", "required": true },
    { "name": "hero", "type": "image", "accept": "image/*", "maxBytes": 8388608 }
  ],
  "entry": "prepublish.html",
  "script": "publish.js",
  "assets": ["preview.css"]
}
```

Field types for the first version:

| Type | UI | Output |
|---|---|---|
| `text` | single-line input | string |
| `textarea` | multi-line input | string |
| `richtext` | simple rich text / Markdown | HTML or Markdown, template's choice |
| `image` | file picker / drop | uploaded file |
| `url` | URL input | string |
| `gallery` | multiple images | uploaded files |

The manifest is for the picker and validation. The template's own HTML decides
how the fields actually look. Relay does not render the form.

### `prepublish.html`

The form. It is served at:

```
/_relay/templates/<id>/prepublish.html
```

It runs on the node origin, so it can use relative URLs. It should:

- render the fields from the manifest (or hard-code them);
- preview the result in an iframe or a local DOM preview;
- load `publish.js`;
- call `publish()` on submit;
- show errors from the publish endpoint.

The page must work without a build step. Plain HTML, CSS, and JavaScript.

### `publish.js`

Runs in the member's browser. It receives the form values and returns the files
that make up the post:

```js
// Contract: publish(values, files) -> { title, summary, site?, files[] }
export async function publish(values, files) {
  const html = renderArticle(values); // template's own code
  return {
    title: values.title,
    summary: values.standfirst,
    site: slug(values.title), // optional; the server can derive one
    files: [
      { path: "index.html", content: html },
      ...files.images.map((f) => ({ path: `assets/${f.name}`, bytes: f.bytes })),
    ],
  };
}
```

The endpoint accepts text content as a string and binary content as base64. A
template that needs to build a tar itself should use the existing deploy
endpoint instead; the template publish endpoint is deliberately small.

### What Relay does and does not do

Relay:

- lists templates;
- serves `prepublish.html` and its assets;
- validates the member, template, file paths, names, and size;
- writes the site atomically and snapshots the previous version;
- writes `relay-post.json` and updates `posts.json`;
- appends the comments embed snippet if the template asks for it;
- returns the post URL.

Relay does **not**:

- execute `publish.js` on the server;
- execute arbitrary template HTML in the node process;
- fetch remote assets or templates;
- install a template from a public marketplace;
- let a template write outside its own site.

Templates are installed by the host and trusted like any page on the node. A
future untrusted-template feature needs a separate origin and a sandbox; it is
not in this pivot.

## The publish endpoint

```
POST /_relay/v1/publish
Authorization: Bearer <invite token>      (optional if the member cookie is set)
Content-Type: application/json
```

Request:

```json
{
  "template": "article",
  "title": "The canal in October",
  "summary": "A walk, six photos, and a map.",
  "site": "canal-in-october",
  "files": [
    { "path": "index.html", "content": "<!doctype html>…" },
    { "path": "assets/map.png", "base64": "iVBORw0KGgo…" }
  ]
}
```

Response:

```json
{
  "post": {
    "id": "post-…",
    "site": "canal-in-october",
    "title": "The canal in October",
    "url": "/mynode/canal-in-october/",
    "postUrl": "/_relay/posts/canal-in-october"
  },
  "files": 2,
  "bytes": 184320
}
```

Validation, in order:

1. The member is live and not revoked.
2. The template exists and is installed.
3. `site` is valid and not reserved (`_relay`). If omitted, derive a slug from
   `title` and add a short suffix if taken.
4. An explicit `site` follows the current rule: any live member may replace
   it, and the index records the last author. An omitted `site` is derived from
   the title and made unique with a numeric suffix, so two posts with the same
   title do not overwrite each other.
5. Every path is relative, has no `..`, and stays under the site root.
6. File count and total decoded bytes are under the limits.
7. A site called `_relay` or a template id cannot be used as a member site.

Then:

1. Snapshot the previous site version.
2. Write the new site to a temporary directory and rename it into place.
3. Write `relay-post.json`.
4. Upsert the post index.
5. Reload the site registry.
6. Return the URLs.

If any step fails, the live site is untouched.

## Plain uploads

The existing endpoint stays:

```
POST /_relay/v1/sites/<site>/deploy
Authorization: Bearer <invite token>
Content-Type: application/x-tar
```

A plain upload:

- creates or updates a post with `template: "custom"`;
- derives the title from the site name unless `relay-post.json` is present in
  the archive;
- appears on the home and profile of the uploading member;
- does not get a generated comments snippet (the wrapper still provides
  comments).

This is the escape hatch for people who already have a site, and for AI
assistants following the deploy skill.

## The post wrapper

Every post has a generated wrapper at:

```
/_relay/posts/<site>
```

It shows:

- title, author display name, date;
- a link to the raw site;
- the site in an iframe (or a large preview card);
- the comments panel.

The wrapper is not the canonical URL for the content. The canonical URL remains
`/<node>/<site>/`. The feed links to the wrapper so the group sees comments and
author context; "open the page" links to the real address.

## Comments on the wrapper

Static posts are served under a CSP sandbox: their scripts cannot call member
APIs with the viewer's cookie. Comments therefore live on the post wrapper at
`/_relay/posts/<site>`, which is a trusted member page, not inside the static
post itself.

The post wrapper shows the comment list and form. A template-generated post
links back to the wrapper for comments; it does not embed a comment script.

An embeddable panel would need an isolated origin or a scoped comment token so
that a post script cannot act as the signed-in member. That is deferred; the
wrapper is the comment surface for this phase.

## Built-in templates

| Template | Fields | Output |
|---|---|---|
| **Note** | title, body | a single readable page |
| **Link** | url, title, note | a card with the link and the member's note |
| **Gallery** | title, images, caption | a photo grid |
| **Article** | title, standfirst, hero, body | a long-form reading page |
| **Songs** | title, description, audio files | a playlist with native audio players |
| **Videos** | title, description, video files | a stack of native video players |
| **Images carousel** | title, images, caption | a swipeable image carousel |
| **Jet Set Radio** | title, description, audio files | a loud street-radio playlist |
| **Longform** | title, byline, standfirst, hero, body | a clean Medium-like long read |
| **Reply** | url, body | a reply to someone else's post |
| **Event** | title, when, where, details | an invitation: what, when, where |

The built-ins live in one embedded folder tree and are discovered by reading
`relay-template.json`; there is no server-side list of template ids.

Each is a normal template folder using the shared browser runtime at
`/_relay/template/runtime.js` and `runtime.css`. The set reads like the same
product: constructivist type, paper, red, ink, and ochre. No template needs a
framework or a build step.

## Creating and sharing a template

A template is a folder:

```
my-template/
  relay-template.json   manifest (id, name, fields, entry, script)
  prepublish.html       the form the member fills in
  publish.js            browser-side code that produces {title, summary, files}
```

Members can share the folder as a tar and install it from the publish page
(the **Install a template from a .tar** panel) or the host can install a folder
with `relay template install`. The new template appears in the picker and in
`GET /_relay/v1/templates` immediately.

Template code runs in the browser on the node origin, so it is trusted like a
page on the node. Install only templates from people you trust, and read
`publish.js` before installing. Built-in ids cannot be shadowed by a member
install; the host CLI can still override one intentionally.

## Open questions

- Should templates be versioned independently of the node? (Manifest has a
  `version`; no update mechanism yet.)
- Should a template be able to declare extra assets that are copied into every
  post? (Yes, via `assets`, but keep it small.)
- Should the publish endpoint accept a prebuilt tar for large galleries?
  (The deploy endpoint already does; templates can fall back to it.)
- How do we let a member edit a post later? (Phase 5: re-open the template with
  the previous values. The site sidecar makes this possible.)

These are tracked in [`pivot-plan.md`](pivot-plan.md) §7.
