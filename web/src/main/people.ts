/** Profile page: the member id comes from /_relay/people/<id>. */
import { Effect } from "effect";
import { mount } from "svelte";
import { get, writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { initialPeople } from "../people/state";
import { addComment as addCommentOp, deleteComment as deleteCommentOp, fetchProfile } from "../member/ops";
import PeopleApp from "../ui/People.svelte";

const root = document.getElementById("app");
if (root === null) throw new Error("people.html is missing #app");

const store = writable(initialPeople());
const set = (patch: Partial<ReturnType<typeof initialPeople>>): void => store.update((s) => ({ ...s, ...patch }));

const memberID = decodeURIComponent(location.pathname.replace(/^.*\/people\//, "").replace(/\/$/, ""));
const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

const loadMore = async (): Promise<void> => {
  const state = get(store);
  if (!state.next || state.loadingMore || !state.profile) return;
  set({ loadingMore: true });
  try {
    const next = await run(fetchProfile(memberID, state.next));
    store.update((s) => ({
      ...s,
      profile: s.profile ? { ...s.profile, posts: [...s.profile.posts, ...next.posts] } : s.profile,
      next: next.next,
      loadingMore: false,
    }));
  } catch (err) {
    set({ loadingMore: false, error: (err as ApiError).message });
  }
};

const comment = async (site: string, body: string): Promise<void> => {
  try {
    const reply = await run(addCommentOp(site, body));
    store.update((s) => ({
      ...s,
      profile: s.profile
        ? {
            ...s.profile,
            posts: s.profile.posts.map((p) =>
              p.site === site ? { ...p, comments: [...p.comments, reply.comment] } : p,
            ),
          }
        : s.profile,
    }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

const removeComment = async (site: string, id: string): Promise<void> => {
  try {
    await run(deleteCommentOp(id));
    store.update((s) => ({
      ...s,
      profile: s.profile
        ? {
            ...s.profile,
            posts: s.profile.posts.map((p) =>
              p.site === site ? { ...p, comments: p.comments.filter((c) => c.id !== id) } : p,
            ),
          }
        : s.profile,
    }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

mount(PeopleApp, {
  target: root,
  props: {
    people: store,
    onLoadMore: () => void loadMore(),
    onComment: comment,
    onDeleteComment: removeComment,
  },
});

if (memberID === "") {
  set({ phase: "error", error: "No member in this address." });
} else {
  void run(fetchProfile(memberID)).then(
    (profile) => set({ phase: "ready", profile, next: profile.next, error: undefined }),
    (err: ApiError) => set({
      phase: err.status === 401 ? "signedOut" : "error",
      error: err.status === 401 ? undefined : err.status === 404 ? "No such member on this node." : err.message,
    }),
  );
}
