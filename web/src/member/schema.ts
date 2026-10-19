import { Schema } from "effect";

/** A member as the node describes them. */
export const Member = Schema.Struct({
  id: Schema.String,
  user: Schema.String,
  displayName: Schema.String,
  bio: Schema.optional(Schema.String),
  created: Schema.String,
  invitedBy: Schema.optional(Schema.String),
  revoked: Schema.optional(Schema.Boolean),
});

/** The node's small member-facing description. */
export const NodeView = Schema.Struct({
  name: Schema.String,
  members: Schema.Number,
});

/** Reply from POST /_relay/v1/session and GET/PATCH /_relay/v1/me. */
export const Me = Schema.Struct({
  member: Member,
  node: NodeView,
});

/** The author summary attached to a post or comment. */
export const Author = Schema.Struct({
  id: Schema.String,
  displayName: Schema.String,
  revoked: Schema.optional(Schema.Boolean),
});

/** One comment under a post. */
export const Comment = Schema.Struct({
  id: Schema.String,
  post: Schema.String,
  member: Schema.String,
  author: Author,
  body: Schema.String,
  html: Schema.String,
  createdAt: Schema.String,
  mine: Schema.optional(Schema.Boolean),
});

export const Comments = Schema.Struct({
  comments: Schema.Array(Comment),
});

export const CommentReply = Schema.Struct({
  comment: Comment,
});

/** One indexed post. The content itself is served at `url`. */
export const Post = Schema.Struct({
  id: Schema.String,
  site: Schema.String,
  member: Schema.String,
  title: Schema.String,
  summary: Schema.optional(Schema.String),
  template: Schema.String,
  image: Schema.optional(Schema.String),
  imageUrl: Schema.optional(Schema.String),
  publishedAt: Schema.String,
  updatedAt: Schema.String,
  hidden: Schema.optional(Schema.Boolean),
  author: Author,
  url: Schema.String,
  postUrl: Schema.String,
  authorUrl: Schema.String,
  comments: Schema.Array(Comment),
});

export const Feed = Schema.Struct({
  posts: Schema.Array(Post),
  next: Schema.optional(Schema.String),
});

export const Profile = Schema.Struct({
  member: Member,
  posts: Schema.Array(Post),
  total: Schema.Number,
  next: Schema.optional(Schema.String),
});

export const PostReply = Schema.Struct({
  post: Post,
});

export type Member = (typeof Member)["Type"];
export type NodeView = (typeof NodeView)["Type"];
export type Me = (typeof Me)["Type"];
export type Author = (typeof Author)["Type"];
export type Post = (typeof Post)["Type"];
export type Comment = (typeof Comment)["Type"];
export type Comments = (typeof Comments)["Type"];
export type CommentReply = (typeof CommentReply)["Type"];
export type Feed = (typeof Feed)["Type"];
export type Profile = (typeof Profile)["Type"];
export type PostReply = (typeof PostReply)["Type"];
