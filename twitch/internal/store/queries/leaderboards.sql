-- name: GetStatLeaderboard :one
SELECT
  (SELECT username FROM stats ORDER BY strength DESC LIMIT 1)     AS top_strength_username,
  (SELECT strength FROM stats ORDER BY strength DESC LIMIT 1)     AS top_strength_value,

  (SELECT username FROM stats ORDER BY intelligence DESC LIMIT 1) AS top_intelligence_username,
  (SELECT intelligence FROM stats ORDER BY intelligence DESC LIMIT 1) AS top_intelligence_value,

  (SELECT username FROM stats ORDER BY charisma DESC LIMIT 1)     AS top_charisma_username,
  (SELECT charisma FROM stats ORDER BY charisma DESC LIMIT 1)     AS top_charisma_value,

  (SELECT username FROM stats ORDER BY luck DESC LIMIT 1)         AS top_luck_username,
  (SELECT luck FROM stats ORDER BY luck DESC LIMIT 1)             AS top_luck_value,

  (SELECT username FROM stats ORDER BY dexterity DESC LIMIT 1)    AS top_dexterity_username,
  (SELECT dexterity FROM stats ORDER BY dexterity DESC LIMIT 1)   AS top_dexterity_value,

  (SELECT username FROM stats ORDER BY penis DESC LIMIT 1)        AS top_penis_username,
  (SELECT penis FROM stats ORDER BY penis DESC LIMIT 1)           AS top_penis_value;
