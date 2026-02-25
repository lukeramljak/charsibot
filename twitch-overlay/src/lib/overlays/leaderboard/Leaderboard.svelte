<script lang="ts">
  import { charsibot } from '$lib/charsibot.svelte';
  import type { Leaderboard, Stat } from '$lib/types';
  import { onMount } from 'svelte';
  import { fly } from 'svelte/transition';

  let leaderboard = $state<Leaderboard | null>(null);
  let timeoutId: ReturnType<typeof setTimeout> | undefined;

  onMount(() => {
    charsibot.connect();
  });

  $effect(() => {
    const lastMsg = charsibot.lastMessage;
    if (!lastMsg || lastMsg.type !== 'leaderboard') return;

    leaderboard = lastMsg.data;

    if (timeoutId) clearTimeout(timeoutId);

    timeoutId = setTimeout(() => {
      leaderboard = null;
    }, 4000);
  });

  const statEmojis: Record<Stat, string> = {
    STR: '💪',
    INT: '🧠',
    CHA: '✨',
    DEX: '🎯',
    LUCK: '🍀',
    PENIS: '🍆',
  };

  function getEmoji(stat: string) {
    return statEmojis[stat as Stat] ?? '✨';
  }
</script>

{#if leaderboard}
  <div class="overlay">
    <div class="card" in:fly={{ y: 200, duration: 350 }} out:fly>
      <div class="header">
        <h2 class="title">Leaderboard</h2>
      </div>
      <div class="rows">
        {#each Object.entries(leaderboard) as [stat, { username, value }] (stat)}
          <div class="row">
            <span class="emoji">{getEmoji(stat)}</span>
            <span class="stat">{stat}</span>
            <span class="username">{username}</span>
            <span class="value">{value}</span>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
    pointer-events: none;
  }

  .card {
    background: linear-gradient(90deg, #f5918a 50%, #f5b58a 100%);
    border: 2px solid rgba(255, 255, 255, 0.4);
    border-radius: 16px;
    padding: 20px 28px;
    min-width: 380px;
    font-family: 'Nunito', sans-serif;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding-bottom: 14px;
    border-bottom: 2px solid rgba(255, 255, 255, 0.35);
    margin-bottom: 14px;
  }

  .title {
    font-size: 22px;
    font-weight: 700;
    color: #fff;
    margin: 0;
    letter-spacing: 0.5px;
  }

  .rows {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 7px 10px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.15);
  }

  .emoji {
    font-size: 18px;
    width: 24px;
    text-align: center;
  }

  .stat {
    font-size: 12px;
    font-weight: 700;
    color: rgba(255, 255, 255, 0.8);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    min-width: 52px;
  }

  .username {
    flex: 1;
    font-size: 15px;
    font-weight: 600;
    color: #fff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .value {
    font-size: 14px;
    font-weight: 700;
    color: #fff;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    padding: 3px 10px;
    min-width: 40px;
    text-align: center;
  }
</style>
