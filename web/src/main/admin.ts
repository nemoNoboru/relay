/**
 * Host console — the only wiring point for this page.
 *
 * Holds the admin-password session (sessionStorage), runs the API effects
 * from `admin/ops.ts`, and folds results into a Svelte store that
 * Admin.svelte renders. No polling: actions and a manual refresh drive
 * re-renders.
 */
import { Effect } from "effect";
import { mount } from "svelte";
import { get, writable } from "svelte/store";
import type { Writable } from "svelte/store";

import { Client } from "../api/client";
import type { ApiError } from "../api/client";
import { changeToken, createKey, deleteComment, deletePost, fetchOverview, revokeKey, rotateKey } from "../admin/ops";
import { initialAdmin } from "../admin/state";
import type { AdminState } from "../admin/state";
import AdminApp from "../ui/Admin.svelte";

/** Where the admin password lives between reloads (never in localStorage/URL). */
const TOKEN_KEY = "relay.admin.token";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("admin.html is missing #app");
}

const store: Writable<AdminState> = writable(initialAdmin(sessionStorage.getItem(TOKEN_KEY) ?? ""));

const set = (patch: Partial<AdminState>): void => store.update((s) => ({ ...s, ...patch }));

/** Run an admin API effect against the live node, as a promise. */
const run = <A>(eff: Effect.Effect<A, ApiError, Client>): Promise<A> =>
  Effect.runPromise(eff.pipe(Effect.provide(Client.layer())));

const badToken = (err: ApiError): boolean => err.status === 401;

/** Fetch the overview for `tok` and fold it into the store. */
const load = (tok: string): void => {
  set({ phase: "loading", error: undefined });
  void run(fetchOverview(tok)).then(
    (overview) => set({ phase: "ready", overview, error: undefined }),
    (err: ApiError) => {
      if (badToken(err)) {
        sessionStorage.removeItem(TOKEN_KEY);
        set({ token: "", phase: "error", error: "bad admin token" });
      } else {
        set({ phase: "error", error: err.message });
      }
    },
  );
};

/** Refresh the overview after an action, without disturbing the phase. */
const refreshQuietly = (tok: string): void => {
  void run(fetchOverview(tok)).then(
    (overview) => set({ overview }),
    () => {}, // the action already succeeded; a stale overview is not fatal
  );
};

const connect = (tok: string): void => {
  const t = tok.trim();
  if (t === "") return;
  sessionStorage.setItem(TOKEN_KEY, t);
  set({ token: t, phase: "loading", error: undefined, actionError: undefined });
  load(t);
};

const create = (user: string): void => {
  const tok = get(store).token;
  const u = user.trim();
  if (tok === "" || u === "") return;
  set({ busy: true, actionError: undefined });
  void run(createKey(tok, u)).then(
    (res) => {
      set({ busy: false, newToken: res.token, notice: `Key for ${res.key.user} created`, actionError: undefined });
      refreshQuietly(tok);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const revoke = (id: string): void => {
  const tok = get(store).token;
  if (tok === "") return;
  set({ busy: true, actionError: undefined });
  void run(revokeKey(tok, id)).then(
    (res) => {
      set({ busy: false, notice: `Key for ${res.key.user} revoked`, actionError: undefined });
      refreshQuietly(tok);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const rotate = (id: string): void => {
  const tok = get(store).token;
  if (tok === "") return;
  set({ busy: true, actionError: undefined });
  void run(rotateKey(tok, id)).then(
    (res) => {
      set({ busy: false, newToken: res.token, notice: `Key for ${res.key.user} rotated`, actionError: undefined });
      refreshQuietly(tok);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const removeComment = (id: string): void => {
  const tok = get(store).token;
  if (tok === "") return;
  set({ busy: true, actionError: undefined });
  void run(deleteComment(tok, id)).then(
    () => {
      set({ busy: false, notice: "Comment deleted", actionError: undefined });
      refreshQuietly(tok);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const removePost = (site: string): void => {
  const tok = get(store).token;
  if (tok === "") return;
  set({ busy: true, actionError: undefined });
  void run(deletePost(tok, site)).then(
    () => {
      set({ busy: false, notice: `Page ${site} removed`, actionError: undefined });
      refreshQuietly(tok);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const refresh = (): void => {
  const tok = get(store).token;
  if (tok === "") return;
  void run(fetchOverview(tok)).then(
    (overview) => set({ overview, phase: "ready" }),
    (err: ApiError) => set({ actionError: err.message }),
  );
};

/** Replace the node's admin token and carry the session over to the new one. */
const change = (next: string): void => {
  const tok = get(store).token;
  const t = next.trim();
  if (tok === "" || t === "") return;
  set({ busy: true, actionError: undefined });
  void run(changeToken(tok, t)).then(
    () => {
      sessionStorage.setItem(TOKEN_KEY, t);
      set({ token: t, busy: false, actionError: undefined });
      load(t);
    },
    (err: ApiError) => set({ busy: false, actionError: err.message }),
  );
};

const dismiss = (): void => set({ newToken: undefined, notice: undefined });

const logout = (): void => {
  sessionStorage.removeItem(TOKEN_KEY);
  store.set(initialAdmin(""));
};

mount(AdminApp, {
  target: root,
  props: {
    admin: store,
    onConnect: connect,
    onCreate: create,
    onRevoke: revoke,
    onRotate: rotate,
    onDeleteComment: removeComment,
    onDeletePost: removePost,
    onRefresh: refresh,
    onDismiss: dismiss,
    onLogout: logout,
    onChangeToken: change,
  },
});

// Resume a saved session.
const saved = sessionStorage.getItem(TOKEN_KEY) ?? "";
if (saved !== "") load(saved);
