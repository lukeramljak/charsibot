import type { BlindBoxSeries } from '$lib/contracts/catalog';
import type { ViewerCollection } from '$lib/contracts/collections';
import { ApplicationError } from '$lib/server/application/errors';
import type {
  BlindBoxService,
  CompletedCollection,
  GrantPlushieResult,
} from '$lib/server/application/ports';
import type { DatabaseConnection } from '$lib/server/db/client';
import { toSafeInteger } from '$lib/server/db/integer';
import { userPlushies } from '$lib/server/db/schema';
import { and, eq, sql } from 'drizzle-orm';

export interface BlindBoxServiceDependencies {
  db: DatabaseConnection['db'];
  series: readonly BlindBoxSeries[];
}

const compareStrings = (left: string, right: string): number => {
  if (left < right) {
    return -1;
  }

  if (left > right) {
    return 1;
  }

  return 0;
};

export const createBlindBoxService = ({
  db,
  series: inputSeries,
}: BlindBoxServiceDependencies): BlindBoxService => {
  if (inputSeries.length === 0) {
    throw new Error('blind-box series must not be empty');
  }

  const series = [...inputSeries].sort((left, right) => compareStrings(left.series, right.series));
  const byIdentifier = new Map(series.map((config) => [config.series, config]));

  const getConfig = (identifier: string): BlindBoxSeries => {
    const config = byIdentifier.get(identifier);

    if (!config) {
      throw new ApplicationError('invalid_input', 'unknown series');
    }

    return config;
  };

  const sortedCollection = (userID: string, identifier: string): string[] => {
    const config = getConfig(identifier);
    const collected = new Set(
      db
        .select({ key: userPlushies.key })
        .from(userPlushies)
        .where(and(eq(userPlushies.userID, userID), eq(userPlushies.series, identifier)))
        .all()
        .map((row) => row.key),
    );

    return config.plushies
      .filter((plushie) => collected.has(plushie.key))
      .map((plushie) => plushie.key);
  };

  return {
    getViewerCollections: async (userID): Promise<ViewerCollection[]> =>
      series.map((config) => ({ config, collected: sortedCollection(userID, config.series) })),
    getCollection: async (userID, identifier) => sortedCollection(userID, identifier),
    grant: async (userID, username, identifier, key): Promise<GrantPlushieResult> => {
      const config = getConfig(identifier);

      if (!config.plushies.some((plushie) => plushie.key === key)) {
        throw new ApplicationError('invalid_input', 'unknown series or plushie');
      }

      const result = db.transaction((tx) => {
        const insert = tx
          .insert(userPlushies)
          .values({ userID, username, series: identifier, key })
          .onConflictDoNothing()
          .run();
        const isNew = toSafeInteger(insert.changes, 'plushie insert change count') > 0;

        if (!isNew) {
          tx.update(userPlushies)
            .set({ username })
            .where(
              and(
                eq(userPlushies.userID, userID),
                eq(userPlushies.series, identifier),
                eq(userPlushies.key, key),
              ),
            )
            .run();
        }

        const collection = tx
          .select({ key: userPlushies.key })
          .from(userPlushies)
          .where(and(eq(userPlushies.userID, userID), eq(userPlushies.series, identifier)))
          .all()
          .map((row) => row.key);

        return { isNew, collection };
      });

      return {
        isNew: result.isNew,
        collection: config.plushies
          .filter((plushie) => result.collection.includes(plushie.key))
          .map((plushie) => plushie.key),
      };
    },
    remove: async (userID, identifier, key) => {
      const config = getConfig(identifier);

      if (!config.plushies.some((plushie) => plushie.key === key)) {
        throw new ApplicationError('invalid_input', 'unknown series or plushie');
      }

      db.delete(userPlushies)
        .where(
          and(
            eq(userPlushies.userID, userID),
            eq(userPlushies.series, identifier),
            eq(userPlushies.key, key),
          ),
        )
        .run();
    },
    reset: async (userID, identifier) => {
      getConfig(identifier);

      db.delete(userPlushies)
        .where(and(eq(userPlushies.userID, userID), eq(userPlushies.series, identifier)))
        .run();
    },
    completed: async (): Promise<CompletedCollection[]> => {
      const completed = new Map<string, string[]>();

      const counts = db.all<{
        user_id: string;
        username: string;
        series: string;
        count: bigint;
      }>(sql`
        SELECT user_id, CAST(MAX(username) AS TEXT) AS username, series, COUNT(*) AS count
        FROM user_plushies
        GROUP BY series, user_id
        ORDER BY series, username COLLATE NOCASE, username, user_id`);

      for (const count of counts) {
        const config = byIdentifier.get(count.series);

        if (
          !config ||
          toSafeInteger(count.count, `${count.series} collection count`) !== config.plushies.length
        ) {
          continue;
        }

        const usernames = completed.get(config.name) ?? [];

        usernames.push(count.username);
        completed.set(config.name, usernames);
      }

      return [...completed.entries()]
        .sort(([left], [right]) => compareStrings(left, right))
        .map(([seriesName, usernames]) => ({
          seriesName,
          usernames: usernames.sort(compareStrings).join(', '),
        }));
    },
  };
};
