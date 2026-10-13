/**
 * SSR render tests for the Relay admin mock page (spec: PLAN.md, AC-1..AC-10).
 *
 * The admin page is a mock: it renders a fixed fixture, never calls the API,
 * and has no live state. `render()` from `svelte/server` runs the real
 * component code, so every panel is asserted as HTML. The production bundle
 * (and its never-fetch guarantee) is covered by the `admin` scenario in
 * `tools/smoke.mjs`.
 *
 * Copy is pinned by these tests; changing mock copy requires changing these
 * assertions in the same commit.
 */
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { describe, it } from "@effect/vitest";
import { expect } from "vitest";
import { render } from "svelte/server";

import { ADMIN_FIXTURE, EMPTY_ADMIN, type AdminData } from "../src/admin/fixture";
import AdminApp from "../src/ui/Admin.svelte";

/** The web/ directory (this file lives in web/test/). */
const webDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

/** SSR-render the admin app with the given data, returning the HTML body. */
const renderAdmin = (data: AdminData): PromiseLike<string> =>
  render(AdminApp, { props: { data } }).then((out) => out.body);

/** All `<tr>…</tr>` rows of an HTML document. */
const rows = (html: string): string[] => html.match(/<tr[^>]*>[\s\S]*?<\/tr>/g) ?? [];

// ---------------------------------------------------------------------------
// AC-1 — the shell
// ---------------------------------------------------------------------------

describe("admin.html shell", () => {
  it("the admin shell is a static mock page", () => {
    const shell = readFileSync(path.join(webDir, "admin.html"), "utf8");
    expect(shell).toContain("<title>Admin — Relay</title>");
    expect(shell).toContain('<link rel="stylesheet" href="./shared/styles.css">');
    expect(shell).toContain('<span class="stamp">Relay admin · Mock</span>');
    expect(shell).toContain('src="./src/main/admin.ts"');
    // exactly one stylesheet link and exactly one script tag (the entry)
    expect(shell.match(/<link rel="stylesheet"/g)).toHaveLength(1);
    expect(shell.match(/<script/g)).toHaveLength(1);
    // static mock: <main id="app"> without live-update semantics
    expect(shell).toMatch(/<main id="app">\s*<\/main>/);
    expect(shell).not.toContain("aria-busy");
  });
});

// ---------------------------------------------------------------------------
// AC-2 — head
// ---------------------------------------------------------------------------

describe("Admin.svelte head", () => {
  it("renders the head (name, visibility, mock note)", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h1>home-pi</h1>");
    expect(html).toContain("Private");
    expect(html).toContain("behind NAT");
    expect(html).toContain("Mock data · no functionality");
  });
});

// ---------------------------------------------------------------------------
// AC-3 — public friends
// ---------------------------------------------------------------------------

describe("Admin.svelte public friends", () => {
  it("renders the public friends picker", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h3>Public friends</h3>");
    expect(html).toContain("not reachable from the open internet");
    // three radios, all in the public-friend group, exactly one selected
    expect(html.match(/type="radio"/g)).toHaveLength(3);
    expect(html.match(/name="public-friend"/g)).toHaveLength(3);
    expect(html.match(/checked/g)).toHaveLength(1);
    expect(html).toContain("Add a public friend");
    expect(html).toContain("relay.cloudpunk.org/raspberry/");
    expect(html).toContain("<button class=\"btn solid\">Connect</button>");
  });
});

// ---------------------------------------------------------------------------
// AC-4 — friends
// ---------------------------------------------------------------------------

describe("Admin.svelte friends", () => {
  it("renders the friends list (prefixes, status)", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h3>Friends</h3>");
    // each friend's prefix renders as its own code chip
    expect(html).toContain("<code>/raspberry/</code>");
    expect(html).toContain("<code>/bookshelf/</code>");
    expect(html).toContain("<code>/workshop/</code>");
    expect(html).toContain("connected");
    expect(html).toContain("offline");
    expect(html).toContain("<button class=\"btn\">Invite a node</button>");
  });

  it("renders the friends empty state", async () => {
    const html = await renderAdmin(EMPTY_ADMIN);
    expect(html).toContain("No friends connected.");
    expect(html).not.toContain('class="friend"');
  });
});

// ---------------------------------------------------------------------------
// AC-5 — key creation (form + masked key file)
// ---------------------------------------------------------------------------

describe("Admin.svelte keys", () => {
  it("renders the key form and a masked key file", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h3>Keys</h3>");
    expect(html).toMatch(/<label[^>]*for="key-name"[^>]*>/);
    const input = html.match(/<input[^>]*id="key-name"[^>]*>/);
    expect(input).not.toBeNull();
    expect(input?.[0]).toContain('type="text"');
    expect(html).toMatch(/<button[^>]*>Create key<\/button>/);
    // the key file carries the expected fields
    expect(html).toContain('"relay_key"');
    expect(html).toContain('"name"');
    expect(html).toContain('"node"');
    expect(html).toContain('"public_key"');
    expect(html).toContain('"private_key"');
    // the private half is masked
    const pk = html.match(/"private_key"\s*:\s*"([^"]*)"/);
    expect(pk).not.toBeNull();
    expect(pk?.[1]).toContain("•");
    // and no unmasked key material (40+ char token) appears anywhere on the page
    expect(html).not.toMatch(/[A-Za-z0-9_-]{40,}/);
  });
});

// ---------------------------------------------------------------------------
// AC-6 — no-recovery warning
// ---------------------------------------------------------------------------

describe("Admin.svelte no-recovery warning", () => {
  it("renders the no-recovery warning", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("shown once, stored nowhere, and impossible to recover");
  });
});

// ---------------------------------------------------------------------------
// AC-7 — issued keys
// ---------------------------------------------------------------------------

describe("Admin.svelte issued keys", () => {
  it("renders issued keys (claimed and open)", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    for (const th of ["name", "key", "created", "status"]) {
      expect(html).toContain(`<th>${th}</th>`);
    }
    const trs = rows(html);
    expect(trs.some((r) => r.includes("bookshelf") && r.includes("claimed"))).toBe(true);
    expect(trs.some((r) => r.includes("zine") && r.includes("open"))).toBe(true);
    expect(html).toContain("9f2c…c41a");
    expect(html).toContain("b7e1…08fd");
  });

  it("renders the keys empty state", async () => {
    const html = await renderAdmin(EMPTY_ADMIN);
    expect(html).toContain("No keys issued.");
    expect(html).not.toContain("<table");
  });
});

// ---------------------------------------------------------------------------
// AC-8 — claimed names
// ---------------------------------------------------------------------------

describe("Admin.svelte names", () => {
  it("renders the claimed names table", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h3>Names</h3>");
    expect(html).toContain("relay.cloudpunk.org/raspberry/bookshelf/");
    expect(html).toContain("9f2c…c41a");
    expect(html).toContain("1.2 MB");
  });

  it("renders the names empty state", async () => {
    const html = await renderAdmin(EMPTY_ADMIN);
    expect(html).toContain("No names claimed on this node.");
    expect(html).not.toContain("<table");
  });
});

// ---------------------------------------------------------------------------
// AC-9 — rules + node capacity
// ---------------------------------------------------------------------------

describe("Admin.svelte rules and node", () => {
  it("renders rules and node capacity", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toContain("<h3>Rules</h3>");
    expect(html.match(/class="rule( |")/g)).toHaveLength(3);
    expect(html).toContain("Allow");
    expect(html).toContain("Review");
    expect(html).toContain("Deny");
    expect(html).toContain("rules/community.txt");
    expect(html).toContain("<h3>Node</h3>");
    expect(html).toContain('class="statrow"');
    expect((html.match(/class="stat"/g) ?? []).length).toBeGreaterThanOrEqual(4);
    expect(html).toContain("12/64");
    expect(html).toContain("41 d");
  });
});

// ---------------------------------------------------------------------------
// AC-10 — shared design vocabulary
// ---------------------------------------------------------------------------

describe("Admin.svelte design vocabulary", () => {
  it("uses the shared design vocabulary", async () => {
    const html = await renderAdmin(ADMIN_FIXTURE);
    expect(html).toMatch(/class="panel( |")/);
    expect(html).toContain('class="statrow"');
    expect(html).toMatch(/class="(?:tbl|friends)/);
    expect(html).toContain('class="hint"');
    expect(html).toContain('class="warn"');
    expect(html).toContain('class="draft-note"');
    expect(html).toContain("Mock data · no functionality");
    expect(html).toContain("2 of 3 friends connected. 2 keys issued, 1 name claimed.");
  });
});
