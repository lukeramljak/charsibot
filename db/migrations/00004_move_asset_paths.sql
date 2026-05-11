-- +goose Up
-- Move blind box asset paths from /blind-box/ to /assets/blind-box/ to avoid
-- colliding with the SvelteKit /blind-box SPA route.

UPDATE blind_box_series SET
  reveal_sound   = REPLACE(reveal_sound,   '/blind-box/', '/assets/blind-box/'),
  box_front_face = REPLACE(box_front_face, '/blind-box/', '/assets/blind-box/'),
  box_side_face  = REPLACE(box_side_face,  '/blind-box/', '/assets/blind-box/');

UPDATE blind_box_plushies SET
  image       = REPLACE(image,       '/blind-box/', '/assets/blind-box/'),
  empty_image = REPLACE(empty_image, '/blind-box/', '/assets/blind-box/');

-- +goose Down
UPDATE blind_box_series SET
  reveal_sound   = REPLACE(reveal_sound,   '/assets/blind-box/', '/blind-box/'),
  box_front_face = REPLACE(box_front_face, '/assets/blind-box/', '/blind-box/'),
  box_side_face  = REPLACE(box_side_face,  '/assets/blind-box/', '/blind-box/');

UPDATE blind_box_plushies SET
  image       = REPLACE(image,       '/assets/blind-box/', '/blind-box/'),
  empty_image = REPLACE(empty_image, '/assets/blind-box/', '/blind-box/');
