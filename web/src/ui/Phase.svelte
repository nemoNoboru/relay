<script lang="ts" generics="T">
  import type { Snippet } from "svelte";

  import { ago } from "../fold";
  import type { LiveState } from "../fold";

  /**
   * The shared presentation of the live state machine
   * (references/relay-ui.md): loading, error, and stale (with its warning
   * plus the last data) are identical on every polled page, so they live
   * here. The page decides only two things: the copy (`loading`, `error`)
   * and how its decoded data looks (`children`, invoked with the data and
   * the `updatedAt` of the last successful poll).
   */
  let {
    state,
    loading,
    error,
    children,
  }: {
    /** The current live state (a reactive prop — re-renders on every fold). */
    state: LiveState<T>;
    /** Shown while `loading`, e.g. "Connecting to the node…". */
    loading: string;
    /** Prefix of the `error` message, e.g. "Cannot reach the node". */
    error: string;
    /** The page's data view, rendered only when data is present. */
    children: Snippet<[{ data: T; updatedAt: number | undefined }]>;
  } = $props();
</script>

{#if state.phase === "loading"}
  <p class="hint">{loading}</p>
{:else if state.phase === "error"}
  <div class="warn">
    {error}: {state.lastError ?? "unknown error"}. Retrying.
  </div>
{:else}
  {#if state.phase === "stale"}
    <div class="warn">
      Connection lost — {state.lastError ?? "unknown error"}. Last update
      {ago(state.updatedAt, Date.now())}; retrying.
    </div>
  {/if}
  {#if state.data === undefined}
    <p class="hint">No data yet.</p>
  {:else}
    {@render children({ data: state.data, updatedAt: state.updatedAt })}
  {/if}
{/if}