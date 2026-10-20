/**
 * SSR-render tests for the public feed (served at /).
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { FeedState } from "../src/feed/state";
import FeedApp from "../src/ui/Feed.svelte";

const noop = (): void => {};

const renderFeed = (state: FeedState): PromiseLike<string> =>
  render(FeedApp, {
    props: {
      feed: writable<FeedState>(state),
      onSignIn: noop,
      onSignOut: noop,
      onLoadMore: noop,
      onComment: async () => {},
      onDeleteComment: async () => {},
    },
  }).then((out) => out.body);

const ME = {
  member: {
    id: "dk_1",
    user: "alice",
    displayName: "Alice",
    bio: "Runs the reading group.",
    created: "2026-10-19T12:00:00Z",
  },
  node: { name: "my-group", members: 12 },
} as const;

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

describe("Feed.svelte", () => {
  it("renders the public feed for a guest", async () => {
    const html = await renderFeed({ phase: "ready", posts: [POST] });
    expect(html).toContain("Sunday walk");
    expect(html).toContain("Six photos from the canal.");
    expect(html).toContain("Gallery");
    expect(html).toContain("Sign in");
    expect(html).toContain("Run your own Relay");
    expect(html).not.toContain("Add a comment");
  });

  it("renders member identity and a comment form", async () => {
    const html = await renderFeed({ phase: "ready", me: ME, posts: [POST] });
    expect(html).toContain("Alice");
    expect(html).toContain("my-group");
    expect(html).toContain("sign out");
    expect(html).toContain("Add a comment…");
  });

  it("shows an empty state with a publish action for members", async () => {
    const html = await renderFeed({ phase: "ready", me: ME, posts: [] });
    expect(html).toContain("Nothing here yet.");
    expect(html).toContain("Publish");
  });

  it("shows the loading state", async () => {
    const html = await renderFeed({ phase: "loading", posts: [] });
    expect(html).toContain("…");
  });

  it("shows the error state", async () => {
    const html = await renderFeed({ phase: "error", posts: [], error: "The node did not answer." });
    expect(html).toContain("The node did not answer.");
  });
});
