<script lang="ts">
  import type { Writable } from "svelte/store";

  import { avatarInitial, avatarStyle } from "../avatar";
  import { shortDate } from "../date";
  import type { PeopleState } from "../people/state";
  import FeedList from "./FeedList.svelte";
  import Shell from "./Shell.svelte";

  /** The people page: the member directory, or one member's profile. */
  let {
    people,
    onSignIn,
    onSignOut,
    onLoadMore,
    onComment,
    onDeleteComment,
    onSave,
  }: {
    people: Writable<PeopleState>;
    onSignIn: (token: string) => void;
    onSignOut: () => void;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
    onSave: (displayName: string, bio: string) => void;
  } = $props();
</script>

<div class="page">
  <Shell
    brand={$people.me?.node.name ?? "Relay"}
    me={$people.me}
    active="people"
    {onSignIn}
    {onSignOut}
  />

  <main>
    {#if $people.phase === "loading"}
      <p class="feed-loading">…</p>
    {:else if $people.phase === "error"}
      <p class="warn">{$people.error ?? "No such member on this node."}</p>
    {:else if $people.members}
      <div class="profile-head">
        <div>
          <span class="kick">People</span>
          <h1>Who's here</h1>
          <p class="profile-sub">{$people.members.length} member{$people.members.length === 1 ? "" : "s"}</p>
        </div>
      </div>
      <div class="people-grid">
        {#each $people.members as member (member.id)}
          <a class="personcard" href="/_relay/people/{member.id}">
            <span class="avatar" style={avatarStyle(member.id)}>{avatarInitial(member.displayName)}</span>
            <b>{member.displayName}</b>
          </a>
        {/each}
      </div>
    {:else if $people.profile}
      <div class="profile-head">
        <span class="avatar" style={avatarStyle($people.profile.member.id)}>{avatarInitial($people.profile.member.displayName)}</span>
        <div>
          <h1>{$people.profile.member.displayName}</h1>
          <p class="profile-sub">
            {$people.profile.total} post{$people.profile.total === 1 ? "" : "s"}
            · joined {shortDate($people.profile.member.created)}
            {#if $people.profile.member.revoked}<span class="former">· former member</span>{/if}
          </p>
        </div>
      </div>

      {#if $people.profile.member.bio}
        <p class="profile-bio">{$people.profile.member.bio}</p>
      {/if}

      {#if $people.me && $people.me.member.id === $people.profile.member.id}
        <details class="edit-profile">
          <summary>Edit profile</summary>
          <form
            class="profileform"
            onsubmit={(e) => {
              e.preventDefault();
              const data = new FormData(e.currentTarget);
              onSave(String(data.get("displayName") ?? ""), String(data.get("bio") ?? ""));
            }}
          >
            <div class="field">
              <label for="display-name">display name</label>
              <input id="display-name" name="displayName" type="text" value={$people.profile.member.displayName} maxlength="63" disabled={$people.busy} />
            </div>
            <div class="field">
              <label for="bio">a line about you</label>
              <textarea id="bio" name="bio" value={$people.profile.member.bio ?? ""} maxlength="280" rows="3" disabled={$people.busy}></textarea>
            </div>
            <button class="btn solid" type="submit" disabled={$people.busy}>Save profile</button>
          </form>
          {#if $people.notice}
            <p class="notice" role="status">{$people.notice}</p>
          {/if}
          {#if $people.actionError}
            <p class="warn">{$people.actionError}</p>
          {/if}
        </details>
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
            signedIn={$people.me !== undefined}
            {onLoadMore}
            {onComment}
            {onDeleteComment}
          />
        {/if}
      </section>
    {/if}
  </main>
</div>
