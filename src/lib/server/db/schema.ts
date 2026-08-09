import { index, integer, primaryKey, sqliteTable, text } from 'drizzle-orm/sqlite-core';

export const userStats = sqliteTable(
  'user_stats',
  {
    userID: text('user_id').notNull(),
    username: text('username').notNull(),
    statName: text('stat_name').notNull(),
    value: integer('value').notNull().default(3),
  },
  (table) => [primaryKey({ columns: [table.userID, table.statName] })],
);

export const userPlushies = sqliteTable(
  'user_plushies',
  {
    userID: text('user_id').notNull(),
    username: text('username').notNull(),
    series: text('series').notNull(),
    key: text('key').notNull(),
  },
  (table) => [primaryKey({ columns: [table.userID, table.series, table.key] })],
);

export const viewerActivity = sqliteTable(
  'viewer_activity',
  {
    userID: text('user_id').primaryKey(),
    username: text('username').notNull(),
    lastActiveAt: text('last_active_at').notNull(),
  },
  (table) => [index('viewer_activity_last_active_at_idx').on(table.lastActiveAt)],
);
