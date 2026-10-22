import { Schema } from "effect";

/**
 * Wire contract for the operator admin surface (`/_relay/v1/admin`).
 *
 * The admin password is sent as a bearer token on every call. An invitation
 * key is per-member; its plaintext token is returned so the host can copy the
 * invite link.
 */
export const DeployKey = Schema.Struct({
  id: Schema.String,
  user: Schema.String,
  /** The plaintext token — returned only to the admin-authenticated console. */
  token: Schema.String,
  /** RFC 3339 timestamp, rendered verbatim. */
  created: Schema.String,
  revoked: Schema.Boolean,
  /** Login username, when the member has set one. */
  username: Schema.optional(Schema.String),
});

export const SiteOwner = Schema.Struct({
  site: Schema.String,
  user: Schema.String,
});

/** One recent comment in the host overview, already sanitized by the node. */
export const AdminComment = Schema.Struct({
  id: Schema.String,
  post: Schema.String,
  member: Schema.String,
  author: Schema.Struct({
    id: Schema.String,
    displayName: Schema.String,
    revoked: Schema.optional(Schema.Boolean),
  }),
  body: Schema.String,
  html: Schema.String,
  createdAt: Schema.String,
});

export const AdminOverview = Schema.Struct({
  keys: Schema.Array(DeployKey),
  sites: Schema.Array(SiteOwner),
  comments: Schema.Array(AdminComment),
  /** The node is still on the default admin password and must be given a new one. */
  mustChangeToken: Schema.Boolean,
});

/** POST /_relay/v1/admin/keys and …/rotate — the one-time token is here. */
export const KeyCreated = Schema.Struct({
  key: DeployKey,
  token: Schema.String,
});

/** POST /_relay/v1/admin/keys/{id}/revoke. */
export const KeyUpdated = Schema.Struct({
  key: DeployKey,
});

/** POST /_relay/v1/admin/token — the admin token has been replaced. */
export const TokenChanged = Schema.Struct({
  changed: Schema.Boolean,
});

export type DeployKey = (typeof DeployKey)["Type"];
export type SiteOwner = (typeof SiteOwner)["Type"];
export type AdminComment = (typeof AdminComment)["Type"];
export type AdminOverview = (typeof AdminOverview)["Type"];
export type KeyCreated = (typeof KeyCreated)["Type"];
export type KeyUpdated = (typeof KeyUpdated)["Type"];
export type TokenChanged = (typeof TokenChanged)["Type"];
