<script lang="ts">
  import type { Writable } from "svelte/store";

  import { ago } from "../fold";
  import type { LiveState } from "../fold";
  import type { Tunnels } from "../tunnels/schema";

  import Panel from "./Panel.svelte";
  import Phase from "./Phase.svelte";

  /**
   * Node admin page.
   *
   * Pure rendering of `live` — the component's only reactive input. Polling,
   * decoding, and the state fold live in `src/live.ts`; the phase
   * presentation (loading/error/stale) lives in `Phase.svelte`; this file
   * decides only what the tunnel data looks like.
   */
  let { live }: { live: Writable<LiveState<Tunnels>> } = $props();

  /** Text for the sr-only aria-live region. Pure in (state, now). */
  const summarize = (s: LiveState<Tunnels>, now: number): string => {
    const total = s.data?.tunnels.length ?? 0;
    const online = s.data?.tunnels.filter((t) => t.online).length ?? 0;
    const base = `${online} of ${total} tunnels online`;
    switch (s.phase) {
      case "loading":
        return "Connecting to the node";
      case "ready":
        return `${base}. Updated ${ago(s.updatedAt, now)}.`;
      case "stale":
        return `Stale — last update ${ago(s.updatedAt, now)}. ${base}.`;
      case "error":
        return `Cannot reach the node. ${s.lastError ?? ""}`.trim();
    }
  };
</script>

<div class="nodeapp">
  <div class="wrap">
    <div class="node-head">
      <h1>Node</h1>
      <p class="node-sub">Relay node admin</p>
    </div>

    <div class="panels">
      <Panel title="Tunnels" wide>
        <Phase state={$live} loading="Connecting to the node…" error="Cannot reach the node">
          {#snippet children({ data, updatedAt })}
            {#if data.tunnels.length === 0}
              <p class="hint">No tunnels registered.</p>
            {:else}
              <table class="tbl">
                <thead>
                  <tr>
                    <th>Peer</th>
                    <th>Prefixes</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {#each data.tunnels as tunnel (tunnel.id)}
                    <tr>
                      <td><code>{tunnel.peer}</code></td>
                      <td><code>{tunnel.prefixes.join(" · ")}</code></td>
                      <td><code>{tunnel.online ? "online" : "offline"}</code></td>
                    </tr>
                  {/each}
                </tbody>
              </table>
              <p class="hint">Updated {ago(updatedAt, Date.now())}.</p>
            {/if}
          {/snippet}
        </Phase>
      </Panel>
    </div>

    <div class="sr-only" aria-live="polite">{summarize($live, Date.now())}</div>
  </div>
</div>