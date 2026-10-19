/**
 * SSR-render tests for the profile page (`/_relay/people/<id>`).
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { PeopleState } from "../src/people/state";
import PeopleApp from "../src/ui/People.svelte";

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

const renderPeople = (state: PeopleState): PromiseLike<string> =>
  render(PeopleApp, {
    props: {
      people: writable<PeopleState>(state),
      onLoadMore: () => {},
      onComment: async () => {},
      onDeleteComment: async () => {},
    },
  }).then((out) => out.body);

describe("People.svelte", () => {
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

  it("shows the empty profile state", async () => {
    const html = await renderPeople({
      phase: "ready",
      profile: {
        member: {
          id: "dk_1",
          user: "alice",
          displayName: "Alice",
          created: "2026-10-19T12:00:00Z",
        },
        posts: [],
        total: 0,
      },
    });
    expect(html).toContain("Nothing published yet.");
  });

  it("shows the signed-out state", async () => {
    const html = await renderPeople({ phase: "signedOut" });
    expect(html).toContain("an invitation");
  });
});
