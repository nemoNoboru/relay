/**
 * The public feed — the only wiring point for feed.html (served at /).
 *
 * The feed itself is public; /me is resolved opportunistically so the chrome
 * can show the member's identity when signed in. Invite links carry
 * `#key=…` and are exchanged for the cookie on load.
 */
import { Effect } from "effect";
import { mount } from "svelte";
import { get, writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { initialFeed } from "../feed/state";
import type { FeedState } from "../feed/state";
import { readToken } from "../invite";
import { addComment as addCommentOp, deleteComment as deleteCommentOp, deletePost as deletePostOp, establishSession, fetchFeed, fetchMe, fetchMembers, signOut } from "../member/ops";
import FeedApp from "../ui/Feed.svelte";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("feed.html is missing #app");
}

const store = writable<FeedState>(initialFeed());
const set = (patch: Partial<FeedState>): void => store.update((s) => ({ ...s, ...patch }));

const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

/** Resolve the viewer; a 401 simply means "guest". */
const loadMe = async (): Promise<FeedState["me"]> => {
  try {
    return await run(fetchMe());
  } catch (err) {
    if ((err as ApiError).status === 401) return undefined;
    throw err;
  }
};

/** Load the first feed page without failing the page. */
const loadFeed = async (): Promise<{ posts: FeedState["posts"]; next?: string; feedError?: string }> => {
  try {
    const feed = await run(fetchFeed());
    return { posts: feed.posts, next: feed.next };
  } catch (err) {
    return { posts: [], feedError: (err as ApiError).message };
  }
};

/** Load the people strip; a failure is a quiet empty strip, not a page error. */
const loadMembers = async (): Promise<FeedState["members"]> => {
  try {
    return (await run(fetchMembers())).members;
  } catch {
    return undefined;
  }
};

const loadMore = async (): Promise<void> => {
  const state = get(store);
  if (!state.next || state.loadingMore) return;
  set({ loadingMore: true });
  try {
    const feed = await run(fetchFeed(state.next));
    store.update((s) => ({
      ...s,
      posts: [...s.posts, ...feed.posts],
      next: feed.next,
      loadingMore: false,
      feedError: undefined,
    }));
  } catch (err) {
    set({ loadingMore: false, feedError: (err as ApiError).message });
  }
};

const comment = async (site: string, body: string): Promise<void> => {
  try {
    const reply = await run(addCommentOp(site, body));
    store.update((s) => ({
      ...s,
      posts: s.posts.map((p) =>
        p.site === site ? { ...p, comments: [...p.comments, reply.comment] } : p,
      ),
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
      posts: s.posts.map((p) =>
        p.site === site ? { ...p, comments: p.comments.filter((c) => c.id !== id) } : p,
      ),
    }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

const removePost = async (site: string): Promise<void> => {
  try {
    await run(deletePostOp(site));
    store.update((s) => ({ ...s, posts: s.posts.filter((p) => p.site !== site) }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

const load = async (): Promise<void> => {
  set({ phase: "loading", error: undefined, signInError: undefined });
  try {
    const me = await loadMe();
    const feed = await loadFeed();
    const members = await loadMembers();
    set({ phase: "ready", me, members, ...feed, busy: false, error: undefined });
  } catch (err) {
    set({ phase: "error", busy: false, error: (err as ApiError).message });
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
    const feed = await loadFeed();
    const members = await loadMembers();
    set({ phase: "ready", me, members, ...feed, busy: false });
  } catch (err) {
    const api = err as ApiError;
    set({
      busy: false,
      signInError:
        api.status === 401
          ? "That invitation is not valid or has been revoked. Ask the host for a new link."
          : api.message,
    });
  }
};

const leave = async (): Promise<void> => {
  set({ busy: true });
  try {
    await run(signOut());
  } catch {
    // The browser cookie is cleared by the response; a network failure is
    // not worth blocking the user on. The next /me call will settle it.
  }
  const me = await loadMe().catch(() => undefined);
  set({ me, busy: false, signInError: undefined });
};

const boot = async (): Promise<void> => {
  const invite = readToken(location.hash);
  if (invite !== "") {
    history.replaceState(null, "", location.pathname + location.search);
    await signIn(invite);
    return;
  }
  await load();
};

mount(FeedApp, {
  target: root,
  props: {
    feed: store,
    onSignIn: (token: string) => void signIn(token),
    onSignOut: () => void leave(),
    onLoadMore: () => void loadMore(),
    onComment: comment,
    onDeleteComment: removeComment,
    onDeletePost: removePost,
  },
});

void boot();
