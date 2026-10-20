import type { Me, Member, Post } from "../member/schema";

export type FeedPhase = "loading" | "ready" | "error";

/** The public feed's state. `me` is undefined for a guest. */
export interface FeedState {
  phase: FeedPhase;
  me?: Me;
  /** The people strip: all live members. */
  members?: ReadonlyArray<Member>;
  /** The chronological feed, newest first. */
  posts: ReadonlyArray<Post>;
  /** Cursor for the next page, absent at the end. */
  next?: string;
  loadingMore?: boolean;
  /** Why the page could not load. */
  error?: string;
  /** Why the feed could not load. */
  feedError?: string;
  /** Why the last inline sign-in failed. */
  signInError?: string;
  busy?: boolean;
}

export const initialFeed = (): FeedState => ({ phase: "loading", posts: [] });
