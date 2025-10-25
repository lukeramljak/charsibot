import { Client, createClient } from "@libsql/client";
import { log } from "./logger";
import type { CollectionType, RewardColumn } from "@charsibot/shared/types";

export interface Tokens {
  access_token: string;
  refresh_token: string;
}

export interface Stats {
  id: string;
  username: string;
  strength: number;
  intelligence: number;
  charisma: number;
  luck: number;
  dexterity: number;
  penis: number;
}

export type TokenType = "bot" | "streamer";

const CREATE_TOKEN_TABLE = `CREATE TABLE IF NOT EXISTS oauth_tokens (
  token_type TEXT PRIMARY KEY,
  access_token TEXT NOT NULL,
  refresh_token TEXT NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`;

const CREATE_STATS_TABLE = `CREATE TABLE IF NOT EXISTS stats (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL,
  strength INTEGER DEFAULT 3,
  intelligence INTEGER DEFAULT 3,
  charisma INTEGER DEFAULT 3,
  luck INTEGER DEFAULT 3,
  dexterity INTEGER DEFAULT 3,
  penis INTEGER DEFAULT 3
)`;

const CREATE_USER_COLLECTIONS_TABLE = `CREATE TABLE IF NOT EXISTS user_collections (
  user_id TEXT NOT NULL,
  username TEXT NOT NULL,
  collection_type TEXT NOT NULL,
  reward1 INTEGER DEFAULT 0,
  reward2 INTEGER DEFAULT 0,
  reward3 INTEGER DEFAULT 0,
  reward4 INTEGER DEFAULT 0,
  reward5 INTEGER DEFAULT 0,
  reward6 INTEGER DEFAULT 0,
  reward7 INTEGER DEFAULT 0,
  reward8 INTEGER DEFAULT 0,
  PRIMARY KEY (user_id, collection_type)
)`;

const REWARD_COLUMNS = [
  "reward1",
  "reward2",
  "reward3",
  "reward4",
  "reward5",
  "reward6",
  "reward7",
  "reward8",
] as const;

const ALL_COLUMNS = REWARD_COLUMNS.join(", ");

export class Store {
  public db: Client;

  constructor(url: string, authToken?: string) {
    if (authToken) {
      this.db = createClient({ url, authToken });
    } else {
      this.db = createClient({ url: `file:${url}` });
    }
  }

  async init() {
    await this.db.execute(CREATE_TOKEN_TABLE);
    await this.db.execute(CREATE_STATS_TABLE);
    await this.db.execute(CREATE_USER_COLLECTIONS_TABLE);
  }

  async getTokens(tokenType: TokenType): Promise<Tokens | null> {
    const res = await this.db.execute({
      sql: "SELECT access_token, refresh_token FROM oauth_tokens WHERE token_type = ?",
      args: [tokenType],
    });

    if (!res.rows.length) {
      return null;
    }

    const row = res.rows[0];

    return {
      access_token: String(row[0]),
      refresh_token: String(row[1]),
    };
  }

  async saveTokens(tokenType: TokenType, access: string, refresh: string) {
    await this.db.execute({
      sql: `INSERT INTO oauth_tokens (token_type, access_token, refresh_token, updated_at)
      VALUES (?, ?, ?, CURRENT_TIMESTAMP)
      ON CONFLICT(token_type) DO UPDATE SET access_token = excluded.access_token, refresh_token = excluded.refresh_token, updated_at = CURRENT_TIMESTAMP`,
      args: [tokenType, access, refresh],
    });
  }

  async getStats(userId: string, username: string): Promise<Stats> {
    await this.db.execute({
      sql: "INSERT INTO stats (id, username) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET username=excluded.username",
      args: [userId, username],
    });

    const res = await this.db.execute({
      sql: "SELECT id, username, strength, intelligence, charisma, luck, dexterity, penis FROM stats WHERE id=?",
      args: [userId],
    });

    const row = res.rows[0];

    return {
      id: String(row[0]),
      username: String(row[1]),
      strength: Number(row[2]),
      intelligence: Number(row[3]),
      charisma: Number(row[4]),
      luck: Number(row[5]),
      dexterity: Number(row[6]),
      penis: Number(row[7]),
    };
  }

  async modifyStat(
    userId: string,
    username: string,
    column: string,
    delta: number
  ) {
    const valid = [
      "strength",
      "intelligence",
      "charisma",
      "luck",
      "dexterity",
      "penis",
    ];

    if (!valid.includes(column)) {
      throw new Error(`invalid stat column: ${column}`);
    }

    await this.db.execute({
      sql: "INSERT INTO stats (id, username) VALUES (?, ?) ON CONFLICT(id) DO NOTHING",
      args: [userId, username],
    });

    await this.db.execute({
      sql: `UPDATE stats SET ${column} = ${column} + ? WHERE id = ?`,
      args: [delta, userId],
    });
  }

  async getUserCollections(
    userId: string,
    collectionType: CollectionType
  ): Promise<string[] | undefined> {
    try {
      const result = await this.db.execute({
        sql: `SELECT ${ALL_COLUMNS} FROM user_collections WHERE user_id = ? AND collection_type = ?`,
        args: [userId, collectionType],
      });

      const collection: string[] = [];
      if (result.rows.length > 0) {
        const row = result.rows[0];
        // Convert boolean columns to array of reward keys
        for (const column of REWARD_COLUMNS) {
          if (row[column]) {
            collection.push(column);
          }
        }
      }

      log.info(`getUserCollection - Success: ${collection.length} items found`);
      return collection;
    } catch (error) {
      log.error({ error }, "getUserCollection");
    }
  }

  async addPlushieToCollection(
    userId: string,
    username: string,
    collectionType: CollectionType,
    rewardColumn: RewardColumn
  ): Promise<{ collection: string[]; isNew: boolean } | undefined> {
    log.info(
      `addPlushieToCollection - userId: ${userId}, username: ${username}, collectionType: ${collectionType}, rewardColumn: ${rewardColumn}`
    );

    try {
      // Check if user already has this reward
      const existingResult = await this.db.execute({
        sql: `SELECT ${rewardColumn} FROM user_collections WHERE user_id = ? AND collection_type = ?`,
        args: [userId, collectionType],
      });

      const alreadyHas =
        existingResult.rows.length > 0 &&
        existingResult.rows[0][rewardColumn] === 1;

      if (!alreadyHas) {
        log.info(
          `addPlushieToCollection - Adding new plushie: ${rewardColumn}`
        );

        // Insert or update the user's collection
        await this.db.execute({
          sql: `INSERT INTO user_collections (user_id, username, collection_type, ${rewardColumn})
              VALUES (?, ?, ?, 1)
              ON CONFLICT(user_id, collection_type) DO UPDATE SET ${rewardColumn} = 1, username = ?`,
          args: [userId, username, collectionType, username],
        });
      } else {
        log.info(
          `addPlushieToCollection - Plushie already exists: ${rewardColumn}`
        );
      }

      // Get updated collection
      const updatedResult = await this.db.execute({
        sql: `SELECT ${ALL_COLUMNS} FROM user_collections WHERE user_id = ? AND collection_type = ?`,
        args: [userId, collectionType],
      });

      const collection: string[] = [];
      if (updatedResult.rows.length > 0) {
        const row = updatedResult.rows[0];
        for (const column of REWARD_COLUMNS) {
          if (row[column]) {
            collection.push(column);
          }
        }
      }

      log.info(
        `addPlushieToCollection - Success: ${
          collection.length
        } items, isNew: ${!alreadyHas}`
      );

      return { collection, isNew: !alreadyHas };
    } catch (error) {
      log.error({ error }, "addPlushieToCollection");
    }
  }
}

export const formatStats = (username: string, s: Stats): string => {
  return `${username}'s stats: STR: ${s.strength} | INT: ${s.intelligence} | CHA: ${s.charisma} | LUCK: ${s.luck} | DEX: ${s.dexterity} | PENIS: ${s.penis}`;
};

export interface Stat {
  display: string;
  column: string;
}

export const statList: Stat[] = [
  { display: "Strength", column: "strength" },
  { display: "Intelligence", column: "intelligence" },
  { display: "Charisma", column: "charisma" },
  { display: "Luck", column: "luck" },
  { display: "Dexterity", column: "dexterity" },
  { display: "Penis", column: "penis" },
];
