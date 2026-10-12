/**
 * Tests for the node client (`api/`) — the only wire-aware code. The layer
 * takes a `fetch` stub, so these run without a network and without a node.
 */
import { Effect } from "effect";
import { describe, effect, layer } from "@effect/vitest";
import { ApiError, Client } from "../src/api/client";
import { Tunnels } from "../src/tunnels/schema";

type FetchFn = typeof globalThis.fetch;

const stub = (
  response: { status?: number; body?: unknown } | Error,
): FetchFn =>
  async (_input, _init) => {
    if (response instanceof Error) throw response;
    return new Response(JSON.stringify(response.body ?? null), {
      status: response.status ?? 200,
      headers: { "content-type": "application/json" },
    });
  };

/** Effect 4 has no `Effect.either`; match into a plain tagged result. */
type Outcome<A, E> =
  | { readonly _tag: "ok"; readonly value: A }
  | { readonly _tag: "fail"; readonly value: E };

const outcome = <A, E>(self: Effect.Effect<A, E>): Effect.Effect<Outcome<A, E>, never> =>
  self.pipe(
    Effect.match({
      onSuccess: (value) => ({ _tag: "ok" as const, value }),
      onFailure: (value) => ({ _tag: "fail" as const, value }),
    }),
  );

describe("Client", () => {
  layer(Client.layer(stub({ body: { tunnels: [{ id: "t1", peer: "friend-a", prefixes: ["/raspberry/"], online: true }] } })))("happy path", (it) => {
    it.effect("decodes a well-formed payload with the schema", (ctx) =>
      Effect.gen(function* () {
        const client = yield* Client;
        const data = yield* client.get("/api/tunnels", Tunnels);
        ctx.expect(data.tunnels).toEqual([{ id: "t1", peer: "friend-a", prefixes: ["/raspberry/"], online: true }]);
      }),
    );
  });

  layer(Client.layer(stub({ status: 500, body: {} })))("http error", (it) => {
    it.effect("surfaces a typed ApiError with the status", (ctx) =>
      Effect.gen(function* () {
        const client = yield* Client;
        const result = yield* outcome(client.get("/api/tunnels", Tunnels));
        ctx.expect(result._tag).toBe("fail");
        if (result._tag === "fail") {
          ctx.expect(result.value._tag).toBe("ApiError");
          ctx.expect(result.value.message).toContain("500");
        }
      }),
    );
  });

  layer(Client.layer(stub({ body: { tunnels: "not-an-array" } })))("decode error", (it) => {
    it.effect("fails when the payload shape is wrong", (ctx) =>
      Effect.gen(function* () {
        const client = yield* Client;
        const result = yield* outcome(client.get("/api/tunnels", Tunnels));
        ctx.expect(result._tag).toBe("fail");
        if (result._tag === "fail") ctx.expect(result.value._tag).toBe("ApiError");
      }),
    );
  });

  layer(Client.layer(stub(new Error("boom"))))("network error", (it) => {
    it.effect("wraps fetch failures in ApiError", (ctx) =>
      Effect.gen(function* () {
        const client = yield* Client;
        const result = yield* outcome(client.get("/api/tunnels", Tunnels));
        ctx.expect(result._tag).toBe("fail");
        if (result._tag === "fail") {
          ctx.expect(result.value._tag).toBe("ApiError");
          ctx.expect(result.value.message).toContain("boom");
        }
      }),
    );
  });
});

describe("ApiError", () => {
  effect("is a Schema-backed tagged error with an explicit shape", (ctx) =>
    Effect.sync(() => {
      const e = new ApiError({ message: "x" });
      ctx.expect(e._tag).toBe("ApiError");
      ctx.expect(e.message).toBe("x");
    }),
  );
});