/**
 * Client-side smoke test for the production bundles (no browser available).
 *
 * Usage: node tools/smoke.mjs <node|site|admin|bundle.js> <error|ready>
 *
 * `admin` is the mock admin page: it has only a `ready` scenario, and its
 * fetch stub rejects on call — the bundle must never call fetch.
 *        page/bundle paths are relative to web/ (node|site auto-resolve
 *        dist/assets/<page>-*.js; an explicit path may also be given)
 *
 * Stubs a minimal DOM + fetch, imports the entry bundle (which mounts the
 * Svelte component and starts the Effect poll stream), waits one tick, and
 * asserts the rendered HTML.
 */
import { readdirSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

// The web/ directory is this script's parent; bundle paths are resolved
// relative to it so the smoke test runs from anywhere.
const webDir = path.dirname(path.dirname(fileURLToPath(import.meta.url)));

// Resolve a page name (node|site|admin) to its entry bundle (dist/assets/<page>-*.js).
const resolveBundle = (page) => {
  const dir = path.join(webDir, "dist/assets");
  const hit = readdirSync(dir).find((f) => f.startsWith(`${page}-`) && f.endsWith(".js"));
  if (!hit) throw new Error(`no dist/assets/${page}-*.js — run \`npm run build\` first`);
  return path.join(dir, hit);
};
const [pageOrBundle, scenario] = process.argv.slice(2);
if (!pageOrBundle || !scenario)
  throw new Error("usage: node tools/smoke.mjs <node|site|admin|bundle.js> <error|ready>");
if (!/^(error|ready)$/.test(scenario)) throw new Error(`unknown scenario ${scenario} (use error|ready)`);
const isPage = (p) => p === "node" || p === "site" || p === "admin";
if (pageOrBundle === "admin" && scenario === "error")
  throw new Error("admin has no error scenario (the mock page never calls the API)");
const bundlePath = isPage(pageOrBundle)
  ? resolveBundle(pageOrBundle)
  : path.isAbsolute(pageOrBundle)
    ? pageOrBundle
    : path.join(webDir, pageOrBundle);

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

const TUNNELS = {
  tunnels: [
    { id: "t1", peer: "12D3KooW…abc", prefixes: ["2606:4700::/32"], online: true },
    { id: "t2", peer: "12D3KooW…def", prefixes: ["2001:db8::/32", "fd00::/8"], online: false },
  ],
};
const DEPLOYS = {
  current: { id: "d42", site: "example.com", version: "v42", deployedAt: "2026-10-01T12:00:00Z", status: "live" },
  versions: [
    { id: "d42", site: "example.com", version: "v42", deployedAt: "2026-10-01T12:00:00Z", status: "live" },
    { id: "d41", site: "example.com", version: "v41", deployedAt: "2026-09-20T09:30:00Z", status: "previous" },
  ],
};

let fetchImpl;
let fetchCalls = 0;
globalThis.fetch = (...args) => {
  fetchCalls++;
  return fetchImpl(...args);
};

const fail = () => Promise.reject(new TypeError("fetch failed"));
const ok = (body) => () => Promise.resolve({ ok: true, status: 200, json: async () => body });

// Derive the page (node|site|admin) from the bundle name if a path was given.
const page = isPage(pageOrBundle)
  ? pageOrBundle
  : path.basename(bundlePath).startsWith("site")
    ? "site"
    : path.basename(bundlePath).startsWith("admin")
      ? "admin"
      : "node";

// The admin page is a mock: any fetch call is a bug, so the stub rejects on
// call, and the assertions below require fetchCalls === 0.
fetchImpl =
  scenario === "error"
    ? fail
    : page === "node"
      ? ok(TUNNELS)
      : page === "site"
        ? ok(DEPLOYS)
        : () => Promise.reject(new Error("admin mock must not call fetch"));

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

if (scenario === "error") {
  assert(
    html.includes(page === "node" ? "Cannot reach the node:" : "Cannot reach the site console:"),
    "error heading rendered",
  );
  assert(html.includes("fetch failed"), "error detail rendered");
  assert(html.includes("Retrying."), "retry note rendered");
} else if (page === "node") {
  assert(html.includes("12D3KooW…abc"), "peer 1 rendered");
  assert(html.includes("2606:4700::/32"), "prefix rendered");
  assert(html.includes("offline"), "offline status rendered");
  assert(html.includes("1 of 2 tunnels online"), "live summary rendered");
} else if (page === "site") {
  assert(html.includes("example.com"), "site name rendered");
  assert(html.includes("v42") && html.includes("v41"), "versions rendered");
  assert(html.includes("previous"), "previous status rendered");
  assert(html.includes("2 site version(s) on record"), "live summary rendered");
} else if (page === "admin") {
  assert(html.includes("home-pi"), "node name rendered");
  assert(html.includes("Public friends"), "public friends panel rendered");
  assert(html.includes("Keys"), "keys panel rendered");
  assert(fetchCalls === 0, "fetch was never called");
}

console.log(`${label} ✓`);
process.exit(0);
