<script lang="ts">
  import type { PlushieData } from '../types';

  interface Props {
    plushies: PlushieData[];
    userCollection: string[];
    emptyPlushieImage: string;
    visible: boolean;
  }

  let { plushies, userCollection, emptyPlushieImage, visible }: Props = $props();
</script>

{#if visible}
  <div class="collection-box">
    <div class="collection-items">
      {#each plushies as plushie (plushie.key)}
        <img
          src={userCollection.includes(plushie.key) ? plushie.image : emptyPlushieImage}
          alt={plushie.name}
          title={plushie.name}
          class:collected={userCollection.includes(plushie.key)}
        />
      {/each}
    </div>
  </div>
{/if}

<style>
  .collection-box {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    width: 100%;
  }

  .collection-items {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    justify-content: center;
  }

  .collection-items img {
    width: 40px;
    object-fit: contain;
    transition: transform 0.2s;
  }

  .collection-items img.collected {
    transform: scale(1.1);
  }
</style>
