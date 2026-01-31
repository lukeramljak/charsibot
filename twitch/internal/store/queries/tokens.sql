-- name: GetTokens :one
SELECT access_token, refresh_token
FROM oauth_tokens
WHERE token_type = ?;

-- name: SaveTokens :exec
INSERT INTO oauth_tokens (
  token_type,
  access_token,
  refresh_token,
  updated_at
) VALUES (
  ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(token_type) DO UPDATE SET
  access_token = excluded.access_token,
  refresh_token = excluded.refresh_token,
  updated_at = CURRENT_TIMESTAMP;
