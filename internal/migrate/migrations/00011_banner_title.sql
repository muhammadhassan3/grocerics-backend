-- +goose Up
-- Banners were image-only, so the admin grid had nothing to tell them apart.
ALTER TABLE banners ADD COLUMN title TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE banners DROP COLUMN IF EXISTS title;
