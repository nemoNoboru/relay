<script lang="ts">
  import type { Snippet } from "svelte";

  /**
   * Shared panel chrome (`.panel`), matching the static mockups.
   * Presentational only — no state, no data.
   */
  let {
    title,
    wide = false,
    children,
  }: { title: string; wide?: boolean; children: Snippet } = $props();
</script>

<section class={wide ? "panel wide" : "panel"}>
  <h3>{title}</h3>
  <div class="panel-body">
    {@render children()}
  </div>
</section>