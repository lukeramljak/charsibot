import { int, numeric, primaryKey, sqliteTable, text } from 'drizzle-orm/sqlite-core';
import { sql } from 'drizzle-orm';

export const tokensTable = sqliteTable('oauth_tokens', {
  tokenType: text().primaryKey(),
  accessToken: text().notNull(),
  refreshToken: text().notNull(),
  updatedAt: numeric().default(sql`CURRENT_TIMESTAMP`),
});

export const statsTable = sqliteTable('stats', {
  id: text().primaryKey(),
  username: text().notNull(),
  strength: int().notNull().default(3),
  intelligence: int().notNull().default(3),
  charisma: int().notNull().default(3),
  luck: int().notNull().default(3),
  dexterity: int().notNull().default(3),
  penis: int().notNull().default(3),
});

export const userCollectionsTable = sqliteTable(
  'user_collections',
  {
    userId: text(),
    username: text().notNull(),
    collectionType: text(),
    reward1: int().default(0),
    reward2: int().default(0),
    reward3: int().default(0),
    reward4: int().default(0),
    reward5: int().default(0),
    reward6: int().default(0),
    reward7: int().default(0),
    reward8: int().default(0),
  },
  (table) => [
    primaryKey({
      name: 'user_collections_user_id_collection_type_pk',
      columns: [table.userId, table.collectionType],
    }),
  ],
);
