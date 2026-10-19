import { Schema } from "effect";

export const TemplateField = Schema.Struct({
  name: Schema.String,
  type: Schema.String,
  required: Schema.optional(Schema.Boolean),
  max: Schema.optional(Schema.Number),
  accept: Schema.optional(Schema.String),
});

export const TemplateInfo = Schema.Struct({
  schema: Schema.Number,
  id: Schema.String,
  name: Schema.String,
  description: Schema.String,
  version: Schema.String,
  fields: Schema.Array(TemplateField),
  entry: Schema.String,
  script: Schema.String,
  builtin: Schema.Boolean,
});

export const Templates = Schema.Struct({
  templates: Schema.Array(TemplateInfo),
});

export type TemplateField = (typeof TemplateField)["Type"];
export type TemplateInfo = (typeof TemplateInfo)["Type"];
export type Templates = (typeof Templates)["Type"];
