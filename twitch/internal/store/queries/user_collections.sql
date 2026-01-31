-- name: GetUserCollectionRow :one
SELECT *
FROM user_collections
WHERE user_id = ?
  AND collection_type = ?;

-- name: AddReward1 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward1
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward1 = 1,
  username = excluded.username;

-- name: AddReward2 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward2
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward2 = 1,
  username = excluded.username;

-- name: AddReward3 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward3
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward3 = 1,
  username = excluded.username;

-- name: AddReward4 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward4
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward4 = 1,
  username = excluded.username;

-- name: AddReward5 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward5
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward5 = 1,
  username = excluded.username;

-- name: AddReward6 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward6
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward6 = 1,
  username = excluded.username;

-- name: AddReward7 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward7
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward7 = 1,
  username = excluded.username;

-- name: AddReward8 :exec
INSERT INTO user_collections (
  user_id,
  username,
  collection_type,
  reward8
) VALUES (
  ?, ?, ?, 1
)
ON CONFLICT(user_id, collection_type) DO UPDATE SET
  reward8 = 1,
  username = excluded.username;

-- name: ResetUserCollection :exec
UPDATE user_collections
SET
  reward1 = 0,
  reward2 = 0,
  reward3 = 0,
  reward4 = 0,
  reward5 = 0,
  reward6 = 0,
  reward7 = 0,
  reward8 = 0
WHERE user_id = ?
  AND collection_type = ?;

-- name: GetCompletedCollections :many
SELECT
  collection_type,
  group_concat(username, ',') AS usernames_csv
FROM user_collections
WHERE
  reward1 = 1 AND
  reward2 = 1 AND
  reward3 = 1 AND
  reward4 = 1 AND
  reward5 = 1 AND
  reward6 = 1 AND
  reward7 = 1 AND
  reward8 = 1
GROUP BY collection_type;
