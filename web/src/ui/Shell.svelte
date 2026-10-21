<script lang="ts">
  import { avatarInitial, avatarStyle } from "../avatar";
  import type { Me } from "../member/schema";

  /**
   * The shared chrome: slim paper bar, brand, primary nav, and the identity
   * slot (inline sign-in for guests, avatar + name + sign out for members).
   */
  let {
    brand,
    me,
    active,
    signInError,
    onSignIn,
    onSignOut,
  }: {
    brand: string;
    me?: Me;
    active: "home" | "publish" | "people" | "post";
    signInError?: string;
    onSignIn: (token: string) => void;
    onSignOut: () => void;
  } = $props();

  let showSignIn = $state(false);
  let tokenInput = $state("");

  const submit = () => {
    const token = tokenInput.trim();
    if (token === "") return;
    onSignIn(token);
  };
</script>

<header class="bar">
  <div class="wrap">
    <a class="brand" href="/"><span class="mark" aria-hidden="true"></span> {brand}</a>
    <nav class="main" aria-label="Primary">
      <a href="/" class:active={active === "home"}>Home</a>
      <a href="/_relay/publish.html" class:active={active === "publish"}>Publish</a>
      <a href="/_relay/people" class:active={active === "people"}>People</a>
    </nav>
    <div class="me">
      {#if me}
        <a class="avatar" href="/_relay/people/{me.member.id}" style={avatarStyle(me.member.id)}>{avatarInitial(me.member.displayName)}</a>
        <a class="name" href="/_relay/people/{me.member.id}">{me.member.displayName}</a>
        <button class="out" type="button" onclick={onSignOut}>sign out</button>
      {:else if showSignIn}
        <form
          class="identity-form"
          onsubmit={(e) => {
            e.preventDefault();
            submit();
          }}
        >
          <label for="signin-key">Invitation</label>
          <input id="signin-key" type="password" bind:value={tokenInput} placeholder="paste key or link" autocomplete="off" />
          <button class="identity-go" type="submit">Open</button>
          <button class="identity-cancel" type="button" onclick={() => (showSignIn = false)}>Cancel</button>
          {#if signInError}<span class="identity-error">{signInError}</span>{/if}
        </form>
      {:else}
        <button class="identity-signin" type="button" onclick={() => (showSignIn = true)}>Sign in</button>
      {/if}
    </div>
  </div>
</header>
