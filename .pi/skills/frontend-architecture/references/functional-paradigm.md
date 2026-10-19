# Functional Paradigm (Relay frontend)

Project decision: functional programming is a **valid paradigm of choice** for Relay's
frontend, and the default for application code under `web/src`. This reference defines
what that means in practice — it is a working discipline, not a school of thought.

## The five commitments

1. **Pure functions are the default unit of logic.** A function that only reads its
   arguments and returns a value is the normal case; anything else must say so.
2. **State is immutable data.** A change to state produces a new value
   (`{...state, members: [...state.members, m]}`), never an in-place edit.
3. **Side effects are named, typed, and confined to the edge.** DOM writes, fetches,
   timers, and storage live in the edge layer, each wrapped in `Effect` with a typed
   error channel.
4. **Data flows one way.** Events in, state out, DOM out. Nothing flows backward.
5. **The UI is a render of state, not a sequence of commands.** Handlers dispatch events;
   views paint state.

## Why functional is the right call for this codebase

- **The UI is operational data, not a document app.** The host console shows invitations,
  members, pages by member, and versions. Action results and derived state are exactly
  where one-way flow and pure derivations pay off: the hard part is knowing what changed
  and recomputing what depends on it, and immutability makes that mechanical.
- **Svelte renders; it does not own the rules.** Svelte is the view layer. The functional
  core — state type, fold, derivations, `Schema` decoding — stays in plain TypeScript and
  remains testable without a browser. Do not move business rules into components.
- **It mirrors the backend.** Relay's Go side is domain-driven: value objects with
  invariants on types, aggregates with behavior, persistence at the edge. A functional
  frontend keeps the same shape across the wire — `Schema` definitions mirror the Go
  domain vocabulary (node, member, invitation, page, version), so both sides read the same.
- **The core is testable without a browser.** Pure folds and derivations are
  table-driven tests in milliseconds. The browser is needed only for edge tests.
- **State is bounded and comprehensible.** A node has dozens of members and pages, not
  millions. One immutable `AdminState` object per page is a design, not an accident.

## The shape

```
        ┌───────────────────┐   events     ┌────────────────────────┐
 user ─►│ DOM events        │─────────────►│  fold (pure)           │
 poll ─►│ poll ticks        │─────────────►│  (State, Event) -> State
 timer► │ timer ticks       │─────────────►│                        │
        └───────────────────┘             └────────────┬───────────┘
                                                       │ new State (immutable)
                                                       ▼
                                              ┌────────────────────────┐
                                              │  view (pure)           │
                                              │  State -> DOM fragment │
                                              └────────────────────────┘
```

- **Events** are a discriminated union (`{ _tag: "OverviewLoaded" } | …`), produced by the
  edge (user clicks, poll results, timers).
- **fold** is a pure function `(state, event) => state`. All business rules live here and
  are the main test surface.
- **view** is a pure function `state => DOM`. It decides what the page looks like for a
  state. No fetches, no timers, no globals.
- The **edge** (Effect world) owns everything impure: the HTTP client, the polling stream,
  event listeners, and the act of writing the rendered fragment to the document.

In Effect terms the page is: a scoped program that runs the event streams, folds them into
a `Ref` of state, and renders on change — see `effect-tsx.md` for the entrypoint pattern.

## Code rules

- `readonly` by default: state types, arrays, and service methods are `readonly`;
  mutation APIs are avoided.
- Copy-on-write for updates; never edit a value two owners can see.
- No getters that do work, no setter-style methods on data. Behavior belongs to services
  (`Context.Service`), data belongs to schema types.
- Classes are used only as `Schema.Class` domain models (invariants + a type) — never as
  mutable god-objects.
- A function that takes state and returns state is pure; if it cannot be, rename it and
  make it an `Effect`.

## Imperative habit → functional shape

| Habit | Functional shape |
|---|---|
| `let state = {...}; state.x = v` | `const next = fold(state, event)` — a new object |
| `fetch().then(r => r.json()).then(d => { dom })` | `api.fetchX(): Effect<X, ApiError>`; render happens on the state change |
| A handler that mutates four things | One event type in, one pure fold, one render |
| `setInterval` + manual `clearInterval` | `Stream.fromEffectSchedule` + scoped subscription (cancelled for you) |
| Reading state back out of the DOM (`el.dataset`, `el.checked`) | State lives in the model; the DOM is output only |
| `try/catch` around everything | Typed `TaggedError`s; `Effect.catchTag` at the edge; every failure becomes a state |
| "Update the UI after I change the thing" | The UI updates because state changed |

## Anti-patterns (smell list)

- **Module-level `let` as a datastore.** If a function outside the fold can see and edit
  shared state, the fold is a lie.
- **The DOM as a datastore.** Reconstructing state from the document is reading your
  output instead of your model.
- **God handlers.** A click handler that validates, fetches, mutates, and patches three
  elements is four jobs; split it into event → fold → render + one effectful action.
- **Callback pyramids deeper than two.** Three `.then`s is an `Effect.gen` wearing a costume.
- **Side effects in derivations.** A "pure" selector that fetches, logs, or touches
  `Date.now()` is impure; push the effect outward.
- **Rebuilding the whole document per tick.** Render the regions that changed; at Relay's
  page size, replacing a section fragment per tick is fine, replacing `<body>` is not.
- **`alert()` / `console.error` as error handling.** Failures are states (`loading |
  ready | stale | error`) and are rendered.

## When the discipline can relax

- Throwaway scripts in `tools/` (one-offs, deploy helpers) are plain TypeScript, no
  structure required. The standard applies to application code under `web/`.
- Inside `dom/`, helper functions may manipulate a *fresh* detached fragment — that is
  constructing output, not mutating shared state.
