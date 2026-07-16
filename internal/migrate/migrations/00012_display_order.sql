-- +goose Up
-- display_order is a user-facing curation knob (mobile UI shows certain items first).
-- Brands need it. Cities don't (short picker; alphabetical is fine). Banners don't
-- either — they're a stack, newest-created shows first (order by created_at).
ALTER TABLE brands ADD COLUMN IF NOT EXISTS display_order INT NOT NULL DEFAULT 0;
ALTER TABLE cities DROP COLUMN IF EXISTS display_order;
ALTER TABLE banners DROP COLUMN IF EXISTS display_order;

-- +goose Down
ALTER TABLE banners ADD COLUMN display_order INT NOT NULL DEFAULT 0;
ALTER TABLE cities ADD COLUMN display_order INT NOT NULL DEFAULT 0;
ALTER TABLE brands DROP COLUMN IF EXISTS display_order;
