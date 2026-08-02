export interface Viewer {
  id: string;
  username: string;
  lastActiveAt?: string;
}

export interface UserStat {
  name: string;
  shortName: string;
  longName: string;
  value: number;
}

export interface LeaderboardRow {
  emoji: string;
  username: string;
  value: number;
}
