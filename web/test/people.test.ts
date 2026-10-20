/**
 * SSR-render tests for the people page: the directory and one profile.
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { PeopleState } from "../src/people/state";
import PeopleApp from "../src/ui/People.svelte";

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
  member: {
    id: "dk_1",
    user: "alice",
    displayName: "Alice",
    bio: "Runs the reading group.",
    created: "2026-10-19T12:00:00Z",
  },
  node: { name: "my-group", members: 12 },
} as const;

const renderPeople = (state: PeopleState): PromiseLike<string> =>
  render(PeopleApp, {
    props: {
      people: writable<PeopleState>(state),
      onSignIn: noop,
      onSignOut: noop,
      onLoadMore: noop,
      onComment: async () => {},
      onDeleteComment: async () => {},
      onSave: noop,
    },
  }).then((out) => out.body);

describe("People.svelte", () => {
  it("renders the member directory", async () => {
    const html = await renderPeople({
      phase: "ready",
      members: [
        { id: "dk_1", user: "alice", displayName: "Alice", bio: "Runs the reading group.", created: "2026-10-19T12:00:00Z" },
        { id: "dk_2", user: "bob", displayName: "Bob", created: "2026-10-20T12:00:00Z" },
      ],
    });
    expect(html).toContain("Who's here");
    expect(html).toContain("Alice");
    expect(html).toContain("Bob");
    expect(html).toContain("2 members");
  });

  it("renders one member's profile and posts", async () => {
    const html = await renderPeople({
      phase: "ready",
      profile: {
        member: {
          id: "dk_1",
          user: "alice",
          displayName: "Alice",
          bio: "Runs the reading group.",
          created: "2026-10-19T12:00:00Z",
        },
        posts: [POST],
        total: 1,
      },
    });
    expect(html).toContain("Alice");
    expect(html).toContain("Runs the reading group.");
    expect(html).toContain("Sunday walk");
    expect(html).toContain("1 post");
  });

  it("offers an edit form only on your own profile", async () => {
    const html = await renderPeople({
      phase: "ready",
      me: ME,
      profile: {
        member: {
          id: "dk_1",
          user: "alice",
          displayName: "Alice",
          bio: "Runs the reading group.",
          created: "2026-10-19T12:00:00Z",
        },
        posts: [],
        total: 0,
      },
    });
    expect(html).toContain("Edit profile");
  });

  it("shows the empty profile state", async () => {
    const html = await renderPeople({
      phase: "ready",
      profile: {
        member: { id: "dk_1", user: "alice", displayName: "Alice", created: "2026-10-19T12:00:00Z" },
        posts: [],
        total: 0,
      },
    });
    expect(html).toContain("Nothing published yet.");
  });

  it("shows the not-found state", async () => {
    const html = await renderPeople({ phase: "error", error: "No such member on this node." });
    expect(html).toContain("No such member on this node.");
  });
});
