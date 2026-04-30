-- name: GetAllSeries :many
SELECT * FROM blind_box_series ORDER BY series;

-- name: GetSeriesByType :one
SELECT * FROM blind_box_series WHERE series = ?;

-- name: GetPlushiesForSeries :many
SELECT * FROM blind_box_plushies WHERE series = ? ORDER BY sort_order;

-- name: GetAllSeriesWithPlushies :many
SELECT
  bbs.series,
  bbs.redemption_title,
  bbs.name,
  bbs.reveal_sound,
  bbs.box_front_face,
  bbs.box_side_face,
  bbs.display_color,
  bbs.text_color,
  bbp.id         AS plushie_id,
  bbp.key        AS plushie_key,
  bbp.sort_order AS plushie_sort_order,
  bbp.weight     AS plushie_weight,
  bbp.name       AS plushie_name,
  bbp.image      AS plushie_image,
  bbp.empty_image AS plushie_empty_image
FROM blind_box_series bbs
LEFT JOIN blind_box_plushies bbp ON bbp.series = bbs.series
ORDER BY bbs.series, bbp.sort_order;
