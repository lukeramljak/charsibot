import { log } from "./logger";
import type { CollectionType, RewardColumn } from "./types";
import { Database } from "bun:sqlite";

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
  public db: Database;

  constructor(dbPath: string) {
    this.db = new Database(dbPath, { strict: true });
  }

  async init() {
    this.db.run("PRAGMA journal_mode = WAL;");
    this.db.run(CREATE_TOKEN_TABLE);
    this.db.run(CREATE_STATS_TABLE);
    this.db.run(CREATE_USER_COLLECTIONS_TABLE);
  }

  getTokens(tokenType: TokenType): Tokens | null {
    const query = this.db.query(
      "SELECT access_token, refresh_token FROM oauth_tokens WHERE token_type = $tokenType"
    );
    const results = query.all({
      tokenType: tokenType,
    }) as any[];

    if (results.length === 0) {
      return null;
    }

    return {
      access_token: String(results[0].access_token),
      refresh_token: String(results[0].refresh_token),
    };
  }

  saveTokens(tokenType: TokenType, access: string, refresh: string) {
    const query = this.db.query(
      `INSERT INTO oauth_tokens (token_type, access_token, refresh_token, updated_at)
      VALUES ($tokenType, $access, $refresh, CURRENT_TIMESTAMP)
      ON CONFLICT(token_type) DO UPDATE SET access_token = excluded.access_token, refresh_token = excluded.refresh_token, updated_at = CURRENT_TIMESTAMP`
    );
    query.run({
      tokenType: tokenType,
      access: access,
      refresh: refresh,
    });
  }

  getStats(userId: string, username: string): Stats {
    const query = this.db.query(
      "INSERT INTO stats (id, username) VALUES ($userId, $username) ON CONFLICT(id) DO UPDATE SET username=excluded.username"
    );
    query.run({
      userId: userId,
      username: username,
    });

    const statsQuery = this.db.query(
      "SELECT id, username, strength, intelligence, charisma, luck, dexterity, penis FROM stats WHERE id = $id"
    );
    const result = statsQuery.get({
      id: userId,
    }) as Stats;

    return {
      id: String(result.id),
      username: String(result.username),
      strength: Number(result.strength),
      intelligence: Number(result.intelligence),
      charisma: Number(result.charisma),
      luck: Number(result.luck),
      dexterity: Number(result.dexterity),
      penis: Number(result.penis),
    };
  }

  modifyStat(userId: string, username: string, column: string, delta: number) {
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

    const insertQuery = this.db.query(
      "INSERT INTO stats (id, username) VALUES ($userId, $username) ON CONFLICT(id) DO NOTHING"
    );
    insertQuery.run({
      userId: userId,
      username: username,
    });

    const updateQuery = this.db.query(
      `UPDATE stats SET ${column} = ${column} + $delta WHERE id = $userId`
    );
    updateQuery.run({
      delta: delta,
      userId: userId,
    });
  }

  getUserCollections(
    userId: string,
    collectionType: CollectionType
  ): string[] | undefined {
    try {
      const query = this.db.query(
        `SELECT ${ALL_COLUMNS} FROM user_collections WHERE user_id = $userId AND collection_type = $collectionType`
      );
      const result = query.get({
        userId: userId,
        collectionType: collectionType,
      }) as any;

      const collection: string[] = [];
      if (result) {
        // Convert boolean columns to array of reward keys
        for (const column of REWARD_COLUMNS) {
          if (result[column]) {
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

  addPlushieToCollection(
    userId: string,
    username: string,
    collectionType: CollectionType,
    rewardColumn: RewardColumn
  ): { collection: string[]; isNew: boolean } | undefined {
    log.info(
      `addPlushieToCollection - userId: ${userId}, username: ${username}, collectionType: ${collectionType}, rewardColumn: ${rewardColumn}`
    );

    try {
      // Check if user already has this reward
      const existingQuery = this.db.query(
        `SELECT ${rewardColumn} FROM user_collections WHERE user_id = $userId AND collection_type = $collectionType`
      );
      const existingResult = existingQuery.get({
        userId: userId,
        collectionType: collectionType,
      }) as any;

      const alreadyHas = existingResult && existingResult[rewardColumn] === 1;

      if (!alreadyHas) {
        log.info(
          `addPlushieToCollection - Adding new plushie: ${rewardColumn}`
        );

        // Insert or update the user's collection
        const insertQuery = this.db.query(
          `INSERT INTO user_collections (user_id, username, collection_type, ${rewardColumn})
              VALUES ($userId, $username, $collectionType, 1)
              ON CONFLICT(user_id, collection_type) DO UPDATE SET ${rewardColumn} = 1, username = $username`
        );
        insertQuery.run({
          userId: userId,
          username: username,
          collectionType: collectionType,
        });
      } else {
        log.info(
          `addPlushieToCollection - Plushie already exists: ${rewardColumn}`
        );
      }

      // Get updated collection
      const query = this.db.query(
        `SELECT ${ALL_COLUMNS} FROM user_collections WHERE user_id = $userId AND collection_type = $collectionType`
      );
      const updatedResult = query.get({
        userId: userId,
        collectionType: collectionType,
      }) as any;

      const collection: string[] = [];
      if (updatedResult) {
        for (const column of REWARD_COLUMNS) {
          if (updatedResult[column]) {
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
