import type { BlindBoxSeries, StatDefinition } from '$lib/contracts/catalog';
import type { OverlayEvent } from '$lib/contracts/overlay';
import type {
  BlindBoxService,
  ChatMessage,
  ChatSender,
  Clock,
  Logger,
  OverlayBus,
  Random,
  StatsService,
} from '$lib/server/application/ports';
import type { BotDependencies } from '$lib/server/bot';

const definitions: readonly StatDefinition[] = [
  {
    name: 'strength',
    shortName: 'STR',
    longName: 'Strength',
    defaultValue: 3,
    sortOrder: 1,
    emoji: 'strong',
  },
  {
    name: 'luck',
    shortName: 'LUCK',
    longName: 'Luck',
    defaultValue: 3,
    sortOrder: 2,
    emoji: 'lucky',
  },
];

export const testSeries: BlindBoxSeries = {
  series: 'coobubu',
  redemptionTitle: 'Cooper Series Blind Box',
  name: 'Coobubus',
  revealSound: '/reveal.mp3',
  boxFrontFace: '/front.png',
  boxSideFace: '/side.png',
  displayColor: '#000000',
  textColor: '#ffffff',
  plushies: [
    {
      series: 'coobubu',
      key: 'cutey',
      sortOrder: 1,
      weight: 2,
      name: 'Cutey',
      image: '/cutey.png',
      emptyImage: '/empty.png',
    },
    {
      series: 'coobubu',
      key: 'secret',
      sortOrder: 2,
      weight: 1,
      name: 'Secret',
      image: '/secret.png',
      emptyImage: '/empty.png',
    },
  ],
};

export const createStatsFake = (overrides: Partial<StatsService> = {}): StatsService => ({
  definitions,
  getOrCreate: async () => [
    { name: 'strength', shortName: 'STR', longName: 'Strength', value: 3 },
    { name: 'luck', shortName: 'LUCK', longName: 'Luck', value: 3 },
  ],
  get: async () => [
    { name: 'strength', shortName: 'STR', longName: 'Strength', value: 4 },
    { name: 'luck', shortName: 'LUCK', longName: 'Luck', value: 3 },
  ],
  leaderboard: async () => [{ emoji: 'strong', username: 'alice', value: 4 }],
  listViewers: async () => [],
  getViewer: async (userID) => ({ id: userID, username: 'alice' }),
  recordActivity: async () => {},
  deleteViewers: async () => {},
  adjust: async () => {},
  set: async () => {},
  reset: async () => {},
  ...overrides,
});

export const createBlindBoxFake = (overrides: Partial<BlindBoxService> = {}): BlindBoxService => ({
  getViewerCollections: async () => [],
  getCollection: async () => ['cutey'],
  grant: async () => ({ isNew: true, collection: ['cutey'] }),
  remove: async () => {},
  reset: async () => {},
  completed: async () => [{ seriesName: 'Coobubus', usernames: 'alice' }],
  ...overrides,
});

export interface ChatFake extends ChatSender {
  messages: ChatMessage[];
  signals: AbortSignal[];
}

export const createChatFake = (): ChatFake => {
  const messages: ChatMessage[] = [];
  const signals: AbortSignal[] = [];

  return {
    messages,
    signals,
    send: async (message, signal) => {
      messages.push(message);
      if (signal) signals.push(signal);
    },
  };
};

export interface OverlayFake extends OverlayBus {
  events: OverlayEvent[];
}

export const createOverlayFake = (): OverlayFake => {
  const events: OverlayEvent[] = [];

  return {
    events,
    publish: (event) => events.push(event),
    subscribe: () => ({
      close: () => {},
      [Symbol.asyncIterator]: () => ({
        next: async () => ({ done: true, value: undefined as never }),
      }),
    }),
    close: () => {},
  };
};

export interface LogEntry {
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  attributes?: Record<string, unknown>;
}

export interface LoggerFake extends Logger {
  entries: LogEntry[];
}

export const createLoggerFake = (): LoggerFake => {
  const entries: LogEntry[] = [];
  const append = (
    level: LogEntry['level'],
    message: string,
    attributes?: Record<string, unknown>,
  ): void => {
    entries.push({ level, message, ...(attributes ? { attributes } : {}) });
  };

  return {
    entries,
    debug: (message, attributes) => append('debug', message, attributes),
    info: (message, attributes) => append('info', message, attributes),
    warn: (message, attributes) => append('warn', message, attributes),
    error: (message, attributes) => append('error', message, attributes),
  };
};

export interface RandomFake extends Random {
  maximums: number[];
}

export const createRandomFake = (values: readonly number[] = [0]): RandomFake => {
  const queue = [...values];
  const maximums: number[] = [];

  return {
    maximums,
    integer: (maxExclusive) => {
      maximums.push(maxExclusive);

      return queue.shift() ?? 0;
    },
  };
};

interface PendingSleep {
  milliseconds: number;
  resolve: () => void;
}

export interface ClockFake extends Clock {
  pending: () => readonly number[];
  resolveNext: (milliseconds: number) => void;
}

export const createClockFake = (): ClockFake => {
  const sleeps: PendingSleep[] = [];

  return {
    now: () => new Date('2026-08-02T01:02:03.456Z'),
    sleep: (milliseconds, signal) =>
      new Promise<void>((resolve, reject) => {
        if (signal?.aborted) {
          reject(signal.reason);

          return;
        }

        const pending: PendingSleep = {
          milliseconds,
          resolve: () => {
            signal?.removeEventListener('abort', abort);
            sleeps.splice(sleeps.indexOf(pending), 1);
            resolve();
          },
        };
        const abort = (): void => {
          sleeps.splice(sleeps.indexOf(pending), 1);
          reject(signal?.reason);
        };
        signal?.addEventListener('abort', abort, { once: true });
        sleeps.push(pending);
      }),
    pending: () => sleeps.map((sleep) => sleep.milliseconds),
    resolveNext: (milliseconds) => {
      const sleep = sleeps.find((candidate) => candidate.milliseconds === milliseconds);
      if (!sleep) throw new Error(`no pending ${milliseconds}ms sleep`);
      sleep.resolve();
    },
  };
};

export interface BotHarness {
  dependencies: BotDependencies;
  stats: StatsService;
  blindBox: BlindBoxService;
  chat: ChatFake;
  overlay: OverlayFake;
  clock: ClockFake;
  random: RandomFake;
  logger: LoggerFake;
}

export const createBotHarness = (overrides: Partial<BotDependencies> = {}): BotHarness => {
  const stats = overrides.stats ?? createStatsFake();
  const blindBox = overrides.blindBox ?? createBlindBoxFake();
  const chat = (overrides.chat as ChatFake | undefined) ?? createChatFake();
  const overlay = (overrides.overlay as OverlayFake | undefined) ?? createOverlayFake();
  const clock = (overrides.clock as ClockFake | undefined) ?? createClockFake();
  const random = (overrides.random as RandomFake | undefined) ?? createRandomFake();
  const logger = (overrides.logger as LoggerFake | undefined) ?? createLoggerFake();
  const dependencies: BotDependencies = {
    botUserID: 'bot-user',
    series: [testSeries],
    stats,
    blindBox,
    chat,
    overlay,
    clock,
    random,
    logger,
    ...overrides,
  };

  return { dependencies, stats, blindBox, chat, overlay, clock, random, logger };
};
