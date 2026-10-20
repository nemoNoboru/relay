<script lang="ts">
  import type { Post } from "../member/schema";

  /** The little page preview for a text template. The whole peek is the link. */
  let { post }: { post: Post } = $props();
</script>

<a class="peek" href={post.postUrl}>
  <span class="ptitle">{post.title}</span>
  {#if post.summary}
    <p>{post.summary}</p>
  {/if}
  <span class="skeleton w80"></span>
  <span class="skeleton w60"></span>
</a>
