<script lang="ts">
  import type { Writable } from "svelte/store";

  import type { AdminOverview, DeployKey } from "../admin/schema";
  import type { AdminState } from "../admin/state";
  import { inviteLink } from "../publish";

  /**
   * Host console (admin.html). Pure rendering of the `admin` store; all
   * effects (fetch, create, revoke, rotate) live in `main/admin.ts` and arrive
   * here as callbacks, so this component never touches the network.
   * Invitation tokens are shown masked until revealed; `revealed`/`copied`
   * are local presentation state only.
   */
  let {
    admin,
    onConnect,
    onCreate,
    onRevoke,
    onRotate,
    onDeleteComment,
    onRefresh,
    onDismiss,
    onLogout,
    onChangeToken,
  }: {
    admin: Writable<AdminState>;
    onConnect: (token: string) => void;
    onCreate: (user: string) => void;
    onRevoke: (id: string) => void;
    onRotate: (id: string) => void;
    onDeleteComment: (id: string) => void;
    onRefresh: () => void;
    onDismiss: () => void;
    onLogout: () => void;
    onChangeToken: (token: string) => void;
  } = $props();

  let tokenInput = $state("");
  let userInput = $state("");
  let newTokenInput = $state("");
  let revealed = $state<Record<string, boolean>>({});

  /** Pages grouped by member. Pure. */
  const sitesByUser = (ov: AdminOverview | undefined): { user: string; sites: string[] }[] => {
    if (ov === undefined) return [];
    const m = new Map<string, string[]>();
    for (const s of ov.sites) {
      const list = m.get(s.user) ?? [];
      list.push(s.site);
      m.set(s.user, list);
    }
    return [...m.entries()]
      .map(([user, sites]) => ({ user, sites }))
      .sort((a, b) => a.user.localeCompare(b.user));
  };

  /** Invitations grouped by member. Pure. */
  const keysByUser = (ov: AdminOverview | undefined): { user: string; keys: DeployKey[] }[] => {
    if (ov === undefined) return [];
    const m = new Map<string, DeployKey[]>();
    for (const k of ov.keys) {
      const list = m.get(k.user) ?? [];
      list.push(k);
      m.set(k.user, list);
    }
    return [...m.entries()]
      .map(([user, keys]) => ({ user, keys }))
      .sort((a, b) => a.user.localeCompare(b.user));
  };

  const toggleReveal = (id: string): void => {
    revealed = { ...revealed, [id]: !revealed[id] };
  };

  const copy = (text: string): void => {
    void navigator.clipboard?.writeText(text);
  };

  let invited = $state("");
  const invite = (k: DeployKey): void => {
    copy(inviteLink(location.origin, k.token));
    invited = k.id;
    setTimeout(() => (invited = ""), 1500);
  };
</script>

<main class="adminapp">
  <div class="wrap">
    <div class="admin-head">
      <div>
        <span class="kick">Relay host</span>
        <h1>Host console</h1>
        <p class="admin-sub">Invitations, members &amp; pages</p>
      </div>
    </div>

    {#if $admin.phase === "locked" || $admin.phase === "error"}
      <section class="panel">
        <h3>Sign in</h3>
        <div class="panel-body">
          <p class="hint">Enter the node's admin password.</p>
          <p class="hint">A fresh node starts on the password <code>relay</code> — change it before inviting anyone.</p>
          <form onsubmit={(e) => { e.preventDefault(); onConnect(tokenInput); }}>
            <div class="field">
              <label for="admin-token">admin password</label>
              <input id="admin-token" type="password" bind:value={tokenInput} placeholder="…" autocomplete="off" />
              <button class="btn solid" type="submit">Connect</button>
            </div>
          </form>
          {#if $admin.phase === "error"}
            <p class="warn">Cannot connect: {$admin.error ?? "unknown error"}.</p>
          {/if}
        </div>
      </section>
    {:else if $admin.phase === "loading"}
      <p class="hint">Loading…</p>
    {:else if $admin.overview?.mustChangeToken}
      <!-- The node came up on the default token; nothing else until it is replaced. -->
      <section class="panel">
        <h3>Choose a password</h3>
        <div class="panel-body">
          <p class="hint">This node is still on its shipped password. Pick your own — eight characters or more.</p>
          <form onsubmit={(e) => { e.preventDefault(); onChangeToken(newTokenInput); newTokenInput = ""; }}>
            <div class="field">
              <label for="new-token">new password</label>
              <input id="new-token" type="password" bind:value={newTokenInput} placeholder="…" autocomplete="new-password" disabled={$admin.busy} />
              <button class="btn solid" type="submit" disabled={$admin.busy}>Save it</button>
            </div>
          </form>
          {#if $admin.actionError}
            <p class="warn">{$admin.actionError}</p>
          {/if}
        </div>
      </section>
    {:else}
      {#if $admin.newToken}
        <section class="panel wide">
          <h3>New invitation — shown once</h3>
          <div class="panel-body">
            <p class="hint">{$admin.notice}. Copy the invite link or token now; it is also listed under the member below.</p>
            <pre class="keyfile">{$admin.newToken}</pre>
            <div class="row-actions">
              <button class="btn solid" onclick={() => copy($admin.newToken ?? "")}>Copy</button>
              <button class="btn" onclick={onDismiss}>Done</button>
            </div>
          </div>
        </section>
      {/if}

      {#if $admin.actionError}
        <p class="warn">{$admin.actionError}</p>
      {/if}

      <section class="panel wide">
        <h3>Invitations &amp; members</h3>
        <div class="panel-body">
          <p class="hint">Each invitation is a member's identity. Send them the link. Revoking a member also revokes everyone below them.</p>
          <form onsubmit={(e) => { e.preventDefault(); onCreate(userInput); userInput = ""; }}>
            <div class="field">
              <label for="key-user">member name</label>
              <input id="key-user" type="text" bind:value={userInput} placeholder="e.g. name" disabled={$admin.busy} />
              <button class="btn solid" type="submit" disabled={$admin.busy}>Create invitation</button>
            </div>
          </form>

          {#if keysByUser($admin.overview).length === 0}
            <p class="hint">No invitations yet.</p>
          {:else}
            <div class="keygroups">
              {#each keysByUser($admin.overview) as g (g.user)}
                <div class="keygroup">
                  <h4>{g.user}</h4>
                  <ul class="keylist">
                    {#each g.keys as k (k.id)}
                      <li class="keyrow">
                        <div class="keymeta">
                          <code>{k.id}</code>
                          <small>{k.created} · {k.revoked ? "revoked" : "active"}</small>
                        </div>
                        <div class="keytoken">
                          {#if revealed[k.id]}
                            <code class="tokensecret">{k.token}</code>
                            <button class="btn" onclick={() => copy(k.token)}>Copy</button>
                            <button class="btn" onclick={() => toggleReveal(k.id)}>Hide</button>
                          {:else}
                            <code class="tokensecret">••••••••••••••</code>
                            <button class="btn" onclick={() => toggleReveal(k.id)}>Reveal</button>
                          {/if}
                        </div>
                        <div class="keyactions">
                          {#if !k.revoked}
                            <button class="btn solid invite-long" onclick={() => invite(k)}>{invited === k.id ? "Copied" : "Copy invite link"}</button>
                            <button class="btn solid invite-short" onclick={() => invite(k)}>{invited === k.id ? "Copied" : "Invite"}</button>
                            <button class="btn" onclick={() => onRotate(k.id)} disabled={$admin.busy}>Rotate</button>
                            <button class="btn" onclick={() => onRevoke(k.id)} disabled={$admin.busy}>Revoke</button>
                          {:else}
                            <span class="hint">revoked</span>
                          {/if}
                        </div>
                      </li>
                    {/each}
                  </ul>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </section>

      <section class="panel wide">
        <h3>Recent comments</h3>
        <div class="panel-body">
          {#if ($admin.overview?.comments.length ?? 0) === 0}
            <p class="hint">No comments yet.</p>
          {:else}
            <ul class="commentlist admincomments">
              {#each $admin.overview?.comments ?? [] as c (c.id)}
                <li class="comment">
                  <div class="comment-head">
                    <a href={"/_relay/people/" + c.author.id}>{c.author.displayName}</a>
                    <span>on <code>{c.post}</code> · {c.createdAt}</span>
                    <button class="comment-delete" type="button" onclick={() => onDeleteComment(c.id)} disabled={$admin.busy}>Delete</button>
                  </div>
                  <div class="comment-body">{c.body}</div>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </section>

      <section class="panel wide">
        <h3>Pages by member</h3>
        <div class="panel-body">
          {#if sitesByUser($admin.overview).length === 0}
            <p class="hint">Nothing published yet.</p>
          {:else}
            <ul class="friends">
              {#each sitesByUser($admin.overview) as u (u.user)}
                <li class="friend">
                  <span>
                    <b>{u.user}</b>
                    <small>{u.sites.length} page{u.sites.length === 1 ? "" : "s"}</small>
                  </span>
                  <code>{u.sites.join(" · ")}</code>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </section>

      <div class="row-actions">
        <button class="btn" onclick={onRefresh} disabled={$admin.busy}>Refresh</button>
        <button class="btn" onclick={onLogout}>Sign out</button>
      </div>
    {/if}
  </div>
</main>
