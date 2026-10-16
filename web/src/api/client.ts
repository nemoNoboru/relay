/**
 * The node HTTP client — the only wire-aware code in the frontend
 * (frontend-architecture skill, `api/` rule).
 *
 * - Relative URLs only: the node serves the UI and the API same-origin
 *   (references/relay-ui.md).
 * - Every response body is decoded with a Schema; nothing crosses into app
 *   code as `unknown`, and no JSON.parse happens outside this file.
 * - `token`, when given, is sent as a bearer token (the admin/deploy
 *   credentials). It is never logged or put in a URL.
 */
import { Context, Effect, Layer, Schema } from "effect";
import type { Constraint, SchemaError } from "effect/Schema";

/**
 * A typed failure of a node request. `message` is safe to render in the UI.
 * `status` is present when the node answered with an error status.
 */
export class ApiError extends Schema.TaggedError<ApiError>()("ApiError", {
  message: Schema.String,
  status: Schema.optional(Schema.Number),
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
    token?: string,
  ) => Effect.Effect<S["Type"], ApiError, S["DecodingServices"]>;
  /** POST `url` with an optional JSON `body` and decode the response with `schema`. */
  readonly post: <S extends Constraint>(
    url: string,
    body: unknown,
    schema: S,
    token?: string,
  ) => Effect.Effect<S["Type"], ApiError, S["DecodingServices"]>;
}

/** One request, as a plain function; the service methods are thin wrappers. */
const request = <S extends Constraint>(
  fetchFn: typeof globalThis.fetch,
  url: string,
  init: { readonly method: string; readonly token?: string; readonly body?: unknown },
  schema: S,
): Effect.Effect<S["Type"], ApiError, S["DecodingServices"]> =>
  Effect.gen(function* () {
    const headers: Record<string, string> = { accept: "application/json" };
    if (init.token !== undefined) headers.authorization = `Bearer ${init.token}`;
    const hasBody = init.body !== undefined;
    if (hasBody) headers["content-type"] = "application/json";

    const res = yield* Effect.tryPromise({
      try: () =>
        fetchFn(url, {
          method: init.method,
          headers,
          body: hasBody ? JSON.stringify(init.body) : undefined,
        }),
      catch: (error: unknown) =>
        new ApiError({ message: `cannot reach the API at ${url}: ${describeError(error)}` }),
    });
    if (!res.ok) {
      return yield* new ApiError({
        message: `the API answered ${res.status} ${res.statusText} for ${url}`,
        status: res.status,
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
  });

export class Client extends Context.Service<Client, ClientShape>()("relay.web.Client") {
  /**
   * Live layer: real `fetch` against the same-origin node. Tests pass a stub
   * `fetch` instead. The layer itself cannot fail; requests fail with
   * `ApiError`.
   */
  static readonly layer = (fetchFn: typeof globalThis.fetch = globalThis.fetch): Layer.Layer<Client> =>
    Layer.succeed(
      Client,
      Client.of({
        get: (url, schema, token) => request(fetchFn, url, { method: "GET", token }, schema),
        post: (url, body, schema, token) => request(fetchFn, url, { method: "POST", body, token }, schema),
      }),
    );
}
