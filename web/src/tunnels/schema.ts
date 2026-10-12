/**
 * `GET /api/tunnels` — draft wire contract for the node page.
 *
 * REVIEW.md §7.5: the tunnel protocol spec is still open. This is the UI's
 * draft of the shape, in Relay's domain vocabulary (tunnel, friend, prefix),
 * and will be finalized when the spec lands.
 */
import { Schema } from "effect";

export const Tunnel = Schema.Struct({
  /** Stable identifier of the tunnel. */
  id: Schema.String,
  /** The friend node this tunnel connects through. */
  peer: Schema.String,
  /** Path prefixes routed through this tunnel. */
  prefixes: Schema.Array(Schema.String),
  /** Whether the tunnel is currently carrying traffic. */
  online: Schema.Boolean,
});

export const Tunnels = Schema.Struct({
  tunnels: Schema.Array(Tunnel),
});

export type Tunnel = (typeof Tunnel)["Type"];
export type Tunnels = (typeof Tunnels)["Type"];