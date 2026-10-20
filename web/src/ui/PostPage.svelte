<script lang="ts">
  import type { Writable } from "svelte/store";

  import { shortDate } from "../date";
  import type { PostState } from "../post/state";
  import Comments from "./Comments.svelte";
  import Shell from "./Shell.svelte";

  /** A post: back link, metadata, the page itself, and public comments. */
  let {
    postState,
    onSignIn,
    onSignOut,
    onComment,
    onDeleteComment,
  }: {
    postState: Writable<PostState>;
    onSignIn: (token: string) => void;
    onSignOut: () => void;
    onComment: (body: string) => void;
    onDeleteComment: (id: string) => void;
  } = $props();
</script>

<div class="page">
  <Shell
    brand={$postState.me?.node.name ?? "Relay"}
    me={$postState.me}
    active="post"
    {onSignIn}
    {onSignOut}
  />

  <main>
    {#if $postState.phase === "loading"}
      <p class="feed-loading">…</p>
    {:else if $postState.phase === "error"}
      <p class="warn">{$postState.error ?? "No such post on this node."}</p>
    {:else if $postState.post}
      <a class="back" href="/">← Back</a>

      <header class="post-head">
        <span class="chip">{$postState.post.templateName ?? $postState.post.template}</span>
        <h1>{$postState.post.title}</h1>
        {#if $postState.post.summary}<p class="post-lede">{$postState.post.summary}</p>{/if}
        <p class="post-sub">
          by <a href={$postState.post.authorUrl}>{$postState.post.author.displayName}</a>
          · {shortDate($postState.post.publishedAt)}
          · <a href={$postState.post.url}>open the page on its own</a>
        </p>
      </header>

      <iframe class="post-frame" src={$postState.post.url} title={$postState.post.title} sandbox="allow-scripts allow-popups allow-forms allow-downloads"></iframe>

      {#if $postState.commentError}<p class="warn">{$postState.commentError}</p>{/if}
      {#if $postState.notice}<p class="notice" role="status">{$postState.notice}</p>{/if}

      <section class="profile-posts">
        <h2>Comments</h2>
        <Comments
          site={$postState.post.site}
          comments={$postState.comments ?? []}
          signedIn={$postState.me !== undefined}
          onComment={async (_site, body) => {
            onComment(body);
          }}
          onDeleteComment={async (_site, id) => {
            onDeleteComment(id);
          }}
        />
      </section>
    {/if}
  </main>
</div>
