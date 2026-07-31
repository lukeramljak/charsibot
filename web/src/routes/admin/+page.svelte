<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { BlindBoxOverlayConfig } from '$lib/types';

  interface User {
    id: string;
    username: string;
  }

  interface UserStat {
    name: string;
    shortName: string;
    longName: string;
    value: number;
  }

  interface Collection {
    config: BlindBoxOverlayConfig;
    collected: string[];
  }

  interface PendingPlushie {
    series: string;
    key: string;
    name: string;
  }

  interface UserDetail {
    user: User;
    stats: UserStat[];
    collections: Collection[];
  }

  let users = $state.raw<User[]>([]);
  let usernameFilter = $state('');
  let filteredUsers = $derived(
    users.filter((user) =>
      user.username.toLowerCase().includes(usernameFilter.trim().toLowerCase()),
    ),
  );
  let selected = $state.raw<UserDetail | null>(null);
  let loading = $state(false);
  let mutatingPlushie = $state<string | null>(null);
  let pendingRandomCollection = $state.raw<Collection | null>(null);
  let randomPlushieDialog = $state<HTMLDialogElement | undefined>(undefined);
  let pendingPlushie = $state.raw<PendingPlushie | null>(null);
  let plushieDialog = $state<HTMLDialogElement | undefined>(undefined);
  let pendingResetCollection = $state.raw<Collection | null>(null);
  let resetDialog = $state<HTMLDialogElement | undefined>(undefined);
  let explodeDialog = $state<HTMLDialogElement | undefined>(undefined);
  let undoExplodeDialog = $state<HTMLDialogElement | undefined>(undefined);
  let randomStatDialog = $state<HTMLDialogElement | undefined>(undefined);
  let error = $state('');
  let statusMessage = $state('');
  let selectedUserHeading = $state<HTMLHeadingElement | undefined>(undefined);

  async function request<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error((await response.text()) || 'Request failed');
    }
    return response.json() as Promise<T>;
  }

  async function loadUsers() {
    loading = true;
    error = '';
    statusMessage = 'Loading viewers…';
    try {
      const response = await request<{ users: User[] }>('/api/admin/users');
      users = response.users;
      statusMessage = `Loaded ${users.length} viewers.`;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Could not search users';
      statusMessage = '';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void initialise();
  });

  afterNavigate(() => {
    if (users.length > 0) void syncSelectedUserFromURL();
  });

  async function initialise() {
    await loadUsers();
    await syncSelectedUserFromURL();
  }

  async function syncSelectedUserFromURL() {
    const userID = page.url.searchParams.get('user');
    const user = users.find((candidate) => candidate.id === userID);
    if (user && selected?.user.id !== user.id) {
      await selectUser(user, false);
    } else if (!user) {
      selected = null;
    }
  }

  async function selectUser(user: User, updateURL = true) {
    if (updateURL) {
      const href = resolve(`/admin?user=${encodeURIComponent(user.id)}`);
      if (new URL(href, page.url).href !== page.url.href) {
        await goto(resolve(`/admin?user=${encodeURIComponent(user.id)}`), {
          keepFocus: true,
          noScroll: true,
        });
        return;
      }
    }
    loading = true;
    error = '';
    statusMessage = `Loading ${user.username}…`;
    try {
      selected = await request<UserDetail>(`/api/admin/users/${encodeURIComponent(user.id)}`);
      statusMessage = `Loaded ${user.username}.`;
      await tick();
      selectedUserHeading?.focus();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Could not load user';
      statusMessage = '';
    } finally {
      loading = false;
    }
  }

  async function updateStat(
    stat: UserStat,
    value: number,
    mode: 'set' | 'adjust',
  ): Promise<boolean> {
    if (!selected || !Number.isFinite(value)) return false;
    return mutate(
      `/api/admin/users/${encodeURIComponent(selected.user.id)}/stats/${encodeURIComponent(stat.name)}`,
      {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mode, value }),
      },
    );
  }

  function openRandomStatDialog() {
    randomStatDialog?.showModal();
  }

  function closeRandomStatDialog() {
    randomStatDialog?.close();
  }

  async function grantRandomStat(displayInChat: boolean) {
    if (!selected) return;
    closeRandomStatDialog();
    await mutate(
      `/api/admin/users/${encodeURIComponent(selected.user.id)}/stats/random`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ displayInChat }),
      },
    );
  }

  function closeExplodeDialog() {
    explodeDialog?.close();
  }

  async function explode() {
    if (!selected) return;
    closeExplodeDialog();
    await mutate(`/api/admin/users/${encodeURIComponent(selected.user.id)}/stats/explode`, {
      method: 'POST',
    });
  }

  function closeUndoExplodeDialog() {
    undoExplodeDialog?.close();
  }

  async function undoExplode() {
    if (!selected) return;
    closeUndoExplodeDialog();
    await mutate(`/api/admin/users/${encodeURIComponent(selected.user.id)}/stats/explode/undo`, {
      method: 'POST',
    });
  }

  async function setPlushie(series: string, key: string, name: string, owned: boolean) {
    if (!selected || mutatingPlushie) return;
    if (!owned) {
      pendingPlushie = { series, key, name };
      plushieDialog?.showModal();
      return;
    }
    const plushieID = `${series}:${key}`;
    mutatingPlushie = plushieID;
    const base = `/api/admin/users/${encodeURIComponent(selected.user.id)}/collections/${encodeURIComponent(series)}/${encodeURIComponent(key)}`;
    try {
      await mutate(base, { method: owned ? 'DELETE' : 'PUT' });
    } finally {
      mutatingPlushie = null;
    }
  }

  function closePlushieDialog() {
    plushieDialog?.close();
    pendingPlushie = null;
  }

  async function grantPlushie(triggerOverlay: boolean) {
    if (!selected) return;
    const plushie = pendingPlushie;
    closePlushieDialog();
    if (!plushie) return;
    const plushieID = `${plushie.series}:${plushie.key}`;
    mutatingPlushie = plushieID;
    try {
      await mutate(
        `/api/admin/users/${encodeURIComponent(selected.user.id)}/collections/${encodeURIComponent(plushie.series)}/${encodeURIComponent(plushie.key)}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ triggerOverlay }),
        },
      );
    } finally {
      mutatingPlushie = null;
    }
  }

  function openRandomPlushieDialog(collection: Collection) {
    pendingRandomCollection = collection;
    randomPlushieDialog?.showModal();
  }

  function closeRandomPlushieDialog() {
    randomPlushieDialog?.close();
    pendingRandomCollection = null;
  }

  async function grantRandomPlushie(triggerOverlay: boolean) {
    if (!selected) return;
    const collection = pendingRandomCollection;
    closeRandomPlushieDialog();
    if (!collection) return;
    await mutate(
      `/api/admin/users/${encodeURIComponent(selected.user.id)}/collections/${encodeURIComponent(collection.config.series)}/random`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ triggerOverlay }),
      },
    );
  }

  function openResetDialog(collection: Collection) {
    pendingResetCollection = collection;
    resetDialog?.showModal();
  }

  function closeResetDialog() {
    resetDialog?.close();
    pendingResetCollection = null;
  }

  async function resetSeries() {
    if (!selected) return;
    const collection = pendingResetCollection;
    closeResetDialog();
    if (!collection) return;
    await mutate(
      `/api/admin/users/${encodeURIComponent(selected.user.id)}/collections/${encodeURIComponent(collection.config.series)}`,
      {
        method: 'DELETE',
      },
    );
  }

  async function mutate(url: string, options: RequestInit): Promise<boolean> {
    loading = true;
    error = '';
    statusMessage = 'Saving changes…';
    try {
      selected = await request<UserDetail>(url, options);
      statusMessage = 'Changes saved.';
      return true;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Could not save change';
      statusMessage = '';
      return false;
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Charsibot Admin</title>
</svelte:head>

<main class="admin-shell p-6 sm:p-10" aria-busy={loading}>
  <div class="admin-frame mx-auto max-w-[96rem]">
    <header class="admin-header mb-8">
      <a class="back-link" href={resolve('/')}>← Overlay</a>
      <p class="eyebrow mt-6">Control room</p>
      <h1 class="admin-title mt-2">Charsibot Admin</h1>
      <p class="admin-subtitle mt-2">
        Local-only controls for viewer stats and blind-box collections.
      </p>
    </header>

    <p class="sr-only" role="status">{statusMessage}</p>

    <div class="admin-layout">
      <aside class="viewer-directory">
        <section class="admin-surface p-5">
          <div class="flex items-center justify-between gap-4">
            <div>
              <p class="eyebrow">Viewer directory</p>
              <h2 class="section-title mt-1">Viewers</h2>
              <p class="admin-muted text-sm">
                {filteredUsers.length} of {users.length} known users
              </p>
            </div>
            <button class="button button-secondary" onclick={loadUsers} disabled={loading}>
              Refresh
            </button>
          </div>

          {#if users.length > 0}
            <label class="sr-only" for="username-filter">Filter by username</label>
            <input
              id="username-filter"
              class="admin-input mt-4 w-full"
              bind:value={usernameFilter}
              placeholder="Filter by username…"
            />
            <ul class="viewer-list mt-4" aria-label="Viewers">
              {#each filteredUsers as user (user.id)}
                <li>
                  <button
                    class={['viewer-row', selected?.user.id === user.id && 'is-selected']}
                    onclick={() => selectUser(user)}
                    disabled={loading}
                    aria-current={selected?.user.id === user.id ? 'true' : undefined}
                    aria-label={`Manage ${user.username}`}
                  >
                    <span class="truncate font-medium">{user.username}</span>
                    <span class="viewer-row-action" aria-hidden="true">Manage</span>
                  </button>
                </li>
              {/each}
            </ul>
            {#if filteredUsers.length === 0}
              <p class="admin-muted mt-3 text-sm">No usernames match that filter.</p>
            {/if}
          {:else if !loading}
            <p class="admin-muted mt-4 text-sm">No users have been recorded yet.</p>
          {/if}
        </section>
      </aside>

      <div class="admin-workspace">
        {#if error}
          <p class="admin-error px-4 py-3" role="alert">
            {error}
          </p>
        {/if}

        {#if selected}
          <section class="user-detail">
            <div
              class="user-detail-header mb-4 flex flex-wrap items-baseline justify-between gap-3"
            >
              <h2 class="section-title text-2xl" bind:this={selectedUserHeading} tabindex="-1">
                {selected.user.username}
              </h2>
              <div class="flex items-center gap-3">
                <button
                  class="button button-secondary"
                  onclick={openRandomStatDialog}
                  disabled={loading}
                >
                  Grant random stat
                </button>
                <button class="button button-danger" onclick={() => explodeDialog?.showModal()} disabled={loading}>
                  Explode
                </button>
                <button class="button button-secondary" onclick={() => undoExplodeDialog?.showModal()} disabled={loading}>
                  Undo explode
                </button>
                <span class="admin-muted font-mono text-xs">{selected.user.id}</span>
              </div>
            </div>

            <section aria-labelledby="stats-heading">
              <h3 class="detail-section-title" id="stats-heading">Stats</h3>
              <div class="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {#each selected.stats as stat (stat.name)}
                  <div class="stat-card p-4 text-center">
                    <p class="font-semibold">
                      {stat.longName} <span class="admin-muted">({stat.shortName})</span>
                    </p>
                    <div class="mt-3 flex items-center justify-center gap-2">
                      <button
                        class="stat-stepper"
                        onclick={() => updateStat(stat, -1, 'adjust')}
                        disabled={loading}
                        aria-label={`Decrease ${stat.longName} by 1`}>−</button
                      >
                      <input
                        class="stat-input w-20 appearance-none px-3 py-2 text-center tabular-nums [-moz-appearance:textfield] [&::-webkit-inner-spin-button]:m-0 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:m-0 [&::-webkit-outer-spin-button]:appearance-none"
                        type="number"
                        value={stat.value}
                        aria-label={`Set ${stat.longName}`}
                        onchange={async (event) => {
                          const input = event.currentTarget;
                          const saved = await updateStat(stat, Number(input.value), 'set');
                          if (!saved) input.value = String(stat.value);
                        }}
                        disabled={loading}
                      />
                      <button
                        class="stat-stepper"
                        onclick={() => updateStat(stat, 1, 'adjust')}
                        disabled={loading}
                        aria-label={`Increase ${stat.longName} by 1`}>+</button
                      >
                    </div>
                  </div>
                {/each}
              </div>
            </section>

            <section class="mt-8" aria-labelledby="blind-boxes-heading">
              <h3 class="detail-section-title" id="blind-boxes-heading">Blind boxes</h3>
              <div class="mt-4 grid gap-6 lg:grid-cols-2">
                {#each selected.collections as collection (collection.config.series)}
                  <section class="collection-card p-5">
                    <div class="mb-4 flex items-center justify-between gap-4">
                      <div>
                        <h3 class="font-bold">{collection.config.name}</h3>
                        <p class="admin-muted text-sm">
                          {collection.collected.length}/{collection.config.plushies.length} collected
                        </p>
                      </div>
                      <div class="flex flex-wrap justify-end gap-2">
                        <button
                          class="button button-secondary"
                          onclick={() => openRandomPlushieDialog(collection)}
                          disabled={loading}
                        >
                          Grant random
                        </button>
                        <button
                          class="button button-danger"
                          onclick={() => openResetDialog(collection)}
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
                          class={['plushie-button p-2 text-left', !owned && 'is-unowned']}
                          onclick={() => setPlushie(collection.config.series, plushie.key, plushie.name, owned)}
                          disabled={mutatingPlushie === plushieID}
                          title={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
                          aria-label={owned ? `Remove ${plushie.name}` : `Grant ${plushie.name}`}
                          aria-pressed={owned}
                        >
                          <img
                            class="mx-auto h-16 w-16 object-contain"
                            src={plushie.image}
                            alt=""
                          />
                          <span class="mt-1 block truncate text-center text-xs">{plushie.name}</span
                          >
                        </button>
                      {/each}
                    </div>
                  </section>
                {/each}
              </div>
            </section>
          </section>
        {:else if !loading}
          <section class="empty-workspace">
            <p class="eyebrow">Ready when you are</p>
            <h2 class="section-title mt-2 text-2xl">Choose a viewer to manage</h2>
            <p class="admin-muted mt-2 max-w-md">
              Their stats and blind-box collections will appear here.
            </p>
          </section>
        {/if}
      </div>
    </div>
  </div>

<dialog
  class="admin-dialog p-6"
  bind:this={randomPlushieDialog}
  aria-labelledby="random-plushie-dialog-title"
  oncancel={() => {
    pendingRandomCollection = null;
  }}
>
  <p class="eyebrow">Blind-box redemption</p>
  <h2 class="section-title mt-2 text-2xl" id="random-plushie-dialog-title">
    Grant a random plushie?
  </h2>
  <p class="admin-muted mt-2">
    {pendingRandomCollection?.config.name ?? 'This series'} uses its normal weighted drop chances.
  </p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closeRandomPlushieDialog}>Cancel</button>
    <button class="button button-secondary" onclick={() => grantRandomPlushie(false)}>
      Grant silently
    </button>
    <button class="button button-primary" onclick={() => grantRandomPlushie(true)}>
      Grant &amp; show overlay
    </button>
  </div>
</dialog>

<dialog
  class="admin-dialog p-6"
  bind:this={resetDialog}
  aria-labelledby="reset-dialog-title"
  oncancel={() => {
    pendingResetCollection = null;
  }}
>
  <p class="eyebrow">Blind-box collection</p>
  <h2 class="section-title mt-2 text-2xl" id="reset-dialog-title">
    Reset {pendingResetCollection?.config.name ?? 'this collection'}?
  </h2>
  <p class="admin-muted mt-2">
    This permanently removes all {pendingResetCollection?.collected.length ?? 0} collected plushies.
  </p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closeResetDialog}>Cancel</button>
    <button class="button button-danger" onclick={resetSeries}>Reset collection</button>
  </div>
</dialog>

<dialog
  class="admin-dialog p-6"
  bind:this={plushieDialog}
  aria-labelledby="plushie-dialog-title"
  oncancel={() => {
    pendingPlushie = null;
  }}
>
  <p class="eyebrow">Blind-box redemption</p>
  <h2 class="section-title mt-2 text-2xl" id="plushie-dialog-title">
    Grant {pendingPlushie?.name ?? 'this plushie'}?
  </h2>
  <p class="admin-muted mt-2">Choose whether to announce this redemption on the overlay.</p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closePlushieDialog}>Cancel</button>
    <button class="button button-secondary" onclick={() => grantPlushie(false)}>Grant silently</button>
    <button class="button button-primary" onclick={() => grantPlushie(true)}>
      Grant &amp; show overlay
    </button>
  </div>
</dialog>

<dialog
  class="admin-dialog p-6"
  bind:this={explodeDialog}
  aria-labelledby="explode-dialog-title"
>
  <p class="eyebrow">Viewer stat</p>
  <h2 class="section-title mt-2 text-2xl" id="explode-dialog-title">
    Explode {selected?.user.username ?? 'this viewer'}?
  </h2>
  <p class="admin-muted mt-2">This sets their PENIS stat to -1000 and displays their updated stats in chat.</p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closeExplodeDialog}>Cancel</button>
    <button class="button button-danger" onclick={explode}>Explode</button>
  </div>
</dialog>

<dialog
  class="admin-dialog p-6"
  bind:this={undoExplodeDialog}
  aria-labelledby="undo-explode-dialog-title"
>
  <p class="eyebrow">Viewer stat</p>
  <h2 class="section-title mt-2 text-2xl" id="undo-explode-dialog-title">
    Undo {selected?.user.username ?? 'this viewer'}'s explosion?
  </h2>
  <p class="admin-muted mt-2">This restores their PENIS stat to its configured default and displays their updated stats in chat.</p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closeUndoExplodeDialog}>Cancel</button>
    <button class="button button-primary" onclick={undoExplode}>Undo explode</button>
  </div>
</dialog>

<dialog
  class="admin-dialog p-6"
  bind:this={randomStatDialog}
  aria-labelledby="random-stat-dialog-title"
>
  <p class="eyebrow">Random stat</p>
  <h2 class="section-title mt-2 text-2xl" id="random-stat-dialog-title">Grant a random stat?</h2>
  <p class="admin-muted mt-2">Choose whether to display the viewer's updated stats in chat.</p>
  <div class="dialog-actions mt-6">
    <button class="button button-secondary" onclick={closeRandomStatDialog}>Cancel</button>
    <button class="button button-secondary" onclick={() => grantRandomStat(false)}>
      Grant silently
    </button>
    <button class="button button-primary" onclick={() => grantRandomStat(true)}>
      Grant &amp; display stats
    </button>
  </div>
</dialog>
</main>

<style>
  .admin-shell {
    --ink: #15111c;
    --panel: #211a2a;
    --panel-raised: #2b2236;
    --line: #4b3c58;
    --text: #fff8ff;
    --muted: #d6c6df;
    --accent: #f2a1ba;
    --accent-strong: #ffbfce;
    --danger: #ff9a9a;
    min-height: 100vh;
    color: var(--text);
    background:
      radial-gradient(circle at 8% 0%, rgb(148 80 127 / 28%), transparent 30rem),
      radial-gradient(circle at 92% 15%, rgb(237 153 124 / 16%), transparent 26rem), var(--ink);
  }

  .admin-frame {
    position: relative;
  }

  .admin-layout {
    display: grid;
    gap: 1.5rem;
    align-items: start;
  }

  .viewer-directory,
  .admin-workspace {
    min-width: 0;
  }

  .admin-workspace {
    display: grid;
    gap: 1.25rem;
  }

  .admin-header {
    border-bottom: 1px solid rgb(214 198 223 / 24%);
    padding-bottom: 2rem;
  }

  .eyebrow {
    color: var(--accent-strong);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.18em;
    text-transform: uppercase;
  }

  .admin-title,
  .section-title {
    font-family: Nunito, sans-serif;
    font-weight: 800;
    letter-spacing: -0.035em;
  }

  .admin-title {
    font-size: clamp(2.25rem, 6vw, 3.5rem);
    line-height: 0.95;
  }

  .section-title {
    line-height: 1.05;
  }

  .detail-section-title {
    color: var(--muted);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }

  .admin-subtitle,
  .admin-muted {
    color: var(--muted);
  }

  .back-link {
    color: var(--muted);
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .back-link:hover {
    color: var(--accent-strong);
  }

  .admin-surface,
  .stat-card,
  .collection-card {
    border: 1px solid rgb(214 198 223 / 18%);
    background: linear-gradient(145deg, rgb(43 34 54 / 96%), rgb(31 24 41 / 96%));
    box-shadow: 0 24px 60px rgb(5 3 9 / 26%);
  }

  .admin-surface {
    border-radius: 1.25rem;
  }

  .stat-card {
    border-radius: 1rem;
  }

  .collection-card {
    border-radius: 1.25rem;
  }

  .button,
  .stat-stepper,
  .plushie-button {
    transition:
      transform 150ms ease,
      background-color 150ms ease,
      border-color 150ms ease,
      color 150ms ease;
  }

  .button:hover:not(:disabled),
  .stat-stepper:hover:not(:disabled),
  .plushie-button:hover:not(:disabled) {
    transform: translateY(-1px);
  }

  .button:focus-visible,
  .stat-stepper:focus-visible,
  .plushie-button:focus-visible,
  .admin-input:focus-visible,
  .stat-input:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: 3px;
  }

  .button:disabled,
  .stat-stepper:disabled,
  .plushie-button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .button-secondary,
  .stat-stepper {
    border: 1px solid var(--line);
    background: rgb(255 255 255 / 6%);
    color: var(--text);
    font-weight: 800;
  }

  .button-secondary {
    border-radius: 999px;
    padding: 0.5rem 1rem;
    font-size: 0.75rem;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .button-secondary:hover:not(:disabled),
  .stat-stepper:hover:not(:disabled) {
    border-color: var(--accent);
    background: rgb(242 161 186 / 16%);
  }

  .button-primary {
    border: 1px solid var(--accent);
    border-radius: 999px;
    padding: 0.5rem 1rem;
    background: var(--accent);
    color: var(--ink);
    font-size: 0.75rem;
    font-weight: 900;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .button-primary:hover:not(:disabled) {
    border-color: var(--accent-strong);
    background: var(--accent-strong);
  }

  .admin-input,
  .stat-input {
    border: 1px solid var(--line);
    border-radius: 0.7rem;
    background: rgb(10 7 15 / 44%);
    color: var(--text);
  }

  .admin-input {
    padding: 0.65rem 0.9rem;
  }

  .admin-input::placeholder {
    color: var(--muted);
  }

  .viewer-list {
    max-height: 22rem;
    overflow-y: auto;
    border: 1px solid var(--line);
    border-radius: 0.85rem;
  }

  .viewer-list li + li {
    border-top: 1px solid rgb(214 198 223 / 12%);
  }

  .viewer-row {
    display: flex;
    width: 100%;
    min-width: 0;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.85rem 1rem;
    color: var(--text);
    text-align: left;
  }

  .viewer-row:hover:not(:disabled) {
    background: rgb(255 255 255 / 4%);
  }

  .viewer-row.is-selected {
    background: rgb(242 161 186 / 13%);
    box-shadow: inset 3px 0 var(--accent);
  }

  .viewer-row:focus-visible {
    outline: 2px solid var(--accent-strong);
    outline-offset: -2px;
  }

  .viewer-row:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .viewer-row-action {
    flex: none;
    color: var(--accent-strong);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .user-detail-header {
    border-bottom: 1px solid rgb(214 198 223 / 22%);
    padding-bottom: 1rem;
  }

  .empty-workspace {
    min-height: 18rem;
    border: 1px dashed rgb(214 198 223 / 30%);
    border-radius: 1.25rem;
    padding: clamp(2rem, 8vw, 5rem);
    background: rgb(43 34 54 / 34%);
  }

  .stat-stepper {
    min-width: 2.5rem;
    border-radius: 0.55rem;
    padding: 0.5rem 0.75rem;
  }

  .button-danger {
    border: 1px solid rgb(255 154 154 / 60%);
    border-radius: 999px;
    padding: 0.45rem 0.8rem;
    color: #ffd1d1;
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }

  .button-danger:hover:not(:disabled) {
    background: rgb(255 108 108 / 15%);
  }

  .admin-dialog {
    position: fixed;
    inset: 0;
    width: min(100% - 2rem, 32rem);
    height: fit-content;
    margin: auto;
    border: 1px solid rgb(214 198 223 / 26%);
    border-radius: 1.25rem;
    background: linear-gradient(145deg, rgb(43 34 54), rgb(31 24 41));
    color: var(--text);
    box-shadow: 0 24px 60px rgb(5 3 9 / 52%);
  }

  .admin-dialog::backdrop {
    background: rgb(10 7 15 / 70%);
  }

  .dialog-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .plushie-button {
    border: 1px solid rgb(214 198 223 / 18%);
    border-radius: 0.75rem;
    background: rgb(255 255 255 / 5%);
    color: var(--text);
  }

  .plushie-button[aria-pressed='true'] {
    border-color: rgb(242 161 186 / 80%);
    background: linear-gradient(145deg, rgb(242 161 186 / 18%), rgb(255 255 255 / 5%));
  }

  .plushie-button.is-unowned {
    opacity: 0.48;
  }

  .admin-error {
    border: 1px solid rgb(255 154 154 / 56%);
    border-radius: 0.85rem;
    background: rgb(116 38 55 / 42%);
    color: #ffe0e0;
  }

  @media (min-width: 900px) {
    .admin-layout {
      grid-template-columns: minmax(17rem, 21rem) minmax(0, 1fr);
      gap: 2rem;
    }

    .viewer-directory {
      position: sticky;
      top: 1.25rem;
    }

    .viewer-list {
      max-height: calc(100vh - 19rem);
      min-height: 16rem;
    }
  }
</style>
