/**
 * Table-driven tests for the live-data fold (the state machine every polled
 * page runs on) and the `ago` display helper.
 */
import { Effect } from "effect";
import { describe, effect } from "@effect/vitest";
import { STALE_AFTER_MS, ago, fold, initial } from "../src/fold";
import type { LiveEvent, LiveState } from "../src/fold";

type Payload = { readonly n: number };

const ok = (n: number, at: number): LiveEvent<Payload> => ({ _tag: "ok", data: { n }, at });
const fail = (message: string, at: number): LiveEvent<Payload> => ({ _tag: "fail", message, at });
const ready = (n: number, updatedAt: number): LiveState<Payload> => ({
  phase: "ready",
  data: { n },
  updatedAt,
  lastError: undefined,
});
const stale = (n: number, updatedAt: number, lastError: string): LiveState<Payload> => ({
  phase: "stale",
  data: { n },
  updatedAt,
  lastError,
});
const errored = (n: number, updatedAt: number, lastError: string): LiveState<Payload> => ({
  phase: "error",
  data: { n },
  updatedAt,
  lastError,
});

interface Case {
  readonly name: string;
  readonly start: LiveState<Payload>;
  readonly event: LiveEvent<Payload>;
  readonly wantPhase: LiveState<Payload>["phase"];
  readonly wantData?: number;
}

const cases: ReadonlyArray<Case> = [
  { name: "loading + ok → ready, data kept", start: initial(), event: ok(1, 1000), wantPhase: "ready", wantData: 1 },
  { name: "loading + fail → error (nothing to keep showing)", start: initial(), event: fail("boom", 1000), wantPhase: "error" },
  { name: "ready + fresh fail → stale, data kept", start: ready(1, 1000), event: fail("boom", 1000 + STALE_AFTER_MS - 1), wantPhase: "stale", wantData: 1 },
  { name: "ready + expired fail → error", start: ready(1, 1000), event: fail("boom", 1000 + STALE_AFTER_MS + 1), wantPhase: "error" },
  { name: "stale + ok → ready again", start: stale(1, 1000, "boom"), event: ok(2, 5000), wantPhase: "ready", wantData: 2 },
  { name: "error + ok → ready (recovery)", start: errored(1, 1000, "boom"), event: ok(3, 999_999), wantPhase: "ready", wantData: 3 },
];

describe("fold", () => {
  effect.each(cases)("{name}", (c, ctx) =>
    Effect.sync(() => {
      const s = fold(c.start, c.event);
      ctx.expect(s.phase).toBe(c.wantPhase);
      if (c.wantData !== undefined) ctx.expect(s.data?.n).toBe(c.wantData);
      if (c.event._tag === "fail") ctx.expect(s.lastError).toBe(c.event.message);
      if (c.event._tag === "ok") ctx.expect(s.lastError).toBeUndefined();
    }),
  );
});

describe("ago", () => {
  effect("formats relative time", (ctx) =>
    Effect.sync(() => {
      ctx.expect(ago(undefined, 0)).toBe("never");
      ctx.expect(ago(0, 3_000)).toBe("3s");
      ctx.expect(ago(0, 125_000)).toBe("2m");
      ctx.expect(ago(0, 3_600_000)).toBe("1h");
      ctx.expect(ago(10_000, 5_000)).toBe("0s"); // future timestamps clamp
    }),
  );
});