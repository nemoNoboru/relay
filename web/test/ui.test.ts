/**
 * SSR-rendering tests for the Svelte page components.
 *
 * `render()` from `svelte/server` runs the real component code (blocks,
 * `@const`, each-iteration) without a browser, so every phase of the live
 * state machine gets rendered and asserted on as HTML. Client-side reactivity
 * is covered by the bundle smoke test instead (see web/README.md).
 */
import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import type { Component } from "svelte";
import type { Writable } from "svelte/store";
import { writable } from "svelte/store";

import type { LiveState } from "../src/fold";
import type { Deploys } from "../src/deploys/schema";
import type { Tunnels } from "../src/tunnels/schema";

import NodeApp from "../src/ui/Node.svelte";
import SiteApp from "../src/ui/Site.svelte";
// ---------------------------------------------------------------------------
// fixtures
// ---------------------------------------------------------------------------

const TUNNELS: Tunnels = {
  tunnels: [
    { id: "t1", peer: "12D3KooW…abc", prefixes: ["2606:4700::/32"], online: true },
    { id: "t2", peer: "12D3KooW…def", prefixes: ["2001:db8::/32", "fd00::/8"], online: false },
  ],
};

const EMPTY_TUNNELS: Tunnels = { tunnels: [] };

const DEPLOYS: Deploys = {
  current: {
    id: "d42",
    site: "example.com",
    version: "v42",
    deployedAt: "2026-10-01T12:00:00Z",
    status: "live",
  },
  versions: [
    {
      id: "d42",
      site: "example.com",
      version: "v42",
      deployedAt: "2026-10-01T12:00:00Z",
      status: "live",
    },
    {
      id: "d41",
      site: "example.com",
      version: "v41",
      deployedAt: "2026-09-20T09:30:00Z",
      status: "previous",
    },
  ],
};

type Phase = LiveState<unknown>["phase"];

/** A live state in any phase; non-loading phases carry a fixed timestamp. */
const state = <T>(phase: Phase, data?: T, lastError?: string): LiveState<T> => ({
  phase,
  data,
  updatedAt: phase === "loading" ? undefined : 1_760_000_000_000,
  lastError: lastError ?? undefined,
});

/**
 * SSR-render a page app with the given live state, returning the HTML body.
 * Both page apps take exactly one prop — the live store — so the shared
 * helper is typed on that shape (svelte-check sees the real prop types).
 */
const renderApp = <T>(
  App: Component<{ live: Writable<LiveState<T>> }>,
  liveState: LiveState<T>,
): Promise<string> =>
  Promise.resolve(render(App, { props: { live: writable<LiveState<T>>(liveState) } }).then((out) => out.body));

const renderNode = (s: LiveState<Tunnels>): Promise<string> => renderApp(NodeApp, s);
const renderSite = (s: LiveState<Deploys>): Promise<string> => renderApp(SiteApp, s);

// ---------------------------------------------------------------------------
// node page
// ---------------------------------------------------------------------------

describe("Node.svelte", () => {
  it("renders the loading phase", async () => {
    const html = await renderNode(state("loading"));
    expect(html).toContain("Connecting to the node…");
    expect(html).toContain("aria-live");
    expect(html).not.toContain("<table");
  });

  it("renders ready with tunnels as a table", async () => {
    const html = await renderNode(state("ready", TUNNELS));
    expect(html).toContain("12D3KooW…abc");
    expect(html).toContain("2606:4700::/32");
    expect(html).toContain("online");
    expect(html).toContain("offline");
    expect(html).toContain("1 of 2 tunnels online");
  });

  it("renders ready with no tunnels", async () => {
    const html = await renderNode(state("ready", EMPTY_TUNNELS));
    expect(html).toContain("No tunnels registered.");
    expect(html).not.toContain("<table");
  });

  it("renders stale: warning plus the last data", async () => {
    const html = await renderNode(state("stale", TUNNELS, "fetch failed"));
    expect(html).toContain("Connection lost — fetch failed");
    expect(html).toContain("12D3KooW…abc");
    expect(html).toContain("Stale — last update");
  });

  it("renders error without data", async () => {
    const html = await renderNode(state<Tunnels>("error", undefined, "cannot reach the API at /api/tunnels"));
    expect(html).toContain("Cannot reach the node: cannot reach the API at /api/tunnels");
    expect(html).toContain("Retrying.");
    expect(html).not.toContain("<table");
  });
});

// ---------------------------------------------------------------------------
// site page
// ---------------------------------------------------------------------------

describe("Site.svelte", () => {
  it("renders the loading phase", async () => {
    const html = await renderSite(state("loading"));
    expect(html).toContain("Connecting…");
    expect(html).not.toContain("statrow");
  });

  it("renders ready: stats, site name, and the version table", async () => {
    const html = await renderSite(state("ready", DEPLOYS));
    expect(html).toContain("example.com");
    expect(html).toContain("v42");
    expect(html).toContain("v41");
    expect(html).toContain("live");
    expect(html).toContain("previous");
    expect(html).toContain("2 site version(s) on record");
  });

  it("renders ready with no versions", async () => {
    const html = await renderSite(state("ready", { current: undefined, versions: [] }));
    expect(html).toContain("No deployments on record.");
  });

  it("renders stale: warning plus the last data", async () => {
    const html = await renderSite(state("stale", DEPLOYS, "fetch failed"));
    expect(html).toContain("Connection lost — fetch failed");
    expect(html).toContain("v42");
  });

  it("renders error without data", async () => {
    const html = await renderSite(state<Deploys>("error", undefined, "cannot reach the API at /api/site/deploy"));
    expect(html).toContain("Cannot reach the site console: cannot reach the API at /api/site/deploy");
    expect(html).toContain("Retrying.");
  });
});