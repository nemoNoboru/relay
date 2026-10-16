/**
 * SSR render tests for the live Relay admin page (`admin.html`).
 *
 * The page renders the `admin` store (Admin.svelte); the API effects live in
 * `main/admin.ts` and are not exercised here. These tests pin the three
 * render states — locked (sign-in), ready (keys + sites), and the one-time
 * token reveal — using `svelte/server` `render()`, which runs the real
 * component code.
 */
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";
import { writable } from "svelte/store";

import type { AdminOverview } from "../src/admin/schema";
import type { AdminState } from "../src/admin/state";
import AdminApp from "../src/ui/Admin.svelte";

const webDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

const noop = (): void => {};

const renderAdmin = (state: AdminState): PromiseLike<string> =>
  render(AdminApp, {
    props: {
      admin: writable<AdminState>(state),
      onConnect: noop,
      onCreate: noop,
      onRevoke: noop,
      onRotate: noop,
      onRefresh: noop,
      onDismiss: noop,
      onLogout: noop,
    },
  }).then((out) => out.body);

const locked = (): AdminState => ({ token: "", phase: "locked", busy: false });

const OVERVIEW: AdminOverview = {
  keys: [
    { id: "dk_1", user: "alice", token: "tok-alice-1234567890", created: "2026-01-01T00:00:00Z", revoked: false },
    { id: "dk_2", user: "bob", token: "tok-bob-9876543210", created: "2026-01-02T00:00:00Z", revoked: true },
  ],
  sites: [
    { site: "alice-blog", user: "alice" },
    { site: "zine", user: "alice" },
    { site: "bob-site", user: "bob" },
  ],
};

const ready = (overview: AdminOverview, extra: Partial<AdminState> = {}): AdminState => ({
  token: "admin",
  phase: "ready",
  busy: false,
  overview,
  ...extra,
});

describe("admin.html shell", () => {
  it("is the operator console shell", () => {
    const shell = readFileSync(path.join(webDir, "admin.html"), "utf8");
    expect(shell).toContain("<title>Admin — Relay</title>");
    expect(shell).toContain('<span class="stamp">Relay operator</span>');
    expect(shell).toContain('src="./src/main/admin.ts"');
  });
});

describe("Admin.svelte sign-in", () => {
  it("renders the token form when locked", async () => {
    const html = await renderAdmin(locked());
    expect(html).toContain("Sign in");
    expect(html).toContain("admin token");
    expect(html).toContain("Connect");
    expect(html).not.toContain("Create key");
    expect(html).not.toContain("Sites by user");
  });

  it("renders the rejection reason when the token is bad", async () => {
    const html = await renderAdmin(ready(OVERVIEW, { phase: "error", error: "bad admin token" }));
    expect(html).toContain("Sign in");
    expect(html).toContain("bad admin token");
  });
});

describe("Admin.svelte ready", () => {
  it("renders deploy keys grouped by user, with status and actions", async () => {
    const html = await renderAdmin(ready(OVERVIEW));
    expect(html).toContain("Deploy keys");
    expect(html).toContain("alice");
    expect(html).toContain("bob");
    expect(html).toContain("active");
    expect(html).toContain("revoked");
    expect(html).toContain("Rotate");
    expect(html).toContain("Revoke");
    expect(html).toContain("Create key");
  });

  it("masks tokens until revealed", async () => {
    const html = await renderAdmin(ready(OVERVIEW));
    expect(html).toContain("Reveal");
    expect(html).toContain("••••••••••••••");
    // plaintext tokens stay hidden in the initial (masked) render
    expect(html).not.toContain("tok-alice-1234567890");
    expect(html).not.toContain("tok-bob-9876543210");
  });

  it("groups sites by user", async () => {
    const html = await renderAdmin(ready(OVERVIEW));
    expect(html).toContain("Sites by user");
    expect(html).toContain("alice-blog");
    expect(html).toContain("zine");
    expect(html).toContain("bob-site");
    expect(html).toContain("2 sites");
    expect(html).toContain("1 site");
  });

  it("renders the empty states", async () => {
    const html = await renderAdmin(ready({ keys: [], sites: [] }));
    expect(html).toContain("No keys issued.");
    expect(html).toContain("No sites deployed yet.");
  });
});

describe("Admin.svelte one-time token", () => {
  it("shows the new token once with copy and dismiss", async () => {
    const html = await renderAdmin(
      ready(OVERVIEW, { newToken: "sekret-token", notice: "Key for alice created" }),
    );
    expect(html).toContain("New deploy key — shown once");
    expect(html).toContain("Key for alice created");
    expect(html).toContain("sekret-token");
    expect(html).toContain("Copy");
    expect(html).toContain("Done");
  });
});
