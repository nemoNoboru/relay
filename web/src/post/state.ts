import type { Comment, Me, Post } from "../member/schema";

export type PostPhase = "loading" | "ready" | "error";

export interface PostState {
  phase: PostPhase;
  me?: Me;
  post?: Post;
  comments?: ReadonlyArray<Comment>;
  error?: string;
  commentError?: string;
  notice?: string;
  busy?: boolean;
  /** Why the last inline sign-in failed. */
  signInError?: string;
}

export const initialPost = (): PostState => ({ phase: "loading" });
