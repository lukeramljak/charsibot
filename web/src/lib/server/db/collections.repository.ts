import type BetterSqlite3 from 'better-sqlite3';

import { toSafeInteger } from '$lib/server/db/integer';

export interface CollectionCountRecord {
  userID: string;
  username: string;
  series: string;
  count: number;
}

export interface CollectionsRepository {
  get: (userID: string, series: string) => string[];
  grant: (
    userID: string,
    username: string,
    series: string,
    key: string,
  ) => { isNew: boolean; collection: string[] };
  remove: (userID: string, series: string, key: string) => void;
  reset: (userID: string, series: string) => void;
  counts: () => CollectionCountRecord[];
}

interface KeyRow {
  key: string;
}

interface CountRow {
  user_id: string;
  username: string;
  series: string;
  count: bigint;
}

export const createCollectionsRepository = (
  database: BetterSqlite3.Database,
): CollectionsRepository => {
  const insert = database.prepare(`
    INSERT OR IGNORE INTO user_plushies (user_id, username, series, key)
    VALUES (?, ?, ?, ?)`);
  const updateUsername = database.prepare(`
    UPDATE user_plushies SET username = ?
    WHERE user_id = ? AND series = ? AND key = ?`);
  const get = database.prepare('SELECT key FROM user_plushies WHERE user_id = ? AND series = ?');
  const remove = database.prepare(`
    DELETE FROM user_plushies WHERE user_id = ? AND series = ? AND key = ?`);
  const reset = database.prepare('DELETE FROM user_plushies WHERE user_id = ? AND series = ?');
  const counts = database.prepare(`
    SELECT user_id, CAST(MAX(username) AS TEXT) AS username, series, COUNT(*) AS count
    FROM user_plushies
    GROUP BY series, user_id
    ORDER BY series, username COLLATE NOCASE, username, user_id`);

  const getKeys = (userID: string, series: string): string[] =>
    (get.all(userID, series) as KeyRow[]).map((row) => row.key);

  const grantTransaction = database.transaction(
    (userID: string, username: string, series: string, key: string) => {
      const result = insert.run(userID, username, series, key);
      const changes = toSafeInteger(result.changes, 'plushie insert change count');
      const isNew = changes > 0;

      if (!isNew) {
        updateUsername.run(username, userID, series, key);
      }

      return { isNew, collection: getKeys(userID, series) };
    },
  );

  return {
    get: getKeys,
    grant: (userID, username, series, key) => grantTransaction(userID, username, series, key),
    remove: (userID, series, key) => {
      remove.run(userID, series, key);
    },
    reset: (userID, series) => {
      reset.run(userID, series);
    },
    counts: () =>
      (counts.all() as CountRow[]).map((row) => ({
        userID: row.user_id,
        username: row.username,
        series: row.series,
        count: toSafeInteger(row.count, `${row.series} collection count`),
      })),
  };
};
