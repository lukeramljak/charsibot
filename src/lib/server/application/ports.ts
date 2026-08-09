import type { StatDefinition } from '$lib/contracts/catalog';
import type { ViewerCollection } from '$lib/contracts/collections';
import type { OverlayEvent } from '$lib/contracts/overlay';
import type { LeaderboardRow, UserStat, Viewer } from '$lib/contracts/viewer';

export interface Clock {
  now: () => Date;
  sleep: (milliseconds: number, signal?: AbortSignal) => Promise<void>;
}

export interface Random {
  integer: (maxExclusive: number) => number;
}

export interface Logger {
  debug: (message: string, attributes?: Record<string, unknown>) => void;
  info: (message: string, attributes?: Record<string, unknown>) => void;
  warn: (message: string, attributes?: Record<string, unknown>) => void;
  error: (message: string, attributes?: Record<string, unknown>) => void;
}

export interface ChatMessage {
  message: string;
  replyParentMessageID?: string;
}

export interface ChatSender {
  send: (message: ChatMessage, signal?: AbortSignal) => Promise<void>;
}

export interface OverlaySubscription extends AsyncIterable<OverlayEvent> {
  close: () => void;
}

export interface OverlayBus {
  publish: (event: OverlayEvent) => void;
  subscribe: () => OverlaySubscription;
  close: () => void;
}

export interface StatsService {
  readonly definitions: readonly StatDefinition[];

  getOrCreate: (userID: string, username: string) => Promise<UserStat[]>;
  get: (userID: string) => Promise<UserStat[]>;
  leaderboard: () => Promise<LeaderboardRow[]>;
  listViewers: () => Promise<Viewer[]>;
  getViewer: (userID: string) => Promise<Viewer>;
  recordActivity: (userID: string, username: string, at: Date) => Promise<void>;
  deleteViewers: (userIDs: readonly string[]) => Promise<void>;
  adjust: (userID: string, statName: string, amount: number) => Promise<void>;
  set: (userID: string, statName: string, value: number) => Promise<void>;
  reset: (userID: string) => Promise<void>;
}

export interface GrantPlushieResult {
  isNew: boolean;
  collection: string[];
}

export interface CompletedCollection {
  seriesName: string;
  usernames: string;
}

export interface BlindBoxService {
  getViewerCollections: (userID: string) => Promise<ViewerCollection[]>;
  getCollection: (userID: string, series: string) => Promise<string[]>;
  grant: (
    userID: string,
    username: string,
    series: string,
    key: string,
  ) => Promise<GrantPlushieResult>;
  remove: (userID: string, series: string, key: string) => Promise<void>;
  reset: (userID: string, series: string) => Promise<void>;
  completed: () => Promise<CompletedCollection[]>;
}
