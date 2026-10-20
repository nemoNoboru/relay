// Static server for the Relay UI experiments.
// Serves the repository root so the mockups in ./ui-experiments/ can load the
// self-hosted fonts at ../web/fonts/*.woff2 relative to each file.
//
//   node ui-experiments/serve.mjs
//   open http://127.0.0.1:8787/ui-experiments/v1-quiet-feed.html
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize, resolve, sep } from "node:path";

const ROOT = resolve(process.argv[2] ?? ".");
const PORT = Number(process.argv[3] ?? process.env.PORT ?? 8787);

const TYPES = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".mjs": "text/javascript; charset=utf-8",
  ".woff2": "font/woff2",
  ".png": "image/png",
  ".svg": "image/svg+xml",
  ".json": "application/json",
};

createServer(async (req, res) => {
  try {
    let pathname = decodeURIComponent((req.url ?? "/").split("?")[0]);
    if (pathname === "/") pathname = "/index.html";
    const file = normalize(join(ROOT, pathname));
    if (file !== ROOT && !file.startsWith(ROOT + sep)) {
      res.writeHead(403).end("forbidden");
      return;
    }
    const data = await readFile(file);
    res.writeHead(200, {
      "content-type": TYPES[extname(file).toLowerCase()] ?? "application/octet-stream",
      "cache-control": "no-store",
    });
    res.end(data);
  } catch {
    res.writeHead(404).end("not found");
  }
}).listen(PORT, () => {
  console.log(`serving ${ROOT} at http://127.0.0.1:${PORT}`);
  console.log(`experiments:`);
  for (const v of ["v1-quiet-feed", "v2-warm-grid", "v3-zine-timeline", "v4-soft-sans"]) {
    console.log(`  http://127.0.0.1:${PORT}/ui-experiments/${v}.html`);
  }
});
