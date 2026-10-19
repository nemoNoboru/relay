import type { AdminOverview } from "./schema";

/**
 * The host console's UI state. The page is interactive: the phase tracks
 * the admin-password session, and actions update `notice`/`newToken`/
 * `actionError` without leaving the ready phase.
 */
export type AdminPhase = "locked" | "loading" | "ready" | "error";

export interface AdminState {
  /** The admin token in use (kept in sessionStorage between reloads). */
  token: string;
  phase: AdminPhase;
  overview?: AdminOverview;
  /** Why the token was rejected (sign-in failure). */
  error?: string;
  /** Why the last action (create/revoke/rotate) failed. */
  actionError?: string;
  /** An action is in flight. */
  busy: boolean;
  /** Success line for the last action. */
  notice?: string;
  /** The one-time token from create/rotate, shown until dismissed. */
  newToken?: string;
}

export const initialAdmin = (token: string): AdminState => ({
  token,
  phase: token === "" ? "locked" : "loading",
  busy: false,
});
