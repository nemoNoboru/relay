/**
 * The state machine every live (polled) page follows
 * (references/relay-ui.md):
 *
 *   loading → ready
 *   ready   → stale    a poll failed; last data is still shown
 *   stale   → error    the last data is older than `staleAfterMs`
 *   *       → ready    any successful poll
 *
 * Pure: `(state, event) → state`. The edge (the live-page driver in
 * `src/live.ts`) feeds events, holds the current state, and publishes the
 * result. Nothing here touches the DOM, clocks, or the network — so it is
 * trivially testable with plain values.
 */

/** Default: after a minute without fresh data, "stale" becomes "error". */
export const STALE_AFTER_MS = 60_000;

export interface LiveState<T> {
  /** Where the page is right now. */
  readonly phase: "loading" | "ready" | "stale" | "error";
  /** Last successfully decoded data. Kept on stale and error (the views just don't render it in the error phase). */
  readonly data: T | undefined;
  /** Epoch ms of the last successful poll. */
  readonly updatedAt: number | undefined;
  /** Human-readable message of the most recent failure, if any. */
  readonly lastError: string | undefined;
}

export const initial = <T>(): LiveState<T> => ({
  phase: "loading",
  data: undefined,
  updatedAt: undefined,
  lastError: undefined,
});

/** One poll attempt, as observed by the fold. `at` is the epoch ms the edge saw it. */
export type LiveEvent<T> =
  | { readonly _tag: "ok"; readonly data: T; readonly at: number }
  | { readonly _tag: "fail"; readonly message: string; readonly at: number };

export function fold<T>(
  state: LiveState<T>,
  event: LiveEvent<T>,
  staleAfterMs: number = STALE_AFTER_MS,
): LiveState<T> {
  if (event._tag === "ok") {
    return {
      phase: "ready",
      data: event.data,
      updatedAt: event.at,
      lastError: undefined,
    };
  }

  if (state.data === undefined) {
    // We never had data: there is nothing to keep showing.
    return { ...state, phase: "error", lastError: event.message };
  }

  const age = event.at - (state.updatedAt ?? event.at);
  return {
    ...state,
    phase: age > staleAfterMs ? "error" : "stale",
    lastError: event.message,
  };
}

/**
 * Human-readable elapsed time. Pure, so views and tests share it.
 * (Behavior pinned by test/fold.test.ts: `undefined → "never"`, future clamps.)
 */
export function ago(ts: number | undefined, now: number): string {
  if (ts === undefined) return "never";
  const s = Math.max(0, Math.round((now - ts) / 1000));
  if (s < 60) return `${s}s`;
  const m = Math.round(s / 60);
  if (m < 60) return `${m}m`;
  const h = Math.round(m / 60);
  return `${h}h`;
}