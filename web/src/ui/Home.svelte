<script lang="ts">
  import type { Writable } from "svelte/store";

  import type { HomeState } from "../home/state";
  import FeedList from "./FeedList.svelte";

  /**
   * The member welcome page. Pure rendering of the `home` store; the invite
   * exchange, /me load, feed pagination, comments, sign-out, and profile save
   * arrive as callbacks from `main/home.ts`.
   */
  let {
    home,
    onSignIn,
    onSignOut,
    onSave,
    onLoadMore,
    onComment,
    onDeleteComment,
  }: {
    home: Writable<HomeState>;
    onSignIn: (token: string) => void;
    onSignOut: () => void;
    onSave: (displayName: string, bio: string) => void;
    onLoadMore: () => void;
    onComment: (site: string, body: string) => Promise<void>;
    onDeleteComment: (site: string, id: string) => Promise<void>;
  } = $props();

  let tokenInput = $state("");
</script>

<main class="homeapp">
  <div class="wrap">
    {#if $home.phase === "loading"}
      <p class="hint">Opening your invitation…</p>
    {:else if $home.phase === "signedOut"}
      <section class="welcome signedout">
        <span class="kick">Your place in the group</span>
        <h1>You need<br><span class="red">an invitation</span></h1>
        <p class="welcome-sub">
          Relay has no sign-up. Open the invite link you were sent, or paste
          its key below. A host or another member makes invitations.
        </p>
        <form
          class="field invitefield"
          onsubmit={(e) => {
            e.preventDefault();
            onSignIn(tokenInput);
          }}
        >
          <label for="invite-key">invitation key or link</label>
          <input
            id="invite-key"
            type="password"
            bind:value={tokenInput}
            placeholder="paste the key or the whole link"
            autocomplete="off"
          />
          <button class="btn solid" type="submit" disabled={$home.busy}>Open my home</button>
        </form>
        {#if $home.error}
          <p class="warn">{$home.error}</p>
        {/if}
        <p class="aside">
          No invitation yet? <a href="../">Read what Relay is</a> · a host can
          <a href="../#keep">keep a node</a>.
        </p>
      </section>
    {:else if $home.phase === "error"}
      <section class="welcome signedout">
        <span class="kick">Cannot open your home</span>
        <h1>Something<br><span class="red">went wrong</span></h1>
        <p class="welcome-sub">{$home.error ?? "The node did not answer."}</p>
        <p class="aside"><a href="./home.html">Try again</a> · <a href="../">Home</a></p>
      </section>
    {:else}
      <header class="home-head">
        <div>
          <span class="kick">Your home</span>
          <h1>Welcome,<br><span class="red">{$home.me?.member.displayName ?? "friend"}</span></h1>
          <p class="home-sub">
            {$home.me?.node.name ?? "this node"} · {$home.me?.node.members ?? 1}
            member{($home.me?.node.members ?? 1) === 1 ? "" : "s"}
          </p>
        </div>
        <div class="home-actions">
          <a class="btn solid" href="./publish.html">Publish a page</a>
          <button class="btn" type="button" onclick={onSignOut} disabled={$home.busy}>Sign out</button>
        </div>
      </header>

      <div class="homegrid">
        <section class="homecard feedcard">
          <span class="tag">Recent posts</span>
          <h2>From your people</h2>
          {#if ($home.posts?.length ?? 0) > 0}
            <FeedList
              posts={$home.posts ?? []}
              next={$home.next}
              loading={$home.loadingMore ?? false}
              error={$home.feedError}
              {onLoadMore}
              {onComment}
              {onDeleteComment}
            />
          {:else}
            <div class="emptyfeed">
              <h3>Nothing here yet</h3>
              <p>Publish a page and it will appear here, alongside everyone else's.</p>
              <a class="btn solid" href="./publish.html">Publish the first one</a>
            </div>
          {/if}
        </section>

        <section class="homecard">
          <span class="tag">Your profile</span>
          <h2>How you show up</h2>
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
              <input
                id="display-name"
                name="displayName"
                type="text"
                value={$home.me?.member.displayName ?? ""}
                maxlength="63"
                disabled={$home.busy}
              />
            </div>
            <div class="field">
              <label for="bio">a line about you</label>
              <textarea
                id="bio"
                name="bio"
                value={$home.me?.member.bio ?? ""}
                maxlength="280"
                rows="3"
                disabled={$home.busy}
              ></textarea>
            </div>
            <button class="btn solid" type="submit" disabled={$home.busy}>Save profile</button>
          </form>
          {#if $home.notice}
            <p class="notice" role="status">{$home.notice}</p>
          {/if}
          {#if $home.actionError}
            <p class="warn">{$home.actionError}</p>
          {/if}
        </section>
      </div>

      <p class="home-note">
        Signed in with an invitation. There is no password to reset: if you
        lose this browser, open your invite link again or ask the host to
        rotate it.
      </p>
    {/if}
  </div>
</main>
