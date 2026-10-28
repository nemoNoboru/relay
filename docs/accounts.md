# Relay — Accounts (member sign-in) Research

> **Note:** written against the old Go node, which has been removed. The implementation is now [`backend/`](../backend/README.md) (FastAPI) and [`mobile/`](../mobile/README.md) (Expo); endpoint, CLI and file names below describe the old node and are kept as design context.

**Status:** proposal, not implemented. This document works out how members sign
in, and whether IndieAuth can give Relay a "more normal" account system. It
changes nothing except itself. It builds on [`indieauth.md`](indieauth.md),
which should be read first.

---

## 1. TL;DR

IndieAuth is a **sign-in protocol**, not an account system. It standardises the
"app gets a token for a user" flow, but it does not provide passwords, passkeys,
or email — those are a separate decision that Relay still has to make.

The recommendation:

- **Keep the invitation as the identity core.** It is what powers onboarding,
  "any member can invite", and cascade revoke.
- **Add passkeys (WebAuthn)** as the *returning* sign-in, so the day-to-day
  experience feels normal (Face ID / Touch ID, one tap) without introducing
  passwords or a reset flow.
- **Expose both through IndieAuth**, so Relay's authorization endpoint becomes
  one normal sign-in page that accepts a passkey *or* an invitation link (and,
  optionally, "sign in with your own domain" for indieweb-native members) —
  all resolving to the same member URL.

## 2. The two questions being conflated

"More normal account system" is really two questions:

1. **What credential proves a member is who they say they are?** — invitation
   link, password, passkey, their own domain, or some combination.
2. **What protocol hands out tokens once they have proved it?** — IndieAuth.

IndieAuth answers only question 2, and does it well. Question 1 is a product
decision. The IndieAuth spec is explicit that the actual login is out of scope:

> "It is up to the authorization endpoint how to authenticate the user. This
> step is out of scope of OAuth 2.0 … Some authorization servers use typical
> username/password authentication, and others use alternative forms of
> authentication such as RelMeAuth, or delegate to other identity providers."

So IndieAuth gives Relay the *frame* of a normal login page and standard token
machinery; Relay supplies the credential inside it.

## 3. Today

Relay's account system is the invitation link:

- Opening an invite link sets the member cookie; the token is also the bearer
  credential for the CLI and assistants.
- There is no password, email, or recovery. A lost link is a host rotation.

This is elegant but esoteric: the first thing a returning member sees is a
paste-a-link form, and mainstream users expect *something* they can type or
biometrically confirm.

## 4. The options

| Credential | Feels normal? | Cost | Trade-off |
|---|---|---|---|
| Invitation link only (today) | ✗ esoteric | zero | already built; the thing we want to soften |
| **Passkey (WebAuthn)** | ✓ Face ID / Touch ID, one tap | small | store one public key + sign counter per member; no reset needed; phishing-resistant |
| Email + password | ✓✓ most mainstream | high | email verification, password hashing, reset flow — the exact recovery machinery Relay deliberately avoided |
| Sign in with your own domain (RelMeAuth) | ✓ for indieweb folks | small | most members don't have a domain; a niche option, not the default |
| Magic link by email | ✓ familiar | medium | needs email delivery; still a link, just sent to an inbox |

## 5. Recommendation: invitation + passkeys, composed with IndieAuth

Keep the invitation as the **onboarding, recovery, and revocation** primitive,
and add passkeys as the **returning** credential.

```
onboarding:  invite link (unchanged) ──► member cookie ──► "add a passkey"
returning:   tap "Sign in" ──► Face ID / Touch ID ──► member cookie
recovery:    lose the passkey ──► open the invite link again ──► re-add
revoke:      host revokes the member ──► passkey and invitation both die
```

Why passkeys rather than email+password:

- **No password to reset.** A passkey is a key pair on the device; losing it is
  recovered by the invitation, which already exists.
- **No email.** Relay stays "no email, no account recovery" at its core; email
  becomes an optional notification/recovery field later, never the login.
- **Fits the ethos.** "Own your identity" without a central user table — the
  key lives on the member's device, the node stores only the public half.
- **Small surface.** WebAuthn registration + assertion is a well-trodden,
  library-supported flow; the node stores a credential id, public key, and
  sign counter per member.

## 6. How IndieAuth composes with this

Relay's IndieAuth **authorization endpoint** is the single sign-in page:

- It accepts a **passkey** (normal path) or an **invitation link** (onboarding /
  recovery path) or, optionally, **RelMeAuth domain** sign-in.
- All three resolve to the same member and the same `me` URL
  (`/_relay/people/<id>`).
- The consent screen then asks the member to approve the client's scopes
  (`create`, `media`, …) exactly as in [`indieauth.md`](indieauth.md).

This means one login experience serves three audiences:

1. **Returning members** — passkey, one tap.
2. **New members** — the invite link they were sent.
3. **Third-party indieweb apps** — "sign in with your Relay profile URL" → the
   same page → passkey or invite.

The node's own member pages (feed, profile) use the same sign-in page, so there
is one normal login everywhere, not a bespoke one per surface.

## 7. Data-model changes

`deploy-keys.json` already holds members (id, display name, bio, token hash,
revoked, invitedBy). Add, all optional:

- `passkeys`: array of WebAuthn credentials — `{credentialId, publicKey,
  signCount, createdAt, name}`. A member may have several (phone + laptop).
- `email`: optional, for future notifications/recovery only; **not** a login.
- `relMe`: optional array of URLs the member controls, for domain sign-in.

No migration: existing members simply have empty lists and keep signing in by
invitation. The key store's revocation already cascades; passkeys and tokens
resolve through the same member record, so revoking a member kills all three.

## 8. What we deliberately do *not* do

- **No email+password by default.** It drags in verification, resets, and a
  support surface Relay exists to avoid. It can stay a later opt-in if a host
  demands it.
- **No central user table.** Members remain rows in `deploy-keys.json`, keyed
  by invitation; passkeys hang off those rows.
- **IndieAuth does not replace the invitation.** It packages the invitation (or
  passkey) for standard clients.

## 9. Open questions

| # | Question | Stance |
|---|---|---|
| 1 | Is a passkey required, or optional per member? | Optional; invitation remains the universal path. |
| 2 | Should the host also get passkeys for the console? | Separate; the host console keeps the admin token for now. |
| 3 | Email for recovery? | Optional later; never a login. |
| 4 | Does "sign in with your domain" risk impersonation? | No — RelMeAuth proves control of the domain; the host maps it to a member explicitly. |
| 5 | WebAuthn library in Go? | `github.com/go-webauthn/webauthn` is the common choice; evaluate against the existing store. |

## 10. Acceptance criteria

1. A returning member signs in with a passkey in one tap, no link pasted.
2. A new member opens an invite link and is offered "add a passkey"; declining
   changes nothing.
3. A member can hold passkeys on two devices and remove one without losing
   access.
4. Revoking a member invalidates their passkeys, tokens, and invitation.
5. The same sign-in page works for the node's own pages and for a third-party
   IndieAuth client.
6. No password or email is required anywhere in the flow.

---

*This document changes no code. It is the source note for a future account-system
change, composed with [`indieauth.md`](indieauth.md) and
[`indieweb-syndication.md`](indieweb-syndication.md).*
