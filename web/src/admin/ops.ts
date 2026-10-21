import { Effect } from "effect";

import { ApiError, Client } from "../api/client";
import { AdminOverview, KeyCreated, KeyUpdated, TokenChanged } from "./schema";

/**
 * Operator admin operations. Each returns the Effect for one API call; the
 * page wiring (`main/admin.ts`) provides the Client layer and runs them.
 */

export const fetchOverview = (token: string): Effect.Effect<AdminOverview, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get("/_relay/v1/admin", AdminOverview, token);
  });

export const createKey = (token: string, user: string): Effect.Effect<KeyCreated, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post("/_relay/v1/admin/keys", { user }, KeyCreated, token);
  });

export const revokeKey = (token: string, id: string): Effect.Effect<KeyUpdated, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post(`/_relay/v1/admin/keys/${id}/revoke`, undefined, KeyUpdated, token);
  });

export const rotateKey = (token: string, id: string): Effect.Effect<KeyCreated, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post(`/_relay/v1/admin/keys/${id}/rotate`, undefined, KeyCreated, token);
  });

export const changeToken = (token: string, next: string): Effect.Effect<TokenChanged, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post("/_relay/v1/admin/token", { token: next }, TokenChanged, token);
  });

export const deleteComment = (token: string, id: string): Effect.Effect<void, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.del(`/_relay/v1/admin/comments/${encodeURIComponent(id)}`, token);
  });

export const deletePost = (token: string, site: string): Effect.Effect<void, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.del(`/_relay/v1/admin/posts/${encodeURIComponent(site)}`, token);
  });
