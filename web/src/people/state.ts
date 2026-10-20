import type { Me, Member, Profile } from "../member/schema";

export type PeoplePhase = "loading" | "ready" | "error";

/** The people page serves two shapes: the directory (members) or one profile. */
export interface PeopleState {
  phase: PeoplePhase;
  me?: Me;
  /** Directory mode: all live members. */
  members?: ReadonlyArray<Member>;
  /** Profile mode: one member and their posts. */
  profile?: Profile;
  /** Cursor for the next profile page, absent at the end. */
  next?: string;
  loadingMore?: boolean;
  error?: string;
  /** Why the last inline sign-in failed. */
  signInError?: string;
  /** Why the last profile save failed. */
  actionError?: string;
  /** A short success line after saving the profile. */
  notice?: string;
  busy?: boolean;
}

export const initialPeople = (): PeopleState => ({ phase: "loading" });
