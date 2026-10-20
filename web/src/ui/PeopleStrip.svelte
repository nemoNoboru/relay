<script lang="ts">
  import { avatarInitial, avatarStyle } from "../avatar";
  import type { Member } from "../member/schema";

  /** The row of people at the top of the feed, plus `+ invite` for members. */
  let {
    members,
    signedIn,
  }: {
    members: ReadonlyArray<Member>;
    signedIn: boolean;
  } = $props();
</script>

{#if members.length > 0 || signedIn}
  <div class="people" aria-label="People in this group">
    {#each members as member (member.id)}
      <a class="person" href="/_relay/people/{member.id}">
        <span class="avatar" style={avatarStyle(member.id)}>{avatarInitial(member.displayName)}</span>
        <span class="name">{member.displayName}</span>
      </a>
    {/each}
    {#if signedIn}
      <a class="person invite" href="/_relay/publish.html">
        <span class="avatar">+</span>
        <span class="name">invite</span>
      </a>
    {/if}
  </div>
{/if}
