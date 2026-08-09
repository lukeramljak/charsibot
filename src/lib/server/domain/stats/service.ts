import type { StatDefinition } from '$lib/contracts/catalog';
import type { LeaderboardRow, UserStat, Viewer } from '$lib/contracts/viewer';
import { ApplicationError } from '$lib/server/application/errors';
import type { StatsService } from '$lib/server/application/ports';
import type { DatabaseConnection } from '$lib/server/db/client';
import { requireSafeInteger, toSafeInteger } from '$lib/server/db/integer';
import { userPlushies, userStats, viewerActivity } from '$lib/server/db/schema';
import { desc, eq, sql } from 'drizzle-orm';

export interface StatsServiceDependencies {
  db: DatabaseConnection['db'];
  definitions: readonly StatDefinition[];
}

interface ViewerRow {
  user_id: string;
  username: string;
  last_active_at: string | null;
}

const validateTimestamp = (value: string): string => {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-](\d{2}):(\d{2}))$/.exec(
      value,
    );
  if (!match) {
    throw new Error(`invalid viewer activity timestamp ${JSON.stringify(value)}`);
  }

  const [, yearText, monthText, dayText, hourText, minuteText, secondText, zoneHour, zoneMinute] =
    match;
  const [year, month, day, hour, minute, second] = [
    yearText,
    monthText,
    dayText,
    hourText,
    minuteText,
    secondText,
  ].map(Number);
  const daysInMonth = new Date(Date.UTC(year, month, 0)).getUTCDate();

  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > daysInMonth ||
    hour > 23 ||
    minute > 59 ||
    second > 59 ||
    (zoneHour !== undefined && Number(zoneHour) > 23) ||
    (zoneMinute !== undefined && Number(zoneMinute) > 59) ||
    Number.isNaN(Date.parse(value))
  ) {
    throw new Error(`invalid viewer activity timestamp ${JSON.stringify(value)}`);
  }

  return value;
};

const toViewer = (row: ViewerRow): Viewer => ({
  id: row.user_id,
  username: row.username,
  ...(row.last_active_at === null ? {} : { lastActiveAt: validateTimestamp(row.last_active_at) }),
});

export const createStatsService = ({
  db,
  definitions: inputDefinitions,
}: StatsServiceDependencies): StatsService => {
  if (inputDefinitions.length === 0) {
    throw new Error('stat definitions must not be empty');
  }

  const definitions = [...inputDefinitions].sort((left, right) => left.sortOrder - right.sortOrder);

  const getViewer = (userID: string): Viewer | undefined => {
    const row = db.get<ViewerRow | undefined>(sql`
      SELECT user_id, CAST(MAX(username) AS TEXT) AS username, MAX(last_active_at) AS last_active_at
      FROM (
        SELECT user_id, username, last_active_at FROM viewer_activity
        UNION ALL SELECT user_id, username, NULL FROM user_stats
        UNION ALL SELECT user_id, username, NULL FROM user_plushies
      )
      WHERE user_id = ${userID}
      GROUP BY user_id`);

    return row ? toViewer(row) : undefined;
  };

  const get = async (userID: string): Promise<UserStat[]> => {
    const values = new Map(
      db
        .select({ statName: userStats.statName, value: userStats.value })
        .from(userStats)
        .where(eq(userStats.userID, userID))
        .all()
        .map((value) => [value.statName, toSafeInteger(value.value, `${value.statName} value`)]),
    );

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
      const values = definitions.map((definition) => ({
        statName: definition.name,
        value: requireSafeInteger(definition.defaultValue, `${definition.name} default value`),
      }));

      db.transaction((tx) => {
        for (const value of values) {
          tx.insert(userStats)
            .values({ userID, username, statName: value.statName, value: value.value })
            .onConflictDoNothing()
            .run();
        }

        tx.update(userStats).set({ username }).where(eq(userStats.userID, userID)).run();
      });

      return get(userID);
    },
    get,
    leaderboard: async (): Promise<LeaderboardRow[]> => {
      const best = new Map<string, { username: string; value: number }>();
      const values = db
        .select({
          userID: userStats.userID,
          username: userStats.username,
          statName: userStats.statName,
          value: userStats.value,
        })
        .from(userStats)
        .orderBy(
          userStats.statName,
          desc(userStats.value),
          sql`${userStats.username} COLLATE NOCASE`,
          userStats.username,
          userStats.userID,
        )
        .all();

      for (const value of values) {
        if (!best.has(value.statName)) {
          best.set(value.statName, {
            username: value.username,
            value: toSafeInteger(value.value, `${value.statName} leaderboard value`),
          });
        }
      }

      return definitions.flatMap((definition) => {
        const row = best.get(definition.name);

        return row ? [{ emoji: definition.emoji, ...row }] : [];
      });
    },
    listViewers: async (): Promise<Viewer[]> =>
      db
        .all<ViewerRow>(
          sql`
          SELECT user_id, CAST(MAX(username) AS TEXT) AS username, MAX(last_active_at) AS last_active_at
          FROM (
            SELECT user_id, username, last_active_at FROM viewer_activity
            UNION ALL SELECT user_id, username, NULL FROM user_stats
            UNION ALL SELECT user_id, username, NULL FROM user_plushies
          )
          GROUP BY user_id
          ORDER BY username COLLATE NOCASE, username, user_id`,
        )
        .map(toViewer),
    getViewer: async (userID): Promise<Viewer> => {
      const viewer = getViewer(userID);

      if (!viewer) {
        throw new ApplicationError('not_found', 'user not found');
      }

      return viewer;
    },
    recordActivity: async (userID, username, at) => {
      if (Number.isNaN(at.getTime())) {
        throw new TypeError('activity timestamp must be valid');
      }

      db.insert(viewerActivity)
        .values({ userID, username, lastActiveAt: at.toISOString() })
        .onConflictDoUpdate({
          target: viewerActivity.userID,
          set: { username, lastActiveAt: at.toISOString() },
        })
        .run();
    },
    deleteViewers: async (userIDs) => {
      db.transaction((tx) => {
        for (const userID of userIDs) {
          tx.delete(viewerActivity).where(eq(viewerActivity.userID, userID)).run();
          tx.delete(userStats).where(eq(userStats.userID, userID)).run();
          tx.delete(userPlushies).where(eq(userPlushies.userID, userID)).run();
        }
      });
    },
    adjust: async (userID, statName, amount) => {
      db.update(userStats)
        .set({ value: sql`${userStats.value} + ${requireSafeInteger(amount, 'stat adjustment')}` })
        .where(sql`${userStats.userID} = ${userID} AND ${userStats.statName} = ${statName}`)
        .run();
    },
    set: async (userID, statName, value) => {
      db.update(userStats)
        .set({ value: requireSafeInteger(value, 'stat value') })
        .where(sql`${userStats.userID} = ${userID} AND ${userStats.statName} = ${statName}`)
        .run();
    },
    reset: async (userID) => {
      const values = definitions.map((definition) => ({
        statName: definition.name,
        value: requireSafeInteger(definition.defaultValue, `${definition.name} default value`),
      }));

      db.transaction((tx) => {
        for (const value of values) {
          tx.update(userStats)
            .set({ value: value.value })
            .where(
              sql`${userStats.userID} = ${userID} AND ${userStats.statName} = ${value.statName}`,
            )
            .run();
        }
      });
    },
  };
};
