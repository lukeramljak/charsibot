import type BetterSqlite3 from 'better-sqlite3';

export interface ViewerRecord {
  userID: string;
  username: string;
  lastActiveAt?: string;
}

export interface ViewersRepository {
  recordActivity: (userID: string, username: string, at: Date) => void;
  list: () => ViewerRecord[];
  get: (userID: string) => ViewerRecord | undefined;
  deleteMany: (userIDs: readonly string[]) => void;
}

interface ViewerRow {
  user_id: string;
  username: string;
  last_active_at: string | null;
}

const viewerUnion = `
  SELECT user_id, CAST(MAX(username) AS TEXT) AS username, MAX(last_active_at) AS last_active_at
  FROM (
    SELECT user_id, username, last_active_at FROM viewer_activity
    UNION ALL SELECT user_id, username, NULL FROM user_stats
    UNION ALL SELECT user_id, username, NULL FROM user_plushies
  )`;

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

const toViewerRecord = (row: ViewerRow): ViewerRecord => ({
  userID: row.user_id,
  username: row.username,
  ...(row.last_active_at === null ? {} : { lastActiveAt: validateTimestamp(row.last_active_at) }),
});

export const createViewersRepository = (database: BetterSqlite3.Database): ViewersRepository => {
  const record = database.prepare(`
    INSERT INTO viewer_activity (user_id, username, last_active_at)
    VALUES (?, ?, ?)
    ON CONFLICT(user_id) DO UPDATE SET
      username = excluded.username,
      last_active_at = excluded.last_active_at`);
  const list = database.prepare(`${viewerUnion}
    GROUP BY user_id
    ORDER BY username COLLATE NOCASE, username, user_id`);
  const get = database.prepare(`${viewerUnion}
    WHERE user_id = ?
    GROUP BY user_id`);
  const deleteActivity = database.prepare('DELETE FROM viewer_activity WHERE user_id = ?');
  const deleteStats = database.prepare('DELETE FROM user_stats WHERE user_id = ?');
  const deletePlushies = database.prepare('DELETE FROM user_plushies WHERE user_id = ?');

  const deleteTransaction = database.transaction((userIDs: readonly string[]) => {
    for (const userID of userIDs) {
      deleteActivity.run(userID);
      deleteStats.run(userID);
      deletePlushies.run(userID);
    }
  });

  return {
    recordActivity: (userID, username, at) => {
      if (Number.isNaN(at.getTime())) {
        throw new TypeError('activity timestamp must be valid');
      }

      record.run(userID, username, at.toISOString());
    },
    list: () => (list.all() as ViewerRow[]).map(toViewerRecord),
    get: (userID) => {
      const row = get.get(userID) as ViewerRow | undefined;

      return row ? toViewerRecord(row) : undefined;
    },
    deleteMany: (userIDs) => deleteTransaction(userIDs),
  };
};
