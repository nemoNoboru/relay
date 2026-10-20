import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

/**
 * Relay owner UI: member home, profiles, post wrappers, host console, publish, deploy.
 *
 * The built `dist/` is embedded into the Go node binary (see `embed.go`) and
 * always mounted at `/_relay/`, including nested shells such as
 * `/_relay/people/<id>`. Assets therefore use the fixed mount path, not a
 * relative base: keep `base: "/_relay/"`.
 *
 * There is no dev-server proxy on purpose: the host console renders its
 * sign-in state, and the publish page needs no API until a submit (see
 * `.pi/skills/frontend-architecture/references/relay-ui.md`).
 */
export default defineConfig({
  plugins: [svelte()],
  base: "/_relay/",
  build: {
    outDir: "dist",
    emptyOutDir: true,
    target: "es2022",
    // No sourcemaps: they bloat the embedded binary and the pages are short.
    sourcemap: false,
    rollupOptions: {
      input: {
        feed: "feed.html",
        people: "people.html",
        post: "post.html",
        admin: "admin.html",
        deploy: "deploy.html",
        publish: "publish.html",
      },
    },
  },
});
