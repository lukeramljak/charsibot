import type { CollectionType, RewardColumn } from '@/blind-box/types';
import { log } from '@/logger';
import type { Stats } from '@/stats/types';
import * as schema from '@/storage/schema';
import { statsTable, tokensTable, userCollectionsTable } from '@/storage/schema';
import type { Database } from 'bun:sqlite';
import { and, eq, sql } from 'drizzle-orm';
import { drizzle } from 'drizzle-orm/bun-sqlite';
import { migrate } from 'drizzle-orm/bun-sqlite/migrator';

interface Tokens {
  accessToken: string;
  refreshToken: string;
}

export type TokenType = 'bot' | 'streamer';

const REWARD_COLUMNS = [
  'reward1',
  'reward2',
  'reward3',
  'reward4',
  'reward5',
  'reward6',
  'reward7',
  'reward8',
] as const;

export class Store {
  public db;

  constructor(dbInstance: Database) {
    this.db = drizzle(dbInstance, { schema, casing: 'snake_case' });
    migrate(this.db, { migrationsFolder: './drizzle' });
  }

  async init() {
    this.db.run('PRAGMA journal_mode = WAL;');
    log.info('storage ready');
  }

  async getTokens(tokenType: TokenType): Promise<Tokens | null> {
    const [tokens] = await this.db
      .select()
      .from(tokensTable)
      .where(eq(tokensTable.tokenType, tokenType));

    if (!tokens) {
      return null;
    }

    return {
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
    };
  }

  async saveTokens(tokenType: TokenType, accessToken: string, refreshToken: string): Promise<void> {
    await this.db
      .insert(tokensTable)
      .values({
        tokenType,
        accessToken,
        refreshToken,
        updatedAt: sql`CURRENT_TIMESTAMP`,
      })
      .onConflictDoUpdate({
        target: tokensTable.tokenType,
        set: {
          accessToken,
          refreshToken,
          updatedAt: sql`CURRENT_TIMESTAMP`,
        },
      });
  }

  async getStats(userId: string, username: string): Promise<Stats> {
    await this.db
      .insert(statsTable)
      .values({
        id: userId,
        username,
      })
      .onConflictDoUpdate({ target: statsTable.id, set: { username } });

    const [stats] = await this.db.select().from(statsTable).where(eq(statsTable.id, userId));

    return stats;
  }

  async modifyStat(userId: string, username: string, column: string, delta: number): Promise<void> {
    const valid = ['strength', 'intelligence', 'charisma', 'luck', 'dexterity', 'penis'];

    if (!valid.includes(column)) {
      throw new Error(`invalid stat column: ${column}`);
    }

    await this.db
      .insert(statsTable)
      .values({
        id: userId,
        username,
      })
      .onConflictDoNothing({ target: statsTable.id });

    await this.db
      .update(statsTable)
      .set({
        username,
        [column]: sql`${sql.identifier(column)} + ${delta}`,
      })
      .where(eq(statsTable.id, userId));
  }

  async getUserCollections(userId: string, collectionType: CollectionType): Promise<string[]> {
    const [result] = await this.db
      .select()
      .from(userCollectionsTable)
      .where(
        and(
          eq(userCollectionsTable.userId, userId),
          eq(userCollectionsTable.collectionType, collectionType),
        ),
      );

    if (!result) return [];

    const collection = REWARD_COLUMNS.filter((col) => result[col] === 1);

    log.info(`getUserCollection - Success: ${collection.length} items found`);

    return collection;
  }

  async addPlushieToCollection(
    userId: string,
    username: string,
    collectionType: CollectionType,
    rewardColumn: RewardColumn,
  ): Promise<{ collection: string[]; isNew: boolean } | undefined> {
    log.info(
      `addPlushieToCollection - userId: ${userId}, username: ${username}, collectionType: ${collectionType}, rewardColumn: ${rewardColumn}`,
    );

    const [existing] = await this.db
      .select()
      .from(userCollectionsTable)
      .where(
        and(
          eq(userCollectionsTable.userId, userId),
          eq(userCollectionsTable.collectionType, collectionType),
        ),
      );

    const alreadyHas = existing?.[rewardColumn] === 1;

    if (!alreadyHas) {
      log.info(`addPlushieToCollection - Adding new plushie: ${rewardColumn}`);

      await this.db
        .insert(userCollectionsTable)
        .values({
          userId,
          username,
          collectionType,
          [rewardColumn]: 1,
        })
        .onConflictDoUpdate({
          target: [userCollectionsTable.userId, userCollectionsTable.collectionType],
          set: { [rewardColumn]: 1, username },
        });
    } else {
      log.info(`addPlushieToCollection - Plushie already exists: ${rewardColumn}`);
    }

    let collection: string[];
    if (alreadyHas && existing) {
      collection = REWARD_COLUMNS.filter((col) => existing[col] === 1);
    } else {
      collection = await this.getUserCollections(userId, collectionType);
    }

    log.info(`addPlushieToCollection - Success: ${collection.length} items, isNew: ${!alreadyHas}`);

    return { collection, isNew: !alreadyHas };
  }

  async resetUserCollection(userId: string, collectionType: CollectionType) {
    const resetValues = Object.fromEntries(REWARD_COLUMNS.map((col) => [col, 0]));

    return this.db
      .update(userCollectionsTable)
      .set(resetValues)
      .where(
        and(
          eq(userCollectionsTable.userId, userId),
          eq(userCollectionsTable.collectionType, collectionType),
        ),
      );
  }

  async getCompletedCollections() {
    const rows = await this.db
      .select({
        collectionType: userCollectionsTable.collectionType,
        usernamesCsv: sql<string>`group_concat(${userCollectionsTable.username}, ',')`,
      })
      .from(userCollectionsTable)
      .where(
        sql`
      (${userCollectionsTable.reward1},
       ${userCollectionsTable.reward2},
       ${userCollectionsTable.reward3},
       ${userCollectionsTable.reward4},
       ${userCollectionsTable.reward5},
       ${userCollectionsTable.reward6},
       ${userCollectionsTable.reward7},
       ${userCollectionsTable.reward8})
       = (1,1,1,1,1,1,1,1)
    `,
      )
      .groupBy(userCollectionsTable.collectionType);

    return rows.map((r) => ({
      collectionType: r.collectionType,
      usernames: r.usernamesCsv?.split(',') ?? [],
    }));
  }

  async getStatLeaderboard() {
    return this.db
      .select({
        top_strength_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY strength DESC LIMIT 1)`,
        top_strength_value: sql<number>`(SELECT strength FROM ${statsTable} ORDER BY strength DESC LIMIT 1)`,
        top_intelligence_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY intelligence DESC LIMIT 1)`,
        top_intelligence_value: sql<number>`(SELECT intelligence FROM ${statsTable} ORDER BY intelligence DESC LIMIT 1)`,
        top_charisma_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY charisma DESC LIMIT 1)`,
        top_charisma_value: sql<number>`(SELECT charisma FROM ${statsTable} ORDER BY charisma DESC LIMIT 1)`,
        top_luck_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY luck DESC LIMIT 1)`,
        top_luck_value: sql<number>`(SELECT luck FROM ${statsTable} ORDER BY luck DESC LIMIT 1)`,
        top_dexterity_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY dexterity DESC LIMIT 1)`,
        top_dexterity_value: sql<number>`(SELECT dexterity FROM ${statsTable} ORDER BY dexterity DESC LIMIT 1)`,
        top_penis_username: sql<string>`(SELECT username FROM ${statsTable} ORDER BY penis DESC LIMIT 1)`,
        top_penis_value: sql<number>`(SELECT penis FROM ${statsTable} ORDER BY penis DESC LIMIT 1)`,
      })
      .from(statsTable)
      .limit(1);
  }
}
