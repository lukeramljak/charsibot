import { randomInt } from 'node:crypto';

import type { ChatSender, Logger, OverlayBus, Random } from '$lib/server/application/ports';
import { loadCatalog } from '$lib/server/catalog/load';
import { openDatabase, type DatabaseConnection } from '$lib/server/db/client';
import { createCollectionsRepository } from '$lib/server/db/collections.repository';
import { createStatsRepository } from '$lib/server/db/stats.repository';
import { createViewersRepository } from '$lib/server/db/viewers.repository';
import { createBlindBoxService } from '$lib/server/domain/blind-box/service';
import { createStatsService } from '$lib/server/domain/stats/service';
import { createOverlayBus } from '$lib/server/events/overlay-bus';
import { setReadiness, setServices } from '$lib/server/runtime/container';
import type { ApplicationRuntime } from '$lib/server/runtime/contracts';
import { createReadiness } from '$lib/server/runtime/readiness';
import { createTwitchChatSender } from '$lib/server/twitch/chat-sender';
import { createHelixClient, createTwitchChatClient } from '$lib/server/twitch/helix';
import { createAppTokenProvider } from '$lib/server/twitch/token';

export interface TwitchConfig {
  clientId: string;
  clientSecret: string;
  botUserId: string;
  channelUserId: string;
}

export interface RuntimeConfig {
  dbPath: string;
  logger: Logger;
  twitch?: TwitchConfig;
}

const createLogger = (): Logger => ({
  debug: (message, attributes) => console.debug(message, attributes ?? ''),
  info: (message, attributes) => console.info(message, attributes ?? ''),
  warn: (message, attributes) => console.warn(message, attributes ?? ''),
  error: (message, attributes) => console.error(message, attributes ?? ''),
});

const createRandom = (): Random => ({
  integer: (maxExclusive) => randomInt(maxExclusive),
});

const readTwitchConfig = (environment: NodeJS.ProcessEnv): TwitchConfig | undefined => {
  const clientId = environment.TWITCH_CLIENT_ID;
  const clientSecret = environment.TWITCH_CLIENT_SECRET;
  const botUserId = environment.TWITCH_BOT_USER_ID;
  const channelUserId = environment.TWITCH_CHANNEL_USER_ID;

  if (!clientId || !clientSecret || !botUserId || !channelUserId) {
    return undefined;
  }

  return { clientId, clientSecret, botUserId, channelUserId };
};

const createChatSender = (twitch: TwitchConfig, logger: Logger): ChatSender => {
  const tokenProvider = createAppTokenProvider({
    clientId: twitch.clientId,
    clientSecret: twitch.clientSecret,
  });

  const helix = createHelixClient({
    clientId: twitch.clientId,
    tokenProvider,
  });

  const client = createTwitchChatClient({ helix, logger });

  return createTwitchChatSender({
    client,
    broadcasterId: twitch.channelUserId,
    senderId: twitch.botUserId,
  });
};

const createLoggingChatStub = (logger: Logger): ChatSender => ({
  send: async (message) => {
    logger.warn('chat not connected, message dropped', { message: message.message });
  },
});

export const readRuntimeConfig = (environment: NodeJS.ProcessEnv): RuntimeConfig => {
  const dbPath = environment.DB_PATH;
  if (!dbPath) {
    throw new Error('DB_PATH environment variable is required');
  }

  return { dbPath, logger: createLogger(), twitch: readTwitchConfig(environment) };
};

export const createApplicationRuntime = async (): Promise<ApplicationRuntime> => {
  const config = readRuntimeConfig(process.env);
  let db: DatabaseConnection | undefined;
  let bus: OverlayBus | undefined;

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

    const statsRepo = createStatsRepository(db.database);
    const viewersRepo = createViewersRepository(db.database);
    const collectionsRepo = createCollectionsRepository(db.database);

    const stats = createStatsService({
      repository: statsRepo,
      viewers: viewersRepo,
      definitions: catalog.stats,
    });

    const blindBox = createBlindBoxService({
      repository: collectionsRepo,
      series: catalog.series,
    });

    bus = createOverlayBus(config.logger);

    const chat = config.twitch
      ? createChatSender(config.twitch, config.logger)
      : createLoggingChatStub(config.logger);

    setServices({
      stats,
      blindBox,
      chat,
      overlay: bus,
      random: createRandom(),
      catalog,
    });

    config.logger.info('application services initialized', {
      chat: config.twitch ? 'twitch' : 'stub',
    });
  } catch (error) {
    bus?.close();
    db?.close();
    throw error;
  }

  const capturedDb = db;
  const capturedBus = bus;

  return {
    stop: async (reason) => {
      config.logger.info('application stopping', { reason });
      capturedBus.close();
      capturedDb.checkpoint();
      capturedDb.close();
      config.logger.info('application stopped');
    },
  };
};
