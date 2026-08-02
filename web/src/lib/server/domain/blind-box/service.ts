import type { BlindBoxSeries } from '$lib/contracts/catalog';
import type { ViewerCollection } from '$lib/contracts/collections';
import { ApplicationError } from '$lib/server/application/errors';
import type {
  BlindBoxService,
  CompletedCollection,
  GrantPlushieResult,
} from '$lib/server/application/ports';
import type { CollectionsRepository } from '$lib/server/db/collections.repository';

export interface BlindBoxServiceDependencies {
  repository: CollectionsRepository;
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
  repository,
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
    const collected = new Set(repository.get(userID, identifier));

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

      const result = repository.grant(userID, username, identifier, key);

      return {
        isNew: result.isNew,
        collection: sortedCollection(userID, identifier),
      };
    },
    remove: async (userID, identifier, key) => {
      const config = getConfig(identifier);

      if (!config.plushies.some((plushie) => plushie.key === key)) {
        throw new ApplicationError('invalid_input', 'unknown series or plushie');
      }

      repository.remove(userID, identifier, key);
    },
    reset: async (userID, identifier) => {
      getConfig(identifier);

      repository.reset(userID, identifier);
    },
    completed: async (): Promise<CompletedCollection[]> => {
      const completed = new Map<string, string[]>();

      for (const count of repository.counts()) {
        const config = byIdentifier.get(count.series);

        if (!config || count.count !== config.plushies.length) {
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
