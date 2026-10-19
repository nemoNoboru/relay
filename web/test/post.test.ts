/**
 * SSR-render tests for the post wrapper (`/_relay/posts/<site>`).
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { PostState } from "../src/post/state";
import PostApp from "../src/ui/PostPage.svelte";

const POST = {
  id: "post_1",
  site: "sunday-walk",
  member: "dk_1",
  title: "Sunday walk",
  summary: "Six photos from the canal.",
  template: "gallery",
  publishedAt: "2026-10-19T16:04:00Z",
  updatedAt: "2026-10-19T16:04:00Z",
  author: { id: "dk_1", displayName: "Alice" },
  url: "/my-group/sunday-walk/",
  postUrl: "/_relay/posts/sunday-walk",
  authorUrl: "/_relay/people/dk_1",
  comments: [],
} as const;

const noop = (): void => {};

const renderPost = (state: PostState): PromiseLike<string> =>
  render(PostApp, {
    props: {
      postState: writable<PostState>(state),
      onComment: noop,
      onDeleteComment: noop,
    },
  }).then((out) => out.body);

describe("PostPage.svelte", () => {
  it("renders the post metadata, iframe and comment form", async () => {
    const html = await renderPost({ phase: "ready", post: POST });
    expect(html).toContain("Sunday walk");
    expect(html).toContain("Six photos from the canal.");
    expect(html).toContain("Alice");
    expect(html).toContain("gallery");
    expect(html).toContain('src="/my-group/sunday-walk/"');
    expect(html).toContain("Talk under the post");
    expect(html).toContain("Post comment");
    expect(html).toContain("No comments yet.");
  });

  it("renders comments with the sanitized body and delete for own comments", async () => {
    const html = await renderPost({
      phase: "ready",
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

  it("shows the signed-out state", async () => {
    const html = await renderPost({ phase: "signedOut" });
    expect(html).toContain("an invitation");
  });
});
