-- +goose Up
CREATE TABLE viewer_activity (
  user_id        TEXT PRIMARY KEY,
  username       TEXT NOT NULL,
  last_active_at TEXT NOT NULL
);

CREATE INDEX viewer_activity_last_active_at_idx ON viewer_activity(last_active_at);
