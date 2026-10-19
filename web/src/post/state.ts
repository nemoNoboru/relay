import type { Comment, Post } from "../member/schema";

export type PostPhase = "loading" | "ready" | "signedOut" | "error";

export interface PostState {
  phase: PostPhase;
  post?: Post;
  comments?: ReadonlyArray<Comment>;
  error?: string;
  commentError?: string;
  notice?: string;
  busy?: boolean;
}

export const initialPost = (): PostState => ({ phase: "loading" });
