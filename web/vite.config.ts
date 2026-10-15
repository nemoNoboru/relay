import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

/**
 * Relay owner UI (node admin + site console + admin mock).
 *
 * The built `dist/` is embedded into the Go node binary (see `embed.go`) and
 * served under an arbitrary path prefix — Relay's addressing is path-based.
 * Every asset URL must therefore be relative: keep `base: "./"`.
 *
 * There is no dev-server proxy on purpose: when the node API is unreachable,
 * the pages render their designed `stale`/`error` state (see
 * `.pi/skills/frontend-architecture/references/relay-ui.md`).
 */
export default defineConfig({
  plugins: [svelte()],
  base: "./",
  build: {
    outDir: "dist",
    emptyOutDir: true,
    target: "es2022",
    // No sourcemaps: they bloat the embedded binary and the pages are short.
    sourcemap: false,
    rollupOptions: {
      input: {
        node: "node.html",
        site: "site.html",
        admin: "admin.html",
        deploy: "deploy.html",
      },
    },
  },
});
