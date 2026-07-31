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

-- name: ListUsers :many
SELECT user_stats.user_id, CAST(MAX(user_stats.username) AS TEXT) AS username FROM user_stats
GROUP BY user_id
UNION
SELECT user_plushies.user_id, CAST(MAX(user_plushies.username) AS TEXT) AS username FROM user_plushies
GROUP BY user_id
ORDER BY username COLLATE NOCASE
;

-- name: GetUserByID :one
SELECT user_stats.user_id, user_stats.username FROM user_stats WHERE user_stats.user_id = sqlc.arg(user_id)
UNION
SELECT user_plushies.user_id, user_plushies.username FROM user_plushies WHERE user_plushies.user_id = sqlc.arg(user_id)
LIMIT 1;
