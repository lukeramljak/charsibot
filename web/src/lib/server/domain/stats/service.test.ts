import { describe, expect, it } from 'vitest';

import { loadCatalog } from '$lib/server/catalog/load';
import { openDatabase } from '$lib/server/db/client';
import { createStatsRepository } from '$lib/server/db/stats.repository';
import { createViewersRepository } from '$lib/server/db/viewers.repository';

import { formatStats } from '$lib/server/domain/stats/format';
import { createStatsService } from '$lib/server/domain/stats/service';

const setup = () => {
  const connection = openDatabase(':memory:');
  const repository = createStatsRepository(connection.database);
  const viewers = createViewersRepository(connection.database);
  const catalog = loadCatalog();
  const service = createStatsService({ repository, viewers, definitions: catalog.stats });
  return { connection, repository, viewers, catalog, service };
};

describe('stats service', () => {
  it('creates defaults idempotently, preserves values, and synchronizes usernames', async () => {
    const { connection, service } = setup();
    const created = await service.getOrCreate('viewer-1', 'oldname');
    expect(created).toHaveLength(6);
    expect(created[0]).toEqual({
      name: 'strength',
      shortName: 'STR',
      longName: 'Strength',
      value: 3,
    });
    await service.set('viewer-1', 'strength', 9);
    const existing = await service.getOrCreate('viewer-1', 'newname');
    expect(existing[0].value).toBe(9);
    expect(
      connection.database
        .prepare('SELECT DISTINCT username FROM user_stats WHERE user_id = ?')
        .pluck()
        .all('viewer-1'),
    ).toEqual(['newname']);
    connection.close();
  });

  it('adjusts signed values and resets every configured stat transactionally', async () => {
    const { connection, service } = setup();
    await service.getOrCreate('viewer-1', 'viewer');
    await service.adjust('viewer-1', 'luck', -5);
    expect((await service.get('viewer-1')).find((stat) => stat.name === 'luck')?.value).toBe(-2);
    await service.reset('viewer-1');
    expect((await service.get('viewer-1')).map((stat) => stat.value)).toEqual([3, 3, 3, 3, 3, 3]);
    connection.close();
  });

  it('rolls back multi-row initialization and reset failures', async () => {
    const { connection, catalog, repository, service } = setup();
    expect(() =>
      repository.initialize('broken', 'broken', [
        catalog.stats[0],
        { ...catalog.stats[1], defaultValue: Number.MAX_SAFE_INTEGER + 1 },
      ]),
    ).toThrow(/safe integer/);
    expect(
      connection.database
        .prepare("SELECT COUNT(*) FROM user_stats WHERE user_id = 'broken'")
        .pluck()
        .get(),
    ).toBe(0n);

    await service.getOrCreate('viewer', 'viewer');
    await service.set('viewer', 'strength', 8);
    expect(() =>
      repository.reset('viewer', [
        catalog.stats[0],
        { ...catalog.stats[1], defaultValue: Number.MAX_SAFE_INTEGER + 1 },
      ]),
    ).toThrow(/safe integer/);
    expect((await service.get('viewer')).find((stat) => stat.name === 'strength')?.value).toBe(8);
    connection.close();
  });

  it('lists the union of viewers case-insensitively and preserves activity timestamps', async () => {
    const { connection, service } = setup();
    await service.getOrCreate('stats-user', 'Zulu');
    connection.database
      .prepare('INSERT INTO user_plushies (user_id, username, series, key) VALUES (?, ?, ?, ?)')
      .run('collection-user', 'alpha', 'coobubu', 'cutey');
    await service.recordActivity('activity-user', 'Bravo', new Date('2026-08-02T01:02:03.456Z'));
    expect(await service.listViewers()).toEqual([
      { id: 'collection-user', username: 'alpha' },
      { id: 'activity-user', username: 'Bravo', lastActiveAt: '2026-08-02T01:02:03.456Z' },
      { id: 'stats-user', username: 'Zulu' },
    ]);
    connection.close();
  });

  it('uses the current lexicographic username merge and deterministic leaderboard ties', async () => {
    const { connection, service } = setup();
    await service.getOrCreate('viewer-b', 'beta');
    await service.getOrCreate('viewer-a', 'Alpha');
    await service.recordActivity('viewer-a', 'aardvark', new Date('2026-08-02T00:00:00Z'));
    expect((await service.getViewer('viewer-a')).username).toBe('aardvark');
    const leaderboard = await service.leaderboard();
    expect(leaderboard[0]).toEqual({ emoji: '💪', username: 'Alpha', value: 3 });
    connection.close();
  });

  it('deletes multiple viewers from activity, stats, and collections', async () => {
    const { connection, service } = setup();
    for (const [id, name] of [
      ['one', 'one'],
      ['two', 'two'],
    ]) {
      await service.getOrCreate(id, name);
      await service.recordActivity(id, name, new Date('2026-08-02T00:00:00Z'));
      connection.database
        .prepare('INSERT INTO user_plushies (user_id, username, series, key) VALUES (?, ?, ?, ?)')
        .run(id, name, 'coobubu', 'cutey');
    }
    await service.deleteViewers(['one', 'two']);
    for (const table of ['viewer_activity', 'user_stats', 'user_plushies']) {
      expect(connection.database.prepare(`SELECT COUNT(*) FROM ${table}`).pluck().get()).toBe(0n);
    }
    connection.close();
  });

  it('rolls back the entire viewer batch when one deletion fails', async () => {
    const { connection, service } = setup();
    await service.getOrCreate('one', 'one');
    await service.getOrCreate('two', 'two');
    connection.database.exec(`
      CREATE TRIGGER reject_second_viewer
      BEFORE DELETE ON user_stats
      WHEN OLD.user_id = 'two'
      BEGIN
        SELECT RAISE(ABORT, 'test deletion failure');
      END`);
    await expect(service.deleteViewers(['one', 'two'])).rejects.toThrow(/test deletion failure/);
    expect(
      connection.database.prepare('SELECT COUNT(DISTINCT user_id) FROM user_stats').pluck().get(),
    ).toBe(2n);
    connection.database.exec('DROP TRIGGER reject_second_viewer');
    connection.close();
  });

  it('rejects invalid persisted activity timestamps', async () => {
    const { connection, service } = setup();
    connection.database
      .prepare('INSERT INTO viewer_activity (user_id, username, last_active_at) VALUES (?, ?, ?)')
      .run('viewer', 'viewer', '2026-02-30T00:00:00Z');
    await expect(service.listViewers()).rejects.toThrow(/invalid viewer activity timestamp/);
    connection.close();
  });

  it('fails rather than rounding an out-of-range SQLite stat value', async () => {
    const { connection, service } = setup();
    connection.database.exec(`
      INSERT INTO user_stats (user_id, username, stat_name, value)
      VALUES ('viewer', 'viewer', 'strength', 9007199254740992)`);
    await expect(service.get('viewer')).rejects.toThrow(/safe integer range/);
    connection.close();
  });
});

describe('formatStats', () => {
  it('matches the Go chat formatting for populated and empty stats', () => {
    expect(
      formatStats('testuser', [
        { name: 'strength', shortName: 'STR', longName: 'Strength', value: 5 },
        { name: 'luck', shortName: 'LUCK', longName: 'Luck', value: -2 },
      ]),
    ).toBe("testuser's stats: STR: 5 | LUCK: -2");
    expect(formatStats('testuser', [])).toBe('No stats found for testuser');
  });
});
