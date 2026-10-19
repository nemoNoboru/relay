/**
 * Member welcome page — the only wiring point for home.html.
 *
 * On load it either exchanges the invite fragment (`#key=…`) for the member
 * cookie, or asks the node for the current member. The Svelte view renders
 * the resulting state; it never fetches.
 */
import { Effect } from "effect";
import { mount } from "svelte";
import { get, writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { initialHome } from "../home/state";
import type { HomeState } from "../home/state";
import { addComment as addCommentOp, deleteComment as deleteCommentOp, establishSession, fetchFeed, fetchMe, saveProfile, signOut } from "../member/ops";
import HomeApp from "../ui/Home.svelte";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("home.html is missing #app");
}

const store = writable<HomeState>(initialHome());
const set = (patch: Partial<HomeState>): void => store.update((s) => ({ ...s, ...patch }));

const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

/**
 * Read a token from a fragment, a full invite link, or a pasted key. The
 * fragment is never sent to the server; this is the only place it is read.
 */
export const readToken = (raw: string): string => {
  const s = raw.trim();
  if (s === "") return "";
  if (s.startsWith("key=") || s.startsWith("token=")) {
    return new URLSearchParams(s).get("key") ?? new URLSearchParams(s).get("token") ?? "";
  }
  try {
    const u = new URL(s, location.origin);
    if (u.hash !== "") {
      const p = new URLSearchParams(u.hash.replace(/^#/, ""));
      return p.get("key") ?? p.get("token") ?? "";
    }
    // A plain token resolves as a relative path; return the original text.
    if (u.pathname === "/" || u.pathname === "") return "";
  } catch {
    return s;
  }
  return s;
};

/** Load the first feed page for a signed-in member without failing the page. */
const loadFeed = async (): Promise<{ posts: HomeState["posts"]; next?: string; feedError?: string }> => {
  try {
    const feed = await run(fetchFeed());
    return { posts: feed.posts, next: feed.next };
  } catch (err) {
    return { posts: [], feedError: (err as ApiError).message };
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
      posts: [...(s.posts ?? []), ...feed.posts],
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
      posts: s.posts?.map((p) =>
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
      posts: s.posts?.map((p) =>
        p.site === site ? { ...p, comments: p.comments.filter((c) => c.id !== id) } : p,
      ),
    }));
  } catch (err) {
    throw new Error((err as ApiError).message);
  }
};

const load = async (): Promise<void> => {
  set({ phase: "loading", error: undefined, actionError: undefined });
  try {
    const me = await run(fetchMe());
    const feed = await loadFeed();
    set({ phase: "ready", me, ...feed, busy: false, error: undefined });
  } catch (err) {
    const api = err as ApiError;
    if (api.status === 401) {
      set({ phase: "signedOut", busy: false, error: undefined });
    } else {
      set({ phase: "error", busy: false, error: api.message });
    }
  }
};

const signIn = async (raw: string): Promise<void> => {
  const token = readToken(raw);
  if (token === "") {
    set({ phase: "signedOut", busy: false, error: "Paste an invitation key or link." });
    return;
  }
  set({ phase: "signedOut", busy: true, error: undefined });
  try {
    const me = await run(establishSession(token));
    const feed = await loadFeed();
    set({ phase: "ready", me, ...feed, busy: false, error: undefined });
  } catch (err) {
    const api = err as ApiError;
    set({
      phase: "signedOut",
      busy: false,
      error:
        api.status === 401
          ? "That invitation is not valid or has been revoked. Ask the host for a new link."
          : api.message,
    });
  }
};

const save = async (displayName: string, bio: string): Promise<void> => {
  set({ busy: true, actionError: undefined, notice: undefined });
  try {
    const me = await run(saveProfile(displayName, bio));
    set({ phase: "ready", me, busy: false, notice: "Profile saved." });
  } catch (err) {
    set({ busy: false, actionError: (err as ApiError).message });
  }
};

const leave = async (): Promise<void> => {
  set({ busy: true, actionError: undefined });
  try {
    await run(signOut());
  } catch {
    // The browser cookie is cleared by the response; a network failure is
    // not worth blocking the user on. The next /me call will settle it.
  }
  store.set({ phase: "signedOut", busy: false });
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

mount(HomeApp, {
  target: root,
  props: {
    home: store,
    onSignIn: (token: string) => void signIn(token),
    onSignOut: () => void leave(),
    onSave: (displayName: string, bio: string) => void save(displayName, bio),
    onLoadMore: () => void loadMore(),
    onComment: comment,
    onDeleteComment: removeComment,
  },
});

void boot();
