/**
 * Publishing from the browser: invite links, Markdown pages, and the tar
 * archive the deploy endpoint expects.
 */
import { marked } from "marked";

export interface SiteFile {
  readonly path: string;
  readonly data: Uint8Array;
}

const enc = new TextEncoder();
const dec = new TextDecoder();

/** The key rides in the fragment, which browsers never send to the server. */
export const inviteLink = (origin: string, key: string): string =>
  `${origin}/_relay/publish.html#${new URLSearchParams({ key })}`;

export const readInvite = (hash: string) => {
  const p = new URLSearchParams(hash.replace(/^#/, ""));
  return { key: p.get("key") ?? "", site: p.get("site") ?? "" };
};

/** A message to paste into Claude or another assistant, ready to run. */
export const assistantPrompt = (origin: string, key: string, site: string): string =>
  [
    "Please publish my website with Relay.",
    "",
    "Upload my site's folder as a tar archive with this command:",
    "",
    `tar -cf - -C SITE_FOLDER . | curl --data-binary @- -H "Authorization: Bearer ${key}" ${origin}/_relay/v1/sites/${site || "SITE_NAME"}/deploy`,
    "",
    "- SITE_FOLDER is the folder with my site; it should have an index.html. Turn any Markdown files into HTML pages first.",
    site ? `- The site name is ${site}.` : "- Ask me for a site name: lowercase letters, numbers and hyphens.",
    `- The JSON reply's "url" field is the address. Give me the full link: ${origin} followed by that url.`,
    "- Every upload replaces the whole site.",
    "- The key after Bearer is private. Don't repeat it back to me or put it anywhere else.",
  ].join("\n");

const octal = (n: number, width: number) => n.toString(8).padStart(width, "0") + "\0";

const header = (path: string, size: number): Uint8Array => {
  const h = new Uint8Array(512);
  const put = (s: string, at: number) => h.set(enc.encode(s), at);
  put(path, 0);
  put(octal(0o644, 7), 100);
  put(octal(size, 11), 124);
  put(octal(Math.floor(Date.now() / 1000), 11), 136);
  put("        0", 148); // checksum counts as spaces; then typeflag "0" (file)
  put("ustar\x0000", 257);
  put(octal(h.reduce((s, b) => s + b, 0), 6) + " ", 148);
  return h;
};

/** A ustar archive. Paths must fit the 100-byte name field. */
export const tar = (files: ReadonlyArray<SiteFile>): Uint8Array => {
  const blocks = files.flatMap((f) => [header(f.path, f.data.length), f.data, new Uint8Array(-f.data.length & 511)]);
  const out = new Uint8Array(blocks.reduce((n, b) => n + b.length, 1024));
  blocks.reduce((at, b) => (out.set(b, at), at + b.length), 0);
  return out;
};

const escape = (s: string) => s.replace(/&/g, "&amp;").replace(/</g, "&lt;");

const STYLE = `body{margin:0;background:#f0e6d2;color:#141210;font:18px/1.6 Georgia,serif}
main{max-width:40rem;margin:0 auto;padding:3rem 1.25rem}
h1,h2,h3{font-family:Impact,"Arial Narrow",sans-serif;line-height:1.1;text-transform:uppercase}
h1{font-size:2.6rem;border-bottom:6px solid #c8202f;padding-bottom:.3rem}
a{color:#c8202f}img{max-width:100%}
pre{overflow:auto;background:#141210;color:#f0e6d2;padding:1rem}
blockquote{margin:0;padding-left:1rem;border-left:4px solid #c8202f}`;

const page = (md: string) =>
  `<!doctype html><html lang="en"><head><meta charset="utf-8">` +
  `<meta name="viewport" content="width=device-width, initial-scale=1">` +
  `<title>${escape(md.match(/^#\s+(.+)$/m)?.[1] ?? "Page")}</title><style>${STYLE}</style></head>` +
  `<body><main>${marked.parse(md, { async: false })}</main></body></html>`;

/** Drop a folder and its own name is not part of the site. */
const unwrap = (files: ReadonlyArray<SiteFile>) => {
  const top = files[0]?.path.split("/")[0] + "/";
  return files.every((f) => f.path.startsWith(top)) ? files.map((f) => ({ ...f, path: f.path.slice(top.length) })) : files;
};

/** Markdown becomes pages; a lone Markdown file becomes the home page. */
export const toSite = (dropped: ReadonlyArray<SiteFile>): ReadonlyArray<SiteFile> => {
  const files = unwrap(dropped);
  const isMd = (f: SiteFile) => f.path.endsWith(".md");
  const home = files.filter(isMd).length === 1 && !files.some((f) => f.path === "index.html");
  return files.map((f) =>
    isMd(f) ? { path: home ? "index.html" : f.path.replace(/\.md$/, ".html"), data: enc.encode(page(dec.decode(f.data))) } : f,
  );
};

/** What to upload: a dropped tar as it is, anything else packed. */
export const body = (files: ReadonlyArray<SiteFile>): Uint8Array => {
  const [only] = files;
  return files.length === 1 && only?.path.endsWith(".tar") ? only.data : tar(toSite(files));
};
