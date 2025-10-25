import { Client, createClient } from "@libsql/client";

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

export class Store {
  public db: Client;

  constructor(dbPath: string = "charsibot.db") {
    this.db = createClient({
      url: `file:${dbPath}`,
    });
  }

  async init() {
    await this.db.execute(CREATE_TOKEN_TABLE);
    await this.db.execute(CREATE_STATS_TABLE);
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
