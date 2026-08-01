<script lang="ts">
  import type { components } from '$lib/api.generated';

  type UserStat = components['schemas']['AdminStat'];

  let {
    stats,
    loading,
    onUpdateStat,
    onDisplayStats,
    onOpenRandomStat,
    onOpenExplode,
    onOpenUndoExplode,
    onOpenResetStats,
  }: {
    stats: UserStat[];
    loading: boolean;
    onUpdateStat: (stat: UserStat, value: number, mode: 'set' | 'adjust') => Promise<boolean>;
    onDisplayStats: () => void;
    onOpenRandomStat: () => void;
    onOpenExplode: () => void;
    onOpenUndoExplode: () => void;
    onOpenResetStats: () => void;
  } = $props();
</script>

<section aria-labelledby="stats-heading">
  <div class="flex flex-wrap items-center justify-between gap-3">
    <h3 class="detail-section-title" id="stats-heading">Stats</h3>
    <div class="user-actions flex flex-wrap items-center justify-end gap-2">
      <button class="button button-secondary" onclick={onDisplayStats} disabled={loading}
        >Display stats in chat</button
      >
      <button class="button button-secondary" onclick={onOpenRandomStat} disabled={loading}
        >Grant random stat</button
      >
      <button class="button button-danger" onclick={onOpenExplode} disabled={loading}
        >Explode</button
      >
      <button class="button button-secondary" onclick={onOpenUndoExplode} disabled={loading}
        >Undo explode</button
      >
      <button class="button button-danger" onclick={onOpenResetStats} disabled={loading}
        >Reset stats</button
      >
    </div>
  </div>
  <div class="mt-4 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
    {#each stats as stat (stat.name)}
      <div class="stat-card p-4 text-center">
        <p class="font-semibold">
          {stat.longName} <span class="admin-muted">({stat.shortName})</span>
        </p>
        <div class="mt-3 flex items-center justify-center gap-2">
          <button
            class="stat-stepper"
            onclick={() => onUpdateStat(stat, -1, 'adjust')}
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
              const saved = await onUpdateStat(stat, Number(input.value), 'set');
              if (!saved) input.value = String(stat.value);
            }}
            disabled={loading}
          />
          <button
            class="stat-stepper"
            onclick={() => onUpdateStat(stat, 1, 'adjust')}
            disabled={loading}
            aria-label={`Increase ${stat.longName} by 1`}>+</button
          >
        </div>
      </div>
    {/each}
  </div>
</section>
