<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { AdminUserDetail, ActivityFilter, GrantResult } from '$lib/admin/types';
  import type { ViewerCollection } from '$lib/contracts/collections';
  import type { UserStat, Viewer } from '$lib/contracts/viewer';
  import UserCollections from '$lib/admin/UserCollections.svelte';
  import UserStats from '$lib/admin/UserStats.svelte';
  import ViewerDirectory from '$lib/admin/ViewerDirectory.svelte';
  import {
    listViewers as listViewersRemote,
    getViewer as getViewerRemote,
    deleteViewer as deleteViewerRemote,
    deleteViewers as deleteViewersRemote,
    updateStat as updateStatRemote,
    displayStats as displayStatsRemote,
    grantRandomStat as grantRandomStatRemote,
    explode as explodeRemote,
    undoExplode as undoExplodeRemote,
    resetStats as resetStatsRemote,
    grantPlushie as grantPlushieRemote,
    grantRandomPlushie as grantRandomPlushieRemote,
    removePlushie as removePlushieRemote,
    resetCollection as resetCollectionRemote,
    displayCollection as displayCollectionRemote,
  } from '$lib/admin/admin.remote';

  interface PendingPlushie {
    series: string;
    key: string;
    name: string;
  }

  let users = $state.raw<Viewer[]>([]);
  let usernameFilter = $state('');
  let activityFilter = $state<ActivityFilter>('all');
  let selectedUserIDs = $state.raw<string[]>([]);
  let filteredUsers = $derived.by(() => {
    const query = usernameFilter.trim().toLowerCase();
    const now = Date.now();
    const matchesActivity = (user: Viewer) => {
      if (activityFilter === 'all') return true;
      if (activityFilter === 'unknown') return !user.lastActiveAt;
      if (activityFilter === 'recent')
        return !!user.lastActiveAt && now - Date.parse(user.lastActiveAt) < 30 * 86400000;
      const days = activityFilter === 'inactive30' ? 30 : 90;
      return !!user.lastActiveAt && now - Date.parse(user.lastActiveAt) >= days * 86400000;
    };
    return users
      .filter((user) => user.username.toLowerCase().includes(query) && matchesActivity(user))
      .toSorted(
        (a, b) =>
          (a.lastActiveAt ?? '').localeCompare(b.lastActiveAt ?? '') ||
          a.username.localeCompare(b.username),
      );
  });
  let selected = $state.raw<AdminUserDetail | null>(null);
  let loading = $state(false);
  let mutatingPlushie = $state<string | null>(null);
  let pendingRandomCollection = $state.raw<ViewerCollection | null>(null);
  let randomPlushieDialog = $state<HTMLDialogElement | undefined>(undefined);
  let pendingPlushie = $state.raw<PendingPlushie | null>(null);
  let plushieDialog = $state<HTMLDialogElement | undefined>(undefined);
  let pendingResetCollection = $state.raw<ViewerCollection | null>(null);
  let resetDialog = $state<HTMLDialogElement | undefined>(undefined);
  let explodeDialog = $state<HTMLDialogElement | undefined>(undefined);
  let undoExplodeDialog = $state<HTMLDialogElement | undefined>(undefined);
  let resetStatsDialog = $state<HTMLDialogElement | undefined>(undefined);
  let deleteUserDialog = $state<HTMLDialogElement | undefined>(undefined);
  let bulkDeleteDialog = $state<HTMLDialogElement | undefined>(undefined);
  let randomStatDialog = $state<HTMLDialogElement | undefined>(undefined);
  let error = $state('');
  let statusMessage = $state('');
  let lastGrant = $state.raw<GrantResult | null>(null);
  let selectedUserHeading = $state<HTMLHeadingElement | undefined>(undefined);
  let viewerManagementDialog = $state<HTMLDialogElement | undefined>(undefined);
  let userSearchDialog = $state<HTMLDialogElement | undefined>(undefined);
  let userSearchInput = $state<HTMLInputElement | undefined>(undefined);
  let userSearchQuery = $state('');
  let userSearchIndex = $state(0);
  let userSearchResultElements = $state.raw<(HTMLButtonElement | undefined)[]>([]);
  let userSearchResults = $derived(
    users
      .filter((user) => user.username.toLowerCase().includes(userSearchQuery.trim().toLowerCase()))
      .toSorted((a, b) => a.username.localeCompare(b.username)),
  );
  let userSearchSelectionID: string | undefined;
  let selectedUserRequest = 0;

  const errorMessage = (err: unknown): string => {
    if (err instanceof Error) return err.message;
    return String(err);
  };

  const isCurrentUserRequest = (requestID: number, userID: string) => {
    return selectedUserRequest === requestID && page.url.searchParams.get('user') === userID;
  };

  const loadUsers = async () => {
    loading = true;
    error = '';
    statusMessage = 'Loading viewers…';
    try {
      users = await listViewersRemote();
      statusMessage = `Loaded ${users.length} viewers.`;
    } catch (err) {
      error = errorMessage(err);
      statusMessage = '';
    } finally {
      loading = false;
    }
  };

  onMount(() => {
    void initialise();
  });

  afterNavigate(() => {
    if (users.length > 0) void syncSelectedUserFromURL();
  });

  const initialise = async () => {
    await loadUsers();
    await syncSelectedUserFromURL();
  };

  const syncSelectedUserFromURL = async () => {
    const userID = page.url.searchParams.get('user');
    const user = users.find((candidate) => candidate.id === userID);
    if (user && selected?.user.id !== user.id) {
      await selectUser(user, false);
    } else if (!user) {
      selectedUserRequest += 1;
      selected = null;
      lastGrant = null;
    }
  };

  const selectUser = async (user: Viewer, updateURL = true) => {
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
    const suppressSelectedUserFocus = userSearchSelectionID === user.id;
    if (suppressSelectedUserFocus) userSearchSelectionID = undefined;
    const requestID = ++selectedUserRequest;
    loading = true;
    error = '';
    lastGrant = null;
    statusMessage = `Loading ${user.username}…`;
    try {
      const response = await getViewerRemote({ userID: user.id });
      if (!isCurrentUserRequest(requestID, user.id)) return;
      selected = response;
      lastGrant = null;
      statusMessage = `Loaded ${user.username}.`;
      await tick();
      if (!suppressSelectedUserFocus) selectedUserHeading?.focus();
    } catch (err) {
      if (isCurrentUserRequest(requestID, user.id)) {
        error = errorMessage(err);
        statusMessage = '';
      }
    } finally {
      if (isCurrentUserRequest(requestID, user.id)) loading = false;
    }
  };

  const openUserSearch = async () => {
    if (!userSearchDialog?.open) userSearchDialog?.showModal();
    await tick();
    userSearchInput?.focus();
    userSearchInput?.select();
  };

  const closeUserSearch = () => {
    userSearchDialog?.close();
    userSearchQuery = '';
    userSearchIndex = 0;
  };

  const selectUserSearchResult = async (user: Viewer) => {
    userSearchSelectionID = user.id;
    closeUserSearch();
    await selectUser(user);
  };

  const selectUserFromManagement = async (user: Viewer) => {
    viewerManagementDialog?.close();
    await selectUser(user);
  };

  const handleUserSearchInput = (value: string) => {
    userSearchQuery = value;
    userSearchIndex = 0;
  };

  const moveUserSearchSelection = async (offset: number) => {
    userSearchIndex =
      (userSearchIndex + offset + userSearchResults.length) % userSearchResults.length;
    await tick();
    userSearchResultElements[userSearchIndex]?.scrollIntoView({ block: 'nearest' });
  };

  const handleUserSearchKeydown = (event: KeyboardEvent) => {
    if (event.key === 'ArrowDown' && userSearchResults.length > 0) {
      event.preventDefault();
      void moveUserSearchSelection(1);
    } else if (event.key === 'ArrowUp' && userSearchResults.length > 0) {
      event.preventDefault();
      void moveUserSearchSelection(-1);
    } else if (event.key === 'Enter') {
      const user = userSearchResults[userSearchIndex];
      if (user) {
        event.preventDefault();
        void selectUserSearchResult(user);
      }
    }
  };

  const handleGlobalKeydown = (event: KeyboardEvent) => {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      void openUserSearch();
    }
  };

  const updateStat = async (
    stat: UserStat,
    value: number,
    mode: 'set' | 'adjust',
  ): Promise<boolean> => {
    if (!selected || !Number.isFinite(value)) return false;
    return mutate(updateStatRemote({ userID: selected.user.id, statName: stat.name, mode, value }));
  };

  const displayStatsInChat = async () => {
    if (!selected) return;
    await mutate(displayStatsRemote({ userID: selected.user.id }));
  };

  const displayCollection = async (collection: ViewerCollection) => {
    if (!selected) return;
    await mutate(
      displayCollectionRemote({ userID: selected.user.id, series: collection.config.series }),
    );
  };

  const closeDeleteUserDialog = () => {
    deleteUserDialog?.close();
  };

  const deleteUser = async () => {
    const user = selected?.user;
    if (!user) return;
    closeDeleteUserDialog();
    loading = true;
    error = '';
    try {
      await deleteViewerRemote({ userID: user.id });
      users = users.filter((candidate) => candidate.id !== user.id);
      selected = null;
      statusMessage = `Deleted ${user.username}.`;
      await goto(resolve('/admin'), { keepFocus: true, noScroll: true });
    } catch (err) {
      error = errorMessage(err);
      statusMessage = '';
    } finally {
      loading = false;
    }
  };

  const toggleUserSelection = (userID: string, checked: boolean) => {
    selectedUserIDs = checked
      ? [...new Set([...selectedUserIDs, userID])]
      : selectedUserIDs.filter((candidate) => candidate !== userID);
  };

  const selectFilteredUsers = () => {
    selectedUserIDs = filteredUsers.map((user) => user.id);
  };

  const clearUserSelection = () => {
    selectedUserIDs = [];
  };

  const closeBulkDeleteDialog = () => {
    bulkDeleteDialog?.close();
  };

  const deleteSelectedUsers = async () => {
    const userIDs = [...selectedUserIDs];
    if (userIDs.length === 0) return;

    closeBulkDeleteDialog();
    loading = true;
    error = '';
    try {
      await deleteViewersRemote({ userIDs });

      users = users.filter((user) => !userIDs.includes(user.id));
      selectedUserIDs = [];
      if (selected && userIDs.includes(selected.user.id)) {
        selected = null;
        await goto(resolve('/admin'), { keepFocus: true, noScroll: true });
      }
      statusMessage = `Deleted ${userIDs.length} viewers.`;
    } catch (err) {
      error = errorMessage(err);
      statusMessage = '';
    } finally {
      loading = false;
    }
  };

  const formatLastActive = (value: string | undefined) => {
    if (!value) return 'Unknown';
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
      new Date(value),
    );
  };

  const openRandomStatDialog = () => {
    randomStatDialog?.showModal();
  };

  const closeRandomStatDialog = () => {
    randomStatDialog?.close();
  };

  const grantRandomStat = async (displayInChat: boolean) => {
    if (!selected) return;
    closeRandomStatDialog();
    lastGrant = null;
    await mutate(grantRandomStatRemote({ userID: selected.user.id, displayInChat }));
  };

  const closeExplodeDialog = () => {
    explodeDialog?.close();
  };

  const handleExplode = async () => {
    if (!selected) return;
    closeExplodeDialog();
    await mutate(explodeRemote({ userID: selected.user.id }));
  };

  const closeUndoExplodeDialog = () => {
    undoExplodeDialog?.close();
  };

  const handleUndoExplode = async () => {
    if (!selected) return;
    closeUndoExplodeDialog();
    await mutate(undoExplodeRemote({ userID: selected.user.id }));
  };

  const closeResetStatsDialog = () => {
    resetStatsDialog?.close();
  };

  const resetStats = async (displayInChat: boolean) => {
    if (!selected) return;
    closeResetStatsDialog();
    await mutate(resetStatsRemote({ userID: selected.user.id, displayInChat }));
  };

  const setPlushie = async (series: string, key: string, name: string, owned: boolean) => {
    if (!selected || mutatingPlushie) return;
    if (!owned) {
      pendingPlushie = { series, key, name };
      plushieDialog?.showModal();
      return;
    }
    const plushieID = `${series}:${key}`;
    mutatingPlushie = plushieID;
    try {
      await mutate(removePlushieRemote({ userID: selected.user.id, series, key }));
    } finally {
      mutatingPlushie = null;
    }
  };

  const closePlushieDialog = () => {
    plushieDialog?.close();
    pendingPlushie = null;
  };

  const grantPlushie = async (triggerOverlay: boolean) => {
    if (!selected) return;
    const plushie = pendingPlushie;
    closePlushieDialog();
    if (!plushie) return;
    const plushieID = `${plushie.series}:${plushie.key}`;
    mutatingPlushie = plushieID;
    try {
      await mutate(
        grantPlushieRemote({
          userID: selected.user.id,
          series: plushie.series,
          key: plushie.key,
          triggerOverlay,
        }),
      );
    } finally {
      mutatingPlushie = null;
    }
  };

  const openRandomPlushieDialog = (collection: ViewerCollection) => {
    pendingRandomCollection = collection;
    randomPlushieDialog?.showModal();
  };

  const closeRandomPlushieDialog = () => {
    randomPlushieDialog?.close();
    pendingRandomCollection = null;
  };

  const grantRandomPlushie = async (triggerOverlay: boolean) => {
    if (!selected) return;
    const collection = pendingRandomCollection;
    closeRandomPlushieDialog();
    if (!collection) return;
    lastGrant = null;
    await mutate(
      grantRandomPlushieRemote({
        userID: selected.user.id,
        series: collection.config.series,
        triggerOverlay,
      }),
    );
  };

  const openResetDialog = (collection: ViewerCollection) => {
    pendingResetCollection = collection;
    resetDialog?.showModal();
  };

  const closeResetDialog = () => {
    resetDialog?.close();
    pendingResetCollection = null;
  };

  const resetSeries = async () => {
    if (!selected) return;
    const collection = pendingResetCollection;
    closeResetDialog();
    if (!collection) return;
    await mutate(
      resetCollectionRemote({ userID: selected.user.id, series: collection.config.series }),
    );
  };

  const mutate = async (operation: Promise<AdminUserDetail>): Promise<boolean> => {
    const userID = selected?.user.id;
    if (!userID) return false;
    const requestID = ++selectedUserRequest;
    loading = true;
    error = '';
    statusMessage = 'Saving changes…';
    try {
      const response = await operation;
      if (!isCurrentUserRequest(requestID, userID)) return false;
      selected = response;
      lastGrant = response.grant ?? null;
      statusMessage = 'Changes saved.';
      return true;
    } catch (err) {
      if (isCurrentUserRequest(requestID, userID)) {
        error = errorMessage(err);
        statusMessage = '';
      }
      return false;
    } finally {
      if (isCurrentUserRequest(requestID, userID)) loading = false;
    }
  };
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<svelte:head>
  <title>Charsibot Admin</title>
</svelte:head>

<main class="admin-shell p-6 sm:p-10" aria-busy={loading}>
  <div class="admin-frame mx-auto">
    <p class="sr-only" role="status">{statusMessage}</p>

    <div class="admin-layout">
      <div class="admin-content">
        <header class="admin-header flex flex-col gap-6">
          <a class="back-link" href={resolve('/')}>← Overlay</a>
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="flex flex-col gap-2">
              <p class="eyebrow">Control room</p>
              <h1 class="admin-title">Charsibot Admin</h1>
              <p class="admin-subtitle">Controls for viewer stats and blind-box collections.</p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <button class="user-search-trigger" onclick={() => void openUserSearch()}>
                Search viewers <kbd>⌘ K</kbd>
              </button>
              <button
                class="button button-secondary"
                onclick={() => viewerManagementDialog?.showModal()}
              >
                Manage viewers
              </button>
            </div>
          </div>
        </header>

        <div class="admin-workspace">
          {#if error}
            <p class="admin-error px-4 py-3" role="alert">
              {error}
            </p>
          {/if}

          {#if selected}
            <section class="user-detail flex flex-col gap-8">
              <div class="flex flex-col gap-4">
                <div class="user-detail-header">
                  <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-2">
                    <div>
                      <div class="flex min-w-0 items-center gap-2">
                        <h2
                          class="section-title truncate text-2xl"
                          bind:this={selectedUserHeading}
                          tabindex="-1"
                        >
                          {selected.user.username}
                        </h2>
                        <span
                          class="admin-muted shrink-0 rounded-full border border-(--line) px-2 py-0.5 font-mono text-[0.65rem]"
                        >
                          {selected.user.id}
                        </span>
                      </div>
                      <p class="admin-muted text-xs">
                        Last active: {formatLastActive(selected.user.lastActiveAt)}
                      </p>
                    </div>
                    <button
                      class="button button-danger"
                      onclick={() => deleteUserDialog?.showModal()}
                      disabled={loading}
                    >
                      Delete viewer
                    </button>
                  </div>
                </div>

                {#if lastGrant}
                  {@const grantedStat = selected.stats.find(
                    (stat) => stat.name === lastGrant?.statName,
                  )}
                  <aside
                    class={[
                      'grant-result px-4 py-3',
                      lastGrant.kind === 'plushie' && lastGrant.isDuplicate && 'is-duplicate',
                    ]}
                    role="status"
                    aria-live="polite"
                  >
                    {#if lastGrant.kind === 'stat'}
                      <p class="eyebrow">Random stat granted</p>
                      <p class="grant-result-title">
                        +1 {grantedStat?.longName ?? lastGrant.statName}
                      </p>
                    {:else if lastGrant.isDuplicate}
                      <p class="eyebrow">Duplicate plushie</p>
                      <p class="grant-result-title">{lastGrant.plushieName}</p>
                      <p class="admin-muted text-sm">
                        Already owned. The collection was unchanged.
                      </p>
                    {:else}
                      <p class="eyebrow">Random plushie granted</p>
                      <p class="grant-result-title">{lastGrant.plushieName}</p>
                    {/if}
                  </aside>
                {/if}

                <UserStats
                  stats={selected.stats}
                  {loading}
                  onUpdateStat={updateStat}
                  onDisplayStats={displayStatsInChat}
                  onOpenRandomStat={openRandomStatDialog}
                  onOpenExplode={() => explodeDialog?.showModal()}
                  onOpenUndoExplode={() => undoExplodeDialog?.showModal()}
                  onOpenResetStats={() => resetStatsDialog?.showModal()}
                />
              </div>

              <UserCollections
                collections={selected.collections}
                {loading}
                {mutatingPlushie}
                onDisplayCollection={displayCollection}
                onOpenRandomPlushie={openRandomPlushieDialog}
                onOpenResetCollection={openResetDialog}
                onSetPlushie={setPlushie}
              />
            </section>
          {:else if !loading}
            <section class="empty-workspace flex flex-col gap-2">
              <p class="eyebrow">Ready when you are</p>
              <h2 class="section-title text-2xl">Choose a viewer to manage</h2>
              <p class="admin-muted max-w-md">
                Their stats and blind-box collections will appear here.
              </p>
            </section>
          {/if}
        </div>
      </div>
    </div>
  </div>

  <dialog
    class="admin-dialog user-search-dialog p-0"
    bind:this={userSearchDialog}
    aria-labelledby="user-search-dialog-title"
    onclose={closeUserSearch}
  >
    <div class="user-search-header p-5">
      <label class="sr-only" for="user-search-input">Search viewers by username</label>
      <input
        id="user-search-input"
        class="user-search-input"
        bind:this={userSearchInput}
        value={userSearchQuery}
        oninput={(event) => handleUserSearchInput(event.currentTarget.value)}
        onkeydown={handleUserSearchKeydown}
        placeholder="Search by username…"
        autocomplete="off"
      />
      <p class="user-search-help" id="user-search-dialog-title">
        <span>Search viewers</span>
        <span>↑↓ to navigate · Enter to open · Esc to close</span>
      </p>
    </div>
    <div
      class={['user-search-results', userSearchResults.length > 5 && 'is-scrollable']}
      role="listbox"
      aria-label="Matching viewers"
    >
      {#each userSearchResults as user, index (user.id)}
        <button
          class={['user-search-result', index === userSearchIndex && 'is-active']}
          role="option"
          aria-selected={index === userSearchIndex}
          bind:this={userSearchResultElements[index]}
          onclick={() => void selectUserSearchResult(user)}
          onmousemove={() => (userSearchIndex = index)}
        >
          <span class="font-semibold">{user.username}</span>
          <span class="user-search-user-id">{user.id}</span>
        </button>
      {:else}
        <p class="user-search-empty">No viewers match "{userSearchQuery}".</p>
      {/each}
    </div>
  </dialog>

  <dialog
    class="viewer-management-dialog p-0"
    bind:this={viewerManagementDialog}
    aria-label="Viewer management"
  >
    <ViewerDirectory
      {users}
      {filteredUsers}
      {usernameFilter}
      {activityFilter}
      {selectedUserIDs}
      selectedUserID={selected?.user.id}
      {loading}
      onRefresh={loadUsers}
      onClose={() => viewerManagementDialog?.close()}
      onUsernameFilterChange={(value) => (usernameFilter = value)}
      onActivityFilterChange={(value) => (activityFilter = value)}
      onSelectFiltered={selectFilteredUsers}
      onClearSelection={clearUserSelection}
      onOpenBulkDelete={() => {
        viewerManagementDialog?.close();
        bulkDeleteDialog?.showModal();
      }}
      onToggleUserSelection={toggleUserSelection}
      onSelectUser={selectUserFromManagement}
    />
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={deleteUserDialog}
    aria-labelledby="delete-user-dialog-title"
  >
    <p class="eyebrow">Prune viewer</p>
    <h2 class="section-title text-2xl" id="delete-user-dialog-title">
      Delete {selected?.user.username ?? 'this viewer'}?
    </h2>
    <p class="admin-muted">
      This permanently removes their activity, stats, and blind-box collections.
    </p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeDeleteUserDialog}>Cancel</button>
      <button class="button button-danger" onclick={deleteUser}>Delete viewer</button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={bulkDeleteDialog}
    aria-labelledby="bulk-delete-dialog-title"
  >
    <p class="eyebrow">Prune viewers</p>
    <h2 class="section-title text-2xl" id="bulk-delete-dialog-title">
      Delete {selectedUserIDs.length} selected viewers?
    </h2>
    <p class="admin-muted">
      This permanently removes their activity, stats, and blind-box collections.
    </p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeBulkDeleteDialog}>Cancel</button>
      <button class="button button-danger" onclick={deleteSelectedUsers}>
        Delete {selectedUserIDs.length} viewers
      </button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={randomPlushieDialog}
    aria-labelledby="random-plushie-dialog-title"
    oncancel={() => {
      pendingRandomCollection = null;
    }}
  >
    <p class="eyebrow">Blind-box redemption</p>
    <h2 class="section-title text-2xl" id="random-plushie-dialog-title">Grant a random plushie?</h2>
    <p class="admin-muted">
      {pendingRandomCollection?.config.name ?? 'This series'} uses its normal weighted drop chances.
    </p>
    <div class="dialog-actions pt-4">
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
    class="admin-dialog gap-2 p-6"
    bind:this={resetStatsDialog}
    aria-labelledby="reset-stats-dialog-title"
  >
    <p class="eyebrow">Viewer stats</p>
    <h2 class="section-title text-2xl" id="reset-stats-dialog-title">
      Reset {selected?.user.username ?? 'this viewer'}'s stats?
    </h2>
    <p class="admin-muted">This restores every stat to its configured default.</p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeResetStatsDialog}>Cancel</button>
      <button class="button button-danger" onclick={() => resetStats(false)}>Reset silently</button>
      <button class="button button-primary" onclick={() => resetStats(true)}>
        Reset &amp; display stats
      </button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={resetDialog}
    aria-labelledby="reset-dialog-title"
    oncancel={() => {
      pendingResetCollection = null;
    }}
  >
    <p class="eyebrow">Blind-box collection</p>
    <h2 class="section-title text-2xl" id="reset-dialog-title">
      Reset {pendingResetCollection?.config.name ?? 'this collection'}?
    </h2>
    <p class="admin-muted">
      This permanently removes all {pendingResetCollection?.collected.length ?? 0} collected plushies.
    </p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeResetDialog}>Cancel</button>
      <button class="button button-danger" onclick={resetSeries}>Reset collection</button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={plushieDialog}
    aria-labelledby="plushie-dialog-title"
    oncancel={() => {
      pendingPlushie = null;
    }}
  >
    <p class="eyebrow">Blind-box redemption</p>
    <h2 class="section-title text-2xl" id="plushie-dialog-title">
      Grant {pendingPlushie?.name ?? 'this plushie'}?
    </h2>
    <p class="admin-muted">Choose whether to announce this redemption on the overlay.</p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closePlushieDialog}>Cancel</button>
      <button class="button button-secondary" onclick={() => grantPlushie(false)}
        >Grant silently</button
      >
      <button class="button button-primary" onclick={() => grantPlushie(true)}>
        Grant &amp; show overlay
      </button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={explodeDialog}
    aria-labelledby="explode-dialog-title"
  >
    <p class="eyebrow">Viewer stat</p>
    <h2 class="section-title text-2xl" id="explode-dialog-title">
      Explode {selected?.user.username ?? 'this viewer'}?
    </h2>
    <p class="admin-muted">
      This sets their PENIS stat to -1000 and displays their updated stats in chat.
    </p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeExplodeDialog}>Cancel</button>
      <button class="button button-danger" onclick={handleExplode}>Explode</button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={undoExplodeDialog}
    aria-labelledby="undo-explode-dialog-title"
  >
    <p class="eyebrow">Viewer stat</p>
    <h2 class="section-title text-2xl" id="undo-explode-dialog-title">
      Undo {selected?.user.username ?? 'this viewer'}'s explosion?
    </h2>
    <p class="admin-muted">
      This restores their PENIS stat to its configured default and displays their updated stats in
      chat.
    </p>
    <div class="dialog-actions pt-4">
      <button class="button button-secondary" onclick={closeUndoExplodeDialog}>Cancel</button>
      <button class="button button-primary" onclick={handleUndoExplode}>Undo explode</button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog gap-2 p-6"
    bind:this={randomStatDialog}
    aria-labelledby="random-stat-dialog-title"
  >
    <p class="eyebrow">Random stat</p>
    <h2 class="section-title text-2xl" id="random-stat-dialog-title">Grant a random stat?</h2>
    <p class="admin-muted">Choose whether to display the viewer's updated stats in chat.</p>
    <div class="dialog-actions pt-4">
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
  :global {
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
      grid-template-areas:
        'header'
        'sidebar'
        'workspace';
    }

    .admin-content {
      display: contents;
    }

    .viewer-directory,
    .admin-workspace {
      min-width: 0;
    }

    .admin-header {
      grid-area: header;
    }

    .viewer-directory {
      grid-area: sidebar;
    }

    .admin-workspace {
      display: grid;
      gap: 1.25rem;
      grid-area: workspace;
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

    .button-secondary,
    .button-primary,
    .button-danger {
      flex: none;
      white-space: nowrap;
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

    .viewer-bulk-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .viewer-bulk-actions .button {
      text-align: center;
    }

    .viewer-list li + li {
      border-top: 1px solid rgb(214 198 223 / 12%);
    }

    .viewer-list-item {
      display: flex;
      min-width: 0;
      align-items: center;
      gap: 1rem;
      padding-inline-start: 1rem;
      transition: background-color 150ms ease;
    }

    .viewer-list-item:has(.viewer-selection:not(:disabled)):hover {
      background: rgb(242 161 186 / 16%);
    }

    .viewer-list-item.is-selected {
      background: rgb(242 161 186 / 13%);
      box-shadow: inset 3px 0 var(--accent);
    }

    .viewer-selection {
      width: 1rem;
      height: 1rem;
      flex: none;
      accent-color: var(--accent);
    }

    .viewer-selection:focus-visible {
      outline: 2px solid var(--accent-strong);
      outline-offset: 3px;
    }

    .viewer-selection:hover:not(:disabled) {
      cursor: pointer;
      filter: brightness(1.2);
    }

    .viewer-row {
      display: flex;
      flex: 1;
      min-width: 0;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      padding: 0.85rem 1rem 0.85rem 0;
      color: var(--text);
      text-align: left;
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

    .admin-dialog[open] {
      display: flex;
      flex-direction: column;
    }

    .admin-dialog::backdrop {
      background: rgb(10 7 15 / 70%);
    }

    .user-search-trigger {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      border: 1px solid var(--line);
      border-radius: 0.7rem;
      padding: 0.55rem 0.65rem 0.55rem 0.85rem;
      background: rgb(10 7 15 / 28%);
      color: var(--muted);
      font-size: 0.75rem;
      font-weight: 700;
    }

    .user-search-trigger:hover,
    .user-search-trigger:focus-visible {
      border-color: var(--accent);
      color: var(--text);
    }

    .user-search-trigger:focus-visible,
    .user-search-input:focus-visible,
    .user-search-result:focus-visible {
      outline: 2px solid var(--accent-strong);
      outline-offset: 3px;
    }

    .user-search-trigger kbd {
      border: 1px solid rgb(214 198 223 / 24%);
      border-radius: 0.35rem;
      padding: 0.12rem 0.32rem;
      background: rgb(255 255 255 / 8%);
      color: var(--text);
      font-family: inherit;
      font-size: 0.68rem;
    }

    .user-search-dialog {
      width: min(100% - 2rem, 38rem);
      overflow: hidden;
    }

    .viewer-management-dialog {
      position: fixed;
      inset: 0;
      width: min(100% - 2rem, 38rem);
      height: fit-content;
      margin: auto;
      border: 0;
      background: transparent;
      color: var(--text);
      overflow: hidden;
    }

    .viewer-management-dialog::backdrop {
      background: rgb(10 7 15 / 70%);
    }

    .user-search-header {
      border-bottom: 1px solid rgb(214 198 223 / 18%);
    }

    .user-search-input {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 0.7rem;
      padding: 0.7rem 0.85rem;
      background: rgb(10 7 15 / 44%);
      color: var(--text);
      font-size: 1.15rem;
      font-weight: 600;
    }

    .user-search-input::placeholder {
      color: var(--muted);
    }

    .user-search-help {
      display: flex;
      justify-content: space-between;
      gap: 1rem;
      margin-top: 1rem;
      color: var(--muted);
      font-size: 0.7rem;
    }

    .user-search-help span:first-child {
      color: var(--accent-strong);
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }

    .user-search-results {
      max-height: min(24rem, 55vh);
      overflow: hidden;
      padding: 0.5rem;
    }

    .user-search-results.is-scrollable {
      overflow-y: auto;
    }

    .user-search-result {
      display: flex;
      width: 100%;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
      border-radius: 0.75rem;
      padding: 0.8rem 0.9rem;
      color: var(--text);
      text-align: left;
    }

    .user-search-result:hover,
    .user-search-result.is-active {
      background: rgb(242 161 186 / 16%);
    }

    .user-search-user-id {
      overflow: hidden;
      color: var(--muted);
      font-family: monospace;
      font-size: 0.7rem;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .user-search-empty {
      padding: 1.25rem 0.9rem;
      color: var(--muted);
      font-size: 0.9rem;
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

    .grant-result {
      border: 1px solid rgb(242 161 186 / 70%);
      border-radius: 0.85rem;
      background: linear-gradient(135deg, rgb(242 161 186 / 20%), rgb(255 255 255 / 5%));
    }

    .grant-result.is-duplicate {
      border-color: rgb(255 191 126 / 76%);
      background: linear-gradient(135deg, rgb(255 191 126 / 24%), rgb(255 255 255 / 5%));
    }

    .grant-result-title {
      margin-top: 0.2rem;
      font-family: Nunito, sans-serif;
      font-size: 1.3rem;
      font-weight: 800;
    }

    @media (min-width: 900px) {
      .admin-shell {
        height: 100dvh;
        min-height: 0;
        overflow: hidden;
      }

      .admin-frame,
      .admin-layout {
        height: 100%;
      }

      .admin-layout {
        grid-template-areas: 'content';
        grid-template-columns: minmax(0, 1fr);
        grid-template-rows: minmax(0, 1fr);
        gap: 2rem;
      }

      .admin-content {
        display: grid;
        height: 100%;
        min-height: 0;
        align-content: start;
        gap: 2rem;
        grid-area: content;
        overflow-y: auto;
        overscroll-behavior: contain;
        padding-right: 0.75rem;
        scrollbar-gutter: stable;
      }

      .admin-content > .admin-header,
      .admin-content > .admin-workspace {
        grid-area: auto;
      }

      .admin-workspace {
        height: auto;
      }
    }
  }
</style>
