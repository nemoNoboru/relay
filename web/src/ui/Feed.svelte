<script lang="ts">
  import type { Writable } from "svelte/store";

  import type { FeedState } from "../feed/state";
  import FeedList from "./FeedList.svelte";
  import PeopleStrip from "./PeopleStrip.svelte";
  import Shell from "./Shell.svelte";

  /** The public feed: people, posts newest first, and a single footer link. */
  let {
    feed,
    onSignIn,
    onSignOut,
    onLoadMore,
    onComment,
    onDeleteComment,
  }: {
    feed: Writable<FeedState>;
    onSignIn: (token: string) => void;
    onSignOut: () => void;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();
</script>

<div class="page">
  <Shell
    brand={$feed.me?.node.name ?? "Relay"}
    me={$feed.me}
    active="home"
    signInError={$feed.signInError}
    {onSignIn}
    {onSignOut}
  />

  <main>
    <section class="masthead">
      <h1 class="masthead-title">Relay</h1>
      <p class="masthead-sub">
        A micro-social network for you and your friends — <a href="/why">see why</a>
      </p>
    </section>

    {#if $feed.phase === "loading"}
      <p class="feed-loading">…</p>
    {:else if $feed.phase === "error"}
      <p class="warn">{$feed.error ?? "The node did not answer."}</p>
    {:else}
      <PeopleStrip members={$feed.members ?? []} signedIn={$feed.me !== undefined} />
      {#if $feed.posts.length > 0}
        <FeedList
          posts={$feed.posts}
          next={$feed.next}
          loading={$feed.loadingMore ?? false}
          error={$feed.feedError}
          signedIn={$feed.me !== undefined}
          {onLoadMore}
          {onComment}
          {onDeleteComment}
        >
          {#snippet children()}
            <a class="publish-cta" href="/_relay/publish.html">Publish your post</a>
          {/snippet}
        </FeedList>
      {:else}
        <div class="feed">
          <a class="publish-cta" href="/_relay/publish.html">Publish your post</a>
          <div class="emptyfeed">
            <h3>Nothing here yet.</h3>
            {#if $feed.me}<a class="btn solid" href="/_relay/publish.html">Publish</a>{/if}
          </div>
        </div>
      {/if}
    {/if}

    <footer><a href="/host">Run your own Relay →</a></footer>
  </main>
</div>
