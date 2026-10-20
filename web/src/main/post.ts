/** Post wrapper: the site name comes from /_relay/posts/<site>. Reads are public. */
import { Effect } from "effect";
import { mount } from "svelte";
import { writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { readToken } from "../invite";
import { addComment, deleteComment, establishSession, fetchComments, fetchMe, fetchPost, signOut } from "../member/ops";
import { initialPost } from "../post/state";
import type { PostState } from "../post/state";
import PostApp from "../ui/PostPage.svelte";

const root = document.getElementById("app");
if (root === null) throw new Error("post.html is missing #app");

const store = writable<PostState>(initialPost());
const set = (patch: Partial<PostState>): void => store.update((s) => ({ ...s, ...patch }));

const site = decodeURIComponent(location.pathname.replace(/^.*\/posts\//, "").replace(/\/$/, ""));
const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

const loadMe = async (): Promise<PostState["me"]> => {
  try {
    return await run(fetchMe());
  } catch (err) {
    if ((err as ApiError).status === 401) return undefined;
    throw err;
  }
};

const loadComments = async (): Promise<void> => {
  try {
    const reply = await run(fetchComments(site));
    set({ comments: reply.comments, commentError: undefined });
  } catch (err) {
    set({ commentError: (err as ApiError).message });
  }
};

const comment = async (body: string): Promise<void> => {
  set({ busy: true, commentError: undefined, notice: undefined });
  try {
    await run(addComment(site, body));
    set({ busy: false, notice: "Comment posted." });
    await loadComments();
  } catch (err) {
    const api = err as ApiError;
    set({
      busy: false,
      commentError: api.status === 429 ? "Wait a moment before commenting again." : api.message,
    });
  }
};

const removeComment = async (id: string): Promise<void> => {
  set({ busy: true, commentError: undefined, notice: undefined });
  try {
    await run(deleteComment(id));
    set({ busy: false, notice: "Comment deleted." });
    await loadComments();
  } catch (err) {
    set({ busy: false, commentError: (err as ApiError).message });
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
    await loadComments();
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
    // The cookie is cleared by the response; the next /me settles it.
  }
  set({ me: undefined, busy: false, signInError: undefined });
  await loadComments();
};

mount(PostApp, {
  target: root,
  props: {
    postState: store,
    onSignIn: (token: string) => void signIn(token),
    onSignOut: () => void leave(),
    onComment: (body: string) => void comment(body),
    onDeleteComment: (id: string) => void removeComment(id),
  },
});

const boot = async (): Promise<void> => {
  set({ phase: "loading" });
  if (site === "") {
    set({ phase: "error", error: "No post in this address." });
    return;
  }
  try {
    const me = await loadMe();
    const reply = await run(fetchPost(site));
    set({ phase: "ready", me, post: reply.post, error: undefined });
    await loadComments();
  } catch (err) {
    const api = err as ApiError;
    set({
      phase: "error",
      error: api.status === 404 ? "No such post on this node." : api.message,
    });
  }
};

void boot();
