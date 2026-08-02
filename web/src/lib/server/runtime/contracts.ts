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

export type RuntimeFactory = () => Promise<ApplicationRuntime>;
