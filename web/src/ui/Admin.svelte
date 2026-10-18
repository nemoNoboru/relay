<script lang="ts">
  import type { Writable } from "svelte/store";

  import type { AdminOverview, DeployKey } from "../admin/schema";
  import type { AdminState } from "../admin/state";
  import { inviteLink } from "../publish";

  /**
   * Operator admin page (admin.html). Pure rendering of the `admin` store;
   * all effects (fetch, create, revoke, rotate) live in `main/admin.ts` and
   * arrive here as callbacks, so this component never touches the network.
   * Key tokens are shown masked until revealed; `revealed`/`copied` are local
   * presentation state only.
   */
  let {
    admin,
    onConnect,
    onCreate,
    onRevoke,
    onRotate,
    onRefresh,
    onDismiss,
    onLogout,
  }: {
    admin: Writable<AdminState>;
    onConnect: (token: string) => void;
    onCreate: (user: string) => void;
    onRevoke: (id: string) => void;
    onRotate: (id: string) => void;
    onRefresh: () => void;
    onDismiss: () => void;
    onLogout: () => void;
  } = $props();

  let tokenInput = $state("");
  let userInput = $state("");
  let revealed = $state<Record<string, boolean>>({});

  /** Sites grouped by owner. Pure. */
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

  /** Keys grouped by owner. Pure. */
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
        <span class="kick">Relay operator</span>
        <h1>Admin</h1>
        <p class="admin-sub">Deploy keys &amp; sites</p>
      </div>
    </div>

    {#if $admin.phase === "locked" || $admin.phase === "error"}
      <section class="panel">
        <h3>Sign in</h3>
        <div class="panel-body">
          <p class="hint">Enter the node's admin token to manage deploy keys and sites.</p>
          <form onsubmit={(e) => { e.preventDefault(); onConnect(tokenInput); }}>
            <div class="field">
              <label for="admin-token">admin token</label>
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
    {:else}
      {#if $admin.newToken}
        <section class="panel wide">
          <h3>New deploy key — shown once</h3>
          <div class="panel-body">
            <p class="hint">{$admin.notice}. Copy the token now; it is also listed under the user below.</p>
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
        <h3>Deploy keys</h3>
        <div class="panel-body">
          <p class="hint">Each key lets one person publish sites. Send them its invite link: they open it and drop their site. Rotate or revoke to cut someone off.</p>
          <form onsubmit={(e) => { e.preventDefault(); onCreate(userInput); userInput = ""; }}>
            <div class="field">
              <label for="key-user">user</label>
              <input id="key-user" type="text" bind:value={userInput} placeholder="e.g. name" disabled={$admin.busy} />
              <button class="btn solid" type="submit" disabled={$admin.busy}>Create key</button>
            </div>
          </form>

          {#if keysByUser($admin.overview).length === 0}
            <p class="hint">No keys issued.</p>
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
                            <button class="btn solid" onclick={() => invite(k)}>{invited === k.id ? "Copied" : "Copy invite link"}</button>
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
        <h3>Sites by user</h3>
        <div class="panel-body">
          {#if sitesByUser($admin.overview).length === 0}
            <p class="hint">No sites deployed yet.</p>
          {:else}
            <ul class="friends">
              {#each sitesByUser($admin.overview) as u (u.user)}
                <li class="friend">
                  <span>
                    <b>{u.user}</b>
                    <small>{u.sites.length} site{u.sites.length === 1 ? "" : "s"}</small>
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
