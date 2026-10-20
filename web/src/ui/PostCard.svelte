<script lang="ts">
  import { avatarInitial, avatarStyle } from "../avatar";
  import { shortDate } from "../date";
  import type { Post } from "../member/schema";
  import Comments from "./Comments.svelte";
  import Peek from "./Peek.svelte";

  /** One post card: author, a preview of the real page, and a comment thread. */
  let {
    post,
    signedIn,
    onComment,
    onDeleteComment,
  }: {
    post: Post;
    signedIn: boolean;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();
</script>

<article class="card">
  <div class="card-head">
    <a class="avatar" href={post.authorUrl} style={avatarStyle(post.author.id)}>{avatarInitial(post.author.displayName)}</a>
    <div class="who">
      <a href={post.authorUrl}>{post.author.displayName}</a>
      <span class="when">{shortDate(post.publishedAt)}</span>
    </div>
    <span class="chip">{post.templateName ?? post.template}</span>
  </div>

  {#if post.imageUrl}
    <a class="cover" href={post.postUrl}>
      <img src={post.imageUrl} alt="" loading="lazy" />
      <span>{post.title}</span>
    </a>
  {:else}
    <Peek {post} />
  {/if}

  <Comments site={post.site} comments={post.comments} {signedIn} {onComment} {onDeleteComment} />
</article>
