-- +goose Up
-- Adjust coobubu and olliepop so the secret plushie has a 1 in 24 chance.
-- Regular plushies share 23/24 of the weight equally (23 each), secret gets 7.

UPDATE blind_box_plushies
SET weight = CASE
  WHEN key = 'secret' THEN 7
  ELSE 23
END
WHERE series IN ('coobubu', 'olliepop');

-- +goose Down
-- Revert to the previous weights: regular plushies at 12, secret at 1.

UPDATE blind_box_plushies
SET weight = CASE
  WHEN key = 'secret' THEN 1
  ELSE 12
END
WHERE series IN ('coobubu', 'olliepop');
