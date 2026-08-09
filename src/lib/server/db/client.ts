import BetterSqlite3 from 'better-sqlite3';
import { drizzle, type BetterSQLite3Database } from 'drizzle-orm/better-sqlite3';
import { migrate } from 'drizzle-orm/better-sqlite3/migrator';
import { resolve } from 'node:path';

import * as schema from '$lib/server/db/schema';

export interface DatabaseConnection {
  database: BetterSqlite3.Database;
  db: BetterSQLite3Database<typeof schema>;
  state: 'created' | 'existing';
  checkpoint: () => void;
  close: () => void;
}

interface CheckpointRow {
  busy: bigint;
}

const migrationsFolder = resolve(process.cwd(), 'drizzle');

const hasApplicationObjects = (database: BetterSqlite3.Database): boolean => {
  const [row] = database
    .prepare(
      `SELECT COUNT(*) AS count FROM sqlite_master
       WHERE name NOT LIKE 'sqlite_%'
         AND name != '__drizzle_migrations'
         AND type IN ('table', 'index', 'view', 'trigger')`,
    )
    .all() as { count: bigint }[];

  return row.count > 0n;
};

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

    const state = hasApplicationObjects(database) ? 'existing' : 'created';
    const db = drizzle(database, { schema });
    migrate(db, { migrationsFolder });
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

    return { database, db, state, checkpoint, close };
  } catch (error) {
    database.close();

    throw error;
  }
};
