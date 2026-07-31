-- +goose Up
-- Catalog data is now loaded from versioned JSON. Rebuild viewer-state tables
-- without catalog foreign keys, then remove the obsolete database catalog.

CREATE TABLE user_stats_new (
  user_id   TEXT NOT NULL,
  username  TEXT NOT NULL,
  stat_name TEXT NOT NULL,
  value     INTEGER NOT NULL DEFAULT 3,
  PRIMARY KEY (user_id, stat_name)
);

INSERT INTO user_stats_new (user_id, username, stat_name, value)
SELECT user_id, username, stat_name, value FROM user_stats;

DROP TABLE user_stats;
ALTER TABLE user_stats_new RENAME TO user_stats;

CREATE TABLE user_plushies_new (
  user_id  TEXT NOT NULL,
  username TEXT NOT NULL,
  series   TEXT NOT NULL,
  key      TEXT NOT NULL,
  PRIMARY KEY (user_id, series, key)
);

INSERT INTO user_plushies_new (user_id, username, series, key)
SELECT user_id, username, series, key FROM user_plushies;

DROP TABLE user_plushies;
ALTER TABLE user_plushies_new RENAME TO user_plushies;

DROP TABLE blind_box_plushies;
DROP TABLE blind_box_series;
DROP TABLE stat_definitions;
