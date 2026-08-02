import type { StatDefinition } from '$lib/contracts/catalog';
import type BetterSqlite3 from 'better-sqlite3';

import { requireSafeInteger, toSafeInteger } from '$lib/server/db/integer';

export interface StatValueRecord {
  statName: string;
  value: number;
}

export interface LeaderboardValueRecord extends StatValueRecord {
  userID: string;
  username: string;
}

export interface StatsRepository {
  initialize: (userID: string, username: string, definitions: readonly StatDefinition[]) => void;
  get: (userID: string) => StatValueRecord[];
  all: () => LeaderboardValueRecord[];
  adjust: (userID: string, statName: string, amount: number) => void;
  set: (userID: string, statName: string, value: number) => void;
  reset: (userID: string, definitions: readonly StatDefinition[]) => void;
}

interface StatRow {
  stat_name: string;
  value: bigint;
}

interface LeaderboardRow extends StatRow {
  user_id: string;
  username: string;
}

export const createStatsRepository = (database: BetterSqlite3.Database): StatsRepository => {
  const ensure = database.prepare(`
    INSERT OR IGNORE INTO user_stats (user_id, username, stat_name, value)
    VALUES (?, ?, ?, ?)`);
  const updateUsername = database.prepare('UPDATE user_stats SET username = ? WHERE user_id = ?');
  const get = database.prepare('SELECT stat_name, value FROM user_stats WHERE user_id = ?');
  const all = database.prepare(`
    SELECT user_id, username, stat_name, value
    FROM user_stats
    ORDER BY stat_name, value DESC, username COLLATE NOCASE, username, user_id`);
  const adjust = database.prepare(`
    UPDATE user_stats SET value = value + ?
    WHERE user_id = ? AND stat_name = ?`);
  const set = database.prepare(`
    UPDATE user_stats SET value = ?
    WHERE user_id = ? AND stat_name = ?`);

  const initializeTransaction = database.transaction(
    (userID: string, username: string, definitions: readonly StatDefinition[]) => {
      for (const definition of definitions) {
        ensure.run(
          userID,
          username,
          definition.name,
          requireSafeInteger(definition.defaultValue, `${definition.name} default value`),
        );
      }

      updateUsername.run(username, userID);
    },
  );

  const resetTransaction = database.transaction(
    (userID: string, definitions: readonly StatDefinition[]) => {
      for (const definition of definitions) {
        set.run(
          requireSafeInteger(definition.defaultValue, `${definition.name} default value`),
          userID,
          definition.name,
        );
      }
    },
  );

  return {
    initialize: (userID, username, definitions) =>
      initializeTransaction(userID, username, definitions),
    get: (userID) =>
      (get.all(userID) as StatRow[]).map((row) => ({
        statName: row.stat_name,
        value: toSafeInteger(row.value, `${row.stat_name} value`),
      })),
    all: () =>
      (all.all() as LeaderboardRow[]).map((row) => ({
        userID: row.user_id,
        username: row.username,
        statName: row.stat_name,
        value: toSafeInteger(row.value, `${row.stat_name} leaderboard value`),
      })),
    adjust: (userID, statName, amount) => {
      adjust.run(requireSafeInteger(amount, 'stat adjustment'), userID, statName);
    },
    set: (userID, statName, value) => {
      set.run(requireSafeInteger(value, 'stat value'), userID, statName);
    },
    reset: (userID, definitions) => resetTransaction(userID, definitions),
  };
};
