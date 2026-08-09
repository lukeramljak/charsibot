import BetterSqlite3 from 'better-sqlite3';
import { mkdtempSync, rmSync, statSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';

import { openDatabase } from '$lib/server/db/client';
import { toSafeInteger } from '$lib/server/db/integer';

const temporaryDirectories: string[] = [];

const databasePath = (): string => {
  const directory = mkdtempSync(join(tmpdir(), 'charsibot-db-'));
  temporaryDirectories.push(directory);
  return join(directory, 'charsibot.db');
};

afterEach(() => {
  for (const directory of temporaryDirectories.splice(0)) {
    rmSync(directory, { recursive: true, force: true });
  }
});

describe('openDatabase', () => {
  it('creates the consolidated Goose v7 schema and exact connection pragmas', () => {
    const connection = openDatabase(databasePath());
    expect(connection.state).toBe('created');
    const { database } = connection;
    expect(database.pragma('foreign_keys', { simple: true })).toBe(1n);
    expect(database.pragma('journal_mode', { simple: true })).toBe('wal');
    expect(database.pragma('page_size', { simple: true })).toBe(4096n);
    expect(database.pragma('cache_size', { simple: true })).toBe(-8000n);
    expect(database.pragma('synchronous', { simple: true })).toBe(1n);
    expect(database.pragma('secure_delete', { simple: true })).toBe(1n);
    expect(database.pragma('busy_timeout', { simple: true })).toBe(30000n);
    const versions = database
      .prepare('SELECT version_id FROM goose_db_version ORDER BY id')
      .pluck()
      .all();
    expect(versions).toEqual([0n, 1n, 2n, 3n, 4n, 5n, 6n, 7n]);
    const tables = database
      .prepare(
        "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name",
      )
      .pluck()
      .all();
    expect(tables).toEqual(['goose_db_version', 'user_plushies', 'user_stats', 'viewer_activity']);
    connection.close();
  });

  it('validates and reopens v7 without changing schema or migration history', () => {
    const path = databasePath();
    const created = openDatabase(path);
    created.database
      .prepare('INSERT INTO user_stats (user_id, username, stat_name, value) VALUES (?, ?, ?, ?)')
      .run('viewer', 'name', 'strength', 9);
    const schemaBefore = created.database
      .prepare('SELECT name, sql FROM sqlite_master WHERE sql IS NOT NULL ORDER BY name')
      .all();
    const historyBefore = created.database
      .prepare('SELECT version_id, is_applied FROM goose_db_version ORDER BY id')
      .all();
    created.close();

    const reopened = openDatabase(path);
    expect(reopened.state).toBe('existing');
    expect(
      reopened.database
        .prepare('SELECT name, sql FROM sqlite_master WHERE sql IS NOT NULL ORDER BY name')
        .all(),
    ).toEqual(schemaBefore);
    expect(
      reopened.database
        .prepare('SELECT version_id, is_applied FROM goose_db_version ORDER BY id')
        .all(),
    ).toEqual(historyBefore);
    expect(
      reopened.database
        .prepare('SELECT value FROM user_stats WHERE user_id = ?')
        .pluck()
        .get('viewer'),
    ).toBe(9n);
    reopened.close();
  });

  it.each([
    ['below v7', 'DELETE FROM goose_db_version WHERE version_id = 7'],
    ['above v7', 'INSERT INTO goose_db_version (version_id, is_applied) VALUES (8, 1)'],
    ['malformed history', 'UPDATE goose_db_version SET is_applied = 0 WHERE version_id = 4'],
    ['missing index', 'DROP INDEX viewer_activity_last_active_at_idx'],
    ['wrong schema', 'ALTER TABLE user_stats ADD COLUMN unexpected TEXT'],
    ['unexpected application objects', 'CREATE TABLE unexpected (id INTEGER PRIMARY KEY)'],
  ])('rejects %s databases', (_name, mutation) => {
    const path = databasePath();
    const connection = openDatabase(path);
    connection.close();
    const database = new BetterSqlite3(path);
    database.exec(mutation);
    database.close();
    expect(() => openDatabase(path)).toThrow();
  });

  it('rejects a partial non-empty database without adding migration tables', () => {
    const path = databasePath();
    const database = new BetterSqlite3(path);
    database.exec('CREATE TABLE unrelated (id INTEGER PRIMARY KEY)');
    database.close();
    expect(() => openDatabase(path)).toThrow();
    const inspected = new BetterSqlite3(path);
    expect(
      inspected
        .prepare(
          "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'goose_db_version'",
        )
        .pluck()
        .get(),
    ).toBe(0);
    inspected.close();
  });

  it('rejects SQLite integers outside the lossless JavaScript range', () => {
    expect(() => toSafeInteger(9_007_199_254_740_992n, 'value')).toThrow(/safe integer range/);
    expect(toSafeInteger(9_007_199_254_740_991n, 'value')).toBe(Number.MAX_SAFE_INTEGER);
  });

  it('survives checkpoint-close-reopen cycle with active WAL data', () => {
    const path = databasePath();
    const connection = openDatabase(path);

    connection.database
      .prepare('INSERT INTO viewer_activity (user_id, username, last_active_at) VALUES (?, ?, ?)')
      .run('u1', 'alice', '2026-08-01T00:00:00Z');
    connection.database
      .prepare('INSERT INTO user_stats (user_id, username, stat_name, value) VALUES (?, ?, ?, ?)')
      .run('u1', 'alice', 'strength', 5);
    connection.database
      .prepare('INSERT INTO user_plushies (user_id, username, series, key) VALUES (?, ?, ?, ?)')
      .run('u1', 'alice', 'coobubu', 'cutey');

    connection.checkpoint();
    connection.close();

    const reopened = openDatabase(path);
    expect(reopened.state).toBe('existing');

    expect(reopened.database.prepare('SELECT COUNT(*) FROM viewer_activity').pluck().get()).toBe(
      1n,
    );
    expect(reopened.database.prepare('SELECT COUNT(*) FROM user_stats').pluck().get()).toBe(1n);
    expect(reopened.database.prepare('SELECT COUNT(*) FROM user_plushies').pluck().get()).toBe(1n);
    expect(
      reopened.database
        .prepare('SELECT username FROM viewer_activity WHERE user_id = ?')
        .pluck()
        .get('u1'),
    ).toBe('alice');

    reopened.close();
  });

  it('checkpoint truncates the WAL file', () => {
    const path = databasePath();
    const connection = openDatabase(path);

    connection.database
      .prepare('INSERT INTO viewer_activity (user_id, username, last_active_at) VALUES (?, ?, ?)')
      .run('u1', 'alice', '2026-08-01T00:00:00Z');

    const walSizeBefore = (() => {
      try {
        return statSync(`${path}-wal`).size;
      } catch {
        return 0;
      }
    })();

    connection.checkpoint();

    const walSizeAfter = (() => {
      try {
        return statSync(`${path}-wal`).size;
      } catch {
        return 0;
      }
    })();

    expect(walSizeAfter).toBeLessThanOrEqual(walSizeBefore);
    expect(walSizeAfter).toBe(0);

    connection.close();
  });
});
