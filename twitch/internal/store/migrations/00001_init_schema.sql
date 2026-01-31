-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS oauth_tokens (
  token_type     TEXT PRIMARY KEY,
  access_token   TEXT NOT NULL,
  refresh_token  TEXT NOT NULL,
  updated_at     NUMERIC DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stats (
  id            TEXT PRIMARY KEY,
  username      TEXT NOT NULL,
  strength      INTEGER NOT NULL DEFAULT 3,
  intelligence  INTEGER NOT NULL DEFAULT 3,
  charisma      INTEGER NOT NULL DEFAULT 3,
  luck          INTEGER NOT NULL DEFAULT 3,
  dexterity     INTEGER NOT NULL DEFAULT 3,
  penis         INTEGER NOT NULL DEFAULT 3
);

CREATE TABLE IF NOT EXISTS user_collections (
  user_id         TEXT,
  username        TEXT NOT NULL,
  collection_type TEXT,
  reward1         INTEGER DEFAULT 0,
  reward2         INTEGER DEFAULT 0,
  reward3         INTEGER DEFAULT 0,
  reward4         INTEGER DEFAULT 0,
  reward5         INTEGER DEFAULT 0,
  reward6         INTEGER DEFAULT 0,
  reward7         INTEGER DEFAULT 0,
  reward8         INTEGER DEFAULT 0,
  CONSTRAINT user_collections_user_id_collection_type_pk
    PRIMARY KEY (user_id, collection_type)
);
-- +goose StatementEnd
