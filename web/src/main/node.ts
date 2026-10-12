/**
 * Node admin page — the only wiring point for this page.
 *
 * Everything shared with the site page (store, poll stream, fold, Svelte
 * mount, pagehide teardown) lives in `src/live.ts`; this file says only what
 * is specific to the node page: its element, cadence, wire shape, and view.
 */
import { startPage } from "../live";
import { Tunnels } from "../tunnels/schema";

import NodeApp from "../ui/Node.svelte";

const root = document.getElementById("app");
if (root === null) {
  throw new Error("node.html is missing #app");
}

// references/relay-ui.md: tunnel status polls at 5s.
startPage({
  target: root,
  pollMs: 5_000,
  component: NodeApp,
  tick: (client) => client.get("/api/tunnels", Tunnels),
});