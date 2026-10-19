<script lang="ts">
  import type { Writable } from "svelte/store";

  import { shortDate } from "../date";
  import type { PeopleState } from "../people/state";
  import FeedList from "./FeedList.svelte";

  /** One member's profile: identity, bio, and their posts. */
  let {
    people,
    onLoadMore,
    onComment,
    onDeleteComment,
  }: {
    people: Writable<PeopleState>;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();
</script>

<main class="profileapp">
  <div class="wrap">
    {#if $people.phase === "loading"}
      <p class="hint">Loading profile…</p>
    {:else if $people.phase === "signedOut"}
      <section class="signedout">
        <span class="kick">Members only</span>
        <h1>You need<br><span class="red">an invitation</span></h1>
        <p class="welcome-sub">Profiles are visible to the group. Open your invitation link first.</p>
        <p class="aside"><a href="/_relay/home.html">Open your home</a> · <a href="/">What is Relay</a></p>
      </section>
    {:else if $people.phase === "error"}
      <section class="signedout">
        <span class="kick">Profile</span>
        <h1>Not<br><span class="red">found</span></h1>
        <p class="welcome-sub">{$people.error ?? "No such member on this node."}</p>
        <p class="aside"><a href="/_relay/home.html">Back to your home</a></p>
      </section>
    {:else if $people.profile}
      <header class="profile-head">
        <div>
          <span class="kick">Profile</span>
          <h1>{$people.profile.member.displayName}</h1>
          <p class="profile-sub">
            {$people.profile.total} post{$people.profile.total === 1 ? "" : "s"}
            · joined {shortDate($people.profile.member.created)}
            {#if $people.profile.member.revoked}<span class="former">· former member</span>{/if}
          </p>
        </div>
        <a class="btn" href="/_relay/home.html">Your home</a>
      </header>

      {#if $people.profile.member.bio}
        <p class="profile-bio">{$people.profile.member.bio}</p>
      {/if}

      <section class="profile-posts">
        <h2>Posts</h2>
        {#if $people.profile.posts.length === 0}
          <p class="hint">Nothing published yet.</p>
        {:else}
          <FeedList
            posts={$people.profile.posts}
            next={$people.next}
            loading={$people.loadingMore ?? false}
            error={$people.error}
            {onLoadMore}
            {onComment}
            {onDeleteComment}
          />
        {/if}
      </section>
    {/if}
  </div>
</main>
