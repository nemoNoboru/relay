import { readFileSync } from "node:fs";
import { expect, test } from "vitest";
import { assistantPrompt, body, cardTitle, inviteLink, readInvite, tar, toSite } from "../src/publish";

const enc = (s: string) => new TextEncoder().encode(s);
const text = (b: Uint8Array) => new TextDecoder().decode(b);
const file = (path: string, s = "x") => ({ path, data: enc(s) });

test("invite link keeps the key in the fragment", () =>
  expect(inviteLink("https://r.org", "k1")).toBe("https://r.org/_relay/publish.html#key=k1"));

test("invite gives back key and site", () =>
  expect(readInvite("#key=k1&site=blog")).toEqual({ key: "k1", site: "blog" }));

test("tar is header, padded data and two end blocks", () =>
  expect(tar([file("a.txt", "hi")]).length).toBe(2048));

test("tar header has name, octal size and ustar magic", () => {
  const h = text(tar([file("a.txt", "hi")]).slice(0, 512));
  expect([h.slice(0, 5), h.slice(124, 135), h.slice(257, 262)]).toEqual(["a.txt", "00000000002", "ustar"]);
});

test("tar checksum matches its header", () => {
  const h = tar([file("a.txt", "hi")]).slice(0, 512);
  const sum = h.reduce((s, b, i) => s + (i >= 148 && i < 156 ? 32 : b), 0);
  expect(parseInt(text(h.slice(148, 155)), 8)).toBe(sum);
});

test("markdown is rendered to html", () =>
  expect(text(toSite([file("notes.md", "# Hi")])[0]!.data)).toContain("<h1>Hi</h1>"));

test("a lone markdown file becomes the home page", () =>
  expect(toSite([file("notes.md")]).map((f) => f.path)).toEqual(["index.html"]));

test("other markdown pages keep their name", () =>
  expect(toSite([file("index.html"), file("about.md")]).map((f) => f.path)).toEqual(["index.html", "about.html"]));

test("a dropped folder loses its own name", () =>
  expect(toSite([file("blog/index.html"), file("blog/img/a.png")]).map((f) => f.path)).toEqual(["index.html", "img/a.png"]));

test("a dropped tar is sent as it is", () =>
  expect(text(body([file("site.tar", "raw")]))).toBe("raw"));

test("a rendered page carries the relay bar linking home", () => {
  const html = text(toSite([file("p.md", "# Hi")])[0]!.data);
  expect(html).toContain('class="relay-bar"');
  expect(html).toContain("Know how");
});

test("a rendered page carries the relay banner", () => {
  const html = text(toSite([file("p.md", "# Hi")])[0]!.data);
  expect(html).toContain("Hosted with Relay");
  expect(html).toContain('class="relay-banner"');
  expect(html).toContain("Share this page");
  expect(html).toContain("QR code");
  expect(html).toContain("Poster");
});

test("the card helper pulls the page title", () =>
  expect(cardTitle("# The Bakery\n\ntext")).toBe("The Bakery"));

test("the card helper falls back to a default title", () =>
  expect(cardTitle("no heading here")).toBe("A page on Relay"));

test("assistant prompt carries the upload line with server, key and site", () =>
  expect(assistantPrompt("https://r.org", "k1", "blog")).toContain(
    '-H "Authorization: Bearer k1" https://r.org/_relay/v1/sites/blog/deploy',
  ));

test("assistant prompt asks for a site name when there is none", () =>
  expect(assistantPrompt("https://r.org", "k1", "")).toContain("Ask me for a site name"));

test("publish page can greet someone with an invite", () =>
  expect(readFileSync("publish.html", "utf8")).toContain('id="invited"'));

test("publish page tells assistants where to upload", () =>
  expect(readFileSync("publish.html", "utf8")).toContain("/_relay/v1/sites/"));
