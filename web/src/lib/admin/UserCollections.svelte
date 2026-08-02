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

  function closeMenu(event: MouseEvent) {
    const menu = (event.currentTarget as HTMLElement).closest('[popover]');
    if (menu instanceof HTMLElement && menu.matches(':popover-open')) menu.hidePopover();
  }

  function menuItems(menu: HTMLElement) {
    return [...menu.querySelectorAll<HTMLButtonElement>('button:not(:disabled)')];
  }

  function focusMenuItem(item: HTMLButtonElement | undefined) {
    item?.focus({ focusVisible: true });
  }

  function focusFirstMenuItem(menuID: string) {
    const menu = document.getElementById(menuID);
    if (!(menu instanceof HTMLElement)) return;

    menu.showPopover();
    requestAnimationFrame(() => focusMenuItem(menuItems(menu)[0]));
  }

  function handleTriggerKeydown(event: KeyboardEvent, menuID: string) {
    if (event.key !== 'ArrowDown') return;

    event.preventDefault();
    focusFirstMenuItem(menuID);
  }

  function handleMenuKeydown(event: KeyboardEvent) {
    const menu = event.currentTarget as HTMLElement;
    const items = menuItems(menu);
    const currentIndex = items.indexOf(document.activeElement as HTMLButtonElement);

    if (event.key === 'Escape') {
      menu.hidePopover();
      document.querySelector<HTMLButtonElement>(`[popovertarget="${menu.id}"]`)?.focus();
      return;
    }

    const nextIndex =
      event.key === 'ArrowDown'
        ? (currentIndex + 1) % items.length
        : event.key === 'ArrowUp'
          ? (currentIndex - 1 + items.length) % items.length
          : event.key === 'Home'
            ? 0
            : event.key === 'End'
              ? items.length - 1
              : null;

    if (nextIndex === null) return;
    event.preventDefault();
    focusMenuItem(items[nextIndex]);
  }
</script>

<section class="flex flex-col gap-4" aria-labelledby="blind-boxes-heading">
  <h3 class="detail-section-title" id="blind-boxes-heading">Blind boxes</h3>
  <div class="grid gap-6 lg:grid-cols-2 2xl:grid-cols-3">
    {#each collections as collection, index (collection.config.series)}
      {@const menuID = `collection-menu-${index}`}
      {@const menuAnchor = `--collection-menu-${index}`}
      <section class="collection-card flex flex-col gap-4 p-5">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h3 class="font-bold">{collection.config.name}</h3>
            <p class="admin-muted text-sm">
              {collection.collected.length}/{collection.config.plushies.length} collected
            </p>
          </div>
          <button
            class="collection-menu-trigger button button-secondary grid h-9 w-9 shrink-0 place-items-center p-0!"
            popovertarget={menuID}
            style={`anchor-name: ${menuAnchor}`}
            disabled={loading}
            onkeydown={(event) => handleTriggerKeydown(event, menuID)}
            aria-haspopup="menu"
            aria-label={`Actions for ${collection.config.name}`}
            title={`Actions for ${collection.config.name}`}
          >
            <svg aria-hidden="true" viewBox="0 0 24 24" fill="currentColor" class="h-5 w-5">
              <circle cx="5" cy="12" r="1.75" />
              <circle cx="12" cy="12" r="1.75" />
              <circle cx="19" cy="12" r="1.75" />
            </svg>
          </button>
          <div
            id={menuID}
            class="collection-actions-menu"
            popover="auto"
            style={`position-anchor: ${menuAnchor}`}
            role="menu"
            tabindex="-1"
            aria-label={`Actions for ${collection.config.name}`}
            onkeydown={handleMenuKeydown}
          >
            <button
              class="button button-secondary w-full text-center"
              role="menuitem"
              onclick={(event) => {
                closeMenu(event);
                onDisplayCollection(collection);
              }}
              disabled={loading}>Display overlay</button
            >
            <button
              class="button button-secondary w-full text-center"
              role="menuitem"
              onclick={(event) => {
                closeMenu(event);
                onOpenRandomPlushie(collection);
              }}
              disabled={loading}>Grant random</button
            >
            <button
              class="button button-danger w-full text-center"
              role="menuitem"
              onclick={(event) => {
                closeMenu(event);
                onOpenResetCollection(collection);
              }}
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
              class={[
                'plushie-button flex flex-col items-center gap-1 p-2 text-left',
                !owned && 'is-unowned',
              ]}
              onclick={() =>
                onSetPlushie(collection.config.series, plushie.key, plushie.name, owned)}
              disabled={mutatingPlushie === plushieID}
              title={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
              aria-label={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
              aria-pressed={owned}
            >
              <img class="size-16 object-contain" src={plushie.image} alt="" />
              <span class="block truncate text-center text-xs w-full">{plushie.name}</span>
            </button>
          {/each}
        </div>
      </section>
    {/each}
  </div>
</section>

<style>
  .collection-actions-menu {
    position: fixed;
    inset: auto;
    top: anchor(bottom);
    right: anchor(right);
    width: max-content;
    min-width: 10rem;
    margin: 0.5rem 0 0;
    border: 1px solid rgb(214 198 223 / 24%);
    border-radius: 0.8rem;
    padding: 0.4rem;
    background: linear-gradient(145deg, rgb(43 34 54), rgb(31 24 41));
    box-shadow: 0 18px 36px rgb(5 3 9 / 40%);
    color: var(--text);
  }

  .collection-actions-menu:popover-open {
    display: grid;
    gap: 0.25rem;
  }

  .collection-menu-trigger:focus-visible,
  .collection-actions-menu .button:focus-visible {
    outline-offset: -2px;
  }
</style>
