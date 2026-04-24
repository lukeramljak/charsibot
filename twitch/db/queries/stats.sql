-- name: GetUserStats :many
SELECT sd.name, sd.short_name, sd.long_name, sv.value
FROM user_stats sv
JOIN stat_definitions sd ON sv.stat_name = sd.name
WHERE sv.user_id = ?
ORDER BY sd.sort_order;

-- name: EnsureUserStats :exec
INSERT INTO user_stats (user_id, username, stat_name, value)
SELECT ?, ?, sd.name, sd.default_value FROM stat_definitions sd
WHERE sd.name NOT IN (
  SELECT us.stat_name FROM user_stats us WHERE us.user_id = ?
);

-- name: UpdateUsername :exec
UPDATE user_stats SET username = ? WHERE user_id = ?;

-- name: ModifyStatValue :exec
UPDATE user_stats SET value = value + ?
WHERE user_id = ? AND stat_name = ?;

-- name: SetStatValue :exec
UPDATE user_stats SET value = ?
WHERE user_id = ? AND stat_name = ?;

-- name: GetStatLeaderboard :many
SELECT sd.emoji, sv.username, CAST(MAX(sv.value) AS INTEGER) AS value
FROM user_stats sv
JOIN stat_definitions sd ON sv.stat_name = sd.name
GROUP BY sv.stat_name
ORDER BY sd.sort_order;

-- name: GetStatDefinitions :many
SELECT * FROM stat_definitions ORDER BY sort_order;

-- name: GetRandomStatDefinition :one
SELECT * FROM stat_definitions ORDER BY RANDOM() LIMIT 1;
