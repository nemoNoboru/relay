<script lang="ts">
  import type { Writable } from "svelte/store";

  import { shortDate } from "../date";
  import type { PostState } from "../post/state";

  /** A post wrapper: metadata, the static page itself, and comments. */
  let {
    postState,
    onComment,
    onDeleteComment,
  }: {
    postState: Writable<PostState>;
    onComment: (body: string) => void;
    onDeleteComment: (id: string) => void;
  } = $props();
</script>

<main class="postapp">
  <div class="wrap">
    {#if $postState.phase === "loading"}
      <p class="hint">Loading post…</p>
    {:else if $postState.phase === "signedOut"}
      <section class="signedout">
        <span class="kick">Members only</span>
        <h1>You need<br><span class="red">an invitation</span></h1>
        <p class="welcome-sub">The post itself is public at its own address. The context around it is for the group.</p>
        <p class="aside"><a href="/_relay/home.html">Open your home</a> · <a href="/">What is Relay</a></p>
      </section>
    {:else if $postState.phase === "error"}
      <section class="signedout">
        <span class="kick">Post</span>
        <h1>Not<br><span class="red">found</span></h1>
        <p class="welcome-sub">{$postState.error ?? "No such post on this node."}</p>
        <p class="aside"><a href="/_relay/home.html">Back to your home</a></p>
      </section>
    {:else if $postState.post}
      <header class="post-head">
        <span class="kick">{$postState.post.template} · {shortDate($postState.post.publishedAt)}</span>
        <h1>{$postState.post.title}</h1>
        {#if $postState.post.summary}<p class="post-lede">{$postState.post.summary}</p>{/if}
        <p class="post-sub">
          by <a href={$postState.post.authorUrl}>{$postState.post.author.displayName}</a>
          {#if $postState.post.author.revoked}<span class="former">former member</span>{/if}
          · <a href={$postState.post.url}>open the page on its own</a>
        </p>
      </header>

      <iframe class="post-frame" src={$postState.post.url} title={$postState.post.title} sandbox="allow-scripts allow-popups allow-forms allow-downloads"></iframe>

      <section class="comments">
        <span class="tag">Comments</span>
        <h2>Talk under the post</h2>
        {#if $postState.commentError}
          <p class="warn">{$postState.commentError}</p>
        {/if}
        {#if $postState.notice}
          <p class="notice" role="status">{$postState.notice}</p>
        {/if}

        {#if ($postState.comments?.length ?? 0) === 0}
          <p class="hint">No comments yet. Be the first.</p>
        {:else}
          <div class="commentlist">
            {#each $postState.comments ?? [] as comment (comment.id)}
              <article class="comment">
                <div class="comment-head">
                  <a href={comment.author.id ? `/_relay/people/${comment.author.id}` : "/_relay/home.html"}>{comment.author.displayName}</a>
                  <span>{shortDate(comment.createdAt)}</span>
                  {#if comment.mine}
                    <button class="comment-delete" type="button" onclick={() => onDeleteComment(comment.id)} disabled={$postState.busy}>
                      Delete
                    </button>
                  {/if}
                </div>
                <!-- html is server-rendered by social.RenderComment: escaped,
                     with only paragraphs, emphasis, and safe links. -->
                <div class="comment-body">{@html comment.html}</div>
              </article>
            {/each}
          </div>
        {/if}

        <form
          class="commentform"
          onsubmit={(e) => {
            e.preventDefault();
            const form = e.currentTarget;
            const body = String(new FormData(form).get("body") ?? "");
            onComment(body);
            form.reset();
          }}
        >
          <label for="comment-body">Add a comment</label>
          <textarea id="comment-body" name="body" rows="3" maxlength="2000" required disabled={$postState.busy}></textarea>
          <div class="commentfoot">
            <button class="btn solid" type="submit" disabled={$postState.busy}>Post comment</button>
            <span class="hint">Visible to members. 2,000 characters max.</span>
          </div>
        </form>
        <p class="aside"><a href="/_relay/home.html">Back to your home</a></p>
      </section>
    {/if}
  </div>
</main>
