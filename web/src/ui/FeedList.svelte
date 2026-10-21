<script lang="ts">
  import type { Snippet } from "svelte";

  import type { Post } from "../member/schema";
  import PostCard from "./PostCard.svelte";

  /** A chronological feed with a scroll sentinel and comments on each card. */
  let {
    posts,
    next,
    loading = false,
    error = "",
    signedIn = false,
    viewerId,
    onLoadMore,
    onComment,
    onDeleteComment,
    onDeletePost,
    children,
  }: {
    posts: ReadonlyArray<Post>;
    next?: string;
    loading?: boolean;
    error?: string;
    signedIn?: boolean;
    viewerId?: string;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
    onDeletePost?: (site: string) => Promise<void>;
    children?: Snippet;
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

<div class="feed">
  {#if children}{@render children()}{/if}
  {#each posts as post (post.id)}
    <PostCard {post} {signedIn} {viewerId} {onComment} {onDeleteComment} {onDeletePost} />
  {/each}
</div>

{#if loading}
  <p class="feed-loading">Loading more…</p>
{/if}
{#if error}
  <p class="warn">{error}</p>
{/if}
{#if next}
  <div class="feed-sentinel" bind:this={sentinel} aria-hidden="true"></div>
{:else if posts.length > 0}
  <p class="feed-end">You're all caught up.</p>
{/if}
