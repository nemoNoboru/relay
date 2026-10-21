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
    viewerId,
    onComment,
    onDeleteComment,
    onDeletePost,
  }: {
    post: Post;
    signedIn: boolean;
    viewerId?: string;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
    onDeletePost?: (site: string) => Promise<void>;
  } = $props();

  let busy = $state(false);
  let error = $state("");

  const own = $derived(viewerId !== undefined && post.member === viewerId);

  const remove = async (): Promise<void> => {
    if (!onDeletePost || busy) return;
    if (!window.confirm("Remove this post? It deletes the page and its comments.")) return;
    busy = true;
    error = "";
    try {
      await onDeletePost(post.site);
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not delete the post.";
    } finally {
      busy = false;
    }
  };
</script>

<article class="card">
  <div class="card-head">
    <a class="avatar" href={post.authorUrl} style={avatarStyle(post.author.id)}>{avatarInitial(post.author.displayName)}</a>
    <div class="who">
      <a href={post.authorUrl}>{post.author.displayName}</a>
      <span class="when">{shortDate(post.publishedAt)}</span>
    </div>
    <span class="chip">{post.templateName ?? post.template}</span>
    {#if own}
      <button class="post-delete" type="button" onclick={() => void remove()} disabled={busy}>
        {busy ? "Deleting…" : "Delete"}
      </button>
    {/if}
  </div>
  {#if error}<span class="identity-error">{error}</span>{/if}

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
