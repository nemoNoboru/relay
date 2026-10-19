<script lang="ts">
  import type { Post } from "../member/schema";
  import PostCard from "./PostCard.svelte";

  /** A chronological feed with a scroll sentinel and comments on each card. */
  let {
    posts,
    next,
    loading = false,
    error = "",
    onLoadMore,
    onComment,
    onDeleteComment,
  }: {
    posts: ReadonlyArray<Post>;
    next?: string;
    loading?: boolean;
    error?: string;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();

  let sentinel = $state<HTMLDivElement>();

  $effect(() => {
    if (!sentinel || !next || loading) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onLoadMore();
      },
      { rootMargin: "700px" },
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  });
</script>

<div class="feedlist">
  {#each posts as post (post.id)}
    <PostCard {post} {onComment} {onDeleteComment} />
  {/each}
</div>

{#if loading}
  <p class="hint feed-loading">Loading more…</p>
{/if}
{#if error}
  <p class="warn">{error}</p>
{/if}
{#if next}
  <div class="feed-sentinel" bind:this={sentinel} aria-hidden="true"></div>
{:else if posts.length > 0}
  <p class="feed-end">You're all caught up.</p>
{/if}
