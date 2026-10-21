/** People page: /_relay/people is the directory, /_relay/people/<id> a profile. */
import { Effect } from "effect";
import { mount } from "svelte";
import { get, writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { readToken } from "../invite";
import { addComment as addCommentOp, deleteComment as deleteCommentOp, deletePost as deletePostOp, establishSession, fetchMe, fetchMembers, fetchProfile, saveProfile, signOut } from "../member/ops";
import { initialPeople } from "../people/state";
import type { PeopleState } from "../people/state";
import PeopleApp from "../ui/People.svelte";

const root = document.getElementById("app");
if (root === null) throw new Error("people.html is missing #app");

const store = writable<PeopleState>(initialPeople());
const set = (patch: Partial<PeopleState>): void => store.update((s) => ({ ...s, ...patch }));

const path = location.pathname.replace(/\/+$/, "");
const memberID = path === "/_relay/people" ? "" : decodeURIComponent(path.replace(/^.*\/people\//, ""));
const isDirectory = memberID === "";

const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

const loadMe = async (): Promise<PeopleState["me"]> => {
  try {
    return await run(fetchMe());
  } catch (err) {
    if ((err as ApiError).status === 401) return undefined;
    throw err;
  }
};

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

const removePost = async (site: string): Promise<void> => {
  try {
    await run(deletePostOp(site));
    store.update((s) => ({
      ...s,
      profile: s.profile
        ? { ...s.profile, posts: s.profile.posts.filter((p) => p.site !== site) }
        : s.profile,
    }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

const save = async (displayName: string, bio: string): Promise<void> => {
  set({ busy: true, actionError: undefined, notice: undefined });
  try {
    const me = await run(saveProfile(displayName, bio));
    const profile = await run(fetchProfile(memberID));
    set({ phase: "ready", me, profile, next: profile.next, busy: false, notice: "Profile saved." });
  } catch (err) {
    set({ busy: false, actionError: (err as ApiError).message });
  }
};

const signIn = async (raw: string): Promise<void> => {
  const token = readToken(raw);
  if (token === "") {
    set({ signInError: "Paste an invitation key or link." });
    return;
  }
  set({ busy: true, signInError: undefined });
  try {
    const me = await run(establishSession(token));
    set({ me, busy: false });
  } catch (err) {
    const api = err as ApiError;
    set({
      busy: false,
      signInError: api.status === 401 ? "That invitation is not valid or has been revoked." : api.message,
    });
  }
};

const leave = async (): Promise<void> => {
  set({ busy: true });
  try {
    await run(signOut());
  } catch {
    // The cookie is cleared by the response; a network failure settles on
    // the next /me call.
  }
  set({ me: undefined, busy: false, signInError: undefined });
};

mount(PeopleApp, {
  target: root,
  props: {
    people: store,
    onSignIn: (token: string) => void signIn(token),
    onSignOut: () => void leave(),
    onLoadMore: () => void loadMore(),
    onComment: comment,
    onDeleteComment: removeComment,
    onDeletePost: removePost,
    onSave: (displayName: string, bio: string) => void save(displayName, bio),
  },
});

const boot = async (): Promise<void> => {
  set({ phase: "loading" });
  try {
    const me = await loadMe();
    if (isDirectory) {
      const reply = await run(fetchMembers());
      set({ phase: "ready", me, members: reply.members, error: undefined });
    } else {
      const profile = await run(fetchProfile(memberID));
      set({ phase: "ready", me, profile, next: profile.next, error: undefined });
    }
  } catch (err) {
    const api = err as ApiError;
    set({
      phase: "error",
      error: api.status === 404 ? (isDirectory ? "No members yet." : "No such member on this node.") : api.message,
    });
  }
};

void boot();
