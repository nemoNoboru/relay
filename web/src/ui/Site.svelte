<script lang="ts">
  import type { Writable } from "svelte/store";

  import { ago } from "../fold";
  import type { LiveState } from "../fold";
  import type { Deploys } from "../deploys/schema";

  import Panel from "./Panel.svelte";
  import Phase from "./Phase.svelte";

  /**
   * Site console page.
   *
   * Pure rendering of `live` — the component's only reactive input. Polling,
   * decoding, and the state fold live in `src/live.ts`; the phase
   * presentation (loading/error/stale) lives in `Phase.svelte`; this file
   * decides only what the deploy data looks like.
   */
  let { live }: { live: Writable<LiveState<Deploys>> } = $props();

  /** "live" for the current deploy, "previous" for the rest. */
  const versionStatus = (currentId: string | undefined, id: string): string =>
    currentId === undefined ? "unknown" : id === currentId ? "live" : "previous";

  /** Text for the sr-only aria-live region. Pure in (state, now). */
  const summarize = (s: LiveState<Deploys>, now: number): string => {
    const count = s.data?.versions.length ?? 0;
    switch (s.phase) {
      case "loading":
        return "Loading site console";
      case "ready":
        return `${count} site version(s) on record. Updated ${ago(s.updatedAt, now)}.`;
      case "stale":
        return `Stale — last update ${ago(s.updatedAt, now)}. ${count} site version(s) on record.`;
      case "error":
        return `Cannot reach the site console. ${s.lastError ?? ""}`.trim();
    }
  };
</script>

<div class="siteapp">
  <div class="wrap">
    <div class="site-head">
      <h1>{$live.data?.current?.site ?? "Site"}</h1>
      <p class="site-sub">Relay site console</p>
    </div>

    <div class="panels">
      {#if $live.data !== undefined}
        {@const data = $live.data}
        <Panel title="Site">
          <div class="statrow">
            <div class="stat"><b>{data.versions.length}</b><span>Versions</span></div>
            <div class="stat"><b>{data.current?.version ?? "—"}</b><span>Live version</span></div>
            <div class="stat"><b>{data.current?.site ?? "—"}</b><span>Site</span></div>
            <div class="stat"><b>{ago($live.updatedAt, Date.now())}</b><span>Last update</span></div>
          </div>
        </Panel>
      {/if}

      <Panel title="Deployments" wide>
        <Phase state={$live} loading="Connecting…" error="Cannot reach the site console">
          {#snippet children({ data })}
            {#if data.versions.length === 0}
              <p class="hint">No deployments on record.</p>
            {:else}
              <table class="tbl">
                <thead>
                  <tr>
                    <th>Version</th>
                    <th>Deployed</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {#each data.versions as version (version.id)}
                    <tr>
                      <td><code>{version.version}</code></td>
                      <td>{new Date(version.deployedAt).toISOString()}</td>
                      <td><code>{versionStatus(data.current?.id, version.id)}</code></td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            {/if}
          {/snippet}
        </Phase>
      </Panel>
    </div>

    <div class="sr-only" aria-live="polite">{summarize($live, Date.now())}</div>
  </div>
</div>