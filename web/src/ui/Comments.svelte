<script lang="ts">
  import { avatarInitial, avatarStyle } from "../avatar";
  import type { Comment } from "../member/schema";

  /**
   * Collapsed comment thread. Renders only when there is something to show:
   * `N comments` when N > 0, or a bare reply form for the first comment when
   * the viewer is a member.
   */
  let {
    site,
    comments,
    signedIn,
    onComment,
    onDeleteComment,
  }: {
    site: string;
    comments: ReadonlyArray<Comment>;
    signedIn: boolean;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();

  let body = $state("");
  let busy = $state(false);
  let error = $state("");

  const submit = async () => {
    const text = body.trim();
    if (text === "" || busy) return;
    busy = true;
    error = "";
    try {
      await onComment(site, text);
      body = "";
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
      await onDeleteComment(site, id);
    } catch (err) {
      error = err instanceof Error ? err.message : "Could not delete the comment.";
    } finally {
      busy = false;
    }
  };
</script>

{#if comments.length > 0}
  <details class="comments">
    <summary>{comments.length} comment{comments.length === 1 ? "" : "s"}</summary>
    <div class="thread">
      {#each comments as comment (comment.id)}
        <div class="cmt">
          <a class="avatar" href={comment.author.id ? `/_relay/people/${comment.author.id}` : "/_relay/people"} style={avatarStyle(comment.author.id)}>{avatarInitial(comment.author.displayName)}</a>
          <div>
            <b>{comment.author.displayName}</b>
            {#if comment.mine}<button class="comment-delete" type="button" onclick={() => remove(comment.id)} disabled={busy}>Delete</button>{/if}
            <!-- Server-rendered and sanitized by social.RenderComment. -->
            <p>{@html comment.html}</p>
          </div>
        </div>
      {/each}
    </div>
    {#if signedIn}
      <form class="add" onsubmit={(e) => { e.preventDefault(); void submit(); }}>
        <input bind:value={body} placeholder="Add a comment…" maxlength="2000" disabled={busy} />
        <button type="submit" disabled={busy || body.trim() === ""}>Reply</button>
      </form>
    {:else}
      <p class="members-note">Members comment here.</p>
    {/if}
    {#if error}<span class="identity-error">{error}</span>{/if}
  </details>
{:else if signedIn}
  <form class="add" onsubmit={(e) => { e.preventDefault(); void submit(); }}>
    <input bind:value={body} placeholder="Add a comment…" maxlength="2000" disabled={busy} />
    <button type="submit" disabled={busy || body.trim() === ""}>Reply</button>
  </form>
  {#if error}<span class="identity-error">{error}</span>{/if}
{/if}
