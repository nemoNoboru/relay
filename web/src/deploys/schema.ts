/**
 * `GET /api/site/deploy` — draft wire contract for the site page.
 *
 * Vocabulary from REVIEW.md §4 (control surface: manifest-first deploys,
 * immutable versions, rollback = re-pointing). Finalized with the control
 * API spec.
 */
import { Schema } from "effect";

export const Deploy = Schema.Struct({
  /** Deploy/version identifier. */
  id: Schema.String,
  /** The site this deploy belongs to. */
  site: Schema.String,
  /** Human-readable version label, e.g. the manifest's version. */
  version: Schema.String,
  /** RFC 3339 timestamp of the deploy. Rendered verbatim by the UI. */
  deployedAt: Schema.String,
  /** `live` is served now; `previous` versions are rollback targets. */
  status: Schema.Union([Schema.Literal("live"), Schema.Literal("previous")]),
});

export const Deploys = Schema.Struct({
  /** The version currently being served, if any. */
  current: Schema.optional(Deploy),
  /** All known versions, newest first. */
  versions: Schema.Array(Deploy),
});

export type Deploy = (typeof Deploy)["Type"];
export type Deploys = (typeof Deploys)["Type"];