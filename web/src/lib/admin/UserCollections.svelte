<script lang="ts">
  import type { components } from '$lib/api.generated';

  type Collection = components['schemas']['AdminCollection'];

  let {
    collections,
    loading,
    mutatingPlushie,
    onDisplayCollection,
    onOpenRandomPlushie,
    onOpenResetCollection,
    onSetPlushie,
  }: {
    collections: Collection[];
    loading: boolean;
    mutatingPlushie: string | null;
    onDisplayCollection: (collection: Collection) => void;
    onOpenRandomPlushie: (collection: Collection) => void;
    onOpenResetCollection: (collection: Collection) => void;
    onSetPlushie: (series: string, key: string, name: string, owned: boolean) => void;
  } = $props();
</script>

<section class="flex flex-col gap-4" aria-labelledby="blind-boxes-heading">
  <h3 class="detail-section-title" id="blind-boxes-heading">Blind boxes</h3>
  <div class="grid gap-6 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
    {#each collections as collection (collection.config.series)}
      <section class="collection-card flex flex-col gap-4 p-5">
        <div class="flex items-center justify-between gap-4">
          <div>
            <h3 class="font-bold">{collection.config.name}</h3>
            <p class="admin-muted text-sm">
              {collection.collected.length}/{collection.config.plushies.length} collected
            </p>
          </div>
          <div class="flex flex-wrap justify-end gap-2">
            <button
              class="button button-secondary"
              onclick={() => onDisplayCollection(collection)}
              disabled={loading}>Display overlay</button
            >
            <button
              class="button button-secondary"
              onclick={() => onOpenRandomPlushie(collection)}
              disabled={loading}>Grant random</button
            >
            <button
              class="button button-danger"
              onclick={() => onOpenResetCollection(collection)}
              disabled={loading || collection.collected.length === 0}
              aria-label={`Reset ${collection.config.name} collection`}>Reset</button
            >
          </div>
        </div>
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
          {#each collection.config.plushies as plushie (plushie.key)}
            {@const owned = collection.collected.includes(plushie.key)}
            {@const plushieID = `${collection.config.series}:${plushie.key}`}
            <button
              class={['plushie-button flex flex-col items-center gap-1 p-2 text-left', !owned && 'is-unowned']}
              onclick={() =>
                onSetPlushie(collection.config.series, plushie.key, plushie.name, owned)}
              disabled={mutatingPlushie === plushieID}
              title={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
              aria-label={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
              aria-pressed={owned}
            >
              <img class="h-16 w-16 object-contain" src={plushie.image} alt="" />
              <span class="block truncate text-center text-xs">{plushie.name}</span>
            </button>
          {/each}
        </div>
      </section>
    {/each}
  </div>
</section>
