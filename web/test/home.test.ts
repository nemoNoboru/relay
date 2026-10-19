/**
 * SSR-render tests for the member welcome page (`home.html`).
 *
 * The page renders the `home` store; the invite exchange and profile effects
 * live in `main/home.ts` and are not exercised here.
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { HomeState } from "../src/home/state";
import HomeApp from "../src/ui/Home.svelte";

const noop = (): void => {};

const renderHome = (state: HomeState): PromiseLike<string> =>
  render(HomeApp, {
    props: {
      home: writable<HomeState>(state),
      onSignIn: noop,
      onSignOut: noop,
      onSave: noop,
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
  publishedAt: "2026-10-19T16:04:00Z",
  updatedAt: "2026-10-19T16:04:00Z",
  author: { id: "dk_1", displayName: "Alice" },
  url: "/my-group/sunday-walk/",
  postUrl: "/_relay/posts/sunday-walk",
  authorUrl: "/_relay/people/dk_1",
  comments: [],
} as const;

describe("Home.svelte", () => {
  it("shows the invitation prompt when signed out", async () => {
    const html = await renderHome({ phase: "signedOut", busy: false });
    expect(html).toContain("You need");
    expect(html).toContain("an invitation");
    expect(html).toContain("invitation key or link");
    expect(html).toContain("Open my home");
  });

  it("shows a rejected invite message", async () => {
    const html = await renderHome({
      phase: "signedOut",
      busy: false,
      error: "That invitation is not valid or has been revoked.",
    });
    expect(html).toContain("not valid or has been revoked");
  });

  it("welcomes the member with their name, node and actions", async () => {
    const html = await renderHome({ phase: "ready", me: ME, busy: false, posts: [] });
    expect(html).toContain("Welcome,");
    expect(html).toContain("Alice");
    expect(html).toContain("my-group");
    expect(html).toContain("12");
    expect(html).toContain("Publish a page");
    expect(html).toContain("Sign out");
    expect(html).toContain("Your profile");
    expect(html).toContain("Runs the reading group.");
    expect(html).toContain("Nothing here yet");
  });

  it("renders the chronological feed with a post card", async () => {
    const html = await renderHome({ phase: "ready", me: ME, busy: false, posts: [POST] });
    expect(html).toContain("From your people");
    expect(html).toContain("Sunday walk");
    expect(html).toContain("Six photos from the canal.");
    expect(html).toContain("/my-group/sunday-walk/");
    expect(html).toContain("/_relay/posts/sunday-walk");
    expect(html).toContain("/_relay/people/dk_1");
    expect(html).toContain("No comments yet.");
    expect(html).toContain("Add a comment");
  });

  it("renders inline comments and an own-comment delete", async () => {
    const withComment = {
      ...POST,
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
    };
    const html = await renderHome({ phase: "ready", me: ME, busy: false, posts: [withComment] });
    expect(html).toContain("The third photo is my favourite.");
    expect(html).toContain("Delete");
  });

  it("shows the loading state", async () => {
    const html = await renderHome({ phase: "loading" });
    expect(html).toContain("Opening your invitation");
  });
});
