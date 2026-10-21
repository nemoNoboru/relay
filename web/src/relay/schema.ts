import { Schema } from "effect";

/** One relay job as the node reports it, from POST and GET. */
export const RelayJob = Schema.Struct({
  id: Schema.String,
  source: Schema.String,
  status: Schema.Literals(["running", "done", "failed"] as const),
  title: Schema.optional(Schema.String),
  kind: Schema.optional(Schema.String),
  site: Schema.optional(Schema.String),
  url: Schema.optional(Schema.String),
  postUrl: Schema.optional(Schema.String),
  error: Schema.optional(Schema.String),
  created: Schema.String,
});

export type RelayJob = (typeof RelayJob)["Type"];
