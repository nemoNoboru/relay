<script lang="ts">
  import { shortDate } from "../date";
  import type { Post } from "../member/schema";

  /** One post in a feed or profile: cover, metadata, comments, and form. */
  let {
    post,
    onComment,
    onDeleteComment,
  }: {
    post: Post;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();

  let commentText = $state("");
  let busy = $state(false);
  let error = $state("");

  const submit = async () => {
    const body = commentText.trim();
    if (body === "" || busy) return;
    busy = true;
    error = "";
    try {
      await onComment(post.site, body);
      commentText = "";
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not post the comment.";
    } finally {
      busy = false;
    }
  };

  const remove = async (id: string) => {
    if (busy) return;
    busy = true;
    error = "";
    try {
      await onDeleteComment(post.site, id);
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not delete the comment.";
    } finally {
      busy = false;
    }
  };
</script>

<article class="postcard">
  <div class="postcard-head">
    <a class="post-title" href={post.postUrl}>{post.title}</a>
    <span class="post-date">{shortDate(post.publishedAt)}</span>
  </div>

  {#if post.imageUrl}
    <img class="post-cover" src={post.imageUrl} alt="" loading="lazy" />
  {/if}

  {#if post.summary}
    <p class="post-summary">{post.summary}</p>
  {/if}

  <p class="post-foot">
    <a href={post.authorUrl}>{post.author.displayName}</a>
    {#if post.author.revoked}<span class="former">former member</span>{/if}
    <span aria-hidden="true">·</span>
    <a href={post.url}>Open page</a>
    <span aria-hidden="true">·</span>
    <span class="post-template">{post.template}</span>
  </p>

  <section class="feed-comments">
    <h3>Comments <span>{post.comments.length}</span></h3>
    {#if post.comments.length > 0}
      <div class="commentlist">
        {#each post.comments as comment (comment.id)}
          <article class="comment">
            <div class="comment-head">
              <a href={comment.author.id ? `/_relay/people/${comment.author.id}` : "/_relay/home.html"}>{comment.author.displayName}</a>
              <span>{shortDate(comment.createdAt)}</span>
              {#if comment.mine}
                <button class="comment-delete" type="button" onclick={() => remove(comment.id)} disabled={busy}>Delete</button>
              {/if}
            </div>
            <!-- Server-rendered and sanitized by social.RenderComment. -->
            <div class="comment-body">{@html comment.html}</div>
          </article>
        {/each}
      </div>
    {:else}
      <p class="hint">No comments yet.</p>
    {/if}

    <form
      class="commentform compact"
      onsubmit={(e) => {
        e.preventDefault();
        void submit();
      }}
    >
      <label for={`comment-${post.id}`}>Add a comment</label>
      <textarea
        id={`comment-${post.id}`}
        bind:value={commentText}
        rows="2"
        maxlength="2000"
        placeholder="Say something to your friend…"
        disabled={busy}
      ></textarea>
      <div class="commentfoot">
        <button class="btn solid" type="submit" disabled={busy || commentText.trim() === ""}>Post</button>
        {#if error}<span class="comment-error">{error}</span>{/if}
      </div>
    </form>
  </section>
</article>
