<script lang="ts">
  import type { Writable } from "svelte/store";

  import type { AdminOverview } from "../admin/schema";
  import type { AdminState } from "../admin/state";

  /**
   * Operator admin page (admin.html). Pure rendering of the `admin` store;
   * all effects (fetch, create, revoke, rotate) live in `main/admin.ts` and
   * arrive here as callbacks, so this component never touches the network.
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

  /** Sites grouped by owner, newest name first within a user. Pure. */
  const byUser = (ov: AdminOverview | undefined): { user: string; sites: string[] }[] => {
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
            <p class="hint">{$admin.notice}. Copy the token now; it will not be shown again.</p>
            <pre class="keyfile">{$admin.newToken}</pre>
            <div class="row-actions">
              <button class="btn solid" onclick={() => void navigator.clipboard?.writeText($admin.newToken ?? "")}>Copy</button>
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
          <p class="hint">Each key lets one user deploy sites. Rotate or revoke a key to cut them off.</p>
          <form onsubmit={(e) => { e.preventDefault(); onCreate(userInput); userInput = ""; }}>
            <div class="field">
              <label for="key-user">user</label>
              <input id="key-user" type="text" bind:value={userInput} placeholder="e.g. alice" disabled={$admin.busy} />
              <button class="btn solid" type="submit" disabled={$admin.busy}>Create key</button>
            </div>
          </form>
          {#if ($admin.overview?.keys.length ?? 0) === 0}
            <p class="hint">No keys issued.</p>
          {:else}
            <table class="tbl">
              <thead>
                <tr><th>user</th><th>created</th><th>status</th><th></th></tr>
              </thead>
              <tbody>
                {#each $admin.overview?.keys ?? [] as k (k.id)}
                  <tr>
                    <td>{k.user}</td>
                    <td>{k.created}</td>
                    <td>{k.revoked ? "revoked" : "active"}</td>
                    <td>
                      {#if !k.revoked}
                        <button class="btn" onclick={() => onRotate(k.id)} disabled={$admin.busy}>Rotate</button>
                        <button class="btn" onclick={() => onRevoke(k.id)} disabled={$admin.busy}>Revoke</button>
                      {:else}
                        <span class="hint">—</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          {/if}
        </div>
      </section>

      <section class="panel wide">
        <h3>Sites by user</h3>
        <div class="panel-body">
          {#if byUser($admin.overview).length === 0}
            <p class="hint">No sites deployed yet.</p>
          {:else}
            <ul class="friends">
              {#each byUser($admin.overview) as u (u.user)}
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
