-- +goose Up
-- +goose StatementBegin

-- Replace the generic oauth_tokens table with a single bot_token table.
-- The bot is the only account whose token needs persisting; the streamer
-- just needs to have authorised the app once (checked by Twitch), and the
-- app access token is fetched fresh on each startup via client credentials.
DROP TABLE oauth_tokens;

-- Blind box series config (includes display fields to fully replace configs.ts)
-- `series` is the canonical identifier (e.g. 'coobubu', 'xmas') and serves as the PK.
CREATE TABLE blind_box_series (
  series           TEXT PRIMARY KEY,
  redemption_title TEXT NOT NULL,
  name             TEXT NOT NULL DEFAULT '',
  reveal_sound     TEXT NOT NULL DEFAULT '',
  box_front_face   TEXT NOT NULL DEFAULT '',
  box_side_face    TEXT NOT NULL DEFAULT '',
  display_color    TEXT NOT NULL DEFAULT '',
  text_color       TEXT NOT NULL DEFAULT ''
);

-- Plushies per series (includes display fields to fully replace configs.ts)
-- `key` is the lowercase command-safe identifier (e.g. 'cutey', 'secret')
-- `name` is the human-readable display name (e.g. 'Cutey', 'Secret')
CREATE TABLE blind_box_plushies (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  series      TEXT NOT NULL REFERENCES blind_box_series(series),
  key         TEXT NOT NULL,
  sort_order  INTEGER NOT NULL DEFAULT 0,
  weight      INTEGER NOT NULL DEFAULT 1,
  name        TEXT NOT NULL DEFAULT '',
  image       TEXT NOT NULL DEFAULT '',
  empty_image TEXT NOT NULL DEFAULT '',
  UNIQUE(series, key)
);

-- Seed blind_box_series from hardcoded config.go and configs.ts values
INSERT INTO blind_box_series (series, redemption_title, name, reveal_sound, box_front_face, box_side_face, display_color, text_color) VALUES
  ('coobubu',    'Cooper Series Blind Box',       'Coobubus',    '/blind-box/coobubu/reveal.mp3',    '/blind-box/coobubu/box-front.png',    '/blind-box/coobubu/box-side.png',    '#ff8c82', '#ffffff'),
  ('olliepop',   'Ollie Series Blind Box',        'Olliepops',   '/blind-box/olliepops/reveal.mp3',   '/blind-box/olliepops/box-front.png',  '/blind-box/olliepops/box-side.png',  '#ff8c82', '#ffffff'),
  ('xmas',       'Christmas Series Blind Box',    'Lil Helpers', '/blind-box/xmas/reveal.mp3',   '/blind-box/xmas/box-front.png',  '/blind-box/xmas/box-side.png',  '#9e0000', '#ffffff'),
  ('valentines', 'Valentine''s Series Blind Box', 'Valentines',  '/blind-box/valentines/reveal.mp3',  '/blind-box/valentines/box-front.png', '/blind-box/valentines/box-side.png', '#ffa7c3', '#ffffff'),
  ('easter',     'Easter Series Blind Box',       'Chocopups',   '/blind-box/easter/reveal.mp3',      '/blind-box/easter/box-front.png',     '/blind-box/easter/box-side.png',     '#ff81a9', '#ffffff');

-- Seed blind_box_plushies from hardcoded config.go weights and configs.ts display data
-- coobubu: rewards 1-7 weight 12, secret weight 1
INSERT INTO blind_box_plushies (series, key, sort_order, weight, name, image, empty_image) VALUES
  ('coobubu', 'cutey',     1, 12, 'Cutey',     '/blind-box/coobubu/cutey.png',     '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'blueberry', 2, 12, 'Blueberry', '/blind-box/coobubu/blueberry.png', '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'lemony',    3, 12, 'Lemony',    '/blind-box/coobubu/lemony.png',    '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'bibi',      4, 12, 'Bibi',      '/blind-box/coobubu/bibi.png',      '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'pinky',     5, 12, 'Pinky',     '/blind-box/coobubu/pinky.png',     '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'minty',     6, 12, 'Minty',     '/blind-box/coobubu/minty.png',     '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'cherry',    7, 12, 'Cherry',    '/blind-box/coobubu/cherry.png',    '/blind-box/coobubu/empty-slot.png'),
  ('coobubu', 'secret',    8,  1, 'Secret',    '/blind-box/coobubu/secret.png',    '/blind-box/coobubu/empty-slot.png');

-- olliepop: same weights as coobubu
INSERT INTO blind_box_plushies (series, key, sort_order, weight, name, image, empty_image) VALUES
  ('olliepop', 'berry',     1, 12, 'Berry',     '/blind-box/olliepops/berry.png',     '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'tangerine', 2, 12, 'Tangerine', '/blind-box/olliepops/tangerine.png', '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'bibble',    3, 12, 'Bibble',    '/blind-box/olliepops/bibble.png',    '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'kiwi',      4, 12, 'Kiwi',      '/blind-box/olliepops/kiwi.png',      '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'crunchy',   5, 12, 'Crunchy',   '/blind-box/olliepops/crunchy.png',   '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'caramel',   6, 12, 'Caramel',   '/blind-box/olliepops/caramel.png',   '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'grape',     7, 12, 'Grape',     '/blind-box/olliepops/grape.png',     '/blind-box/olliepops/empty-slot.png'),
  ('olliepop', 'secret',    8,  1, 'Secret',    '/blind-box/olliepops/secret.png',    '/blind-box/olliepops/empty-slot.png');

-- xmas: rewards 1-7 weight 3, secret weight 1
INSERT INTO blind_box_plushies (series, key, sort_order, weight, name, image, empty_image) VALUES
  ('xmas', 'snowy',  1, 3, 'Snowy',  '/blind-box/xmas/snowy.png',  '/blind-box/xmas/snowy-blank.png'),
  ('xmas', 'piney',  2, 3, 'Piney',  '/blind-box/xmas/piney.png',  '/blind-box/xmas/piney-blank.png'),
  ('xmas', 'starry', 3, 3, 'Starry', '/blind-box/xmas/starry.png', '/blind-box/xmas/starry-blank.png'),
  ('xmas', 'socky',  4, 3, 'Socky',  '/blind-box/xmas/socky.png',  '/blind-box/xmas/socky-blank.png'),
  ('xmas', 'gingey', 5, 3, 'Gingey', '/blind-box/xmas/gingey.png', '/blind-box/xmas/gingey-blank.png'),
  ('xmas', 'nicky',  6, 3, 'Nicky',  '/blind-box/xmas/nicky.png',  '/blind-box/xmas/nicky-blank.png'),
  ('xmas', 'dancey', 7, 3, 'Dancey', '/blind-box/xmas/dancey.png', '/blind-box/xmas/dancey-blank.png'),
  ('xmas', 'secret', 8, 1, 'Secret', '/blind-box/xmas/secret.png', '/blind-box/xmas/secret-blank.png');

-- valentines: same weights as xmas
INSERT INTO blind_box_plushies (series, key, sort_order, weight, name, image, empty_image) VALUES
  ('valentines', 'choccy',  1, 3, 'Choccy',  '/blind-box/valentines/choccy.png',  '/blind-box/valentines/choccy-blank.png'),
  ('valentines', 'cupie',   2, 3, 'Cupie',   '/blind-box/valentines/cupie.png',   '/blind-box/valentines/cupie-blank.png'),
  ('valentines', 'bachie',  3, 3, 'Bachie',  '/blind-box/valentines/bachie.png',  '/blind-box/valentines/bachie-blank.png'),
  ('valentines', 'drinky',  4, 3, 'Drinky',  '/blind-box/valentines/drinky.png',  '/blind-box/valentines/drinky-blank.png'),
  ('valentines', 'sherbie', 5, 3, 'Sherbie', '/blind-box/valentines/sherbie.png', '/blind-box/valentines/sherbie-blank.png'),
  ('valentines', 'lovey',   6, 3, 'Lovey',   '/blind-box/valentines/lovey.png',   '/blind-box/valentines/lovey-blank.png'),
  ('valentines', 'bunchie', 7, 3, 'Bunchie', '/blind-box/valentines/bunchie.png', '/blind-box/valentines/bunchie-blank.png'),
  ('valentines', 'secret',  8, 1, 'Secret',  '/blind-box/valentines/secret.png',  '/blind-box/valentines/secret-blank.png');

-- easter: rewards 1-7 weight 5, secret weight 1
INSERT INTO blind_box_plushies (series, key, sort_order, weight, name, image, empty_image) VALUES
  ('easter', 'bunny',  1, 5, 'Bunny',  '/blind-box/easter/bunny.png',  '/blind-box/easter/bunny-blank.png'),
  ('easter', 'chikky', 2, 5, 'Chikky', '/blind-box/easter/chikky.png', '/blind-box/easter/chikky-blank.png'),
  ('easter', 'nesty',  3, 5, 'Nesty',  '/blind-box/easter/nesty.png',  '/blind-box/easter/nesty-blank.png'),
  ('easter', 'lamby',  4, 5, 'Lamby',  '/blind-box/easter/lamby.png',  '/blind-box/easter/lamby-blank.png'),
  ('easter', 'choccy', 5, 5, 'Choccy', '/blind-box/easter/choccy.png', '/blind-box/easter/choccy-blank.png'),
  ('easter', 'eggy',   6, 5, 'Eggy',   '/blind-box/easter/eggy.png',   '/blind-box/easter/eggy-blank.png'),
  ('easter', 'flowey', 7, 5, 'Flowey', '/blind-box/easter/flowey.png', '/blind-box/easter/flowey-blank.png'),
  ('easter', 'secret', 8, 1, 'Secret', '/blind-box/easter/secret.png', '/blind-box/easter/secret-blank.png');

-- Stat definitions (replaces hardcoded columns and statList in redemptions.go)
-- short_name is used for overlay/leaderboard labels (abbreviated)
-- long_name is used for chat messages (e.g. potion redemption)
CREATE TABLE stat_definitions (
  name        TEXT PRIMARY KEY,
  short_name  TEXT NOT NULL,
  long_name   TEXT NOT NULL,
  default_value INTEGER NOT NULL DEFAULT 3,
  sort_order  INTEGER NOT NULL
);

INSERT INTO stat_definitions (name, short_name, long_name, default_value, sort_order) VALUES
  ('strength',     'STR',   'Strength',     3, 1),
  ('intelligence', 'INT',   'Intelligence', 3, 2),
  ('charisma',     'CHA',   'Charisma',     3, 3),
  ('luck',         'LUCK',  'Luck',         3, 4),
  ('dexterity',    'DEX',   'Dexterity',    3, 5),
  ('penis',        'PENIS', 'Penis',        3, 6);

-- Normalised stat values (replaces fixed-column stats table)
CREATE TABLE user_stats (
  user_id   TEXT NOT NULL,
  username  TEXT NOT NULL,
  stat_name TEXT NOT NULL REFERENCES stat_definitions(name),
  value     INTEGER NOT NULL DEFAULT 3,
  PRIMARY KEY (user_id, stat_name)
);

-- Migrate existing stats data into user_stats
INSERT INTO user_stats (user_id, username, stat_name, value)
SELECT id, username, 'strength',     strength     FROM stats UNION ALL
SELECT id, username, 'intelligence', intelligence FROM stats UNION ALL
SELECT id, username, 'charisma',     charisma     FROM stats UNION ALL
SELECT id, username, 'luck',         luck         FROM stats UNION ALL
SELECT id, username, 'dexterity',    dexterity    FROM stats UNION ALL
SELECT id, username, 'penis',        penis        FROM stats;

-- New normalised user_plushies table
-- Row presence = collected. No `collected` flag needed - reset deletes rows.
-- `key` matches blind_box_plushies.key for the given series.
-- key is not foreign-keyed because blind_box_plushies has a composite unique
-- constraint on (series, key), not key alone. Valid keys are enforced by
-- the application (loaded from DB at startup via GetPlushiesForSeries).
CREATE TABLE user_plushies (
  user_id  TEXT NOT NULL,
  username TEXT NOT NULL,
  series   TEXT NOT NULL REFERENCES blind_box_series(series),
  key      TEXT NOT NULL,
  PRIMARY KEY (user_id, series, key)
);

-- Migrate existing user_collections data
-- Maps old reward1-8 columns to their plushie keys per series.
-- Only rows where the reward column = 1 are migrated (row presence = collected).
-- Note: old collection_type 'christmas' maps to series 'xmas'.
INSERT INTO user_plushies (user_id, username, series, key)
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'cutey'
    WHEN 'olliepop'   THEN 'berry'
    WHEN 'christmas'  THEN 'snowy'
    WHEN 'valentines' THEN 'choccy'
    WHEN 'easter'     THEN 'bunny'
  END
FROM user_collections WHERE reward1 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'blueberry'
    WHEN 'olliepop'   THEN 'tangerine'
    WHEN 'christmas'  THEN 'piney'
    WHEN 'valentines' THEN 'cupie'
    WHEN 'easter'     THEN 'chikky'
  END
FROM user_collections WHERE reward2 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'lemony'
    WHEN 'olliepop'   THEN 'bibble'
    WHEN 'christmas'  THEN 'starry'
    WHEN 'valentines' THEN 'bachie'
    WHEN 'easter'     THEN 'nesty'
  END
FROM user_collections WHERE reward3 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'bibi'
    WHEN 'olliepop'   THEN 'kiwi'
    WHEN 'christmas'  THEN 'socky'
    WHEN 'valentines' THEN 'drinky'
    WHEN 'easter'     THEN 'lamby'
  END
FROM user_collections WHERE reward4 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'pinky'
    WHEN 'olliepop'   THEN 'crunchy'
    WHEN 'christmas'  THEN 'gingey'
    WHEN 'valentines' THEN 'sherbie'
    WHEN 'easter'     THEN 'choccy'
  END
FROM user_collections WHERE reward5 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'minty'
    WHEN 'olliepop'   THEN 'caramel'
    WHEN 'christmas'  THEN 'nicky'
    WHEN 'valentines' THEN 'lovey'
    WHEN 'easter'     THEN 'eggy'
  END
FROM user_collections WHERE reward6 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  CASE collection_type
    WHEN 'coobubu'    THEN 'cherry'
    WHEN 'olliepop'   THEN 'grape'
    WHEN 'christmas'  THEN 'dancey'
    WHEN 'valentines' THEN 'bunchie'
    WHEN 'easter'     THEN 'flowey'
  END
FROM user_collections WHERE reward7 = 1
UNION ALL
SELECT user_id, username,
  CASE collection_type WHEN 'christmas' THEN 'xmas' ELSE collection_type END,
  'secret'
FROM user_collections WHERE reward8 = 1;

-- Drop old tables
DROP TABLE user_collections;
DROP TABLE stats;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore stats table
CREATE TABLE stats (
  id            TEXT PRIMARY KEY,
  username      TEXT NOT NULL,
  strength      INTEGER NOT NULL DEFAULT 3,
  intelligence  INTEGER NOT NULL DEFAULT 3,
  charisma      INTEGER NOT NULL DEFAULT 3,
  luck          INTEGER NOT NULL DEFAULT 3,
  dexterity     INTEGER NOT NULL DEFAULT 3,
  penis         INTEGER NOT NULL DEFAULT 3
);

INSERT INTO stats (id, username, strength, intelligence, charisma, luck, dexterity, penis)
SELECT
  sv.user_id,
  sv.username,
  MAX(CASE WHEN sv.stat_name = 'strength'     THEN sv.value ELSE 3 END),
  MAX(CASE WHEN sv.stat_name = 'intelligence' THEN sv.value ELSE 3 END),
  MAX(CASE WHEN sv.stat_name = 'charisma'     THEN sv.value ELSE 3 END),
  MAX(CASE WHEN sv.stat_name = 'luck'         THEN sv.value ELSE 3 END),
  MAX(CASE WHEN sv.stat_name = 'dexterity'    THEN sv.value ELSE 3 END),
  MAX(CASE WHEN sv.stat_name = 'penis'        THEN sv.value ELSE 3 END)
FROM user_stats sv
GROUP BY sv.user_id, sv.username;

-- Restore user_collections with fixed columns
-- Maps plushie keys back to reward1-8 columns per series.
-- Note: series 'xmas' maps back to collection_type 'christmas'.
CREATE TABLE user_collections (
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

INSERT INTO user_collections (user_id, username, collection_type,
  reward1, reward2, reward3, reward4, reward5, reward6, reward7, reward8)
SELECT
  user_id, username,
  CASE series WHEN 'xmas' THEN 'christmas' ELSE series END,
  MAX(CASE WHEN series = 'coobubu'    AND key = 'cutey'      THEN 1
           WHEN series = 'olliepop'   AND key = 'berry'      THEN 1
           WHEN series = 'xmas'       AND key = 'snowy'      THEN 1
           WHEN series = 'valentines' AND key = 'choccy'     THEN 1
           WHEN series = 'easter'     AND key = 'bunny'      THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'blueberry'  THEN 1
           WHEN series = 'olliepop'   AND key = 'tangerine'  THEN 1
           WHEN series = 'xmas'       AND key = 'piney'      THEN 1
           WHEN series = 'valentines' AND key = 'cupie'      THEN 1
           WHEN series = 'easter'     AND key = 'chikky'     THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'lemony'     THEN 1
           WHEN series = 'olliepop'   AND key = 'bibble'     THEN 1
           WHEN series = 'xmas'       AND key = 'starry'     THEN 1
           WHEN series = 'valentines' AND key = 'bachie'     THEN 1
           WHEN series = 'easter'     AND key = 'nesty'      THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'bibi'       THEN 1
           WHEN series = 'olliepop'   AND key = 'kiwi'       THEN 1
           WHEN series = 'xmas'       AND key = 'socky'      THEN 1
           WHEN series = 'valentines' AND key = 'drinky'     THEN 1
           WHEN series = 'easter'     AND key = 'lamby'      THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'pinky'      THEN 1
           WHEN series = 'olliepop'   AND key = 'crunchy'    THEN 1
           WHEN series = 'xmas'       AND key = 'gingey'     THEN 1
           WHEN series = 'valentines' AND key = 'sherbie'    THEN 1
           WHEN series = 'easter'     AND key = 'choccy'     THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'minty'      THEN 1
           WHEN series = 'olliepop'   AND key = 'caramel'    THEN 1
           WHEN series = 'xmas'       AND key = 'nicky'      THEN 1
           WHEN series = 'valentines' AND key = 'lovey'      THEN 1
           WHEN series = 'easter'     AND key = 'eggy'       THEN 1
           ELSE 0 END),
  MAX(CASE WHEN series = 'coobubu'    AND key = 'cherry'     THEN 1
           WHEN series = 'olliepop'   AND key = 'grape'      THEN 1
           WHEN series = 'xmas'       AND key = 'dancey'     THEN 1
           WHEN series = 'valentines' AND key = 'bunchie'    THEN 1
           WHEN series = 'easter'     AND key = 'flowey'     THEN 1
           ELSE 0 END),
  MAX(CASE WHEN key = 'secret' THEN 1 ELSE 0 END)
FROM user_plushies
GROUP BY user_id, username, series;

DROP TABLE user_plushies;

DROP TABLE user_stats;
DROP TABLE stat_definitions;
DROP TABLE blind_box_plushies;
DROP TABLE blind_box_series;

-- Restore oauth_tokens with NUMERIC affinity
CREATE TABLE oauth_tokens (
  token_type     TEXT PRIMARY KEY,
  access_token   TEXT NOT NULL,
  refresh_token  TEXT NOT NULL,
  updated_at     NUMERIC DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
