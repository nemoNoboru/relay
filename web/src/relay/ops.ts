import { Effect } from "effect";

import { ApiError, Client } from "../api/client";
import { RelayJob } from "./schema";

/** Start a relay job for a url. The member cookie or a bearer token signs it. */
export const relayUrl = (url: string, token?: string): Effect.Effect<RelayJob, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post("/_relay/v1/relay", { url }, RelayJob, token);
  });

/** Poll a relay job until it is done or failed. */
export const relayJob = (id: string, token?: string): Effect.Effect<RelayJob, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get(`/_relay/v1/relay/${encodeURIComponent(id)}`, RelayJob, token);
  });
