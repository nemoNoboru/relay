/**
 * Client-side smoke test for the built owner UI bundles (no browser).
 *
 * Usage:
 *   node tools/smoke.mjs admin ready
 *   node tools/smoke.mjs home signedout
 *   node tools/smoke.mjs home ready
 *
 * The host console renders the sign-in form on load and must not fetch until
 * a password is entered. The member home calls /_relay/v1/me on load; the
 * stub answers either 401 (signed out) or a member payload.
 *
 * Stubs a minimal DOM, imports the entry bundle (which mounts the Svelte
 * component), waits one tick, and asserts the rendered HTML.
 */
import { readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

// The web/ directory is this script's parent; bundle paths are resolved
// relative to it so the smoke test runs from anywhere.
const webDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));

// Resolve a page entry bundle (dist/assets/<page>-*.js).
const resolveBundle = (page) => {
  const dir = path.join(webDir, "dist/assets");
  const hit = readdirSync(dir).find((f) => f.startsWith(`${page}-`) && f.endsWith(".js"));
  if (!hit) throw new Error(`no dist/assets/${page}-*.js — run \`npm run build\` first`);
  return path.join(dir, hit);
};
const [page = "admin", scenario = "ready"] = process.argv.slice(2);
const allowed =
  (page === "admin" && scenario === "ready") ||
  (page === "home" && (scenario === "signedout" || scenario === "ready"));
if (!allowed) throw new Error("usage: node tools/smoke.mjs <admin ready|home signedout|home ready>");
const bundlePath = resolveBundle(page);

class El {
  constructor(tag, nodeType = 1) {
    this.tagName = tag.toUpperCase();
    this.nodeType = nodeType;
    this.children = [];
    this.attributes = new Map();
    this.parentNode = null;
    this._text = null;
  }
  // Real DOM aliases Svelte's runtime reads (e.g. the input `checked`
  // setter probes nodeName/namespaceURI before touching the property).
  get nodeName() {
    return this.tagName;
  }
  get namespaceURI() {
    return this.nodeType === 1 ? "http://www.w3.org/1999/xhtml" : null;
  }
  get ownerDocument() {
    return globalThis.document;
  }
  get firstChild() {
    return this.children[0] ?? null;
  }
  get lastChild() {
    return this.children[this.children.length - 1] ?? null;
  }
  get nextSibling() {
    if (this.parentNode === null) return null;
    const i = this.parentNode.children.indexOf(this);
    return this.parentNode.children[i + 1] ?? null;
  }
  get data() {
    return this.nodeType === 8 ? this._text : null;
  }
  set data(v) {
    if (this.nodeType === 8) this._text = String(v ?? "");
  }
  appendChild(node) {
    this._insert(node, null);
    return node;
  }
  append(...nodes) {
    for (const n of nodes) this.appendChild(n);
    return this;
  }
  insertBefore(node, ref) {
    this._insert(node, ref);
    return node;
  }
  _insert(node, ref) {
    if (node.parentNode !== null) node.parentNode.removeChild(node);
    if (ref === null) this.children.push(node);
    else this.children.splice(this.children.indexOf(ref), 0, node);
    node.parentNode = this;
  }
  before(node) {
    if (this.parentNode !== null) this.parentNode.insertBefore(node, this);
    return this;
  }
  after(node) {
    if (this.parentNode !== null) this.parentNode.insertBefore(node, this.nextSibling);
    return this;
  }
  removeChild(node) {
    const i = this.children.indexOf(node);
    if (i >= 0) this.children.splice(i, 1);
    node.parentNode = null;
    return node;
  }
  replaceChildren(...nodes) {
    for (const c of this.children) c.parentNode = null;
    this.children = [];
    for (const n of nodes) this.appendChild(n);
  }
  remove() {
    if (this.parentNode !== null) this.parentNode.removeChild(this);
  }
  cloneNode(deep) {
    const c = new El(this.tagName, this.nodeType);
    this.attributes.forEach((v, k) => c.setAttribute(k, v));
    c._text = this._text;
    if (deep) for (const ch of this.children) c.appendChild(ch.cloneNode(true));
    return c;
  }
  addEventListener() {}
  removeEventListener() {}
  setAttribute(name, value) {
    this.attributes.set(name, String(value));
  }
  removeAttribute(name) {
    this.attributes.delete(name);
  }
  hasAttribute(name) {
    return this.attributes.has(name);
  }
  getAttribute(name) {
    return this.attributes.has(name) ? this.attributes.get(name) : null;
  }
  get className() {
    return this.attributes.get("class") ?? "";
  }
  set className(v) {
    this.attributes.set("class", String(v));
  }
  get textContent() {
    if (this.nodeType === 3) return this._text;
    return this.children.map((c) => c.textContent).join("");
  }
  set textContent(v) {
    for (const c of this.children) c.parentNode = null;
    this.children = [];
    if (v !== null && v !== "") {
      const t = new El("#text", 3);
      t._text = String(v);
      t.parentNode = this;
      this.children = [t];
    }
    this._text = v === null ? null : String(v ?? "");
  }
  get innerHTML() {
    return this.children.map((c) => c.outerHTML()).join("");
  }
  set innerHTML(v) {
    // Only used for <template> elements (Svelte static-markup chunks).
    // Like a real <template>, the parsed tree stays inside `content`
    // (a fragment), not in the element's own children.
    this.children.forEach((c) => (c.parentNode = null));
    this.children = [];
    this._htmlSrc = String(v);
    const src = this._htmlSrc;
    if (src.length > 0) {
      this._content = parseHTML(src);
    }
  }
  get content() {
    return this._content ?? null;
  }
  get nodeValue() {
    return this.nodeType === 3 ? this._text : null;
  }
  set nodeValue(v) {
    if (this.nodeType === 3) this._text = String(v ?? "");
  }
  querySelectorAll(sel) {
    return [...this.all()].filter(
      (n) => n.nodeType === 1 && (sel === "*" || n.tagName === sel.toUpperCase()),
    );
  }
  *all() {
    for (const c of this.children) {
      yield c;
      yield* c.all();
    }
  }
  outerHTML() {
    if (this.nodeType === 3) return this._text;
    if (this.nodeType === 8) return `<!--${this._text}-->`;
    if (this.nodeType === 11) return this.children.map((c) => c.outerHTML()).join("");
    const attrs = [...this.attributes].map(([k, v]) => ` ${k}="${v.replace(/"/g, "&quot;")}"`).join("");
    const inner = this.children.map((c) => c.outerHTML()).join("");
    return `<${this.tagName.toLowerCase()}${attrs}>${inner}</${this.tagName.toLowerCase()}>`;
  }
}

// Minimal HTML parser for <template> innerHTML (Svelte static-markup chunks).
// Supports the subset these templates contain: element tags with simple
// attributes, comments (incl. `<!>` → `<!---->` anchors), and text.
const decodeEntities = (s) =>
  s
    .replace(/&#x([0-9a-f]+);/gi, (_, h) => String.fromCodePoint(parseInt(h, 16)))
    .replace(/&#(\d+);/g, (_, d) => String.fromCodePoint(Number(d)))
    .replace(/&quot;/g, '"')
    .replace(/&apos;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&");

function parseHTML(src) {
  const root = new El("#fragment");
  const stack = [root];
  const cur = () => stack[stack.length - 1];
  const VOID = new Set(["br", "hr", "img", "input", "meta", "link", "source", "area", "base", "col", "embed", "track", "wbr"]);
  let i = 0;
  while (i < src.length) {
    if (src[i] !== "<") {
      const next = src.indexOf("<", i);
      const end = next === -1 ? src.length : next;
      if (end > i) {
        const t = new El("#text", 3);
        t._text = decodeEntities(src.slice(i, end));
        cur().appendChild(t);
      }
      i = end;
      continue;
    }
    if (src.startsWith("<!--", i)) {
      const end = src.indexOf("-->", i + 4);
      cur().appendChild(new Comment(src.slice(i + 4, end === -1 ? src.length : end)));
      i = end === -1 ? src.length : end + 3;
    } else if (src[i + 1] === "!") {
      const end = src.indexOf(">", i);
      cur().appendChild(new Comment(src.slice(i + 2, end)));
      i = end + 1;
    } else if (src[i + 1] === "/") {
      const end = src.indexOf(">", i);
      const tag = src.slice(i + 2, end).trim().toUpperCase();
      for (let s = stack.length - 1; s > 0; s--) {
        if (stack[s].tagName === tag) {
          stack.length = s;
          break;
        }
      }
      i = end + 1;
    } else {
      const end = src.indexOf(">", i);
      let inner = src.slice(i + 1, end).trim();
      const selfClose = inner.endsWith("/");
      if (selfClose) inner = inner.slice(0, -1).trim();
      const m = inner.match(/^[a-zA-Z][\w-]*/);
      const tag = m[0].toLowerCase();
      const el = new El(tag);
      const rest = inner.slice(m[0].length);
      const attrRe = /([\w-]+)\s*=\s*"([^"]*)"/g;
      let am;
      while ((am = attrRe.exec(rest)) !== null) el.setAttribute(am[1], am[2]);
      cur().appendChild(el);
      if (!selfClose && !VOID.has(tag)) stack.push(el);
      i = end + 1;
    }
  }
  return root;
}

const app = new El("main");
app.setAttribute("id", "app");
const byId = { app };

globalThis.document = {
  getElementById: (id) => byId[id] ?? null,
  createElement: (tag) => new El(tag),
  createTextNode: (text) => {
    const n = new El("#text", 3);
    n._text = String(text);
    return n;
  },
  createComment: (text) => new Comment(text),
  createDocumentFragment: () => new El("#fragment"),
  importNode: (n) => n,
  querySelectorAll: () => [],
  activeElement: null,
  // Svelte's delegated event registration may attach to document as well.
  addEventListener: () => {},
  removeEventListener: () => {},
};

const listeners = {};
globalThis.window = {
  addEventListener: (name, fn) => {
    (listeners[name] ??= []).push(fn);
  },
  removeEventListener: (name, fn) => {
    listeners[name] = (listeners[name] ?? []).filter((f) => f !== fn);
  },
};
// The admin page keeps its token in sessionStorage between reloads.
globalThis.sessionStorage = {
  getItem: () => null,
  setItem: () => {},
  removeItem: () => {},
};
// The member home reads the invite fragment and may clean the URL.
globalThis.location = {
  origin: "http://localhost",
  pathname: "/_relay/home.html",
  search: "",
  hash: "",
};
globalThis.history = { replaceState: () => {} };
globalThis.MutationObserver = class {
  observe() {}
  disconnect() {}
};

// Svelte's mount() probes these globals (vh(): Element/Node/Text prototypes +
// navigator.userAgent for the Firefox class-attribute patch). `navigator`
// already exists as a global in Node 24.
class Comment extends El {
  constructor(data = "") {
    super("#comment", 8);
    this._text = String(data);
  }
}

globalThis.Node = El;
globalThis.Element = El;
globalThis.Text = El;
globalThis.Comment = Comment;
// Svelte's event-delegation cleanup probes `instanceof HTMLMediaElement`
// (for media listeners). Real buttons are not media elements; the stub
// class mirrors that so no smoke element matches it.
globalThis.HTMLMediaElement = class {};

const ME = {
  member: {
    id: "dk_1",
    user: "alice",
    displayName: "Alice",
    bio: "Runs the reading group.",
    created: "2026-10-19T12:00:00Z",
  },
  node: { name: "my-group", members: 12 },
};
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
};

let fetchCalls = 0;
globalThis.fetch = (..._args) => {
  fetchCalls++;
  if (page === "admin") {
    return Promise.reject(new Error("the host console must not fetch before sign-in"));
  }
  if (scenario === "signedout") {
    return Promise.resolve({
      ok: false,
      status: 401,
      statusText: "Unauthorized",
      json: async () => ({ error: { code: "session_required", message: "sign in" } }),
    });
  }
  const url = String(_args[0]);
  if (url.includes("/feed")) {
    return Promise.resolve({ ok: true, status: 200, statusText: "OK", json: async () => ({ posts: [POST] }) });
  }
  return Promise.resolve({ ok: true, status: 200, statusText: "OK", json: async () => ME });
};

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const dumpTree = (n, d = 0) => {
  const pad = "  ".repeat(d);
  const label =
    n.nodeType === 8
      ? `<!--${n._text}-->`
      : n.nodeType === 3
        ? `#text ${JSON.stringify((n._text ?? "").slice(0, 40))}`
        : `<${n.tagName.toLowerCase()}>`;
  console.error(pad + label + " children=" + (n.children ? n.children.length : 0));
  for (const c of n.children ?? []) dumpTree(c, d + 1);
};

await import(fileURLToPath(new URL(`file://${bundlePath}`)));
await sleep(200); // first tick is immediate; Svelte flushes on a microtask
if (process.env.SMOKE_TREE) dumpTree(app);

const html = app.outerHTML();
const label = `${page}-${scenario}`;
const assert = (cond, what) => {
  if (!cond) {
    console.error(`FAIL [${label}]: ${what}\n--- rendered ---\n${html}\n----------------`);
    dumpTree(app);
    process.exit(1);
  }
};

if (page === "admin") {
  assert(html.includes("Sign in"), "sign-in form rendered");
  assert(html.includes("admin password"), "admin password field rendered");
  assert(html.includes("Connect"), "connect button rendered");
  assert(fetchCalls === 0, "fetch was never called (no saved token)");
} else if (scenario === "signedout") {
  assert(html.includes("You need"), "invitation prompt rendered");
  assert(html.includes("an invitation"), "invitation copy rendered");
  assert(html.includes("Open my home"), "sign-in action rendered");
  assert(fetchCalls === 1, "the page asked for /me once");
} else {
  assert(html.includes("Welcome,"), "welcome heading rendered");
  assert(html.includes("Alice"), "member name rendered");
  assert(html.includes("my-group"), "node name rendered");
  assert(html.includes("Publish a page"), "publish action rendered");
  assert(html.includes("Sunday walk"), "feed post rendered");
  assert(fetchCalls === 2, "the page asked for /me and /feed once");
}

console.log(`${label} ✓`);
process.exit(0);
