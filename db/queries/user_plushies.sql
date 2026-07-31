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

-- name: GetUserPlushieCounts :many
SELECT series, CAST(MAX(username) AS TEXT) AS username, COUNT(*) AS count
FROM user_plushies
GROUP BY series, user_id
ORDER BY series, username;
