import type { StatDefinition } from '$lib/contracts/catalog';
import type { LeaderboardRow, UserStat, Viewer } from '$lib/contracts/viewer';
import { ApplicationError } from '$lib/server/application/errors';
import type { StatsService } from '$lib/server/application/ports';
import type { StatsRepository } from '$lib/server/db/stats.repository';
import type { ViewersRepository } from '$lib/server/db/viewers.repository';

export interface StatsServiceDependencies {
  repository: StatsRepository;
  viewers: ViewersRepository;
  definitions: readonly StatDefinition[];
}

export const createStatsService = ({
  repository,
  viewers,
  definitions: inputDefinitions,
}: StatsServiceDependencies): StatsService => {
  if (inputDefinitions.length === 0) {
    throw new Error('stat definitions must not be empty');
  }

  const definitions = [...inputDefinitions].sort((left, right) => left.sortOrder - right.sortOrder);

  const get = async (userID: string): Promise<UserStat[]> => {
    const values = new Map(repository.get(userID).map((value) => [value.statName, value.value]));

    return definitions.flatMap((definition) => {
      const value = values.get(definition.name);

      return value === undefined
        ? []
        : [
            {
              name: definition.name,
              shortName: definition.shortName,
              longName: definition.longName,
              value,
            },
          ];
    });
  };

  return {
    definitions,
    getOrCreate: async (userID, username) => {
      repository.initialize(userID, username, definitions);

      return get(userID);
    },
    get,
    leaderboard: async (): Promise<LeaderboardRow[]> => {
      const best = new Map<string, { username: string; value: number }>();
      for (const value of repository.all()) {
        if (!best.has(value.statName)) {
          best.set(value.statName, { username: value.username, value: value.value });
        }
      }

      return definitions.flatMap((definition) => {
        const row = best.get(definition.name);

        return row ? [{ emoji: definition.emoji, ...row }] : [];
      });
    },
    listViewers: async (): Promise<Viewer[]> =>
      viewers.list().map((viewer) => ({
        id: viewer.userID,
        username: viewer.username,
        ...(viewer.lastActiveAt ? { lastActiveAt: viewer.lastActiveAt } : {}),
      })),
    getViewer: async (userID): Promise<Viewer> => {
      const viewer = viewers.get(userID);

      if (!viewer) {
        throw new ApplicationError('not_found', 'user not found');
      }

      return {
        id: viewer.userID,
        username: viewer.username,
        ...(viewer.lastActiveAt ? { lastActiveAt: viewer.lastActiveAt } : {}),
      };
    },
    recordActivity: async (userID, username, at) => {
      viewers.recordActivity(userID, username, at);
    },
    deleteViewers: async (userIDs) => {
      viewers.deleteMany(userIDs);
    },
    adjust: async (userID, statName, amount) => {
      repository.adjust(userID, statName, amount);
    },
    set: async (userID, statName, value) => {
      repository.set(userID, statName, value);
    },
    reset: async (userID) => {
      repository.reset(userID, definitions);
    },
  };
};
