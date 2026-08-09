import { randomInt } from 'node:crypto';

import type { ChatSender, Clock, Logger, OverlayBus, Random } from '$lib/server/application/ports';
import { createBot } from '$lib/server/bot/bot';
import type { Bot } from '$lib/server/bot/types';
import { loadCatalog } from '$lib/server/catalog/load';
import { openDatabase, type DatabaseConnection } from '$lib/server/db/client';
import { createBlindBoxService } from '$lib/server/domain/blind-box/service';
import { createStatsService } from '$lib/server/domain/stats/service';
import { createOverlayBus } from '$lib/server/events/overlay-bus';
import { setReadiness, setServices } from '$lib/server/runtime/container';
import type { ApplicationRuntime, Environment } from '$lib/server/runtime/contracts';
import { createReadiness } from '$lib/server/runtime/readiness';
import { createBotNotificationHandler } from '$lib/server/twitch/bot-adapter';
import { createTwitchChatSender } from '$lib/server/twitch/chat-sender';
import { createConduitSessionManager } from '$lib/server/twitch/conduits';
import { createEventSubTransport } from '$lib/server/twitch/eventsub';
import { createHelixClient, createTwitchChatClient } from '$lib/server/twitch/helix';
import { createAppTokenProvider } from '$lib/server/twitch/token';
import type { EventSubTransport } from '$lib/server/twitch/types';
import { createNodeEventSubConnector } from '$lib/server/twitch/websocket';

export interface TwitchConfig {
  clientId: string;
  clientSecret: string;
  botUserId: string;
  channelUserId: string;
}

export interface RuntimeConfig {
  dbPath: string;
  logger: Logger;
  mockTwitch: boolean;
  twitch?: TwitchConfig;
}

const createLogger = (): Logger => ({
  debug: (message, attributes) => console.debug(message, attributes ?? ''),
  info: (message, attributes) => console.info(message, attributes ?? ''),
  warn: (message, attributes) => console.warn(message, attributes ?? ''),
  error: (message, attributes) => console.error(message, attributes ?? ''),
});

const createClock = (): Clock => ({
  now: () => new Date(),
  sleep: (milliseconds, signal) =>
    new Promise((resolve, reject) => {
      if (signal?.aborted) {
        reject(signal.reason);
        return;
      }

      const timeout = setTimeout(resolve, milliseconds);
      signal?.addEventListener(
        'abort',
        () => {
          clearTimeout(timeout);
          reject(signal.reason);
        },
        { once: true },
      );
    }),
});

const createRandom = (): Random => ({
  integer: (maxExclusive) => randomInt(maxExclusive),
});

const readTwitchConfig = (environment: Environment): TwitchConfig => {
  const clientId = environment.TWITCH_CLIENT_ID;
  const clientSecret = environment.TWITCH_CLIENT_SECRET;
  const botUserId = environment.TWITCH_BOT_USER_ID;
  const channelUserId = environment.TWITCH_CHANNEL_USER_ID;

  if (!clientId || !clientSecret || !botUserId || !channelUserId) {
    throw new Error(
      'TWITCH_CLIENT_ID, TWITCH_CLIENT_SECRET, TWITCH_BOT_USER_ID, and TWITCH_CHANNEL_USER_ID are required unless TWITCH_MOCK_MODE=true',
    );
  }

  return {
    clientId: String(clientId),
    clientSecret: String(clientSecret),
    botUserId: String(botUserId),
    channelUserId: String(channelUserId),
  };
};

const createLoggingChatStub = (logger: Logger): ChatSender => ({
  send: async (message) => {
    logger.warn('chat not connected, message dropped', { message: message.message });
  },
});

export const readRuntimeConfig = (environment: Environment): RuntimeConfig => {
  const dbPath = environment.DB_PATH;
  if (!dbPath) {
    throw new Error('DB_PATH environment variable is required');
  }

  const mockTwitch = environment.TWITCH_MOCK_MODE === 'true';
  if (mockTwitch) {
    return { dbPath: String(dbPath), logger: createLogger(), mockTwitch };
  }

  return {
    dbPath: String(dbPath),
    logger: createLogger(),
    mockTwitch,
    twitch: readTwitchConfig(environment),
  };
};

interface TwitchStack {
  chat: ChatSender;
  transport: EventSubTransport;
  bot: Bot;
  maintainToken: (signal: AbortSignal) => Promise<void>;
}

const createTwitchStack = (
  twitch: TwitchConfig,
  deps: {
    stats: ReturnType<typeof createStatsService>;
    blindBox: ReturnType<typeof createBlindBoxService>;
    overlay: OverlayBus;
    random: Random;
    clock: Clock;
    logger: Logger;
    catalog: ReturnType<typeof loadCatalog>;
    readiness: ReturnType<typeof createReadiness>;
  },
): TwitchStack => {
  const tokenProvider = createAppTokenProvider({
    clientId: twitch.clientId,
    clientSecret: twitch.clientSecret,
  });

  const helix = createHelixClient({
    clientId: twitch.clientId,
    tokenProvider,
  });

  const chatClient = createTwitchChatClient({ helix, logger: deps.logger });

  const chat = createTwitchChatSender({
    client: chatClient,
    broadcasterId: twitch.channelUserId,
    senderId: twitch.botUserId,
  });

  const bot = createBot({
    botUserID: twitch.botUserId,
    series: deps.catalog.series,
    stats: deps.stats,
    blindBox: deps.blindBox,
    overlay: deps.overlay,
    chat,
    clock: deps.clock,
    random: deps.random,
    logger: deps.logger,
  });

  const onNotification = createBotNotificationHandler(bot);

  const sessionManager = createConduitSessionManager({
    helix,
    clientId: twitch.clientId,
    botUserId: twitch.botUserId,
    channelUserId: twitch.channelUserId,
    logger: deps.logger,
  });

  const connector = createNodeEventSubConnector();

  const transport = createEventSubTransport({
    connector,
    sessionManager,
    onNotification,
    onReadinessChange: (state) => {
      deps.readiness.set('twitch', state.ready);
    },
    logger: deps.logger,
  });

  return { chat, transport, bot, maintainToken: (signal) => tokenProvider.maintain(signal) };
};

export const createApplicationRuntime = async (
  environment: Environment = process.env,
): Promise<ApplicationRuntime> => {
  const config = readRuntimeConfig(environment);
  let db: DatabaseConnection | undefined;
  let bus: OverlayBus | undefined;
  let twitchStack: TwitchStack | undefined;
  const twitchController = new AbortController();

  const readiness = createReadiness();
  setReadiness(readiness);

  try {
    const catalog = loadCatalog();
    readiness.set('catalog', true);
    config.logger.info('catalog loaded', {
      stats: catalog.stats.length,
      series: catalog.series.length,
    });

    db = openDatabase(config.dbPath);
    readiness.set('database', true);
    config.logger.info('database opened', { path: config.dbPath, state: db.state });

    const stats = createStatsService({
      db: db.db,
      definitions: catalog.stats,
    });

    const blindBox = createBlindBoxService({
      db: db.db,
      series: catalog.series,
    });

    bus = createOverlayBus(config.logger);
    const random = createRandom();
    const clock = createClock();

    let chat: ChatSender;

    if (config.twitch) {
      twitchStack = createTwitchStack(config.twitch, {
        stats,
        blindBox,
        overlay: bus,
        random,
        clock,
        logger: config.logger,
        catalog,
        readiness,
      });
      chat = twitchStack.chat;
    } else {
      chat = createLoggingChatStub(config.logger);
      readiness.set('twitch', true);
    }

    setServices({
      stats,
      blindBox,
      chat,
      overlay: bus,
      random,
      catalog,
    });

    config.logger.info('application services initialized', {
      chat: config.twitch ? 'twitch' : 'mock',
    });

    if (twitchStack) {
      await twitchStack.transport.start(twitchController.signal);
      twitchStack.maintainToken(twitchController.signal).catch((error: unknown) => {
        if (!twitchController.signal.aborted) {
          config.logger.error('token maintenance failed', { error });
        }
      });
      config.logger.info('twitch bot connected');
    }
  } catch (error) {
    twitchController.abort('startup failure');
    if (twitchStack) {
      await twitchStack.bot.stop('startup failure').catch(() => undefined);
      await twitchStack.transport.stop('startup failure').catch(() => undefined);
    }
    bus?.close();
    db?.close();
    throw error;
  }

  const capturedDb = db;
  const capturedBus = bus;
  const capturedTwitch = twitchStack;

  return {
    stop: async (reason) => {
      config.logger.info('application stopping', { reason });
      twitchController.abort(reason);
      if (capturedTwitch) {
        await capturedTwitch.bot.stop(reason);
        await capturedTwitch.transport.stop(reason);
      }
      capturedBus.close();
      capturedDb.checkpoint();
      capturedDb.close();
      config.logger.info('application stopped');
    },
  };
};
