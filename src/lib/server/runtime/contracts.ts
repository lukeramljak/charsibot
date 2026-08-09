export type ReadinessComponent = 'catalog' | 'database' | 'twitch';

export interface ReadinessSnapshot {
  ready: boolean;
  components: Record<ReadinessComponent, boolean>;
}

export interface Readiness {
  snapshot: () => ReadinessSnapshot;
  set: (component: ReadinessComponent, ready: boolean) => void;
}

export interface TwitchRuntime {
  start: (signal: AbortSignal) => Promise<void>;
  stop: (reason: string) => Promise<void>;
}

export interface ApplicationRuntime {
  stop: (reason: string) => Promise<void>;
}

export interface Environment {
  DB_PATH?: string | number;
  TWITCH_CLIENT_ID?: string | number;
  TWITCH_CLIENT_SECRET?: string | number;
  TWITCH_BOT_USER_ID?: string | number;
  TWITCH_CHANNEL_USER_ID?: string | number;
}

export type RuntimeFactory = (environment?: Environment) => Promise<ApplicationRuntime>;
