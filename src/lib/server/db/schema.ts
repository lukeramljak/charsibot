import type BetterSqlite3 from 'better-sqlite3';

import { toSafeInteger } from '$lib/server/db/integer';

interface TableInfoRow {
  name: string;
  type: string;
  notnull: bigint;
  dflt_value: string | null;
  pk: bigint;
}

interface IndexListRow {
  name: string;
  unique: bigint;
}

interface IndexInfoRow {
  seqno: bigint;
  name: string;
}

interface ColumnExpectation {
  name: string;
  type: string;
  notNull: boolean;
  primaryKeyPosition: number;
  defaultValue: string | null;
}

const requiredTables: Record<string, readonly ColumnExpectation[]> = {
  user_stats: [
    { name: 'user_id', type: 'TEXT', notNull: true, primaryKeyPosition: 1, defaultValue: null },
    { name: 'username', type: 'TEXT', notNull: true, primaryKeyPosition: 0, defaultValue: null },
    { name: 'stat_name', type: 'TEXT', notNull: true, primaryKeyPosition: 2, defaultValue: null },
    { name: 'value', type: 'INTEGER', notNull: true, primaryKeyPosition: 0, defaultValue: '3' },
  ],
  user_plushies: [
    { name: 'user_id', type: 'TEXT', notNull: true, primaryKeyPosition: 1, defaultValue: null },
    { name: 'username', type: 'TEXT', notNull: true, primaryKeyPosition: 0, defaultValue: null },
    { name: 'series', type: 'TEXT', notNull: true, primaryKeyPosition: 2, defaultValue: null },
    { name: 'key', type: 'TEXT', notNull: true, primaryKeyPosition: 3, defaultValue: null },
  ],
  viewer_activity: [
    { name: 'user_id', type: 'TEXT', notNull: false, primaryKeyPosition: 1, defaultValue: null },
    { name: 'username', type: 'TEXT', notNull: true, primaryKeyPosition: 0, defaultValue: null },
    {
      name: 'last_active_at',
      type: 'TEXT',
      notNull: true,
      primaryKeyPosition: 0,
      defaultValue: null,
    },
  ],
};

const requiredApplicationObjects = [
  'user_plushies',
  'user_stats',
  'viewer_activity',
  'viewer_activity_last_active_at_idx',
] as const;

const createSchema = (database: BetterSqlite3.Database): void => {
  database.transaction(() => {
    database.exec(`
      CREATE TABLE user_stats (
        user_id TEXT NOT NULL,
        username TEXT NOT NULL,
        stat_name TEXT NOT NULL,
        value INTEGER NOT NULL DEFAULT 3,
        PRIMARY KEY (user_id, stat_name)
      );
      CREATE TABLE user_plushies (
        user_id TEXT NOT NULL,
        username TEXT NOT NULL,
        series TEXT NOT NULL,
        key TEXT NOT NULL,
        PRIMARY KEY (user_id, series, key)
      );
      CREATE TABLE viewer_activity (
        user_id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        last_active_at TEXT NOT NULL
      );
      CREATE INDEX viewer_activity_last_active_at_idx
        ON viewer_activity(last_active_at);
    `);
  })();
};

const dropLegacyGooseTable = (database: BetterSqlite3.Database): void => {
  database.exec('DROP TABLE IF EXISTS goose_db_version');
};

const listApplicationObjects = (database: BetterSqlite3.Database): string[] => {
  const rows = database
    .prepare(
      `SELECT name FROM sqlite_master
       WHERE name NOT LIKE 'sqlite_%' AND type IN ('table', 'index', 'view', 'trigger')`,
    )
    .all() as { name: string }[];

  return rows.map((row) => row.name).sort();
};

export const isEmptyDatabase = (database: BetterSqlite3.Database): boolean =>
  listApplicationObjects(database).length === 0;

const validateTable = (
  database: BetterSqlite3.Database,
  tableName: string,
  expected: readonly ColumnExpectation[],
): void => {
  const escapedName = tableName.replaceAll("'", "''");
  const actual = database.pragma(`table_info('${escapedName}')`) as TableInfoRow[];

  if (actual.length !== expected.length) {
    throw new Error(`database table ${tableName} has an incompatible column count`);
  }

  expected.forEach((column, index) => {
    const row = actual[index];

    if (
      row.name !== column.name ||
      row.type.toUpperCase() !== column.type ||
      toSafeInteger(row.notnull, `${tableName}.${column.name}.notnull`) !==
        Number(column.notNull) ||
      toSafeInteger(row.pk, `${tableName}.${column.name}.pk`) !== column.primaryKeyPosition ||
      row.dflt_value !== column.defaultValue
    ) {
      throw new Error(`database column ${tableName}.${column.name} is incompatible`);
    }
  });
};

const validateActivityIndex = (database: BetterSqlite3.Database): void => {
  const indexes = database.pragma("index_list('viewer_activity')") as IndexListRow[];
  const index = indexes.find(
    (candidate) => candidate.name === 'viewer_activity_last_active_at_idx',
  );

  if (!index || toSafeInteger(index.unique, 'viewer activity index uniqueness') !== 0) {
    throw new Error('database is missing the viewer activity timestamp index');
  }

  const columns = database.pragma(
    "index_info('viewer_activity_last_active_at_idx')",
  ) as IndexInfoRow[];

  if (
    columns.length !== 1 ||
    toSafeInteger(columns[0].seqno, 'viewer activity index position') !== 0 ||
    columns[0].name !== 'last_active_at'
  ) {
    throw new Error('database viewer activity timestamp index is incompatible');
  }
};

const validateSchema = (database: BetterSqlite3.Database): void => {
  const objects = listApplicationObjects(database);

  if (
    objects.length !== requiredApplicationObjects.length ||
    objects.some((name, index) => name !== requiredApplicationObjects[index])
  ) {
    throw new Error('database application objects do not match the expected schema');
  }

  for (const [tableName, columns] of Object.entries(requiredTables)) {
    validateTable(database, tableName, columns);
  }

  validateActivityIndex(database);
};

export const initializeOrValidate = (database: BetterSqlite3.Database): 'created' | 'existing' => {
  dropLegacyGooseTable(database);

  if (isEmptyDatabase(database)) {
    createSchema(database);
    validateSchema(database);

    return 'created';
  }

  validateSchema(database);

  return 'existing';
};
