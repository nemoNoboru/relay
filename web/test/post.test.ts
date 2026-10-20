/**
 * SSR-render tests for the post wrapper (`/_relay/posts/<site>`).
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { PostState } from "../src/post/state";
import PostApp from "../src/ui/PostPage.svelte";

const noop = (): void => {};

const POST = {
  id: "post_1",
  site: "sunday-walk",
  member: "dk_1",
  title: "Sunday walk",
  summary: "Six photos from the canal.",
  template: "gallery",
  templateName: "Gallery",
  publishedAt: "2026-10-19T16:04:00Z",
  updatedAt: "2026-10-19T16:04:00Z",
  author: { id: "dk_1", displayName: "Alice" },
  url: "/my-group/sunday-walk/",
  postUrl: "/_relay/posts/sunday-walk",
  authorUrl: "/_relay/people/dk_1",
  comments: [],
} as const;

const ME = {
  member: { id: "dk_1", user: "alice", displayName: "Alice", bio: "", created: "2026-10-19T12:00:00Z" },
  node: { name: "my-group", members: 12 },
} as const;

const renderPost = (state: PostState): PromiseLike<string> =>
  render(PostApp, {
    props: {
      postState: writable<PostState>(state),
      onSignIn: noop,
      onSignOut: noop,
      onComment: noop,
      onDeleteComment: noop,
    },
  }).then((out) => out.body);

describe("PostPage.svelte", () => {
  it("renders the post publicly with no write controls for a guest", async () => {
    const html = await renderPost({ phase: "ready", post: POST });
    expect(html).toContain("Sunday walk");
    expect(html).toContain("Six photos from the canal.");
    expect(html).toContain("Alice");
    expect(html).toContain("Gallery");
    expect(html).toContain('src="/my-group/sunday-walk/"');
    expect(html).toContain("Comments");
    expect(html).not.toContain("Reply");
  });

  it("shows the comment form to a member", async () => {
    const html = await renderPost({ phase: "ready", me: ME, post: POST });
    expect(html).toContain("Add a comment…");
    expect(html).toContain("Reply");
    expect(html).toContain("sign out");
  });

  it("renders comments with the sanitized body and delete for own comments", async () => {
    const html = await renderPost({
      phase: "ready",
      me: ME,
      post: POST,
      comments: [
        {
          id: "cmt_1",
          post: "sunday-walk",
          member: "dk_1",
          author: { id: "dk_1", displayName: "Alice" },
          body: "The third photo is my favourite.",
          html: "<p>The third photo is my favourite.</p>",
          createdAt: "2026-10-19T16:20:00Z",
          mine: true,
        },
      ],
    });
    expect(html).toContain("The third photo is my favourite.");
    expect(html).toContain("Delete");
  });

  it("shows the not-found state", async () => {
    const html = await renderPost({ phase: "error", error: "No such post on this node." });
    expect(html).toContain("No such post on this node.");
  });
});
