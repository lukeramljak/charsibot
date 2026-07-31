-- name: GetUserStatValues :many
SELECT stat_name, value
FROM user_stats
WHERE user_id = ?;

-- name: EnsureUserStat :exec
INSERT OR IGNORE INTO user_stats (user_id, username, stat_name, value)
VALUES (?, ?, ?, ?);

-- name: UpdateUsername :exec
UPDATE user_stats SET username = ? WHERE user_id = ?;

-- name: ModifyStatValue :exec
UPDATE user_stats SET value = value + ?
WHERE user_id = ? AND stat_name = ?;

-- name: SetStatValue :exec
UPDATE user_stats SET value = ?
WHERE user_id = ? AND stat_name = ?;

-- name: GetAllUserStatValues :many
SELECT username, stat_name, value
FROM user_stats;
