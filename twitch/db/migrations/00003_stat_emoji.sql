-- +goose Up
ALTER TABLE stat_definitions ADD COLUMN emoji TEXT NOT NULL DEFAULT '';

UPDATE stat_definitions SET emoji = '💪' WHERE name = 'strength';
UPDATE stat_definitions SET emoji = '🧠' WHERE name = 'intelligence';
UPDATE stat_definitions SET emoji = '✨' WHERE name = 'charisma';
UPDATE stat_definitions SET emoji = '🍀' WHERE name = 'luck';
UPDATE stat_definitions SET emoji = '🎯' WHERE name = 'dexterity';
UPDATE stat_definitions SET emoji = '🍆' WHERE name = 'penis';

-- +goose Down
ALTER TABLE stat_definitions DROP COLUMN emoji;
