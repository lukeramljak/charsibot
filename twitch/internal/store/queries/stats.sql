-- name: UpsertStatsUser :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO UPDATE SET
  username = excluded.username;

-- name: GetStats :one
SELECT *
FROM stats
WHERE id = ?;

-- name: ModifyStrength :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  strength = strength + ?
WHERE id = ?;

-- name: ModifyIntelligence :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  intelligence = intelligence + ?
WHERE id = ?;

-- name: ModifyCharisma :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  charisma = charisma + ?
WHERE id = ?;

-- name: ModifyLuck :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  luck = luck + ?
WHERE id = ?;

-- name: ModifyDexterity :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  dexterity = dexterity + ?
WHERE id = ?;

-- name: ModifyPenis :exec
INSERT INTO stats (id, username)
VALUES (?, ?)
ON CONFLICT(id) DO NOTHING;

UPDATE stats
SET
  username = ?,
  penis = penis + ?
WHERE id = ?;
