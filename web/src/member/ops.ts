import { Effect } from "effect";

import { ApiError, Client } from "../api/client";
import { CommentReply, Comments, Feed, Me, Members, PostReply, Profile } from "./schema";

/**
 * Member session, profile, and post-index effects. The node sets an
 * HttpOnly cookie on POST /session; the browser never reads or stores the
 * token itself.
 */

export const establishSession = (token: string): Effect.Effect<Me, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post("/_relay/v1/session", { token }, Me);
  });

export const fetchMe = (): Effect.Effect<Me, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get("/_relay/v1/me", Me);
  });

export const signOut = (): Effect.Effect<void, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.del("/_relay/v1/session");
  });

export const saveProfile = (displayName: string, bio: string): Effect.Effect<Me, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.patch("/_relay/v1/me", { displayName, bio }, Me);
  });

export const fetchFeed = (cursor?: string): Effect.Effect<Feed, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    const url = cursor ? `/_relay/v1/feed?cursor=${encodeURIComponent(cursor)}` : "/_relay/v1/feed";
    return yield* client.get(url, Feed);
  });

export const fetchMembers = (): Effect.Effect<Members, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get("/_relay/v1/people", Members);
  });

export const fetchProfile = (memberID: string, cursor?: string): Effect.Effect<Profile, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    const path = `/_relay/v1/people/${encodeURIComponent(memberID)}`;
    const url = cursor ? `${path}?cursor=${encodeURIComponent(cursor)}` : path;
    return yield* client.get(url, Profile);
  });

export const fetchPost = (site: string): Effect.Effect<PostReply, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get(`/_relay/v1/posts/${encodeURIComponent(site)}`, PostReply);
  });

export const fetchComments = (site: string): Effect.Effect<Comments, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.get(`/_relay/v1/posts/${encodeURIComponent(site)}/comments`, Comments);
  });

export const addComment = (site: string, body: string): Effect.Effect<CommentReply, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.post(`/_relay/v1/posts/${encodeURIComponent(site)}/comments`, { body }, CommentReply);
  });

export const deleteComment = (id: string): Effect.Effect<void, ApiError, Client> =>
  Effect.gen(function* () {
    const client = yield* Client;
    return yield* client.del(`/_relay/v1/comments/${encodeURIComponent(id)}`);
  });
