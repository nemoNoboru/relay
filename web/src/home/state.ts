import type { Me, Post } from "../member/schema";

export type HomePhase = "loading" | "signedOut" | "ready" | "error";

/** The member welcome page's state. */
export interface HomeState {
  phase: HomePhase;
  me?: Me;
  /** The chronological feed, newest first. */
  posts?: ReadonlyArray<Post>;
  /** Cursor for the next page, absent at the end. */
  next?: string;
  loadingMore?: boolean;
  /** Why the page could not load (or why the invite was rejected). */
  error?: string;
  /** Why the feed could not load while the member is still signed in. */
  feedError?: string;
  /** Why the last profile save failed. */
  actionError?: string;
  /** A short success line after saving the profile. */
  notice?: string;
  busy?: boolean;
}

export const initialHome = (): HomeState => ({ phase: "loading" });
