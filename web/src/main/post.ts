/** Post wrapper: the site name comes from /_relay/posts/<site>. */
import { Effect } from "effect";
import { mount } from "svelte";
import { writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { initialPost } from "../post/state";
import { addComment, deleteComment, fetchComments, fetchPost } from "../member/ops";
import PostApp from "../ui/PostPage.svelte";

const root = document.getElementById("app");
if (root === null) throw new Error("post.html is missing #app");

const store = writable(initialPost());
const set = (patch: Partial<ReturnType<typeof initialPost>>): void => store.update((s) => ({ ...s, ...patch }));

const site = decodeURIComponent(location.pathname.replace(/^.*\/posts\//, "").replace(/\/$/, ""));
const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

const loadComments = async (): Promise<void> => {
  try {
    const reply = await run(fetchComments(site));
    set({ comments: reply.comments, commentError: undefined });
  } catch (err) {
    const api = err as ApiError;
    if (api.status === 401) {
      set({ phase: "signedOut" });
    } else {
      set({ commentError: api.message });
    }
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

mount(PostApp, {
  target: root,
  props: {
    postState: store,
    onComment: (body: string) => void comment(body),
    onDeleteComment: (id: string) => void removeComment(id),
  },
});

if (site === "") {
  set({ phase: "error", error: "No post in this address." });
} else {
  void run(fetchPost(site)).then(
    (reply) => {
      set({ phase: "ready", post: reply.post, error: undefined });
      void loadComments();
    },
    (err: ApiError) => set({
      phase: err.status === 401 ? "signedOut" : "error",
      error: err.status === 401 ? undefined : err.status === 404 ? "No such post on this node." : err.message,
    }),
  );
}
