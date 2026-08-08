import { randomInt } from 'node:crypto';

import type { Logger, OverlayBus, Random } from '$lib/server/application/ports';
import { loadCatalog } from '$lib/server/catalog/load';
import { openDatabase, type DatabaseConnection } from '$lib/server/db/client';
import { createCollectionsRepository } from '$lib/server/db/collections.repository';
import { createStatsRepository } from '$lib/server/db/stats.repository';
import { createViewersRepository } from '$lib/server/db/viewers.repository';
import { createBlindBoxService } from '$lib/server/domain/blind-box/service';
import { createStatsService } from '$lib/server/domain/stats/service';
import { createOverlayBus } from '$lib/server/events/overlay-bus';
import { setServices } from '$lib/server/runtime/container';
import type { ApplicationRuntime } from '$lib/server/runtime/contracts';

export interface RuntimeConfig {
  dbPath: string;
  logger: Logger;
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

export const readRuntimeConfig = (environment: NodeJS.ProcessEnv): RuntimeConfig => {
  const dbPath = environment.DB_PATH;
  if (!dbPath) {
    throw new Error('DB_PATH environment variable is required');
  }

  return { dbPath, logger: createLogger() };
};

export const createApplicationRuntime = async (): Promise<ApplicationRuntime> => {
  const config = readRuntimeConfig(process.env);
  let db: DatabaseConnection | undefined;
  let bus: OverlayBus | undefined;

  try {
    const catalog = loadCatalog();
    config.logger.info('catalog loaded', {
      stats: catalog.stats.length,
      series: catalog.series.length,
    });

    db = openDatabase(config.dbPath);
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

    setServices({
      stats,
      blindBox,
      chat: {
        send: async (message) => {
          config.logger.warn('chat not connected, message dropped', {
            message: message.message,
          });
        },
      },
      overlay: bus,
      random: createRandom(),
      catalog,
    });

    config.logger.info('application services initialized');
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
