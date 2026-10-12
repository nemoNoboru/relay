/**
 * The node HTTP client — the only wire-aware code in the frontend
 * (frontend-architecture skill, `api/` rule).
 *
 * - Relative URLs only: the node serves the UI and the API same-origin
 *   (references/relay-ui.md).
 * - Every response body is decoded with a Schema; nothing crosses into app
 *   code as `unknown`, and no JSON.parse happens outside this file.
 */
import { Context, Effect, Layer, Schema } from "effect";
import type { Constraint, SchemaError } from "effect/Schema";

/**
 * A typed failure of a node request. `message` is safe to render in the UI.
 */
export class ApiError extends Schema.TaggedError<ApiError>()("ApiError", {
  message: Schema.String,
}) {}

/**
 * Renders an arbitrary failure as a short, UI-safe message. Schema decode
 * errors arrive as arrays of issues, each carrying its own `message`.
 */
const describeError = (error: unknown): string => {
  if (typeof error === "string") return error;
  if (error instanceof Error) return error.message;
  if (Array.isArray(error)) return error.map(describeError).join(", ");
  if (typeof error === "object" && error !== null) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === "string") return message;
  }
  try {
    return JSON.stringify(error);
  } catch {
    return String(error);
  }
};

export interface ClientShape {
  /** GET `url` (relative, same-origin) and decode the JSON body with `schema`. */
  readonly get: <S extends Constraint>(
    url: string,
    schema: S,
  ) => Effect.Effect<S["Type"], ApiError, S["DecodingServices"]>;
}

export class Client extends Context.Service<Client, ClientShape>()("relay.web.Client") {
  /**
   * Live layer: real `fetch` against the same-origin node. Tests pass a stub
   * `fetch` instead. The layer itself cannot fail; requests fail with
   * `ApiError`.
   */
  static readonly layer = (fetchFn: typeof globalThis.fetch = globalThis.fetch): Layer.Layer<Client> =>
    Layer.succeed(Client, Client.of({
      get: (url, schema) =>
        Effect.gen(function* () {
          const res = yield* Effect.tryPromise({
            try: () => fetchFn(url, { headers: { accept: "application/json" } }),
            catch: (error: unknown) =>
              new ApiError({ message: `cannot reach the API at ${url}: ${describeError(error)}` }),
          });
          if (!res.ok) {
            return yield* new ApiError({
              message: `the API answered ${res.status} ${res.statusText} for ${url}`,
            });
          }
          const body: unknown = yield* Effect.tryPromise({
            try: () => res.json(),
            catch: (error: unknown) =>
              new ApiError({ message: `the API sent non-JSON for ${url}: ${describeError(error)}` }),
          });
          return yield* Schema.decodeUnknownEffect(schema)(body).pipe(
            Effect.mapError((issue: SchemaError) =>
              new ApiError({
                message: `the API payload for ${url} does not match the expected shape: ${describeError(issue)}`,
              }),
            ),
          );
        }),
    }));
}