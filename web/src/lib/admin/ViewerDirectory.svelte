<script lang="ts">
  import type { components } from '$lib/api.generated';

  type User = components['schemas']['User'];
  type ActivityFilter = 'all' | 'unknown' | 'inactive30' | 'inactive90' | 'recent';

  let {
    users,
    filteredUsers,
    usernameFilter,
    activityFilter,
    selectedUserIDs,
    selectedUserID,
    loading,
    onRefresh,
    onClose,
    onUsernameFilterChange,
    onActivityFilterChange,
    onSelectFiltered,
    onClearSelection,
    onOpenBulkDelete,
    onToggleUserSelection,
    onSelectUser,
  }: {
    users: User[];
    filteredUsers: User[];
    usernameFilter: string;
    activityFilter: ActivityFilter;
    selectedUserIDs: string[];
    selectedUserID: string | undefined;
    loading: boolean;
    onRefresh: () => void;
    onClose?: () => void;
    onUsernameFilterChange: (value: string) => void;
    onActivityFilterChange: (value: ActivityFilter) => void;
    onSelectFiltered: () => void;
    onClearSelection: () => void;
    onOpenBulkDelete: () => void;
    onToggleUserSelection: (userID: string, checked: boolean) => void;
    onSelectUser: (user: User) => void;
  } = $props();
</script>

<section class="admin-surface flex flex-col gap-4 p-5">
  <div class="flex items-center justify-between gap-4">
    <div class="flex flex-col gap-1">
      <p class="eyebrow">Viewer directory</p>
      <h2 class="section-title">Viewers</h2>
      <p class="admin-muted text-sm">{filteredUsers.length} of {users.length} known users</p>
    </div>
    <div class="flex items-center gap-2">
      {#if onClose}
        <button class="button button-secondary" onclick={onClose}>Close</button>
      {/if}
      <button class="button button-secondary" onclick={onRefresh} disabled={loading}>Refresh</button>
    </div>
  </div>

  {#if users.length > 0}
    <label class="sr-only" for="username-filter">Filter by username</label>
    <div class="flex flex-col gap-2">
      <input
        id="username-filter"
        class="admin-input w-full"
        value={usernameFilter}
        oninput={(event) => onUsernameFilterChange(event.currentTarget.value)}
        placeholder="Filter by username…"
      />
      <label class="sr-only" for="activity-filter">Filter by activity</label>
      <select
        id="activity-filter"
        class="admin-input w-full"
        value={activityFilter}
        onchange={(event) => onActivityFilterChange(event.currentTarget.value as ActivityFilter)}
      >
        <option value="all">All activity</option>
        <option value="unknown">Unknown activity</option>
        <option value="inactive30">Inactive for 30+ days</option>
        <option value="inactive90">Inactive for 90+ days</option>
        <option value="recent">Active in the last 30 days</option>
      </select>
    </div>
    <div class="viewer-bulk-actions">
      <button
        class="button button-secondary"
        onclick={onSelectFiltered}
        disabled={loading || filteredUsers.length === 0}
      >
        Select filtered
      </button>
      {#if selectedUserIDs.length > 0}
        <button class="button button-secondary" onclick={onClearSelection} disabled={loading}
          >Clear selection</button
        >
        <button class="button button-danger" onclick={onOpenBulkDelete} disabled={loading}>
          Prune {selectedUserIDs.length} selected
        </button>
      {/if}
    </div>
    <ul class="viewer-list" aria-label="Viewers">
      {#each filteredUsers as user (user.id)}
        <li class={['viewer-list-item', selectedUserID === user.id && 'is-selected']}>
          <input
            class="viewer-selection"
            type="checkbox"
            checked={selectedUserIDs.includes(user.id)}
            onchange={(event) => onToggleUserSelection(user.id, event.currentTarget.checked)}
            aria-label={`Select ${user.username} for pruning`}
            disabled={loading}
          />
          <button
            class="viewer-row"
            onclick={() => onSelectUser(user)}
            disabled={loading}
            aria-current={selectedUserID === user.id ? 'true' : undefined}
            aria-label={`Manage ${user.username}`}
          >
            <span class="truncate font-medium">{user.username}</span>
            <span class="viewer-row-action" aria-hidden="true">Manage</span>
          </button>
        </li>
      {/each}
    </ul>
    {#if filteredUsers.length === 0}
      <p class="admin-muted text-sm">No usernames match that filter.</p>
    {/if}
  {:else if !loading}
    <p class="admin-muted text-sm">No users have been recorded yet.</p>
  {/if}
</section>
