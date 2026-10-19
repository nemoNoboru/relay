import type { Profile } from "../member/schema";

export type PeoplePhase = "loading" | "ready" | "signedOut" | "error";

export interface PeopleState {
  phase: PeoplePhase;
  profile?: Profile;
  /** Cursor for the next page, absent at the end. */
  next?: string;
  loadingMore?: boolean;
  error?: string;
}

export const initialPeople = (): PeopleState => ({ phase: "loading" });
