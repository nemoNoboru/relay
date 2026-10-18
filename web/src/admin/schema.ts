import { Schema } from "effect";

/**
 * Wire contract for the operator admin surface (`/_relay/v1/admin`).
 *
 * The admin token is sent as a bearer token on every call. A deploy key is
 * per-user; its plaintext token exists only in the create/rotate response.
 */
export const DeployKey = Schema.Struct({
  id: Schema.String,
  user: Schema.String,
  /** The plaintext token — returned only to the admin-authenticated console. */
  token: Schema.String,
  /** RFC 3339 timestamp, rendered verbatim. */
  created: Schema.String,
  revoked: Schema.Boolean,
});

export const SiteOwner = Schema.Struct({
  site: Schema.String,
  user: Schema.String,
});

export const AdminOverview = Schema.Struct({
  keys: Schema.Array(DeployKey),
  sites: Schema.Array(SiteOwner),
  /** The node is still on the default admin token and must be given a new one. */
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
export type AdminOverview = (typeof AdminOverview)["Type"];
export type KeyCreated = (typeof KeyCreated)["Type"];
export type KeyUpdated = (typeof KeyUpdated)["Type"];
export type TokenChanged = (typeof TokenChanged)["Type"];
