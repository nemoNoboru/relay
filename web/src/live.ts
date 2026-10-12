/**
 * The live-page driver — the single place the "poll, decode, fold, render"
 * pipeline lives. Both pages (node admin, site console) are thin options
 * objects handed to `startPage`; this module owns the Effect service wiring,
 * the scoped poll stream, the Svelte store, and pagehide teardown.
 *
 * Data flow per tick:
 *
 *   opts.tick(client) (fetch + decode)  →  LiveEvent  →  fold(state, event)  →  store
 *                                                                        │
 *   Svelte component ◄───────────────────────────────────────────────────┘
 *   (pure rendering of the store; this module is the only mutable state)
 */
import { Duration, Effect, Exit, Schedule, Scope, Stream } from "effect";
import type { Component } from "svelte";
import { mount, unmount } from "svelte";
import type { Writable } from "svelte/store";
import { writable } from "svelte/store";

import { Client } from "./api/client";
import { ApiError } from "./api/client";
import type { ClientShape } from "./api/client";
import { fold, initial } from "./fold";
import type { LiveEvent, LiveState } from "./fold";

/** Everything that makes one page a *live* page. */
export interface LivePageOptions<T> {
  /** The `#app` element to mount into. */
  readonly target: HTMLElement;
  /** Poll cadence (ms). references/relay-ui.md: 5s tunnels, 10s deploys. */
  readonly pollMs: number;
  /** The page view — a pure renderer of the live store. */
  readonly component: Component<{ live: Writable<LiveState<T>> }>;
  /**
   * One tick: fetch + decode the page's data. This is the page's wire shape —
   * the schema is named here, at the only place that matters.
   */
  readonly tick: (client: ClientShape) => Effect.Effect<T, ApiError, never>;
}

/**
 * Mounts the page, forks its poll stream into a page-local scope, and tears
 * both down on `pagehide`. Idempotency is not a goal — each page calls this
 * exactly once at module load.
 */
export function startPage<T>(opts: LivePageOptions<T>): void {
  const live = writable<LiveState<T>>(initial<T>());
  const component = mount(opts.component, { target: opts.target, props: { live } });

  /** One tick: fetch + decode, classify the result, fold it into the store. */
  const tick = (client: ClientShape): Effect.Effect<void, never> =>
    Effect.gen(function* () {
      const at = Date.now();
      const event: LiveEvent<T> = yield* opts.tick(client).pipe(
        Effect.match({
          onSuccess: (data) => ({ _tag: "ok" as const, data, at }),
          onFailure: (error) => ({ _tag: "fail" as const, message: error.message, at }),
        }),
      );
      live.update((state) => fold(state, event));
    });

  const start = (): Effect.Effect<void, never> =>
    Effect.gen(function* () {
      const client = yield* Client;
      const page = yield* Scope.make();

      // Live data is a Stream (skill, non-negotiable): run the tick once,
      // then re-run it on a schedule that re-fires after each completion.
      // Forked into the page scope, it dies when the scope closes.
      yield* Effect.forkScoped(
        Stream.runDrain(Stream.fromEffectSchedule(tick(client), Schedule.spaced(Duration.millis(opts.pollMs)))),
      ).pipe(Effect.provideService(Scope.Scope, page));

      // Teardown: when the page goes away, unmount the view and close the
      // page scope — the poll fiber runs its finalizer. The listener
      // finalizer is registered in the page scope, so closing it also
      // removes the listener; the page's JS context dies with it either way.
      yield* Effect.acquireRelease(
        Effect.sync(() => {
          const onHide = (): void => {
            void unmount(component);
            void Effect.runFork(Scope.close(page, Exit.void));
          };
          window.addEventListener("pagehide", onHide, { once: true });
          return onHide;
        }),
        (onHide) => Effect.sync(() => window.removeEventListener("pagehide", onHide)),
      ).pipe(Effect.provideService(Scope.Scope, page));
    }).pipe(Effect.provide(Client.layer()));

  // Mounting happened above, so the `loading` phase is visible before the
  // first tick fires.
  void Effect.runFork(start());
}