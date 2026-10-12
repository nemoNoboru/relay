<script lang="ts">
  import type { AdminData } from "../admin/fixture";
  import Panel from "./Panel.svelte";

  /**
   * Relay admin mock view (admin.html).
   *
   * A pure function of `{ data }`: fixture in, static HTML out. No store, no
   * polling, no fetch — the smoke test asserts the production bundle never
   * calls the API. Interactions are visual only (no handlers); the
   * `draft-note` chip says so. Promoting the mock to a live page means
   * swapping the fixture for live.ts data — the view stays.
   */
  let { data }: { data: AdminData } = $props();
</script>

<main class="adminapp">
  <div class="wrap">
    <span class="draft-note">Mock data · no functionality</span>

    <div class="admin-head">
      <div>
        <span class="kick">Relay admin</span>
        <h1>{data.node.name}</h1>
        <p class="admin-sub">{data.node.version}</p>
      </div>
      <div class="node-vis">
        <span class="tagline">Visibility</span>
        <strong>{data.node.visibility}</strong>
        <em>{data.node.visibilityNote}</em>
      </div>
    </div>

    <div class="panels">
      <!-- PUBLIC FRIENDS -->
      <Panel title="Public friends" wide>
        <p class="warn">This node is not reachable from the open internet. Pick a public friend to relay through.</p>
        <ul class="pick">
          {#each data.publicFriends as p}
            <li>
              <label>
                <input type="radio" name="public-friend" checked={p.selected} />
                <span>
                  <b>{p.name}</b>
                  <small>{p.note} · {p.latencyMs} ms</small>
                </span>
              </label>
            </li>
          {/each}
          <li>
            <label>
              <input type="radio" name="public-friend" disabled />
              <span>
                <b>Add a public friend…</b>
                <small>any public node will do</small>
              </span>
            </label>
          </li>
        </ul>
        <p class="result">You would be reachable at <code>{data.resultAddress}</code></p>
        <div class="row-actions">
          <button class="btn solid">Connect</button>
          <span class="hint">Connecting dials the friend and registers /raspberry/ with it.</span>
        </div>
      </Panel>

      <!-- FRIENDS -->
      <Panel title="Friends">
        <p class="hint">Nodes this node is connected to. Each owns a path prefix under a public friend.</p>
        {#if data.friends.length > 0}
          <ul class="friends">
            {#each data.friends as f}
              <li class="friend">
                <span>
                  <b>{f.name}</b>
                  <small>{f.role} · {f.online ? "connected" : "offline"}</small>
                </span>
                <code>{f.prefix}</code>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="hint">No friends connected.</p>
        {/if}
        <div class="row-actions">
          <button class="btn">Invite a node</button>
          <span class="hint">Send an invite file, or tell them your address.</span>
        </div>
      </Panel>

      <!-- KEYS -->
      <Panel title="Keys" wide>
        <p class="hint">A key lets a user claim a name and deploy a site on this node. The operator creates the key file and hands it over once.</p>
        <div class="field">
          <label for="key-name">name</label>
          <input id="key-name" type="text" placeholder="e.g. zine" />
          <button class="btn solid">Create key</button>
        </div>
        {#if data.keyFile}
          <pre class="keyfile">{@html JSON.stringify(data.keyFile, null, 2)}</pre>
          <div class="row-actions">
            <button class="btn">Reveal</button>
            <button class="btn solid">Save a copy</button>
          </div>
        {/if}
        <p class="warn">The private half is the deed to the name: shown once, stored nowhere, and impossible to recover. There is no account recovery — ever.</p>
        {#if data.keys.length > 0}
          <table class="tbl">
            <thead>
              <tr>
                <th>name</th>
                <th>key</th>
                <th>created</th>
                <th>status</th>
              </tr>
            </thead>
            <tbody>
              {#each data.keys as k}
                <tr>
                  <td>{k.name}</td>
                  <td><code>{k.fingerprint}</code></td>
                  <td>{k.created}</td>
                  <td>{k.status}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {:else}
          <p class="hint">No keys issued.</p>
        {/if}
      </Panel>

      <!-- NAMES -->
      <Panel title="Names">
        {#if data.names.length > 0}
          <table class="tbl">
            <thead>
              <tr>
                <th>name</th>
                <th>address</th>
                <th>owner</th>
                <th>size</th>
              </tr>
            </thead>
            <tbody>
              {#each data.names as n}
                <tr>
                  <td>{n.name}</td>
                  <td><code>{n.address}</code></td>
                  <td><code>{n.owner}</code></td>
                  <td>{n.size}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {:else}
          <p class="hint">No names claimed on this node.</p>
        {/if}
      </Panel>

      <!-- RULES -->
      <Panel title="Rules">
        <p class="hint">Open deployments on this node must follow the published rule set.</p>
        <ul class="rules">
          {#each data.rules as r}
            <li class="rule">
              <span>{r.tag}</span>
              {r.text}
            </li>
          {/each}
        </ul>
        <p class="result"><code>rules/community.txt</code></p>
        <p class="hint">published to every site on this node</p>
      </Panel>

      <!-- NODE -->
      <Panel title="Node">
        <div class="statrow">
          <div class="stat">
            <b>{data.node.sessionsUsed}/{data.node.sessionsMax}</b>
            <span>sessions</span>
          </div>
          <div class="stat">
            <b>{data.node.storage}</b>
            <span>storage</span>
          </div>
          <div class="stat">
            <b>{data.node.friendsOnline} of {data.node.friendsTotal}</b>
            <span>friends online</span>
          </div>
          <div class="stat">
            <b>{data.node.uptime}</b>
            <span>uptime</span>
          </div>
        </div>
      </Panel>
    </div>

    <p class="sr-only">2 of 3 friends connected. 2 keys issued, 1 name claimed.</p>
  </div>
</main>
