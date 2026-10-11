# Effect-TS for Relay frontend

Effect-TS (<https://effect.website>) is the functional library of choice for the Relay
frontend. This reference covers the current API (v4 line, as of this writing) for the
pieces this codebase uses.

**When in doubt about an API, do not guess.** The Effect repo ships AI-oriented docs that
track the current major: [`LLMS.md`](https://github.com/Effect-TS/effect/blob/main/LLMS.md)
and [`ai-docs/`](https://github.com/Effect-TS/effect/tree/main/ai-docs) on GitHub. The
library has shifted APIs between majors; memory and stale blog posts are not the source
of truth.

## Setup

Dependencies (pinned in the lockfile; upgrade the major deliberately, with a changelog
read-through):

```
effect            # core: Effect, Schema, Stream, Layer, Context, Schedule, …
@effect/platform  # browser runtime (FetchHttpClient)
@effect/vitest    # dev: it.effect, layer(...), assert
```

Import style: `import { Effect, Schema, Stream, Layer, Context } from "effect"`;
HTTP from `effect/http`.

## Writing effects

- `Effect.gen(function*() { … })` for inline effectful code — imperative style with
  `yield*`, like async/await but typed.
- `Effect.fn("Name.method")` for reusable effectful functions (spans/tracing get their
  name from this string). `Effect.fnUntraced` on hot paths.
- Wrapping plain sources:
  - `Effect.succeed(v)` — a value you already have
  - `Effect.sync(f)` — side effect that cannot fail
  - `Effect.try({ try: f, catch: cause => new MyError({ cause }) })` — throwing sync code
  - `Effect.tryPromise({ try: async f, catch })` — Promise APIs that can reject
  - `Effect.callback(resume => { …; return finalizer })` — callback APIs (note the
    returned finalizer: that is how interruption cancels the source)
  - `Effect.fromNullishOr(v).pipe(Effect.mapError(() => new Missing({})))` — nullish values

Always `return yield*` when raising inside a gen, so the type checker knows the flow ends.

## Typed errors

```ts
export class TunnelApiError extends Schema.TaggedError<TunnelApiError>()("TunnelApiError", {
  status: Schema.Int,
  cause: Schema.Defect(),
}) {}
```

- One `TaggedError` class per failure mode the caller can reasonably react to.
- Untyped causes (network failures, unknown throws) go in as `Schema.Defect()`.
- `Effect.catchTag("TunnelApiError", e => …)` to react to one; `Effect.catch(e => …)` as
  the last resort; `Effect.mapError(reason => new OuterError({ reason }))` to wrap.
- The wrap rule mirrors the Go skill: every layer adds context, nothing is swallowed,
  unhandled failures must be impossible by construction.
- At the page edge, failures become **rendered states** (`stale`/`error`), never `alert()`.

## Schema: the wire boundary

All API decoding happens here, and only here (inside `api/`).

```ts
export class Tunnel extends Schema.Class<Tunnel>("relay/Tunnel")({
  id: Schema.String,
  site: Schema.String,
  status: Schema.Literals(["connecting", "open", "closed"]),
  peers: Schema.Array(Schema.String),
}) {}

// decoded / encoded TS representations, when needed:
export type TunnelType = typeof Tunnel["Type"]
```

- `Schema.decodeUnknownEffect(Tunnel)` — the standard decode; errors stay typed in the
  error channel.
- Reuse the parsers at the edge; do not hand-parse JSON anywhere else, and do not
  validate with predicate functions where a Schema field exists
  (`Schema.NonEmptyString`, `Schema.Literals([...])`, `Schema.Array`, …).
- Schema definitions mirror the Go domain vocabulary: `Site`, `Prefix`, `Deploy`,
  `Tunnel`, `Friend`, `Metrics`. If the Go side renames a field, the Schema changes in
  the same PR.

## Services and layers

The structural unit of frontend code is the service.

```ts
// feature service
export class Tunnels extends Context.Service<Tunnels, {
  readonly state: Effect.Effect<NodesTunnelState>
  refresh(): Effect.Effect<void, TunnelApiError>
}>()("relay/Tunnels") {
  static readonly layerNoDeps = Layer.effect(Tunnels, Effect.gen(function*() {
    const api = yield* Api
    // …
    return Tunnels.of({ state, refresh })
  }))

  static readonly layer = Layer.effectDiscard(
    Tunnels.layerNoDeps.pipe(Layer.provide(Api.layer))
  )
}
```

- Identifier strings are `relay/<Feature>` — stable, used by tracing and tests.
- `layerNoDeps` declares dependencies in its type; `layer` composes them away with
  `Layer.provide` (hides dependencies) or `Layer.provideMerge` (keeps them exposed).
- `Context.Reference<T>("relay/PollInterval", { defaultValue: () => "5 seconds" })` for
  configuration and feature flags (poll cadences, dev overrides).
- No module-level singletons. If a value is shared, it is a service; if it is a constant,
  it is a `const`.

## Resources and scope (teardown)

The page is the scope:

```ts
// listener that cleans itself up when the fiber is interrupted
const clicks = Stream.fromEventListener(button, "click")

// or explicit acquire/release for anything with a lifetime
const es = yield* Effect.acquireRelease(
  Effect.sync(() => new EventSource("/api/metrics/stream")),
  (es) => Effect.sync(() => es.close())
)
```

- Start subscriptions with `Effect.forkScoped` (or run them inside a scoped gen):
  interrupting the scope runs every finalizer — listeners removed, streams stopped.
- Never `setInterval` by hand. A scoped polling stream replaces it (next section).

## HTTP (the node API)

Same-origin; relative paths; the node serves its own UI and API.

```ts
export const Http = Layer.effect(
  HttpClient.HttpClient,
  (yield* HttpClient.HttpClient).pipe(
    HttpClient.filterStatusOk,                                   // non-2xx is a typed failure
    HttpClient.retryTransient({ schedule: Schedule.exponential(100), times: 3 }),
  )
).pipe(Layer.provide(FetchHttpClient.layer))

const tunnels = client.get("/api/tunnels").pipe(
  Effect.flatMap(HttpClientResponse.schemaBodyJson(Schema.Array(Tunnel))),  // decoded, typed
  Effect.mapError(cause => new TunnelApiError({ status: 0, cause })),
)
```

- `schemaBodyJson(Schema)` is the standard response decode — never `res.json()` in app
  code.
- `HttpClientRequest` helpers (`setUrlParams`, `bodyJson`, `acceptJson`) for request
  shaping; the site console's mutating calls (deploy, rollback) go through the same
  client so auth headers, retries, and decoding stay consistent.
- The `api/` service exposes `Effect`-returning methods. Promises never leak past it.

## Streams: live data

This is how the owner console feels alive — polling modeled as data, not loops.

```ts
// poll the node API every 5s until the scope is interrupted
const poll = Stream.fromEffectSchedule(
  api.fetchTunnels(),
  Schedule.spaced("5 seconds")
).pipe(
  Stream.map(tunnels => ({ _tag: "TunnelData" as const, tunnels })),
  Stream.mapError(cause => new TunnelApiError({ status: 0, cause }))
)

// user events as a stream
const clicks = Stream.fromEventListener(button, "click")

// merge, fold, render — all scoped
yield* Effect.gen(function*() {
  const state = Ref.make<PageState>(initial)
  yield* Effect.forkScoped(
    Stream.merge([poll, clicks])
      .pipe(Stream.runForEach(e => Ref.update(state, s => fold(s, e))))
  )
  yield* Effect.forkScoped(
    Stream.fromRef(state).pipe(Stream.runForEach(s => view.render(s, mount)))
  )
})
```

- Cadence defaults (overridable via `Context.Reference`): tunnels 5s, deploys 10s,
  metrics 30s — daily-salted counters do not need faster.
- Poll failures do not kill the stream: map them into a `stale` event so the view can
  show "last updated …s ago" instead of going quiet.
- `Stream.take(n)` for bounded streams; `Stream.paginate` for paginated APIs.

## Entrypoint pattern (browser)

```ts
// web/src/main/node.ts — the only place the runtime is touched
const program = Effect.gen(function*() {
  const tunnels = yield* Tunnels
  const sites = yield* Sites
  // …mount, run subscriptions, render loop (see Streams section)
}).pipe(Effect.provide(Layer.succeed([Tunnels.layer, Sites.layer])))

const fiber = Effect.runFork(program)   // keep the handle; scope = page lifetime
window.addEventListener("pagehide", () => {
  Effect.runFork(Fiber.interrupt(fiber)) // runs every finalizer: listeners, polls
}, { once: true })
```

- `Effect.runFork` at the entry; keep the `Fiber`. Page teardown is `Fiber.interrupt`
  on `pagehide`.
- No fire-and-forget `Effect.runPromise` for long-lived work — an untracked promise is an
  orphaned subscription.

## Testing (`@effect/vitest`)

```ts
import { assert, describe, it, layer } from "@effect/vitest"

describe("fold", () => {
  it.effect.each([
    { event: { _tag: "TunnelData", tunnels: [openT] }, expect: 1 },
    { event: { _tag: "TunnelData", tunnels: [] },        expect: 0 },
  ])("tunnel count %#", ({ event, expect }) =>
    Effect.succeed(assert.strictEqual(countOpen(fold(initial, event)), expect)))
})

layer(Tunnels.layerTest)("Tunnels", (it) => {
  it.effect("survives a failed poll as stale", () => Effect.gen(function*() {
    // …with a faked Api layer
  }))
})

it.effect("polling cadence", () => Effect.gen(function*() {
  const fiber = yield* Effect.forkChild(pollOnce)
  yield* TestClock.adjust("5 seconds")   // virtual time: tests do not sleep
  assert.strictEqual(yield* Fiber.join(fiber), 1)
}))

it.effect.prop("fold round-trips", [Schema.Array(Schema.String)], (xs) =>
  Effect.succeed(assert.deepStrictEqual(recount(applyAll(xs, s0)), xs.length)))
```

- `it.effect` for Effect code, `it.live` when real runtime services are needed,
  `it.effect.prop` with Schema arbitraries for property tests.
- `layer(...)` blocks share one layer across a suite (`afterAll` teardown) — the pattern
  for testing services against faked dependencies (`layerTest` + `Context.Reference`/
  test-ref services stand in for `Api`).
- Time is virtual via `TestClock.adjust` — no test ever waits 5 real seconds.
- Table-driven for folds (the main test surface); the DOM itself is tested rarely, and
  then with `it.live`.

## What to avoid

- Leaking `Promise`s out of `api/` or feature services — they return `Effect`.
- `unsafe` APIs (`offerUnsafe`, `runSyncUnsafe`, …) outside of finalizer callbacks where
  an effect is not available.
- Unscoped long-lived work — anything that outlives a tick needs a scope, so it can be
  interrupted.
- Importing `effect/*` deep modules for things the root export already provides; the
  root namespace is the standard surface.