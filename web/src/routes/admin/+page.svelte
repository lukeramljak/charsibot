<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { resolve } from '$app/paths';
  import { page } from '$app/state';
  import type { components } from '$lib/api.generated';
  import UserCollections from '$lib/admin/UserCollections.svelte';
  import UserStats from '$lib/admin/UserStats.svelte';
  import ViewerDirectory from '$lib/admin/ViewerDirectory.svelte';

  type User = components['schemas']['User'];
  type UserStat = components['schemas']['AdminStat'];
  type Collection = components['schemas']['AdminCollection'];
  type UserDetail = components['schemas']['AdminUserResponse'];
  type UsersResponse = components['schemas']['AdminUsersResponse'];
  type ActivityFilter = 'all' | 'unknown' | 'inactive30' | 'inactive90' | 'recent';

  interface PendingPlushie {
    series: string;
    key: string;
    name: string;
  }

  let users = $state.raw<User[]>([]);
  let usernameFilter = $state('');
  let activityFilter = $state<ActivityFilter>('all');
  let selectedUserIDs = $state.raw<string[]>([]);
  let filteredUsers = $derived.by(() => {
    const query = usernameFilter.trim().toLowerCase();
    const now = Date.now();
    const matchesActivity = (user: User) => {
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
  let resetStatsDialog = $state<HTMLDialogElement | undefined>(undefined);
  let deleteUserDialog = $state<HTMLDialogElement | undefined>(undefined);
  let bulkDeleteDialog = $state<HTMLDialogElement | undefined>(undefined);
  let randomStatDialog = $state<HTMLDialogElement | undefined>(undefined);
  let error = $state('');
  let statusMessage = $state('');
  let selectedUserHeading = $state<HTMLHeadingElement | undefined>(undefined);
  let selectedUserRequest = 0;

  function adminUserURL(userID: string, path = '') {
    return `/api/admin/users/${encodeURIComponent(userID)}${path}`;
  }

  function isCurrentUserRequest(requestID: number, userID: string) {
    return selectedUserRequest === requestID && page.url.searchParams.get('user') === userID;
  }

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
      const response = await request<UsersResponse>('/api/admin/users');
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
      selectedUserRequest += 1;
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
    const requestID = ++selectedUserRequest;
    loading = true;
    error = '';
    statusMessage = `Loading ${user.username}…`;
    try {
      const response = await request<UserDetail>(adminUserURL(user.id));
      if (!isCurrentUserRequest(requestID, user.id)) return;
      selected = response;
      statusMessage = `Loaded ${user.username}.`;
      await tick();
      selectedUserHeading?.focus();
    } catch (err) {
      if (isCurrentUserRequest(requestID, user.id)) {
        error = err instanceof Error ? err.message : 'Could not load user';
        statusMessage = '';
      }
    } finally {
      if (isCurrentUserRequest(requestID, user.id)) loading = false;
    }
  }

  async function updateStat(
    stat: UserStat,
    value: number,
    mode: 'set' | 'adjust',
  ): Promise<boolean> {
    if (!selected || !Number.isFinite(value)) return false;
    return mutate(adminUserURL(selected.user.id, `/stats/${encodeURIComponent(stat.name)}`), {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ mode, value }),
    });
  }

  async function displayStatsInChat() {
    if (!selected) return;
    await mutate(adminUserURL(selected.user.id, '/stats/display'), {
      method: 'POST',
    });
  }

  async function displayCollection(collection: Collection) {
    if (!selected) return;
    await mutate(
      adminUserURL(
        selected.user.id,
        `/collections/${encodeURIComponent(collection.config.series)}/display`,
      ),
      { method: 'POST' },
    );
  }

  function closeDeleteUserDialog() {
    deleteUserDialog?.close();
  }

  async function deleteUser() {
    const user = selected?.user;
    if (!user) return;
    closeDeleteUserDialog();
    loading = true;
    error = '';
    try {
      const response = await fetch(adminUserURL(user.id), {
        method: 'DELETE',
      });
      if (!response.ok) throw new Error((await response.text()) || 'Could not delete viewer');
      users = users.filter((candidate) => candidate.id !== user.id);
      selected = null;
      statusMessage = `Deleted ${user.username}.`;
      await goto(resolve('/admin'), { keepFocus: true, noScroll: true });
    } catch (err) {
      error = err instanceof Error ? err.message : 'Could not delete viewer';
      statusMessage = '';
    } finally {
      loading = false;
    }
  }

  function toggleUserSelection(userID: string, checked: boolean) {
    selectedUserIDs = checked
      ? [...new Set([...selectedUserIDs, userID])]
      : selectedUserIDs.filter((candidate) => candidate !== userID);
  }

  function selectFilteredUsers() {
    selectedUserIDs = filteredUsers.map((user) => user.id);
  }

  function clearUserSelection() {
    selectedUserIDs = [];
  }

  function closeBulkDeleteDialog() {
    bulkDeleteDialog?.close();
  }

  async function deleteSelectedUsers() {
    const userIDs = [...selectedUserIDs];
    if (userIDs.length === 0) return;

    closeBulkDeleteDialog();
    loading = true;
    error = '';
    const deletedUserIDs: string[] = [];
    try {
      for (const userID of userIDs) {
        const response = await fetch(adminUserURL(userID), {
          method: 'DELETE',
        });
        if (!response.ok) throw new Error((await response.text()) || 'Could not delete viewers');
        deletedUserIDs.push(userID);
      }

      users = users.filter((user) => !deletedUserIDs.includes(user.id));
      selectedUserIDs = [];
      if (selected && deletedUserIDs.includes(selected.user.id)) {
        selected = null;
        await goto(resolve('/admin'), { keepFocus: true, noScroll: true });
      }
      statusMessage = `Deleted ${deletedUserIDs.length} viewers.`;
    } catch (err) {
      users = users.filter((user) => !deletedUserIDs.includes(user.id));
      selectedUserIDs = selectedUserIDs.filter((userID) => !deletedUserIDs.includes(userID));
      if (selected && deletedUserIDs.includes(selected.user.id)) {
        selected = null;
        await goto(resolve('/admin'), { keepFocus: true, noScroll: true });
      }
      error = err instanceof Error ? err.message : 'Could not delete viewers';
      statusMessage =
        deletedUserIDs.length > 0
          ? `Deleted ${deletedUserIDs.length} viewers before stopping.`
          : '';
    } finally {
      loading = false;
    }
  }

  function formatLastActive(value: string | undefined) {
    if (!value) return 'Unknown';
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(
      new Date(value),
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
    await mutate(adminUserURL(selected.user.id, '/stats/random'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ displayInChat }),
    });
  }

  function closeExplodeDialog() {
    explodeDialog?.close();
  }

  async function explode() {
    if (!selected) return;
    closeExplodeDialog();
    await mutate(adminUserURL(selected.user.id, '/stats/explode'), {
      method: 'POST',
    });
  }

  function closeUndoExplodeDialog() {
    undoExplodeDialog?.close();
  }

  async function undoExplode() {
    if (!selected) return;
    closeUndoExplodeDialog();
    await mutate(adminUserURL(selected.user.id, '/stats/explode/undo'), {
      method: 'POST',
    });
  }

  function closeResetStatsDialog() {
    resetStatsDialog?.close();
  }

  async function resetStats(displayInChat: boolean) {
    if (!selected) return;
    closeResetStatsDialog();
    await mutate(adminUserURL(selected.user.id, '/stats/reset'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ displayInChat }),
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
    const base = adminUserURL(
      selected.user.id,
      `/collections/${encodeURIComponent(series)}/${encodeURIComponent(key)}`,
    );
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
        adminUserURL(
          selected.user.id,
          `/collections/${encodeURIComponent(plushie.series)}/${encodeURIComponent(plushie.key)}`,
        ),
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
      adminUserURL(
        selected.user.id,
        `/collections/${encodeURIComponent(collection.config.series)}/random`,
      ),
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
      adminUserURL(
        selected.user.id,
        `/collections/${encodeURIComponent(collection.config.series)}`,
      ),
      {
        method: 'DELETE',
      },
    );
  }

  async function mutate(url: string, options: RequestInit): Promise<boolean> {
    const userID = selected?.user.id;
    if (!userID) return false;
    const requestID = ++selectedUserRequest;
    loading = true;
    error = '';
    statusMessage = 'Saving changes…';
    try {
      const response = await request<UserDetail>(url, options);
      if (!isCurrentUserRequest(requestID, userID)) return false;
      selected = response;
      statusMessage = 'Changes saved.';
      return true;
    } catch (err) {
      if (isCurrentUserRequest(requestID, userID)) {
        error = err instanceof Error ? err.message : 'Could not save change';
        statusMessage = '';
      }
      return false;
    } finally {
      if (isCurrentUserRequest(requestID, userID)) loading = false;
    }
  }
</script>

<svelte:head>
  <title>Charsibot Admin</title>
</svelte:head>

<main class="admin-shell p-6 sm:p-10" aria-busy={loading}>
  <div class="admin-frame mx-auto max-w-384">
    <p class="sr-only" role="status">{statusMessage}</p>

    <div class="admin-layout">
      <header class="admin-header">
        <a class="back-link" href={resolve('/')}>← Overlay</a>
        <p class="eyebrow mt-6">Control room</p>
        <h1 class="admin-title mt-2">Charsibot Admin</h1>
        <p class="admin-subtitle mt-2">
          Local-only controls for viewer stats and blind-box collections.
        </p>
      </header>

      <aside class="viewer-directory">
        <ViewerDirectory
          {users}
          {filteredUsers}
          {usernameFilter}
          {activityFilter}
          {selectedUserIDs}
          selectedUserID={selected?.user.id}
          {loading}
          onRefresh={loadUsers}
          onUsernameFilterChange={(value) => (usernameFilter = value)}
          onActivityFilterChange={(value) => (activityFilter = value)}
          onSelectFiltered={selectFilteredUsers}
          onClearSelection={clearUserSelection}
          onOpenBulkDelete={() => bulkDeleteDialog?.showModal()}
          onToggleUserSelection={toggleUserSelection}
          onSelectUser={selectUser}
        />
      </aside>

      <div class="admin-workspace">
        {#if error}
          <p class="admin-error px-4 py-3" role="alert">
            {error}
          </p>
        {/if}

        {#if selected}
          <section class="user-detail">
            <div class="user-detail-header mb-4">
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
                      class="admin-muted shrink-0 rounded-full border border-[var(--line)] px-2 py-0.5 font-mono text-[0.65rem]"
                    >
                      {selected.user.id}
                    </span>
                  </div>
                  <p class="admin-muted mt-1 text-xs">
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
    bind:this={deleteUserDialog}
    aria-labelledby="delete-user-dialog-title"
  >
    <p class="eyebrow">Prune viewer</p>
    <h2 class="section-title mt-2 text-2xl" id="delete-user-dialog-title">
      Delete {selected?.user.username ?? 'this viewer'}?
    </h2>
    <p class="admin-muted mt-2">
      This permanently removes their activity, stats, and blind-box collections.
    </p>
    <div class="dialog-actions mt-6">
      <button class="button button-secondary" onclick={closeDeleteUserDialog}>Cancel</button>
      <button class="button button-danger" onclick={deleteUser}>Delete viewer</button>
    </div>
  </dialog>

  <dialog
    class="admin-dialog p-6"
    bind:this={bulkDeleteDialog}
    aria-labelledby="bulk-delete-dialog-title"
  >
    <p class="eyebrow">Prune viewers</p>
    <h2 class="section-title mt-2 text-2xl" id="bulk-delete-dialog-title">
      Delete {selectedUserIDs.length} selected viewers?
    </h2>
    <p class="admin-muted mt-2">
      This permanently removes their activity, stats, and blind-box collections.
    </p>
    <div class="dialog-actions mt-6">
      <button class="button button-secondary" onclick={closeBulkDeleteDialog}>Cancel</button>
      <button class="button button-danger" onclick={deleteSelectedUsers}>
        Delete {selectedUserIDs.length} viewers
      </button>
    </div>
  </dialog>

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
    bind:this={resetStatsDialog}
    aria-labelledby="reset-stats-dialog-title"
  >
    <p class="eyebrow">Viewer stats</p>
    <h2 class="section-title mt-2 text-2xl" id="reset-stats-dialog-title">
      Reset {selected?.user.username ?? 'this viewer'}'s stats?
    </h2>
    <p class="admin-muted mt-2">This restores every stat to its configured default.</p>
    <div class="dialog-actions mt-6">
      <button class="button button-secondary" onclick={closeResetStatsDialog}>Cancel</button>
      <button class="button button-danger" onclick={() => resetStats(false)}>Reset silently</button>
      <button class="button button-primary" onclick={() => resetStats(true)}>
        Reset &amp; display stats
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
      <button class="button button-secondary" onclick={() => grantPlushie(false)}
        >Grant silently</button
      >
      <button class="button button-primary" onclick={() => grantPlushie(true)}>
        Grant &amp; show overlay
      </button>
    </div>
  </dialog>

  <dialog class="admin-dialog p-6" bind:this={explodeDialog} aria-labelledby="explode-dialog-title">
    <p class="eyebrow">Viewer stat</p>
    <h2 class="section-title mt-2 text-2xl" id="explode-dialog-title">
      Explode {selected?.user.username ?? 'this viewer'}?
    </h2>
    <p class="admin-muted mt-2">
      This sets their PENIS stat to -1000 and displays their updated stats in chat.
    </p>
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
    <p class="admin-muted mt-2">
      This restores their PENIS stat to its configured default and displays their updated stats in
      chat.
    </p>
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
      margin-left: 1rem;
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
      padding: 0.85rem 1rem;
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
        grid-template-areas:
          'sidebar header'
          'sidebar workspace';
        grid-template-columns: minmax(17rem, 21rem) minmax(0, 1fr);
        grid-template-rows: auto minmax(0, 1fr);
        gap: 2rem;
      }

      .viewer-directory {
        height: 100%;
        min-height: 0;
      }

      .viewer-directory > .admin-surface {
        display: flex;
        height: 100%;
        min-height: 0;
        flex-direction: column;
      }

      .viewer-list {
        max-height: none;
        min-height: 0;
        flex: 1 1 auto;
      }

      .admin-workspace {
        height: 100%;
        min-height: 0;
        overflow-y: auto;
        overscroll-behavior: contain;
        padding-right: 0.75rem;
        scrollbar-gutter: stable;
      }

      .viewer-bulk-actions {
        flex-direction: column;
      }

      .viewer-bulk-actions .button {
        width: 100%;
        text-align: center;
      }
    }
  }
</style>
