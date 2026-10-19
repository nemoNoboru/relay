import { Effect } from "effect";

import { ApiError, Client } from "../api/client";
import { Templates } from "./schema";

export const fetchTemplates = (): Effect.Effect<Templates, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get("/_relay/v1/templates", Templates);
  });
