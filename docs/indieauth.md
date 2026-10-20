# Relay — IndieAuth Research

**Status:** research proposal, not implemented. This document studies IndieAuth
and how Relay could use it. It changes nothing except itself. It supersedes the
note in [`indieweb-syndication.md`](indieweb-syndication.md) §5, which skipped
IndieAuth; on reflection, IndieAuth is the single most useful building block in
the IndieWeb stack for Relay, because it is what lets standard clients talk to
a Relay node at all.

---

## 1. TL;DR

IndieAuth is an **identity layer on top of OAuth 2.0 where every identity is a
URL**. A Relay member already *has* a URL — their profile,
`/_relay/people/<id>`. So the node can be an **IndieAuth server** for its
members, and that one fact unlocks the rest of the IndieWeb:

- A member signs in to any IndieWeb client (Quill, IndieKit, Micro.blog apps)
  with their Relay profile URL.
- The client discovers the node's authorization/token endpoints, the member
  approves scopes (e.g. `create`), and the client gets a Bearer token.
- The client uses that token against the node's **Micropub** endpoint to post
  on the member's behalf.

This is the missing piece that makes Relay an IndieWeb *publish target*, not
just a feed. The node can also act as an **IndieAuth client**, letting members
sign in to Relay with their own domain — useful later, but not the main win.

In one line: **IndieAuth is how Relay's invitation becomes usable by every
IndieWeb app, without Relay inventing an account system.**

## 2. What IndieAuth is

IndieAuth is a W3C-published extension of OAuth 2.0. Its distinctive choices:

- **Users are URLs.** A "user" is a resolvable profile URL (`me`), e.g.
  `https://example.com/`. There is no username/password registry.
- **Clients are URLs.** A `client_id` is a URL that carries the client's
  metadata, so clients don't register in advance.
- **Discovery by link relations.** Given a profile URL, a client fetches it and
  looks for `rel="indieauth-metadata"` (or the legacy
  `rel="authorization_endpoint"` / `rel="token_endpoint"`), which points at an
  OAuth 2.0 server-metadata document (RFC 8414) naming the endpoints.
- **Standard OAuth machinery.** Authorization code flow with PKCE, Bearer
  tokens, refresh tokens, token introspection (RFC 7662, plus a `me` field),
  and token revocation (RFC 7009).
- **Scopes** like `create`, `update`, `delete`, `media`, `read`, `profile`.

The two endpoints:

- **Authorization endpoint** — the human-facing page that authenticates the
  user and asks them to approve the client and scopes. *How* it authenticates
  the user is deliberately out of scope.
- **Token endpoint** — the machine-facing endpoint that exchanges the
  authorization code for an access token plus the user's canonical `me` URL.

The flow, concretely:

```
member types: https://node.example/_relay/people/alice
        │
        ▼
client fetches that URL ──► finds rel="indieauth-metadata"
        │
        ▼
client redirects browser to authorization_endpoint
        │  (member signs in, approves "create")
        ▼
client exchanges code at token_endpoint ──► access_token + me
        │
        ▼
client calls Micropub with: Authorization: Bearer <token>
```

## 3. Correcting the earlier note

[`indieweb-syndication.md`](indieweb-syndication.md) §5 said IndieAuth "does
not fit Relay's group-first model" because IndieAuth signs in with a domain and
Relay has no per-member domains. That conflates *how a user proves identity*
with *what the identity is*.

IndieAuth's identity is just a URL. Relay's members have URLs
(`/_relay/people/<id>`), and — crucially — the **node already knows how to
prove a member controls theirs**: the invitation link. IndieAuth explicitly
leaves "how the authorization endpoint authenticates the user" out of scope, so
Relay plugs its invitation/cookie into that slot and keeps its identity model
intact.

The earlier proposal (reuse the invitation token as the Micropub bearer) still
works for Relay's own tools, but it does **not** let off-the-shelf IndieWeb
clients talk to Relay, because those clients only know how to do IndieAuth.
IndieAuth is therefore the *enabling* piece, not an alternative.

## 4. The two roles Relay can play

### 4.1 Relay as an IndieAuth **server** (the main win)

The node hosts authorization + token (+ revocation + introspection) endpoints,
and each member's profile URL is an IndieAuth identity.

```
Quill / IndieKit / Micro.blog app
        │  "sign in with https://node.example/_relay/people/alice"
        ▼
Relay node (IndieAuth server)
        │  member signs in with their invitation link (existing)
        │  member approves: "this app may create posts"
        ▼
Bearer token ──► Micropub endpoint on the same node
```

- The member's `me` is `https://node.example/_relay/people/<id>`.
- The "authenticate the user" step reuses Relay's invitation exchange: open the
  invite link (or paste it) → member cookie → approve the client. No password,
  no email, no new identity system.
- A revoked invitation immediately invalidates the member's ability to re-sign
  in and, if tokens are introspected through the same key store, invalidates
  their existing tokens too.

This is what makes Relay an IndieWeb **publish target**: standard clients can
sign in and post.

### 4.2 Relay as an IndieAuth **client** (secondary)

The node (or a member) signs in to a *third-party* IndieWeb service by URL.

- **Members signing in to Relay with their own domain.** A member who already
  has a personal site can use "sign in with your domain" as an *alternative*
  credential; the host maps that domain to the member record. Useful for
  indieweb-native friends, but most members won't have a domain, so it stays an
  optional extra, never the default.
- **The host signing in to other IndieWeb services.** When Relay later POSSEs
  *out*, or the host wants the node to reply to another site via Webmention
  "as the group", the host's own identity (or the node's identity) can be an
  IndieAuth/domain identity.

## 5. Design: Relay as an IndieAuth server

### 5.1 Endpoints

| Path | Role |
|---|---|
| `/_relay/v1/indieauth/metadata` | Server metadata (RFC 8414): `authorization_endpoint`, `token_endpoint`, `revocation_endpoint`, `introspection_endpoint`, `scopes_supported`, `code_challenge_methods_supported` |
| `/_relay/v1/indieauth/authorization` | Authorization endpoint: member signs in (invitation), approves client + scopes, issues a code |
| `/_relay/v1/indieauth/token` | Token endpoint: exchanges the code (PKCE) for `access_token`, `token_type: Bearer`, `scope`, `me`, `expires_in` |
| `/_relay/v1/indieauth/revocation` | Revokes a token (RFC 7009) |
| `/_relay/v1/indieauth/introspection` | Introspects a token (RFC 7662 + `me`) — internal use by the node's Micropub endpoint |

### 5.2 Identity and discovery

- Each profile page `/_relay/people/<id>` emits:
  ```html
  <link rel="indieauth-metadata" href="/_relay/v1/indieauth/metadata">
  <link rel="micropub" href="/_relay/v1/micropub">
  ```
  (plus `h-card` with `p-name` and `u-photo`, from the syndication plan).
- The canonical `me` returned by the token endpoint is
  `https://node.example/_relay/people/<id>`.
- The metadata endpoint advertises the scopes Relay supports.

### 5.3 "Authenticate the user" = the invitation

This is the one Relay-specific part, and it is already built:

1. The authorization endpoint shows "Sign in with your invitation."
2. The member pastes the invite link/key (or the browser already holds the
   member cookie).
3. The node resolves the key to a member and shows the consent screen:
   *"{App} wants to create posts as {display name}. Allow?"*
4. On consent, the node issues an authorization code bound to that member,
   client, redirect URI, and scope set.

No new credential is introduced; IndieAuth is a standard *wrapper* around the
invitation.

### 5.4 Scope mapping

| IndieAuth scope | Relay meaning | Support |
|---|---|---|
| `create` | Publish posts (Micropub create) | ✅ core |
| `media` | Upload media (Micropub media endpoint) | ✅ (files are already part of publishing) |
| `update` | Edit a post (Micropub update) / edit own profile | ⚠️ post-edit is not built; map to profile edit only, or omit |
| `delete` | Delete own post (Micropub delete) | ❌ members can't delete posts today; omit until it exists |
| `read` | Read feed/profile (already public) | optional, mostly for completeness |
| `profile` | Read `name`/`photo`/`url` | ✅ return display name + profile URL |

The node advertises only the scopes it actually honours. A client requesting an
unsupported scope gets a normal OAuth error.

### 5.5 Token handling

- Access tokens are short-lived (minutes to hours) with refresh tokens for
  long-lived clients.
- Tokens are stored server-side (hashed, like invite tokens), so **revoking an
  invitation revokes its member's tokens** through the same key store.
- PKCE is required (`S256`), since the client is not a confidential client.
- The node's Micropub endpoint validates tokens by introspection against its
  own token store (no separate round trip when both live on the node).

## 6. Implementation lever

Relay is Go. [**indielib**](https://go.hacdias.com/indielib) is a Go library
that provides both server and client IndieAuth (and Micropub) tooling, so the
authorization/token/introspection/revocation plumbing does not need to be
written from scratch. The Relay-specific work is the invitation-to-member step
and the scope mapping.

## 7. Security and trade-offs

- **No password, no email.** The invitation remains the only credential;
  IndieAuth only *packages* it for standard clients. There is nothing new to
  phish or reset.
- **Revocation is immediate and cascading.** Revoke a member → their invitation
  and their tokens stop working, because both resolve the same key store.
- **Scopes are conservative.** A client gets exactly what the member approved;
  Relay maps scopes to the few things a member can actually do.
- **PKCE + exact redirect URI.** Standard OAuth protections apply.
- **One deliberate deviation:** IndieAuth normally proves identity via a
  resolvable domain and `rel=me`; Relay proves it via the invitation. This is
  within the spec's "out of scope" authentication step, but we document it so
  nobody assumes domain-verification semantics.

## 8. What we do *not* do

- **We do not become a general-purpose OAuth provider.** The node issues tokens
  only for its own members and its own Micropub endpoint.
- **We do not replace invitations.** IndieAuth is an access layer over the
  invitation, not a second account system.
- **We do not implement `update`/`delete` post actions just to satisfy a scope.**
  Scopes follow what Relay does; Relay does not grow features to match scopes.

The *member-facing* side of this — which credential a member actually uses to
sign in — is in [`accounts.md`](accounts.md): the invitation stays the core,
and passkeys are the "normal" convenience layered on top of IndieAuth.

## 9. Phased plan

1. **Phase A — identity + discovery.** Emit `indieauth-metadata` and `micropub`
   links on profile pages; publish the metadata endpoint (static JSON).
2. **Phase B — server endpoints.** Authorization + token endpoints backed by
   the invitation login; PKCE; short-lived tokens; introspection.
3. **Phase C — Micropub integration.** The Micropub endpoint (from the
   syndication plan) validates IndieAuth tokens instead of a raw invite token.
   This is the moment standard clients start working.
4. **Phase D — client role (optional).** Members can sign in with their own
   domain; host can authenticate the node to third-party IndieWeb services.

Phases A–C together are the "Micropub endpoint" phase of
[`indieweb-syndication.md`](indieweb-syndication.md), made real by standard
auth.

## 10. Open questions

| # | Question | Stance |
|---|---|---|
| 1 | Does every member need to be an IndieAuth identity? | Yes, automatically — every member has a profile URL; the metadata link makes it discoverable. |
| 2 | Should the *host* also be an IndieAuth identity? | Later; the host console stays on the admin token for now. |
| 3 | Are tokens hashed at rest? | Yes, mirroring invite tokens. |
| 4 | Which scopes ship first? | `create`, `media`, `profile`, `read`. `update`/`delete` wait for real post-edit/delete. |
| 5 | Do we support the legacy `authorization_endpoint`/`token_endpoint` links? | Emit both for compatibility with older clients; it is two extra `<link>` tags. |

## 11. Acceptance criteria

1. A standard IndieWeb client (e.g. Quill) signs in with a member's profile URL
   and completes the flow without a Relay-specific step.
2. The member approves `create` and the client posts through Micropub; the post
   appears in the feed attributed to that member.
3. A revoked invitation causes both new sign-ins and existing tokens to fail.
4. A client requesting `delete` is refused (scope unsupported).
5. No password, email, or new credential is introduced anywhere in the flow.

---

## Sources

- [IndieAuth — IndieWeb](https://indieweb.org/IndieAuth)
- [IndieAuth Specification](https://indieauth.spec.indieweb.org/)
- [IndieAuth — W3C](https://www.w3.org/TR/indieauth/)
- [indielib — Go IndieAuth/Micropub toolkit](https://go.hacdias.com/indielib)

*This document changes no code. It is the source note for a future IndieAuth
integration, and it supersedes the "we skip IndieAuth" note in
[`indieweb-syndication.md`](indieweb-syndication.md).*
