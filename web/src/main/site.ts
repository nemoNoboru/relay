/**
 * Site console page — the only wiring point for this page.
 *
 * Everything shared with the node page (store, poll stream, fold, Svelte
 * mount, pagehide teardown) lives in `src/live.ts`; this file says only what
 * is specific to the site page: its element, cadence, wire shape, and view.
 */
import { startPage } from "../live";
import { Deploys } from "../deploys/schema";

import SiteApp from "../ui/Site.svelte";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("site.html is missing #app");
}

// references/relay-ui.md: deploy status polls at 10s.
startPage({
  target: root,
  pollMs: 10_000,
  component: SiteApp,
  tick: (client) => client.get("/api/site/deploy", Deploys),
});