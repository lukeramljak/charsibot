import BetterSqlite3 from 'better-sqlite3';

import { initializeOrValidateV7 } from '$lib/server/db/schema';

export interface DatabaseConnection {
  database: BetterSqlite3.Database;
  state: 'created' | 'existing';
  checkpoint: () => void;
  close: () => void;
}

interface CheckpointRow {
  busy: bigint;
}

const applyPragmas = (database: BetterSqlite3.Database): void => {
  database.pragma('foreign_keys = ON');
  database.pragma('journal_mode = WAL');
  database.pragma('page_size = 4096');
  database.pragma('cache_size = -8000');
  database.pragma('synchronous = NORMAL');
  database.pragma('secure_delete = ON');
  database.pragma('busy_timeout = 30000');
};

export const openDatabase = (path: string): DatabaseConnection => {
  const database = new BetterSqlite3(path);

  database.defaultSafeIntegers(true);

  try {
    applyPragmas(database);

    const state = initializeOrValidateV7(database);
    let closed = false;

    const checkpoint = (): void => {
      if (!closed) {
        const [result] = database.pragma('wal_checkpoint(TRUNCATE)') as CheckpointRow[];

        if (result.busy !== 0n) {
          throw new Error('SQLite WAL checkpoint could not complete');
        }
      }
    };

    const close = (): void => {
      if (closed) {
        return;
      }

      try {
        checkpoint();
      } finally {
        database.close();
        closed = true;
      }
    };

    return { database, state, checkpoint, close };
  } catch (error) {
    database.close();

    throw error;
  }
};
