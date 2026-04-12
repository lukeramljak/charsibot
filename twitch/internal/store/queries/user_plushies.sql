-- name: UpsertUserPlushie :exec
INSERT INTO user_plushies (user_id, username, series, key)
VALUES (?, ?, ?, ?)
ON CONFLICT(user_id, series, key) DO UPDATE SET
  username = excluded.username;

-- name: InsertUserPlushieIfNew :exec
INSERT OR IGNORE INTO user_plushies (user_id, username, series, key)
VALUES (?, ?, ?, ?);

-- name: LastChangeCount :one
SELECT changes() AS n;

-- name: GetCollectedPlushies :many
SELECT key FROM user_plushies
WHERE user_id = ? AND series = ?;

-- name: HasUserPlushie :one
SELECT EXISTS(
  SELECT 1 FROM user_plushies
  WHERE user_id = ? AND series = ? AND key = ?
) AS owned;

-- name: ResetUserPlushies :exec
DELETE FROM user_plushies
WHERE user_id = ? AND series = ?;

-- name: GetCompletedCollections :many
WITH completed AS (
  SELECT up.series, up.username
  FROM user_plushies up
  GROUP BY up.user_id, up.series
  HAVING COUNT(*) = (
    SELECT COUNT(*) FROM blind_box_plushies bp
    WHERE bp.series = up.series
  )
)
SELECT bbs.name AS series_name, GROUP_CONCAT(completed.username, ', ') AS usernames
FROM completed
JOIN blind_box_series bbs ON bbs.series = completed.series
GROUP BY completed.series
ORDER BY bbs.series;
